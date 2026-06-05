// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

import (
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
)

// Range identifies a contiguous span of source code.
type Range struct {
	Filename string `json:"filename"`
	Start    Pos    `json:"start"`
	End      Pos    `json:"end"`
}

// Pos is a single position within a source file.
type Pos struct {
	Line   int `json:"line"`
	Column int `json:"column"`
	Byte   int `json:"byte"`
}

// RangeFromHCL converts an hcl.Range into the model's wire-friendly Range.
//
// Filename is normalized to forward slashes so JSON output is byte-identical
// across Linux, macOS, and Windows. This is the sole chokepoint where
// hcl.Range filenames enter the model, so every downstream Range field
// gets the same canonical form for free.
func RangeFromHCL(rang hcl.Range) Range {
	return Range{
		Filename: filepath.ToSlash(rang.Filename),
		Start:    Pos{Line: rang.Start.Line, Column: rang.Start.Column, Byte: rang.Start.Byte},
		End:      Pos{Line: rang.End.Line, Column: rang.End.Column, Byte: rang.End.Byte},
	}
}
