// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

import "github.com/hashicorp/hcl/v2"

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

// RangeFromHcl converts an hcl.Range into the model's wire-friendly Range.
func RangeFromHcl(rang hcl.Range) Range {
	return Range{
		Filename: rang.Filename,
		Start:    Pos{Line: rang.Start.Line, Column: rang.Start.Column, Byte: rang.Start.Byte},
		End:      Pos{Line: rang.End.Line, Column: rang.End.Column, Byte: rang.End.Byte},
	}
}
