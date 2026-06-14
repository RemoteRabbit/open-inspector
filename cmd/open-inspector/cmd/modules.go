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
	modulesFormat string
	modulesFailOn string
)

// modulesCmd renders the module call graph: which module calls which child
// module, resolved recursively from <dir>.
var modulesCmd = &cobra.Command{
	Use:   "modules <dir>",
	Short: "Render the module call graph rooted at <dir>.",
	Args:  cobra.ExactArgs(1),
	RunE:  runModules,
}

func init() {
	modulesCmd.Flags().StringVar(&modulesFormat, "format", "tree", "output format: tree|dot|mermaid|json")
	modulesCmd.Flags().StringVar(&modulesFailOn, "fail-on", "error",
		"exit nonzero if a diagnostic with this severity is present: error|warning|never")
	rootCmd.AddCommand(modulesCmd)
}

func runModules(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	module, err := inspector.Inspect(args[0], inspector.WithModuleGraph())
	if err != nil {
		return fmt.Errorf("inspect: %w", err)
	}

	var renderErr error
	switch modulesFormat {
	case "tree":
		renderErr = graph.RenderTree(cmd.OutOrStdout(), module)
	case "dot":
		renderErr = graph.RenderDot(cmd.OutOrStdout(), module)
	case "mermaid":
		renderErr = graph.RenderMermaid(cmd.OutOrStdout(), module)
	case "json":
		renderErr = graph.RenderJSON(cmd.OutOrStdout(), module)
	default:
		return fmt.Errorf("unknown format %q", modulesFormat)
	}
	if renderErr != nil {
		return renderErr
	}

	code := ExitCode(module.Diagnostics, FailOnPolicy(modulesFailOn))
	if code != 0 {
		// Trigger a non-zero exit without printing "Error:" spam.
		os.Exit(code)
	}
	return nil
}
