// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package cmd defines the open-inspector cobra command tree: the root
// command, its persistent flags, the subcommands, and the output
// renderers and exit-code policy that back them.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version is set by ldflags at build time; defaults to "dev" for local builds. Matches
// the existing main.go convention.
var Version = "dev"

var (
	flagLogLevel string
	flagNoColor  bool
	flagQuiet    bool
)

// rootCmd is the top-level "open-inspector" command. Subcommands attach
// to it via AddCommand in their respective init functions.
var rootCmd = &cobra.Command{
	Use:   "open-inspector",
	Short: "Inspect Terraform/OpenTofu configurations.",
	Long: `open-inspector parses Terraform and OpenTofu module directories into
a stable, source-range-accurate model.

Run "open-inspector config <dir>" to inspect a module.`,
	SilenceUsage:  true,
	SilenceErrors: false,
}

// Execute runs the root command and returns the process exit code: 0 on
// success, 1 when the command returns an error. cobra prints the error
// itself, so this only maps the error into a status code.
func Execute() int {
	// cobra prints the error itself (SilenceErrors is false), so we only
	// translate a non-nil error into a non-zero exit code here.
	if err := rootCmd.Execute(); err != nil {
		return 1
	}
	return 0
}

// Root returns the root command so external tooling (notably cmd/docgen)
// can walk the command tree to generate documentation.
func Root() *cobra.Command { return rootCmd }

// init registers the persistent flags, wires up the logger, and sets the
// version template on the root command.
func init() {
	// TODO: --log-level is parsed and advertised but not yet honored; the
	// logger only respects --quiet and --no-color. Wire the level into
	// newLogger (gate infof/debugf by severity) or drop the flag.
	rootCmd.PersistentFlags().StringVar(&flagLogLevel, "log-level", "info", "log-level: debug|info|warn|error")
	rootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVar(&flagQuiet, "quiet", false, "suppress informational log output")

	// Configure the logger once flags are parsed, before any RunE fires.
	cobra.OnInitialize(func() {
		stderrLog = newLogger(os.Stderr, flagQuiet, flagNoColor)
	})

	rootCmd.Version = Version
	rootCmd.SetVersionTemplate(fmt.Sprintf("open-inspector %s\n", Version))
}
