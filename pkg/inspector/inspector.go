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
	"github.com/remoterabbit/open-inspector/pkg/schema"
)

// Version is the semantic version of the open-inspector library.
// It is reported by the CLI and embedded in JSON output.
//
// The value is a var (not a const) so that release builds can override
// it via -ldflags "-X github.com/remoterabbit/open-inspector/pkg/inspector.Version=...".
// release-please rewrites the literal below on each release PR; the
// trailing marker comment is required - do not remove it.
var Version = "0.4.1" // x-release-please-version

// Inspect performs a (currently stub) inspection of the Terraform/OpenTofu
// module rooted at dir and returns the resulting model. Future steps will
// add HCL parsing, child module resolution, and optional schema enrichment.
func Inspect(dir string, opts ...Option) (*model.Module, error) {
	defaults := defaultOptions()
	for _, fn := range opts {
		fn(&defaults)
	}
	if defaults.schemaErr != nil {
		return nil, defaults.schemaErr
	}

	module, err := config.Load(dir)
	if err != nil {
		return nil, err
	}

	enrichSchema(module, dir, &defaults)

	if defaults.moduleGraph {
		graph.Build(module, defaults.toGraphOptions())
	}
	return module, nil
}

// enrichSchema applies provider-schema enrichment when the caller opted in
// via WithSchema or WithSchemaAuto. Auto-detection failures are recorded
// as a warning diagnostic on the module rather than aborting inspection.
func enrichSchema(module *model.Module, dir string, defaults *options) {
	if defaults.schemaAuto && defaults.schema == nil {
		loaded, _, err := schema.Auto(dir)
		if err != nil {
			module.Diagnostics = append(module.Diagnostics, model.Diagnostic{
				Severity: model.SeverityWarning,
				Summary:  "schema auto-detection failed",
				Detail:   err.Error(),
			})
		} else {
			defaults.schema = loaded
		}
	}
	if defaults.schema != nil {
		schema.Enrich(module, defaults.schema)
	}
}
