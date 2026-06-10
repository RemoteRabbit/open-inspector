// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

// SchemaFindings groups the schema-derived annotations for a single
// resource, data source, or ephemeral resource. All fields are omitempty:
// a resource with no findings carries a nil *SchemaFindings and adds
// nothing to the JSON output.
type SchemaFindings struct {
	// UnknownAttrs are attributes the user set that the schema does not
	// define (for example a misspelled attribute name).
	UnknownAttrs []AttrFinding `json:"unknown_attrs,omitempty"`
	// DeprecatedAttrs are attributes the user set that the schema marks
	// deprecated.
	DeprecatedAttrs []DeprecatedAttr `json:"deprecated_attrs,omitempty"`
	// MissingRequired are attribute names the schema marks required that
	// the user did not set. Sorted for deterministic output.
	MissingRequired []string `json:"missing_required,omitempty"`
}

// AttrFinding names a single attribute and points at the block that set it.
type AttrFinding struct {
	Name  string `json:"name"`  // attribute name
	Range Range  `json:"range"` // where the attribute was set
}

// DeprecatedAttr names a deprecated attribute and carries the schema's
// deprecation/description message.
type DeprecatedAttr struct {
	Name    string `json:"name"`              // attribute name
	Message string `json:"message,omitempty"` // schema's deprecation message
	Range   Range  `json:"range"`             // where the attribute was set
}
