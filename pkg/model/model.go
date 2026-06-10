// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package model defines the shared data types produced by all open-inspector
// loaders (config, graph, schema). Types are intentionally plain structs so
// they serialize cleanly to JSON and remain stable across releases.
package model

// SchemaVersion identifies the JSON output schema. Treat a bump like a
// database migration: it invalidates every downstream consumer.
//
// Bump when ANY of these happen:
//   - an existing field is renamed or removed;
//   - an existing field's JSON type changes (string -> object, etc.);
//   - an existing field becomes required (was omitempty);
//   - an existing enum value is removed.
//
// Do NOT bump when:
//   - a new optional field is added;
//   - a new enum value is added (consumers should ignore unknowns);
//   - a field gains omitempty (that only ever loosens the contract).
//
// The machine-readable contract lives at docs/schema/v1.json; regenerate
// it with `make jsonschema` after any model change.
const SchemaVersion = 1
