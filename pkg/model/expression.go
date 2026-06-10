// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

// Expression is a snippet of HCL source captured at parse time.
// Range identifies its position;  Source is the verbatim bytes from
// the file. Neither field is evaluated by the loader.
type Expression struct {
	Source string `json:"source"` // verbatim HCL source bytes
	Range  Range  `json:"range"`  // position of the expression in source
}
