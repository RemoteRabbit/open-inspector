// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

// EphemeralResource corresponds to `ephemeral "<type>" "<name>" {}`
// Same meta-args as a managed resource.
type EphemeralResource struct {
	Type      string      `json:"type"`                 // resource type
	Name      string      `json:"name"`                 // local name
	Provider  string      `json:"provider,omitempty"`   // from the provider meta-argument, if set
	Count     *Expression `json:"count,omitempty"`      // count meta-argument expression, if set
	ForEach   *Expression `json:"for_each,omitempty"`   // for_each meta-argument expression, if set
	DependsOn []string    `json:"depends_on,omitempty"` // explicit dependency references

	// AttrNames lists the user-set top-level attribute names that are not
	// meta-arguments, captured at load time and sorted. See
	// Resource.AttrNames for the exact semantics and caveats.
	AttrNames []string `json:"attr_names,omitempty"`

	// SchemaFindings holds schema-derived annotations, populated only when
	// inspection runs with a provider schema. See Resource.SchemaFindings.
	SchemaFindings *SchemaFindings `json:"schema_findings,omitempty"`

	Lifecycle *Lifecycle `json:"lifecycle,omitempty"` // lifecycle block, if present
	Comment   string     `json:"comment,omitempty"`   // leading docstring comment above block, if any.
	Position  Position   `json:"position"`            // source position of the ephemeral block
}
