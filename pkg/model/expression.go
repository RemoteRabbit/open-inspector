// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

// Expression is a snippet of HCL source captured at parse time.
// Range identifies its position; Source is the verbatim bytes from
// the file. Neither field is evaluated by the loader.
type Expression struct {
	Source     string      `json:"source"`               // verbatim HCL source bytes
	Position   Position    `json:"position"`             // position of the expression in source
	References []Reference `json:"references,omitempty"` // values referenced by the expression, in source order, de-duplicated
}

// ReferenceKind classifies the root object a Reference points at.
type ReferenceKind string

// Reference kinds. "other" covers iteration and meta symbols such as
// each, self, count, terraform, and path that are not module-level
// declarations.
const (
	ReferenceVar       ReferenceKind = "var"       // input variable:  var.NAME
	ReferenceLocal     ReferenceKind = "local"     // local value:     local.NAME
	ReferenceModule    ReferenceKind = "module"    // module output:   module.NAME
	ReferenceData      ReferenceKind = "data"      // data source:      data.TYPE.NAME
	ReferenceEphemeral ReferenceKind = "ephemeral" // ephemeral resource: ephemeral.TYPE.NAME
	ReferenceResource  ReferenceKind = "resource"  // managed resource: TYPE.NAME
	ReferenceOther     ReferenceKind = "other"     // each / self / count / path / terraform / ...
)

// Reference is a single value referenced by an Expression, extracted from
// the expression's traversals (not from the verbatim source).
type Reference struct {
	Kind    ReferenceKind `json:"kind"`    // category of the referenced object
	Address string        `json:"address"` // canonical address, e.g. "var.region" or "aws_s3_bucket.b"
	// Attribute is the member accessed on the referenced object, when it is
	// meaningful for resolving the reference more precisely. It is currently
	// captured only for module references: the output name in
	// "module.NAME.OUTPUT" (so Address "module.net" gains Attribute
	// "subnet_id"). Empty for every other kind and for bare "module.NAME".
	Attribute string   `json:"attribute,omitempty"`
	Position  Position `json:"position,omitempty"` // source range of this traversal, if available
}
