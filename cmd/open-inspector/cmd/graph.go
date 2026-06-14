// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/remoterabbit/open-inspector/pkg/graph"
	"github.com/remoterabbit/open-inspector/pkg/inspector"
	"github.com/remoterabbit/open-inspector/pkg/model"
)

var (
	graphFormat    string
	graphRecursive bool
	graphMaxDepth  int
	graphFailOn    string
	graphKinds     []string
	graphContract  bool
)

// validGraphKinds lists the node kinds accepted by --kind.
var validGraphKinds = map[string]model.DependencyNodeKind{
	"resource":  model.DependencyNodeResource,
	"data":      model.DependencyNodeData,
	"ephemeral": model.DependencyNodeEphemeral,
	"variable":  model.DependencyNodeVariable,
	"local":     model.DependencyNodeLocal,
	"output":    model.DependencyNodeOutput,
	"module":    model.DependencyNodeModule,
}

// graphKindNames returns the accepted --kind values in sorted order, for help
// text and error messages.
func graphKindNames() []string {
	names := make([]string, 0, len(validGraphKinds))
	for name := range validGraphKinds {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// parseGraphKinds turns the --kind values into a set, rejecting unknown kinds.
// An empty selection returns a nil set (meaning "all kinds").
func parseGraphKinds(values []string) (map[model.DependencyNodeKind]struct{}, error) {
	if len(values) == 0 {
		return nil, nil
	}
	set := make(map[model.DependencyNodeKind]struct{}, len(values))
	for _, value := range values {
		kind, ok := validGraphKinds[strings.ToLower(strings.TrimSpace(value))]
		if !ok {
			return nil, fmt.Errorf("unknown kind %q (valid: %s)", value, strings.Join(graphKindNames(), ", "))
		}
		set[kind] = struct{}{}
	}
	return set, nil
}

// graphCmd renders the dependency graph of the resources and other
// declarations in a module directory, in the chosen format. It mirrors
// `terraform graph` / `tofu graph` (the resource dependency graph), as
// opposed to the `modules` command (the module call graph).
var graphCmd = &cobra.Command{
	Use:   "graph <dir>",
	Short: "Render the dependency graph of a module directory",
	Long: `graph builds the dependency graph of the module in <dir>: resources,
data sources, locals, outputs, variables, and module calls, with edges
derived from the references between them (including references inside
nested blocks) plus depends_on and replace_triggered_by.

It is a static analysis: no providers, no 'init', and (for local modules)
no network are required. With --recursive, child module calls are resolved
and each module's graph is rendered too, with cross-module edges drawn from
module input arguments and from references to specific child module outputs.

Use --kind to keep only certain node kinds; add --contract to collapse paths
that run through filtered-out nodes instead of breaking them.`,
	Args: cobra.ExactArgs(1),
	RunE: runGraph,
}

// init registers the graph subcommand and its flags on the root command.
func init() {
	graphCmd.Flags().StringVar(&graphFormat, "format", "tree", "output format: tree|dot|mermaid|json")
	graphCmd.Flags().BoolVar(&graphRecursive, "recursive", false,
		"resolve child module calls and render each module's dependency graph")
	graphCmd.Flags().IntVar(&graphMaxDepth, "max-depth", 16,
		"maximum child-module recursion depth (only with --recursive)")
	graphCmd.Flags().StringVar(&graphFailOn, "fail-on", "error",
		"exit nonzero if a diagnostic with this severity is present: error|warning|never")
	graphCmd.Flags().StringSliceVar(&graphKinds, "kind", nil,
		"keep only these node kinds (comma-separated): "+strings.Join(graphKindNames(), "|")+"; default all")
	graphCmd.Flags().BoolVar(&graphContract, "contract", false,
		"with --kind, contract paths through removed nodes instead of breaking them")
	rootCmd.AddCommand(graphCmd)
}

// runGraph inspects the module directory in args[0], building the dependency
// graph (and, with --recursive, the module graph too), then renders it in
// the requested format.
func runGraph(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	kinds, err := parseGraphKinds(graphKinds)
	if err != nil {
		return err
	}

	opts := []inspector.Option{inspector.WithDependencyGraph()}
	if graphRecursive {
		opts = append(opts, inspector.WithModuleGraph(), inspector.WithMaxDepth(graphMaxDepth))
	}

	module, err := inspector.Inspect(args[0], opts...)
	if err != nil {
		return fmt.Errorf("inspect: %w", err)
	}

	if graphContract {
		graph.ContractDependenciesByKind(module, kinds)
	} else {
		graph.FilterDependenciesByKind(module, kinds)
	}

	var renderErr error
	switch graphFormat {
	case "tree":
		renderErr = graph.RenderDepsTree(cmd.OutOrStdout(), module)
	case "dot":
		renderErr = graph.RenderDepsDot(cmd.OutOrStdout(), module)
	case "mermaid":
		renderErr = graph.RenderDepsMermaid(cmd.OutOrStdout(), module)
	case "json":
		renderErr = graph.RenderDepsJSON(cmd.OutOrStdout(), module)
	default:
		return fmt.Errorf("unknown format %q", graphFormat)
	}
	if renderErr != nil {
		return renderErr
	}

	code := ExitCode(module.Diagnostics, FailOnPolicy(graphFailOn))
	if code != 0 {
		// Trigger a non-zero exit without printing "Error:" spam.
		os.Exit(code)
	}
	return nil
}
