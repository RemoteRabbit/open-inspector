// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/remoterabbit/open-inspector/pkg/inspector"
)

var (
	configJSON   bool
	configFailOn string
)

// configCmd inspects a single module directory.
var configCmd = &cobra.Command{
	Use:   "config <dir>",
	Short: "Inspect a Terraform/OpenTofu module directory",
	Long: `config parses every .tf, .tf.json, .tofu, and .tofu.json file
in <dir> (non-recursive) and prints either a human-readable summary
table or a versioned JSON envelope.`,
	Args: cobra.ExactArgs(1),
	RunE: runConfig,
}

func init() {
	configCmd.Flags().BoolVar(&configJSON, "json", false,
		"emit machine-readable JSON instead of a human table")
	configCmd.Flags().StringVar(&configFailOn, "fail-on", "error",
		"exit nonzero if a diagnostic with this severity is present: error|warning|never")
	rootCmd.AddCommand(configCmd)
}

func runConfig(cmd *cobra.Command, args []string) error {
	stderrLog.infof("inspecting %s", args[0])
	mod, err := inspector.Inspect(args[0])
	if err != nil {
		return fmt.Errorf("inspect: %w", err)
	}
	if configJSON {
		if err := renderJSON(cmd, mod); err != nil {
			return err
		}
	} else {
		if err := renderTable(cmd, mod); err != nil {
			return err
		}
	}
	code := ExitCode(mod.Diagnostics, FailOnPolicy(configFailOn))
	if code != 0 {
		// Trigger a non-zero exit without printing "Error:" spam.
		os.Exit(code)
	}
	return nil
}
