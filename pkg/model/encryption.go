// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

// Encryption corresponds to OpenTofu's `terraform { encryption {} }`
// configuration tree (OpenTofu 1.7+). Terraform rejects this block.
type Encryption struct {
	KeyProviders       []EncryptionKeyProvider `json:"key_providers,omitempty"`             // key_provider blocks
	Methods            []EncryptionMethod      `json:"methods,omitempty"`                   // method blocks
	State              *EncryptionTarget       `json:"state,omitempty"`                     // state encryption target, if configured
	Plan               *EncryptionTarget       `json:"plan,omitempty"`                      // plan encryption target, if configured
	RemoteStateSources []EncryptionRemoteState `json:"remote_state_data_sources,omitempty"` // remote_state_data_sources entries
	Position           Position                `json:"position"`                            // source position of the encryption block
}

// EncryptionKeyProvider describes one `key_provider "<type>" "<name>" {}`
// inside an encryption block. The body schema is provider-defined, so
// each attribute is captured verbatim as an Expression.
type EncryptionKeyProvider struct {
	Type     string                `json:"type"`           // key provider type
	Name     string                `json:"name"`           // key provider name
	Body     map[string]Expression `json:"body,omitempty"` // provider-defined attributes, captured verbatim
	Position Position              `json:"position"`       // source position of the block
}

// EncryptionMethod describes one `method "<type>" "<name>" {}` inside
// an encryption block. Body shape is method-defined; captured verbatim.
type EncryptionMethod struct {
	Type     string                `json:"type"`           // method type
	Name     string                `json:"name"`           // method name
	Body     map[string]Expression `json:"body,omitempty"` // method-defined attributes, captured verbatim
	Position Position              `json:"position"`       // source position of the block
}

// EncryptionTarget describes the `state {}` or `plan {}` sub-block:
// a required `method` reference and an optional `fallback`.
type EncryptionTarget struct {
	Method   Expression `json:"method"`             // required method reference
	Fallback Expression `json:"fallback,omitempty"` // optional fallback method reference
	Position Position   `json:"position"`           // source position of the block
}

// EncryptionRemoteState describes a `remote_state_data_sources {}` entry.
// Its body schema is open-ended; captured verbatim.
type EncryptionRemoteState struct {
	Name     string                `json:"name"`           // data source name
	Body     map[string]Expression `json:"body,omitempty"` // open-ended attributes, captured verbatim
	Position Position              `json:"position"`       // source position of the block
}
