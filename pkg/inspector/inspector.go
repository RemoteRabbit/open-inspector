// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package inspector is the top-level facade for open-inspector. Library
// consumers should depend on this package; lower-level packages
// (config, graph, schema) may change without notice until v1.
package inspector

import (
	"github.com/remoterabbit/open-inspector/pkg/config"
	"github.com/remoterabbit/open-inspector/pkg/graph"
	"github.com/remoterabbit/open-inspector/pkg/model"
)

// Version is the semantic version of the open-inspector library.
// It is reported by the CLI and embedded in JSON output.
//
// The value is a var (not a const) so that release builds can override
// it via -ldflags "-X github.com/remoterabbit/open-inspector/pkg/inspector.Version=...".
// release-please rewrites the literal below on each release PR; the
// trailing marker comment is required - do not remove it.
var Version = "0.2.0" // x-release-please-version

// Inspect performs a (currently stub) inspection of the Terraform/OpenTofu
// module rooted at dir and returns the resulting model. Future steps will
// add HCL parsing, child module resolution, and optional schema enrichment.
func Inspect(dir string, opts ...Option) (*model.Module, error) {
	defaults := defaultOptions()
	for _, fn := range opts {
		fn(&defaults)
	}

	module, err := config.Load(dir)
	if err != nil {
		return nil, err
	}

	if defaults.moduleGraph {
		graph.Build(module, defaults.toGraphOptions())
	}
	return module, nil
}
