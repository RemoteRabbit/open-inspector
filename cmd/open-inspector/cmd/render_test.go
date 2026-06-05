// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRenderTable_VariablesAndOutputs(t *testing.T) {
	configJSON = false
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{
		"config", "../../../testdata/fixtures/variables-and-outputs",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	out := buf.String()
	for _, want := range []string{
		"## Variables (5)",
		"## Outputs (3)",
		"## Locals (2)",
		"region",
		"us-east-1",
		"SENSITIVE",   // column header
		"VALIDATIONS", // column header
		"(full detail available with --json)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("table output missing %q:\n%s", want, out)
		}
	}
	// The region variable has exactly one validation block, surfaced as a
	// count rather than dropped. Locate its row and confirm a "1" appears.
	var regionRow string
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "region ") {
			regionRow = line
			break
		}
	}
	if regionRow == "" {
		t.Fatalf("no region row found:\n%s", out)
	}
	if !strings.Contains(regionRow, " 1 ") {
		t.Errorf("expected region row to report 1 validation, got %q", regionRow)
	}
}

func TestRenderTable_ModernBlocksSections(t *testing.T) {
	configJSON = false
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{
		"config", "../../../testdata/fixtures/modern-blocks",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	out := buf.String()
	for _, want := range []string{
		"## Moved (1)",
		"## Imports (1)",
		"## Removed (1)",
		"## Checks (1)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("table output missing %q:\n%s", want, out)
		}
	}
}
