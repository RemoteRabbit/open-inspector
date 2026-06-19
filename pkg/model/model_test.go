// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/hcl/v2"
)

func TestDiagnosticsHasErrors(t *testing.T) {
	tests := []struct {
		name string
		diag Diagnostics
		want bool
	}{
		{"nil", nil, false},
		{"empty", Diagnostics{}, false},
		{"only warnings", Diagnostics{{Severity: SeverityWarning}}, false},
		{"single error", Diagnostics{{Severity: SeverityError}}, true},
		{"mixed", Diagnostics{{Severity: SeverityWarning}, {Severity: SeverityError}}, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.diag.HasErrors(); got != test.want {
				t.Errorf("HasErrors() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestPositionFromHCL(t *testing.T) {
	// filepath.ToSlash converts the host OS separator; on Windows this
	// turns a native path into forward slashes, while a path that already
	// uses forward slashes is unchanged on every OS. Use the latter so the
	// field mapping is asserted portably.
	rang := hcl.Range{
		Filename: "dir/sub/main.tf",
		Start:    hcl.Pos{Line: 3, Column: 5, Byte: 42},
		End:      hcl.Pos{Line: 3, Column: 9, Byte: 46},
	}

	got := PositionFromHCL(rang)

	want := Position{
		Filename: "dir/sub/main.tf",
		Start:    Pos{Line: 3, Column: 5, Byte: 42},
		End:      Pos{Line: 3, Column: 9, Byte: 46},
	}
	if got != want {
		t.Errorf("PositionFromHCL() = %+v, want %+v", got, want)
	}
}

func TestDiagnosticsFromHCL(t *testing.T) {
	subject := &hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 1, Column: 1, Byte: 0}}
	context := &hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 1, Byte: 10}}

	input := hcl.Diagnostics{
		{
			Severity: hcl.DiagError,
			Summary:  "bad thing",
			Detail:   "the details",
			Subject:  subject,
			Context:  context,
		},
		{
			Severity: hcl.DiagWarning,
			Summary:  "minor thing",
		},
	}

	got := DiagnosticsFromHCL(input)

	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}

	if got[0].Severity != SeverityError {
		t.Errorf("got[0].Severity = %q, want %q", got[0].Severity, SeverityError)
	}
	if got[0].Summary != "bad thing" || got[0].Detail != "the details" {
		t.Errorf("got[0] summary/detail = %q/%q", got[0].Summary, got[0].Detail)
	}
	if got[0].Subject == nil || got[0].Subject.Filename != "main.tf" {
		t.Errorf("got[0].Subject = %+v, want non-nil with filename main.tf", got[0].Subject)
	}
	if got[0].Context == nil || got[0].Context.Start.Byte != 10 {
		t.Errorf("got[0].Context = %+v, want non-nil at byte 10", got[0].Context)
	}

	if got[1].Severity != SeverityWarning {
		t.Errorf("got[1].Severity = %q, want %q", got[1].Severity, SeverityWarning)
	}
	if got[1].Subject != nil || got[1].Context != nil {
		t.Errorf("got[1] subject/context = %+v/%+v, want both nil", got[1].Subject, got[1].Context)
	}
}

func TestDiagnosticsFromHCLEmpty(t *testing.T) {
	got := DiagnosticsFromHCL(nil)
	if got == nil {
		t.Fatal("DiagnosticsFromHCL(nil) = nil, want non-nil empty slice")
	}
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}

// jsonKeys collects the top-level object keys produced by marshaling v.
func jsonKeys(t *testing.T, v any) map[string]json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out map[string]json.RawMessage
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	return out
}

func hasKey(keys map[string]json.RawMessage, name string) bool {
	_, ok := keys[name]
	return ok
}

// TestWireFormatErrorMessageKey pins the JSON key for the error message on
// both block types that carry one. Assertion (check { assert { ... } }) and
// Validation (variable { validation { ... } }) describe the same Terraform
// keyword, error_message, and must serialize it identically. This guards
// against the two drifting apart again.
func TestWireFormatErrorMessageKey(t *testing.T) {
	assertion := jsonKeys(t, Assertion{ErrorMessage: Expression{Source: "x"}})
	if !hasKey(assertion, "error_message") {
		t.Errorf("Assertion error message key: got keys %v, want error_message", keysOf(assertion))
	}
	if hasKey(assertion, "expression") {
		t.Error("Assertion still serializes the stale `expression` key")
	}

	validation := jsonKeys(t, Validation{ErrorMessage: Expression{Source: "x"}})
	if !hasKey(validation, "error_message") {
		t.Errorf("Validation error message key: got keys %v, want error_message", keysOf(validation))
	}
}

// TestWireFormatRequiredKeys spot-checks that the stable, non-omitempty
// fields stay present in the marshaled output, since they form the contract
// downstream consumers rely on.
func TestWireFormatRequiredKeys(t *testing.T) {
	tests := []struct {
		name string
		v    any
		want []string
	}{
		{"Expression", Expression{}, []string{"source", "position"}},
		{"Assertion", Assertion{}, []string{"condition", "error_message", "position"}},
		{"Diagnostic", Diagnostic{}, []string{"severity", "summary"}},
		{"Position", Position{}, []string{"filename", "start", "end"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			keys := jsonKeys(t, test.v)
			for _, want := range test.want {
				if !hasKey(keys, want) {
					t.Errorf("%s missing key %q; got %v", test.name, want, keysOf(keys))
				}
			}
		})
	}
}

func keysOf(m map[string]json.RawMessage) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
