// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

// CheckBlock corresponds to `check "<name>" { data ... assert ... }`
type CheckBlock struct {
	Name       string      `json:"name"`                  // check block name
	DataSource *Resource   `json:"data_source,omitempty"` // optional scoped data source the assertions read
	Assertions []Assertion `json:"assertions,omitempty"`  // assert blocks inside the check
	Position   Position    `json:"position"`              // source position of the check block
}

// Assertion is one `assert { condition, error_message }` inside a check block.
type Assertion struct {
	Condition    Expression `json:"condition"`     // boolean condition expression
	ErrorMessage Expression `json:"error_message"` // message shown when the condition fails
	Position     Position   `json:"position"`      // source position of the assert block
}
