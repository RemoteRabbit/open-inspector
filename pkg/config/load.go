// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package config loads Terraform/OpenTofu configuration from a module
// directory into the open-inspector model.
package config

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/hashicorp/hcl/v2"

	"github.com/remoterabbit/open-inspector/pkg/model"
)

// Load parses every Terraform/OpenTofu configuration file directly inside dir and returns
// a Module describing what it found. Errors in the source files are reported as Diagnostics
// on the result; filesystem and other system errors are returned as a Go error.
func Load(dir string) (*model.Module, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolve module path: %w", err)
	}

	fs, walkDiags := walk(abs)
	parsed, parseDiags := parse(fs)

	module := &model.Module{Path: abs}
	module.Diagnostics = append(module.Diagnostics, walkDiags...)
	module.Diagnostics = append(module.Diagnostics, parseDiags...)

	decodeFiles(parsed.files, module)

	// Locals are decoded from a map (JustAttributes), so their order is
	// non-deterministic across runs. Sort by name once at the end so that
	// snapshots, JSON output, and downstream consumers see a stable order.
	sort.Slice(module.Locals, func(i, j int) bool {
		return module.Locals[i].Name < module.Locals[j].Name
	})

	return module, nil
}

// decodeFiles walks every parsed file, applies the root schema, and
// dispatches each recognised top-level block to its decoder. Blocks
// not listed in rootSchema are left in the file's leftover body.
func decodeFiles(files []*hcl.File, module *model.Module) {
	for _, file := range files {
		if file == nil {
			continue
		}
		content, _, hdiag := file.Body.PartialContent(rootSchema)
		module.Diagnostics = append(module.Diagnostics, model.DiagnosticsFromHCL(hdiag)...)

		for _, block := range content.Blocks {
			switch block.Type {
			case "terraform":
				module.Diagnostics = append(module.Diagnostics, decodeTerraformBlock(block, module)...)
			case "provider":
				module.Diagnostics = append(module.Diagnostics, decodeProviderBlock(block, module)...)
			case "variable":
				module.Diagnostics = append(module.Diagnostics, decodeVariableBlock(block, file.Bytes, module)...)
			case "output":
				module.Diagnostics = append(module.Diagnostics, decodeOuputsBlock(block, file.Bytes, module)...)
			case "locals":
				module.Diagnostics = append(module.Diagnostics, decodeLocalsBlock(block, file.Bytes, module)...)
			case "resource":
				module.Diagnostics = append(module.Diagnostics, decodeResourceBlock(block, file.Bytes, model.ManagedResourceMode, module)...)
			case "data":
				module.Diagnostics = append(module.Diagnostics, decodeResourceBlock(block, file.Bytes, model.DataResourceMode, module)...)
			case "module":
				module.Diagnostics = append(module.Diagnostics, decodeModuleCallBlock(block, file.Bytes, module)...)
			case "moved":
				module.Diagnostics = append(module.Diagnostics, decodeMovedBlock(block, module)...)
			case "import":
				module.Diagnostics = append(module.Diagnostics, decodeImportBlock(block, file.Bytes, module)...)
			case "removed":
				module.Diagnostics = append(module.Diagnostics, decodeRemovedBlock(block, module)...)
			}
		}
	}
}
