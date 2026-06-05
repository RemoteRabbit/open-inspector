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
		"SENSITIVE", // column header
	} {
		if !strings.Contains(out, want) {
			t.Errorf("table output missing %q:\n%s", want, out)
		}
	}
}
