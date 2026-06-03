// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

// fixturePath returns the absolute path to a named directory under
// testdata/fixtures/.
func fixturePath(t *testing.T, name string) string {
	t.Helper()
	p, err := filepath.Abs(filepath.Join("..", "..", "testdata", "fixtures", name))
	if err != nil {
		t.Fatalf("resolve fixture path: %v", err)
	}
	return p
}

func TestLoad_Simple_RequiredProviders(t *testing.T) {
	t.Parallel()

	mod, err := Load(fixturePath(t, "simple"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if mod.Diagnostics.HasErrors() {
		t.Fatalf("unexpected error diagnostics: %#v", mod.Diagnostics)
	}

	if got, want := mod.RequiredCore, []string{">= 1.5.0"}; !reflect.DeepEqual(got, want) {
		t.Errorf("RequiredCore = %v, want %v", got, want)
	}

	if got := len(mod.RequiredProviders); got != 1 {
		t.Fatalf("RequiredProviders count = %d, want 1", got)
	}
	null, ok := mod.RequiredProviders["null"]
	if !ok {
		t.Fatalf("RequiredProviders[null] missing; have %v", mod.RequiredProviders)
	}
	if null.Source != "hashicorp/null" {
		t.Errorf("null.Source = %q, want %q", null.Source, "hashicorp/null")
	}
	if got, want := null.VersionConstraints, []string{"~> 3.2"}; !reflect.DeepEqual(got, want) {
		t.Errorf("null.VersionConstraints = %v, want %v", got, want)
	}
	if len(null.ConfigurationAliases) != 0 {
		t.Errorf("null.ConfigurationAliases = %v, want none", null.ConfigurationAliases)
	}
	if null.Range.Filename == "" {
		t.Errorf("null.Range.Filename is empty")
	}

	if len(mod.Providers) != 0 {
		t.Errorf("Providers = %v, want none for simple fixture", mod.Providers)
	}
}

func TestLoad_Providers_FullDecoding(t *testing.T) {
	t.Parallel()

	mod, err := Load(fixturePath(t, "providers"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if mod.Diagnostics.HasErrors() {
		t.Fatalf("unexpected error diagnostics: %#v", mod.Diagnostics)
	}

	if got, want := mod.RequiredCore, []string{">= 1.5.0"}; !reflect.DeepEqual(got, want) {
		t.Errorf("RequiredCore = %v, want %v", got, want)
	}

	wantProvs := map[string]struct {
		source   string
		versions []string
		aliases  []string
	}{
		"aws":    {"hashicorp/aws", []string{"~> 5.0"}, []string{"aws.east", "aws.west"}},
		"random": {"hashicorp/random", []string{">= 3.0, < 4.0"}, nil},
		"http":   {"hashicorp/http", []string{"~> 3.4"}, nil},
	}
	if got, want := len(mod.RequiredProviders), len(wantProvs); got != want {
		t.Fatalf("RequiredProviders count = %d, want %d", got, want)
	}
	for name, want := range wantProvs {
		got, ok := mod.RequiredProviders[name]
		if !ok {
			t.Errorf("RequiredProviders[%s] missing", name)
			continue
		}
		if got.Source != want.source {
			t.Errorf("RequiredProviders[%s].Source = %q, want %q", name, got.Source, want.source)
		}
		if !reflect.DeepEqual(got.VersionConstraints, want.versions) {
			t.Errorf("RequiredProviders[%s].VersionConstraints = %v, want %v",
				name, got.VersionConstraints, want.versions)
		}
		if !reflect.DeepEqual(emptyToNil(got.ConfigurationAliases), want.aliases) {
			t.Errorf("RequiredProviders[%s].ConfigurationAliases = %v, want %v",
				name, got.ConfigurationAliases, want.aliases)
		}
	}

	// Providers: aws (alias east), aws (alias west), random (no alias).
	// File order is deterministic but Providers slice order follows
	// file traversal; sort for a stable assertion.
	type pc struct{ Name, Alias string }
	got := make([]pc, 0, len(mod.Providers))
	for _, p := range mod.Providers {
		got = append(got, pc{p.Name, p.Alias})
	}
	sort.Slice(got, func(i, j int) bool {
		if got[i].Name != got[j].Name {
			return got[i].Name < got[j].Name
		}
		return got[i].Alias < got[j].Alias
	})
	want := []pc{
		{"aws", "east"},
		{"aws", "west"},
		{"random", ""},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Providers = %v, want %v", got, want)
	}

	for i, p := range mod.Providers {
		if p.Range.Filename == "" {
			t.Errorf("Providers[%d].Range.Filename is empty", i)
		}
	}
}

// emptyToNil normalises an empty slice into nil so reflect.DeepEqual
// treats `[]string{}` and `nil` the same way callers expect.
func emptyToNil(s []string) []string {
	if len(s) == 0 {
		return nil
	}
	return s
}
