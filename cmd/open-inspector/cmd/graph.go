// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cmd

import (
	"fmt"
	"os"

	"github.com/remoterabbit/open-inspector/pkg/graph"
	"github.com/remoterabbit/open-inspector/pkg/inspector"
	"github.com/spf13/cobra"
)

var (
	graphFormat string
	graphFailOn string
)

var graphCmd = &cobra.Command{
	Use:   "graph <dir>",
	Short: "Render the module call graph rooted at <dir>.",
	Args:  cobra.ExactArgs(1),
	RunE:  runGraph,
}

func init() {
	graphCmd.Flags().StringVar(&graphFormat, "format", "tree", "output format: tree|dot|mermaid|json")
	graphCmd.Flags().StringVar(&graphFailOn, "fail-on", "error",
		"exit nonzero if a diagnostic with this severity is present: error|warning|never")
	rootCmd.AddCommand(graphCmd)
}

func runGraph(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	module, err := inspector.Inspect(args[0], inspector.WithModuleGraph())
	if err != nil {
		return fmt.Errorf("inspect: %w", err)
	}

	var renderErr error
	switch graphFormat {
	case "tree":
		renderErr = graph.RenderTree(cmd.OutOrStdout(), module)
	case "dot":
		renderErr = graph.RenderDot(cmd.OutOrStdout(), module)
	case "mermaid":
		renderErr = graph.RenderMermaid(cmd.OutOrStdout(), module)
	case "json":
		renderErr = graph.RenderJSON(cmd.OutOrStdout(), module)
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
