// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Command jsonschemagen regenerates the JSON Schema for the open-inspector
// --json output. It reflects pkg/model.Module (the payload carried under
// the envelope's "module" key) and writes both a machine-readable
// docs/schema/v1.json and a human-readable docs/schema/v1.md.
//
// Run via `make jsonschema`; CI diff-checks the committed output.
package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"

	"github.com/remoterabbit/open-inspector/pkg/model"
)

func main() {
	reflector := &jsonschema.Reflector{ExpandedStruct: true}

	// Surface the model's Go doc comments as JSON Schema "description"
	// values so the generated docs explain each type and field. Run from
	// the repo root (via `make jsonschema`); the path is resolved relative
	// to the working directory.
	if err := reflector.AddGoComments("github.com/remoterabbit/open-inspector", "./pkg/model"); err != nil {
		log.Fatalf("read model comments: %v", err)
	}

	schema := reflector.Reflect(&model.Module{})

	bytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		log.Fatalf("marshal schema: %v", err)
	}

	outputDir := filepath.Join("docs", "schema")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		log.Fatalf("mkdir %s: %v", outputDir, err)
	}

	jsonPath := filepath.Join(outputDir, "v1.json")
	if err := os.WriteFile(jsonPath, append(bytes, '\n'), 0o644); err != nil {
		log.Fatalf("write %s: %v", jsonPath, err)
	}

	markdownPath := filepath.Join(outputDir, "v1.md")
	if err := os.WriteFile(markdownPath, []byte(renderMarkdown(schema)), 0o644); err != nil {
		log.Fatalf("write %s: %v", markdownPath, err)
	}
}
