// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

// ValueKind names the category of a decoded literal Value.
type ValueKind string

// Value kinds. "null" is a typed or untyped null; the structural kinds
// carry their children in the matching field.
const (
	ValueNull   ValueKind = "null"
	ValueString ValueKind = "string"
	ValueNumber ValueKind = "number"
	ValueBool   ValueKind = "bool"
	ValueList   ValueKind = "list" // list or set (collections render the same)
	ValueTuple  ValueKind = "tuple"
	ValueMap    ValueKind = "map"
	ValueObject ValueKind = "object"
)

// Value is a decoded constant value. Only the field matching Kind is set.
// Numbers are stored as their canonical decimal string to avoid float
// precision loss across the JSON boundary.
type Value struct {
	Kind   ValueKind         `json:"kind"`
	String string            `json:"string,omitempty"`
	Number string            `json:"number,omitempty"` // canonical decimal text
	Bool   bool              `json:"bool,omitempty"`
	List   []*Value          `json:"list,omitempty"`   // list/set elements
	Tuple  []*Value          `json:"tuple,omitempty"`  // tuple elements
	Map    map[string]*Value `json:"map,omitempty"`    // map entries
	Object map[string]*Value `json:"object,omitempty"` // object attributes
}
