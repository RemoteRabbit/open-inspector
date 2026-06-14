// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

// Body is a generic, schema-less capture of an HCL block body: its
// attributes (as unevaluated expressions) and its nested blocks. It is
// populated only for native HCL; JSON bodies (.tf.json / .tofu.json) are
// not walked, because a schema-less walk requires the concrete native
// syntax tree.
type Body struct {
	Attributes map[string]Expression `json:"attributes,omitempty"` // attribute name -> value expression
	Blocks     []NestedBlock         `json:"blocks,omitempty"`     // nested blocks, in source order
}

// NestedBlock is one nested block within a Body, such as a resource's
// versioning {} block or a dynamic "x" {} block.
type NestedBlock struct {
	Type     string   `json:"type"`             // block type keyword, e.g. "versioning" or "dynamic"
	Labels   []string `json:"labels,omitempty"` // block labels, e.g. ["lifecycle_rule"] for a dynamic block
	Body     Body     `json:"body"`             // the block's own body, recursively
	Position Position `json:"position"`         // source position of the block header
}
