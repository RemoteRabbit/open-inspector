// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package schema

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testSchemaPath returns the absolute path to a checked-in schema fixture
// under pkg/config/testdata/schemas/.
func testSchemaPath(t *testing.T, name string) string {
	t.Helper()
	p, err := filepath.Abs(filepath.Join("..", "config", "testdata", "schemas", name))
	if err != nil {
		t.Fatalf("resolve schema path: %v", err)
	}
	return p
}

// loadTestSchema opens and loads a checked-in schema fixture.
func loadTestSchema(t *testing.T, name string) *Schema {
	t.Helper()
	f, err := os.Open(testSchemaPath(t, name))
	if err != nil {
		t.Fatalf("open schema: %v", err)
	}
	defer func() { _ = f.Close() }()
	s, err := Load(f)
	if err != nil {
		t.Fatalf("load schema: %v", err)
	}
	return s
}

func TestLoad_NullSchema(t *testing.T) {
	t.Parallel()

	s := loadTestSchema(t, "null.json")

	resource, source := s.LookupResource("registry.opentofu.org/hashicorp/null", "null_resource")
	if resource == nil {
		t.Fatalf("null_resource missing; source=%q", source)
		return
	}
	if _, ok := resource.Block.Attributes["triggers"]; !ok {
		t.Errorf("null_resource.triggers missing from schema")
	}
}

func TestLoad_LookupFallsBackAcrossRegistry(t *testing.T) {
	t.Parallel()

	s := loadTestSchema(t, "null.json")

	// The config-declared source is the terraform.io registry, but the
	// schema fixture was emitted by OpenTofu under opentofu.org. The
	// fallback should still find the resource.
	resource, source := s.LookupResource("registry.terraform.io/hashicorp/null", "null_resource")
	if resource == nil {
		t.Fatalf("expected fallback lookup to find null_resource")
	}
	if !strings.Contains(source, "hashicorp/null") {
		t.Errorf("resolved source = %q, want it to contain hashicorp/null", source)
	}
}

func TestLoad_LookupDataSource(t *testing.T) {
	t.Parallel()

	s := loadTestSchema(t, "null.json")

	dataSource, source := s.LookupDataSource("registry.opentofu.org/hashicorp/null", "null_data_source")
	if dataSource == nil {
		t.Fatalf("null_data_source missing; source=%q", source)
		return
	}
	deprecated, ok := dataSource.Block.Attributes["id"]
	if !ok {
		t.Fatalf("null_data_source.id missing from schema")
	}
	if !deprecated.Deprecated {
		t.Errorf("null_data_source.id should be marked deprecated")
	}
}

func TestLoad_MissingFormatVersion(t *testing.T) {
	t.Parallel()

	_, err := Load(strings.NewReader(`{"provider_schemas":{}}`))
	if err == nil {
		t.Fatalf("expected error for document missing format_version")
	}
}
