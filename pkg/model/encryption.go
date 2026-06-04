// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

// Encryption corresponds to OpenTofu's `terraform { encryption {} }`
// configuration tree (OpenTofu 1.7+). Terraform rejects this block.
type Encryption struct {
	KeyProviders       []EncryptionKeyProvider `json:"key_providers,omitempty"`
	Methods            []EncryptionMethod      `json:"methods,omitempty"`
	State              *EncryptionTarget       `json:"state,omitempty"`
	Plan               *EncryptionTarget       `json:"plan,omitempty"`
	RemoteStateSources []EncryptionRemoteState `json:"remote_state_data_sources,omitempty"`
	Range              Range                   `json:"range"`
}

// EncryptionKeyProvider describes one `key_provider "<type>" "<name>" {}`
// inside an encryption block. The body schema is provider-defined, so
// each attribute is captured verbatim as an Expression.
type EncryptionKeyProvider struct {
	Type  string                `json:"type"`
	Name  string                `json:"name"`
	Body  map[string]Expression `json:"body,omitempty"`
	Range Range                 `json:"range"`
}

// EncryptionMethod describes one `method "<type>" "<name>" {}` inside
// an encryption block. Body shape is method-defined; captured verbatim.
type EncryptionMethod struct {
	Type  string                `json:"type"`
	Name  string                `json:"name"`
	Body  map[string]Expression `json:"body,omitempty"`
	Range Range                 `json:"range"`
}

// EncryptionTarget describes the `state {}` or `plan {}` sub-block:
// a required `method` reference and an optional `fallback`.
type EncryptionTarget struct {
	Method   Expression `json:"method"`
	Fallback Expression `json:"fallback,omitempty"`
	Range    Range      `json:"range"`
}

// EncryptionRemoteState describes a `remote_state_data_sources {}` entry.
// Its body schema is open-ended; captured verbatim.
type EncryptionRemoteState struct {
	Name  string                `json:"name"`
	Body  map[string]Expression `json:"body,omitempty"`
	Range Range                 `json:"range"`
}
