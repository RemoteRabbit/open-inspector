// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

// MovedBlock corresponds to `moved { from = X, to = Y }`
type MovedBlock struct {
	From     string   `json:"from"`     // old address
	To       string   `json:"to"`       // new address
	Position Position `json:"position"` // source position of the moved block
}

// ImportBlock corresponds to `import { to = X, id = ..., provider = ... }`
type ImportBlock struct {
	To       string     `json:"to"`                 // address the resource is imported to
	ID       Expression `json:"id"`                 // import ID expression
	Provider string     `json:"provider,omitempty"` // provider reference, if set
	Position Position   `json:"position"`           // source position of the import block
}

// RemovedBlock corresponds to `removed { from, lifecycle {destroy } }`
type RemovedBlock struct {
	From          string   `json:"from"`                      // address being removed
	DestroyOnDrop *bool    `json:"destroy_on_drop,omitempty"` // whether to destroy on removal (lifecycle { destroy })
	Position      Position `json:"position"`                  // source position of the removed block
}
