// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package schema

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/remoterabbit/open-inspector/pkg/config"
	"github.com/remoterabbit/open-inspector/pkg/model"
)

// findResource returns the named managed resource from rs or fails.
func findResource(t *testing.T, rs []model.Resource, typ, name string) model.Resource {
	t.Helper()
	for _, r := range rs {
		if r.Type == typ && r.Name == name {
			return r
		}
	}
	t.Fatalf("resource %s.%s not found", typ, name)
	return model.Resource{}
}

func TestEnrich_UnknownAttr(t *testing.T) {
	t.Parallel()

	s := loadTestSchema(t, "null.json")
	dir, err := filepath.Abs(filepath.Join("..", "..", "testdata", "fixtures", "simple-with-typo"))
	if err != nil {
		t.Fatalf("resolve fixture: %v", err)
	}
	mod, err := config.Load(dir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	Enrich(mod, s)

	example := findResource(t, mod.ManagedResources, "null_resource", "example")
	if example.SchemaFindings == nil {
		t.Fatalf("expected SchemaFindings to be populated")
	}
	if got := len(example.SchemaFindings.UnknownAttrs); got != 1 {
		t.Fatalf("UnknownAttrs count = %d, want 1: %#v", got, example.SchemaFindings)
	}
	if name := example.SchemaFindings.UnknownAttrs[0].Name; name != "trigerz" {
		t.Errorf("UnknownAttrs[0].Name = %q, want %q", name, "trigerz")
	}
	if example.SchemaFindings.UnknownAttrs[0].Range.Filename == "" {
		t.Errorf("UnknownAttrs[0].Range.Filename is empty")
	}
}

func TestEnrich_NoFindingsLeavesNil(t *testing.T) {
	t.Parallel()

	s := loadTestSchema(t, "null.json")
	dir, err := filepath.Abs(filepath.Join("..", "..", "testdata", "fixtures", "simple"))
	if err != nil {
		t.Fatalf("resolve fixture: %v", err)
	}
	mod, err := config.Load(dir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	Enrich(mod, s)

	// simple/ sets only the valid `triggers` attribute, so there is
	// nothing to report and SchemaFindings stays nil (omitted from JSON).
	example := findResource(t, mod.ManagedResources, "null_resource", "example")
	if example.SchemaFindings != nil {
		t.Errorf("expected nil SchemaFindings, got %#v", example.SchemaFindings)
	}
}

func TestEnrich_DeprecatedAttr(t *testing.T) {
	t.Parallel()

	s := loadTestSchema(t, "null.json")
	mod := &model.Module{
		RequiredProviders: map[string]model.ProviderRequirement{
			"null": {Source: "hashicorp/null"},
		},
		DataResources: []model.Resource{{
			Mode:      model.DataResourceMode,
			Type:      "null_data_source",
			Name:      "example",
			AttrNames: []string{"id"}, // deprecated in the null schema
		}},
	}

	Enrich(mod, s)

	findings := mod.DataResources[0].SchemaFindings
	if findings == nil || len(findings.DeprecatedAttrs) != 1 {
		t.Fatalf("expected one DeprecatedAttrs entry, got %#v", findings)
	}
	if findings.DeprecatedAttrs[0].Name != "id" {
		t.Errorf("DeprecatedAttrs[0].Name = %q, want %q", findings.DeprecatedAttrs[0].Name, "id")
	}
	if findings.DeprecatedAttrs[0].Message == "" {
		t.Errorf("expected a non-empty deprecation message")
	}
}

// missingRequiredSchema is a hand-written provider schema with a required
// attribute, used to exercise the missing-required path without depending
// on a real provider that declares one.
const missingRequiredSchema = `{
  "format_version": "1.0",
  "provider_schemas": {
    "registry.terraform.io/hashicorp/example": {
      "resource_schemas": {
        "example_thing": {
          "version": 0,
          "block": {
            "attributes": {
              "name":     {"type": "string", "required": true},
              "size":     {"type": "string", "optional": true},
              "id":       {"type": "string", "computed": true}
            }
          }
        }
      }
    }
  }
}`

func TestEnrich_MissingRequired(t *testing.T) {
	t.Parallel()

	s, err := Load(strings.NewReader(missingRequiredSchema))
	if err != nil {
		t.Fatalf("load schema: %v", err)
	}
	mod := &model.Module{
		RequiredProviders: map[string]model.ProviderRequirement{
			"example": {Source: "hashicorp/example"},
		},
		ManagedResources: []model.Resource{{
			Mode:      model.ManagedResourceMode,
			Type:      "example_thing",
			Name:      "thing",
			AttrNames: []string{"size"}, // sets the optional attr, omits required `name`
		}},
	}

	Enrich(mod, s)

	findings := mod.ManagedResources[0].SchemaFindings
	if findings == nil {
		t.Fatalf("expected SchemaFindings")
		return
	}
	if want := []string{"name"}; !reflect.DeepEqual(findings.MissingRequired, want) {
		t.Errorf("MissingRequired = %v, want %v (computed `id` must be excluded)", findings.MissingRequired, want)
	}
}

// ephemeralSchema is a hand-written provider schema exposing an ephemeral
// resource, used to exercise the ephemeral enrichment path (no real
// provider in the test fixtures declares one).
const ephemeralSchema = `{
  "format_version": "1.0",
  "provider_schemas": {
    "registry.terraform.io/hashicorp/example": {
      "ephemeral_resource_schemas": {
        "example_secret": {
          "version": 0,
          "block": {
            "attributes": {
              "name": {"type": "string", "required": true},
              "ttl":  {"type": "string", "optional": true}
            }
          }
        }
      }
    }
  }
}`

func TestEnrich_EphemeralResource(t *testing.T) {
	t.Parallel()

	s, err := Load(strings.NewReader(ephemeralSchema))
	if err != nil {
		t.Fatalf("load schema: %v", err)
	}
	mod := &model.Module{
		RequiredProviders: map[string]model.ProviderRequirement{
			"example": {Source: "hashicorp/example"},
		},
		EphemeralResources: []model.EphemeralResource{{
			Type:      "example_secret",
			Name:      "token",
			AttrNames: []string{"bogus"}, // unknown; also omits required `name`
		}},
	}

	Enrich(mod, s)

	findings := mod.EphemeralResources[0].SchemaFindings
	if findings == nil {
		t.Fatalf("expected SchemaFindings on the ephemeral resource")
		return
	}
	if len(findings.UnknownAttrs) != 1 || findings.UnknownAttrs[0].Name != "bogus" {
		t.Errorf("UnknownAttrs = %#v, want one entry named %q", findings.UnknownAttrs, "bogus")
	}
	if want := []string{"name"}; !reflect.DeepEqual(findings.MissingRequired, want) {
		t.Errorf("MissingRequired = %v, want %v", findings.MissingRequired, want)
	}
}

func TestProviderName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		meta, resourceType, want string
	}{
		{"", "aws_instance", "aws"},
		{"aws.east", "aws_instance", "aws"},
		{"aws", "aws_instance", "aws"},
		{"", "null_resource", "null"},
	}
	for _, tc := range cases {
		if got := providerName(tc.meta, tc.resourceType); got != tc.want {
			t.Errorf("providerName(%q, %q) = %q, want %q", tc.meta, tc.resourceType, got, tc.want)
		}
	}
}
