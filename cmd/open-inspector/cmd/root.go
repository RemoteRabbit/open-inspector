// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package cmd ...
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

// rootCmd
var rootCmd = &cobra.Command{
	Use:   "open-inspector",
	Short: "Inspect Terraform/OpenTofu configurations.",
	Long: `open-inspector parses Terraform and OpenTofu module directories into
a stable, source-range-accurate model.

Run "open-inspector config <dir>" to inspect a module.`,
	SilenceUsage:  true,
	SilenceErrors: false,
}

// Execute command
func Execute() int {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error: ", err)
		return 1
	}
	return 0
}

// init command
func init() {
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
