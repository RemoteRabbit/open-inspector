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
