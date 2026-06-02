// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package inspector is the top-level facade for open-inspector. Library
// consumers should depend on this package; lower-level packages
// (config, graph, schema) may change without notice until v1.
package inspector

import (
	"fmt"
	"path/filepath"

	"github.com/remoterabbit/open-inspector/pkg/model"
)

// Version is the semantic version of the open-inspector library.
// It is reported by the CLI and embedded in JSON output.
const Version = "0.0.1"

// Inspect performs a (currently stub) inspection of the Terraform/OpenTofu
// module rooted at dir and returns the resulting model. Future steps will
// add HCL parsing, child module resolution, and optional schema enrichment.
func Inspect(dir string) (*model.Module, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolve module path: %w", err)
	}
	return &model.Module{Path: abs}, nil
}
