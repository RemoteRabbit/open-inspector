// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

// CheckBlock corresponds to `check "<name>" { data ... assert ... }`
type CheckBlock struct {
	Name       string      `json:"name"`
	DataSource *Resource   `json:"data_source,omitempty"`
	Assertions []Assertion `json:"assertions,omitempty"`
	Range      Range       `json:"range"`
}

// Assertion is one `assert { condition, error_message }` inside a check block.
type Assertion struct {
	Condition    Expression `json:"condition"`
	ErrorMessage Expression `json:"expression"`
	Range        Range      `json:"range"`
}
