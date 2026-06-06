// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package schema loads `tofu/terraform providers schema -json` documents
// and enriches a model.Module with schema-derived findings: unknown
// attributes, deprecated attributes, and missing required attributes.
//
// Enrichment is best-effort and static-only: it sees the attribute names
// a user wrote, not their values. The terraform-json types are wrapped
// behind Schema so consumers depend on this package's surface rather than
// the upstream library directly.
package schema

import tfjson "github.com/hashicorp/terraform-json"

// Schema wraps terraform-json's ProviderSchemas, exposing only the surface
// the enricher needs. It is constructed by Load or Auto.
type Schema struct {
	underlying *tfjson.ProviderSchemas
}
