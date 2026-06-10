// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestConfig_JSON_RoundTrip(t *testing.T) {
	configJSON = true
	t.Cleanup(func() { configJSON = false })

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{
		"config", "--json",
		"../../../testdata/fixtures/simple",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, buf.String())
	}
	if got["schema_version"] == nil {
		t.Errorf("output missing schema_version: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "null_resource") {
		t.Errorf("output missing expected resource: %s", buf.String())
	}
}

func TestConfig_WithSchema_FindsUnknown(t *testing.T) {
	configJSON = true
	configSchema = "../../../pkg/config/testdata/schemas/null.json"
	t.Cleanup(func() { configJSON = false; configSchema = "" })

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{
		"config", "--json",
		"--schema", configSchema,
		"../../../testdata/fixtures/simple-with-typo",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "schema_findings") {
		t.Errorf("output missing schema_findings: %s", out)
	}
	if !strings.Contains(out, "trigerz") {
		t.Errorf("output missing the unknown attribute name: %s", out)
	}
}

func TestConfig_WithSchema_Table(t *testing.T) {
	configSchema = "../../../pkg/config/testdata/schemas/null.json"
	t.Cleanup(func() { configSchema = "" })

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{
		"config",
		"--schema", configSchema,
		"../../../testdata/fixtures/simple-with-typo",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "FINDINGS") {
		t.Errorf("table missing FINDINGS column: %s", out)
	}
	if !strings.Contains(out, "unknown: trigerz") {
		t.Errorf("table missing findings summary: %s", out)
	}
}

func TestConfig_MissingSchemaFile(t *testing.T) {
	configSchema = "../../../testdata/fixtures/does-not-exist.json"
	t.Cleanup(func() { configSchema = "" })

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{
		"config",
		"--schema", configSchema,
		"../../../testdata/fixtures/simple",
	})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatalf("expected an error for a missing schema file")
		return
	}
	if !strings.Contains(err.Error(), "open schema") {
		t.Errorf("error = %v, want it to mention opening the schema", err)
	}
}
