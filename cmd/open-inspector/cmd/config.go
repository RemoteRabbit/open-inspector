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
	configSchema string
	configDeps   bool
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

// init registers the config subcommand and its flags on the root command.
func init() {
	configCmd.Flags().BoolVar(&configJSON, "json", false,
		"emit machine-readable JSON instead of a human table")
	configCmd.Flags().StringVar(&configFailOn, "fail-on", "error",
		"exit nonzero if a diagnostic with this severity is present: error|warning|never")
	configCmd.Flags().StringVar(&configSchema, "schema", "",
		"enrich resources with findings from a tofu/terraform 'providers schema -json' document; pass a file path or 'auto' to shell out")
	configCmd.Flags().BoolVar(&configDeps, "deps", false,
		"derive the intra-module dependency graph (resources, locals, outputs, ... and their references)")
	rootCmd.AddCommand(configCmd)
}

// runConfig inspects the module directory in args[0], renders it as JSON or
// a table, and then applies the --fail-on policy: when diagnostics match
// the threshold it exits the process directly with the policy's code,
// bypassing cobra's error-to-exit-1 translation.
func runConfig(cmd *cobra.Command, args []string) error {
	stderrLog.infof("inspecting %s", args[0])

	opts, cleanup, err := schemaOptions()
	if err != nil {
		return err
	}
	defer cleanup()

	if configDeps {
		opts = append(opts, inspector.WithDependencyGraph())
	}

	mod, err := inspector.Inspect(args[0], opts...)
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

// schemaOptions translates the --schema flag into inspector options. It
// returns a cleanup function (closing any opened schema file) that the
// caller must defer. With no --schema flag it returns no options.
func schemaOptions() ([]inspector.Option, func(), error) {
	noop := func() {}
	switch configSchema {
	case "":
		return nil, noop, nil
	case "auto":
		return []inspector.Option{inspector.WithSchemaAuto()}, noop, nil
	default:
		file, err := os.Open(configSchema)
		if err != nil {
			return nil, noop, fmt.Errorf("open schema: %w", err)
		}
		return []inspector.Option{inspector.WithSchema(file)}, func() { _ = file.Close() }, nil
	}
}
