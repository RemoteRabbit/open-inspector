// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package inspector_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/remoterabbit/open-inspector/pkg/inspector"
	"github.com/remoterabbit/open-inspector/pkg/model"
)

func TestInspectReturnsAbsolutePath(t *testing.T) {
	t.Parallel()

	mod, err := inspector.Inspect(".")
	if err != nil {
		t.Fatalf("Inspect returned error: %v", err)
	}
	if mod == nil {
		t.Fatal("Inspect returned nil module")
		return
	}
	if !filepath.IsAbs(mod.Path) {
		t.Errorf("module path %q is not absolute", mod.Path)
	}
}

func TestInspect_WithModuleGraph_Local(t *testing.T) {
	// Fixtures live at the repo root; pkg tests reach them via ../../
	// (matches pkg/config's filepath.Join("..", "..", "testdata", ...)).
	module, err := inspector.Inspect("../../testdata/fixtures/multi-module",
		inspector.WithModuleGraph())
	if err != nil {
		t.Fatalf("inspect: %v", err)
	}
	if len(module.Children) != 2 {
		t.Errorf("Children: want 2, got %d", len(module.Children))
	}
	if module.Children["network"].Module == nil {
		t.Errorf("network child not loaded")
	}
}

// findManaged returns the named managed resource from the module or fails.
func findManaged(t *testing.T, module *model.Module, typ, name string) model.Resource {
	t.Helper()
	for _, r := range module.ManagedResources {
		if r.Type == typ && r.Name == name {
			return r
		}
	}
	t.Fatalf("managed resource %s.%s not found", typ, name)
	return model.Resource{}
}

func TestInspect_WithSchema_Enriches(t *testing.T) {
	t.Parallel()

	schemaFile, err := os.Open("../config/testdata/schemas/null.json")
	if err != nil {
		t.Fatalf("open schema: %v", err)
	}
	defer func() { _ = schemaFile.Close() }()

	module, err := inspector.Inspect("../../testdata/fixtures/simple-with-typo",
		inspector.WithSchema(schemaFile))
	if err != nil {
		t.Fatalf("inspect: %v", err)
	}

	example := findManaged(t, module, "null_resource", "example")
	if example.SchemaFindings == nil || len(example.SchemaFindings.UnknownAttrs) != 1 {
		t.Fatalf("expected one unknown-attr finding, got %#v", example.SchemaFindings)
	}
	if got := example.SchemaFindings.UnknownAttrs[0].Name; got != "trigerz" {
		t.Errorf("unknown attr = %q, want %q", got, "trigerz")
	}
}

func TestInspect_WithSchema_NoFindingsForValidConfig(t *testing.T) {
	t.Parallel()

	schemaFile, err := os.Open("../config/testdata/schemas/null.json")
	if err != nil {
		t.Fatalf("open schema: %v", err)
	}
	defer func() { _ = schemaFile.Close() }()

	module, err := inspector.Inspect("../../testdata/fixtures/simple",
		inspector.WithSchema(schemaFile))
	if err != nil {
		t.Fatalf("inspect: %v", err)
	}

	example := findManaged(t, module, "null_resource", "example")
	if example.SchemaFindings != nil {
		t.Errorf("expected no findings for a valid config, got %#v", example.SchemaFindings)
	}
}

func TestInspect_WithSchema_DecodeErrorSurfaces(t *testing.T) {
	t.Parallel()

	_, err := inspector.Inspect("../../testdata/fixtures/simple",
		inspector.WithSchema(strings.NewReader("not valid json")))
	if err == nil {
		t.Fatalf("expected an error for an undecodable schema document")
	}
}

func TestInspect_WithSchemaAuto_UninitializedWarns(t *testing.T) {
	t.Parallel()

	// The fixture has not been `init`-ed, so auto-detection cannot produce
	// a schema; it must surface a warning diagnostic instead of aborting.
	module, err := inspector.Inspect("../../testdata/fixtures/simple",
		inspector.WithSchemaAuto())
	if err != nil {
		t.Fatalf("inspect: %v", err)
	}

	found := false
	for _, diag := range module.Diagnostics {
		if diag.Severity == model.SeverityWarning &&
			strings.Contains(diag.Summary, "schema auto-detection failed") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a schema auto-detection warning, got %#v", module.Diagnostics)
	}
}
