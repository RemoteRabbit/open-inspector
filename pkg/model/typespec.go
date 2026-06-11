// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

// TypeKind names the category of a TypeSpec node.
type TypeKind string

// Type kinds. "dynamic" is cty's any-type (the `any` keyword); "tuple" and "object" are structural.
// The rest are collection or primitive.
const (
	TypeString  TypeKind = "string"
	TypeNumber  TypeKind = "number"
	TypeBool    TypeKind = "bool"
	TypeList    TypeKind = "list"
	TypeSet     TypeKind = "set"
	TypeMap     TypeKind = "map"
	TypeObject  TypeKind = "object"
	TypeTuple   TypeKind = "tuple"
	TypeDynamic TypeKind = "dynamic"
)

// TypeSpec is the structured form of a variable's type constraint. It is a recursive tree:
// collections carry an Element, tuples carry Elements, and objects carry Attributes.
// Primitives carry only Kind.
type TypeSpec struct {
	Kind       TypeKind               `json:"kind"`                 // category of this node
	Element    *TypeSpec              `json:"element,omitempty"`    // element type for list/se/map
	Elements   []*TypeSpec            `json:"elements,omitempty"`   // ordered element types for tuple
	Attributes map[string]*ObjectAttr `json:"attributes,omitempty"` // attributes for object
}

// ObjectAttr describes one attribute of an object typed TypeSpec.
type ObjectAttr struct {
	Type     *TypeSpec `json:"type"`               // the attribute's type
	Optional bool      `json:"optional,omitempty"` // declared with optional(...)
	Default  *Value    `json:"default,omitempty"`  // default supplied to optional(T, default)
}
