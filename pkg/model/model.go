// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package model defines the shared data types produced by all open-inspector
// loaders (config, graph, schema). Types are intentionally plain structs so
// they serialize cleanly to JSON and remain stable across releases.
package model

// SchemaVersion identifies the JSON output schema. Bump on any
// backwards-incompatible change to exported fields.
const SchemaVersion = 1

// Module is the root inspection result for a single Terraform/OpenTofu
// module directory. Fields will grow as loaders are implemented.
type Module struct {
	// Path is the absolute filesystem path of the inspected module.
	Path string `json:"path"`
}
