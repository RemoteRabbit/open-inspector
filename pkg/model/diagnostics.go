// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

import "github.com/hashicorp/hcl/v2"

// Severity classifies a Diagnostic as an error or warning.
type Severity string

// Severity values reported by loaders.
const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

// Diagnostic is a single problem reported by a loader, with optional
// source location information.
type Diagnostic struct {
	Severity Severity  `json:"severity"`          // error or warning
	Summary  string    `json:"summary"`           // short problem description
	Detail   string    `json:"detail,omitempty"`  // optional longer explanation
	Subject  *Position `json:"subject,omitempty"` // optional primary source location
	Context  *Position `json:"context,omitempty"` // optional surrounding source location
}

// Diagnostics is a collection of Diagnostic values.
type Diagnostics []Diagnostic

// HasErrors reports whether any diagnostic in the collection has
// SeverityError.
func (diag Diagnostics) HasErrors() bool {
	for _, error := range diag {
		if error.Severity == SeverityError {
			return true
		}
	}
	return false
}

// DiagnosticsFromHCL translates a slice of hcl.Diagnostic into the
// model's wire-friendly Diagnostics type.
func DiagnosticsFromHCL(hcld hcl.Diagnostics) Diagnostics {
	output := make(Diagnostics, 0, len(hcld))
	for _, diag := range hcld {
		sev := SeverityError
		if diag.Severity == hcl.DiagWarning {
			sev = SeverityWarning
		}
		newDiag := Diagnostic{Severity: sev, Summary: diag.Summary, Detail: diag.Detail}
		if diag.Subject != nil {
			rang := PositionFromHCL(*diag.Subject)
			newDiag.Subject = &rang
		}
		if diag.Context != nil {
			rang := PositionFromHCL(*diag.Context)
			newDiag.Context = &rang
		}
		output = append(output, newDiag)
	}
	return output
}
