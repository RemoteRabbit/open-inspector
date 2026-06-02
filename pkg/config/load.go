// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package config loads Terraform/OpenTofu configuration from a module
// directory into the open-inspector model.
package config

import (
	"fmt"
	"path/filepath"

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
	_, parseDiags := parse(fs)

	module := &model.Module{Path: abs}
	module.Diagnostics = append(module.Diagnostics, walkDiags...)
	module.Diagnostics = append(module.Diagnostics, parseDiags...)

	return module, nil
}
