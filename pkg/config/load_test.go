// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/remoterabbit/open-inspector/pkg/model"
)

// update regenerates the golden snapshot files under
// pkg/config/testdata/golden/ when set. Run:
//
//	go test ./pkg/config -run TestLoad_Snapshots -update
var update = flag.Bool("update", false, "regenerate golden snapshots")

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

// TestLoad_Fixtures is a parameterized smoke table: one entry per
// fixture, asserting only top-level counts. It's cheap to extend and
// catches gross regressions across the whole loader surface.
func TestLoad_Fixtures(t *testing.T) {
	t.Parallel()

	cases := []struct {
		dir                 string
		wantVars            int
		wantOutputs         int
		wantLocals          int
		wantManagedRes      int
		wantDataRes         int
		wantModuleCalls     int
		wantRequiredProvs   int
		wantProviderConfigs int
		wantErrorDiagnostic bool
	}{
		{dir: "simple", wantVars: 1, wantOutputs: 1, wantManagedRes: 1, wantRequiredProvs: 1},
		{dir: "variables-and-outputs", wantVars: 5, wantOutputs: 3, wantLocals: 2},
		{dir: "providers", wantRequiredProvs: 3, wantDataRes: 1, wantManagedRes: 1, wantProviderConfigs: 3},
		{dir: "resources-count-foreach", wantManagedRes: 2, wantDataRes: 1, wantModuleCalls: 2, wantVars: 2},
		{dir: "json-config", wantVars: 1, wantOutputs: 1, wantManagedRes: 1, wantRequiredProvs: 1},
		{dir: "multi-module", wantModuleCalls: 2, wantOutputs: 1, wantVars: 1},
		{dir: "module-sources", wantModuleCalls: 6},
		{dir: "resources-full", wantManagedRes: 2, wantModuleCalls: 1, wantRequiredProvs: 1, wantProviderConfigs: 1},
		{dir: "tofu-extension", wantVars: 2, wantOutputs: 2},
		{dir: "overrides", wantVars: 1, wantManagedRes: 1},
		{dir: "invalid/syntax-error", wantErrorDiagnostic: true},
		{dir: "invalid/malformed-validation", wantErrorDiagnostic: true},
		{dir: "invalid/non-literal-attrs", wantErrorDiagnostic: true},
	}
	for _, tc := range cases {
		t.Run(tc.dir, func(t *testing.T) {
			t.Parallel()
			mod, err := Load(fixturePath(t, tc.dir))
			if err != nil {
				t.Fatalf("Load: %v", err)
			}

			if tc.wantErrorDiagnostic {
				if !mod.Diagnostics.HasErrors() {
					t.Errorf("expected at least one error diagnostic, got none")
				}
				// Don't assert counts on broken fixtures: the parser
				// produces partial output whose exact shape isn't a
				// useful baseline.
				return
			}
			if mod.Diagnostics.HasErrors() {
				t.Errorf("unexpected error diagnostics: %#v", mod.Diagnostics)
			}

			if got := len(mod.Variables); got != tc.wantVars {
				t.Errorf("Variables: want %d, got %d", tc.wantVars, got)
			}
			if got := len(mod.Outputs); got != tc.wantOutputs {
				t.Errorf("Outputs: want %d, got %d", tc.wantOutputs, got)
			}
			if got := len(mod.Locals); got != tc.wantLocals {
				t.Errorf("Locals: want %d, got %d", tc.wantLocals, got)
			}
			if got := len(mod.ManagedResources); got != tc.wantManagedRes {
				t.Errorf("ManagedResources: want %d, got %d", tc.wantManagedRes, got)
			}
			if got := len(mod.DataResources); got != tc.wantDataRes {
				t.Errorf("DataResources: want %d, got %d", tc.wantDataRes, got)
			}
			if got := len(mod.ModuleCalls); got != tc.wantModuleCalls {
				t.Errorf("ModuleCalls: want %d, got %d", tc.wantModuleCalls, got)
			}
			if got := len(mod.RequiredProviders); got != tc.wantRequiredProvs {
				t.Errorf("RequiredProviders: want %d, got %d", tc.wantRequiredProvs, got)
			}
			if got := len(mod.Providers); got != tc.wantProviderConfigs {
				t.Errorf("Providers: want %d, got %d", tc.wantProviderConfigs, got)
			}
		})
	}
}

// snapshotFixtures lists the fixtures whose full JSON output is
// captured under testdata/golden/. Invalid fixtures are excluded —
// their loader output isn't a useful baseline.
var snapshotFixtures = []string{
	"simple",
	"variables-and-outputs",
	"providers",
	"resources-count-foreach",
	"resources-full",
	"json-config",
	"multi-module",
	"module-sources",
	"tofu-extension",
	"overrides",
}

// TestLoad_Snapshots round-trips every fixture in snapshotFixtures
// through Load → JSON and compares against a golden file. Run with
// -update to regenerate the goldens after an intentional loader change.
func TestLoad_Snapshots(t *testing.T) {
	for _, dir := range snapshotFixtures {
		t.Run(dir, func(t *testing.T) {
			mod, err := Load(fixturePath(t, dir))
			if err != nil {
				t.Fatalf("Load: %v", err)
			}

			normalizeForSnapshot(mod, dir)

			got, err := json.MarshalIndent(mod, "", "  ")
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			got = append(got, '\n')

			golden := filepath.Join("testdata", "golden", dir+".json")
			if *update {
				if err := os.MkdirAll(filepath.Dir(golden), 0o755); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
				if err := os.WriteFile(golden, got, 0o644); err != nil {
					t.Fatalf("write golden: %v", err)
				}
				return
			}

			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("read golden: %v (run `go test ./pkg/config -run TestLoad_Snapshots -update` to create)", err)
			}
			if diff := cmp.Diff(string(want), string(got)); diff != "" {
				t.Errorf("snapshot mismatch for %s (-want +got):\n%s", dir, diff)
			}
		})
	}
}

// normalizeForSnapshot rewrites machine-specific fields on a Module so
// the resulting JSON snapshot is reproducible across machines and CI
// runners. Only the absolute paths (Module.Path and every Range.Filename)
// vary between hosts; everything else is already derived from fixture
// content.
func normalizeForSnapshot(mod *model.Module, fixture string) {
	mod.Path = "<fixture>/" + fixture

	rewrite := func(r *model.Range) {
		if r == nil || r.Filename == "" {
			return
		}
		// Keep only the path relative to the fixture root.
		idx := strings.Index(r.Filename, "testdata/fixtures/")
		if idx < 0 {
			return
		}
		rel := r.Filename[idx+len("testdata/fixtures/"):]
		r.Filename = "<fixture>/" + rel
	}

	for k, p := range mod.RequiredProviders {
		rewrite(&p.Range)
		mod.RequiredProviders[k] = p
	}
	for i := range mod.Variables {
		rewrite(&mod.Variables[i].Range)
		if mod.Variables[i].Default != nil {
			rewrite(&mod.Variables[i].Default.Range)
		}
		for j := range mod.Variables[i].Validations {
			rewrite(&mod.Variables[i].Validations[j].Range)
			rewrite(&mod.Variables[i].Validations[j].Condition.Range)
			rewrite(&mod.Variables[i].Validations[j].ErrorMessage.Range)
		}
	}
	for i := range mod.Outputs {
		rewrite(&mod.Outputs[i].Range)
		rewrite(&mod.Outputs[i].Value.Range)
	}
	for i := range mod.Locals {
		rewrite(&mod.Locals[i].Range)
		rewrite(&mod.Locals[i].Value.Range)
	}
	rewriteResources := func(rs []model.Resource) {
		for i := range rs {
			rewrite(&rs[i].Range)
			if rs[i].Count != nil {
				rewrite(&rs[i].Count.Range)
			}
			if rs[i].ForEach != nil {
				rewrite(&rs[i].ForEach.Range)
			}
			if rs[i].Lifecycle != nil {
				for j := range rs[i].Lifecycle.Preconditions {
					rewrite(&rs[i].Lifecycle.Preconditions[j].Range)
					rewrite(&rs[i].Lifecycle.Preconditions[j].Condition.Range)
					rewrite(&rs[i].Lifecycle.Preconditions[j].ErrorMessage.Range)
				}
				for j := range rs[i].Lifecycle.Postconditions {
					rewrite(&rs[i].Lifecycle.Postconditions[j].Range)
					rewrite(&rs[i].Lifecycle.Postconditions[j].Condition.Range)
					rewrite(&rs[i].Lifecycle.Postconditions[j].ErrorMessage.Range)
				}
			}
		}
	}
	rewriteResources(mod.ManagedResources)
	rewriteResources(mod.DataResources)
	for i := range mod.ModuleCalls {
		rewrite(&mod.ModuleCalls[i].Range)
		if mod.ModuleCalls[i].Count != nil {
			rewrite(&mod.ModuleCalls[i].Count.Range)
		}
		if mod.ModuleCalls[i].ForEach != nil {
			rewrite(&mod.ModuleCalls[i].ForEach.Range)
		}
	}
	for i := range mod.Providers {
		rewrite(&mod.Providers[i].Range)
	}
	for i := range mod.Diagnostics {
		rewrite(mod.Diagnostics[i].Subject)
		rewrite(mod.Diagnostics[i].Context)
	}
}

// TestLoad_VariableTypes confirms Decision 2 from the design doc:
// variable types are serialized via typeexpr.TypeString so that
// nested types (objects inside lists, optional() markers, etc.)
// survive a round-trip as the canonical HCL form.
func TestLoad_VariableTypes(t *testing.T) {
	t.Parallel()

	mod, err := Load(fixturePath(t, "variables-and-outputs"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if mod.Diagnostics.HasErrors() {
		t.Fatalf("unexpected error diagnostics: %#v", mod.Diagnostics)
	}

	byName := indexVariables(mod.Variables)

	// The loader captures the verbatim source of the type expression,
	// preserving `optional(T, default)` markers and any other detail the
	// user wrote (see Decision 1 in docs/step-2-config-loader.md). This
	// means whitespace and newlines from the fixture survive too.
	want := map[string]string{
		"region":      "string",
		"tags":        "map(string)",
		"db_password": "string",
		"instance_sizes": "list(object({\n" +
			"    name = string\n" +
			"    cpu  = number\n" +
			"    mem  = number\n" +
			"  }))",
		"feature_flags": "object({\n" +
			"    enable_logging = optional(bool, true)\n" +
			"    enable_metrics = optional(bool, false)\n" +
			"  })",
	}
	for name, wantType := range want {
		got, ok := byName[name]
		if !ok {
			t.Errorf("variable %q missing", name)
			continue
		}
		if got.Type != wantType {
			t.Errorf("variable %q type:\n  got  %q\n  want %q", name, got.Type, wantType)
		}
	}

	// feature_flags must contain the `optional(...)` markers verbatim —
	// that's the whole point of switching from typeexpr.TypeString to
	// source capture.
	if !strings.Contains(byName["feature_flags"].Type, "optional(bool, true)") {
		t.Errorf("feature_flags type missing optional(bool, true): %q", byName["feature_flags"].Type)
	}
}

// TestLoad_MalformedValidation locks in the no-panic guarantee for a
// validation {} block (and a resource precondition {}) that is missing
// its required attributes. The loader must surface error diagnostics
// for the missing attributes instead of nil-dereferencing.
func TestLoad_MalformedValidation(t *testing.T) {
	t.Parallel()

	mod, err := Load(fixturePath(t, "invalid/malformed-validation"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !mod.Diagnostics.HasErrors() {
		t.Fatalf("expected error diagnostics for malformed validation/precondition, got none")
	}

	// The malformed validation block must NOT be appended to the
	// variable's Validations slice (we skip it after recording the
	// diagnostic).
	v := indexVariables(mod.Variables)["name"]
	if len(v.Validations) != 0 {
		t.Errorf("variable.Validations should be empty for malformed block, got %d entries", len(v.Validations))
	}

	// Same for the resource's lifecycle.Preconditions.
	r := findResource(t, mod.ManagedResources, "null_resource", "checked")
	if r.Lifecycle != nil && len(r.Lifecycle.Preconditions) != 0 {
		t.Errorf("Lifecycle.Preconditions should be empty for malformed block, got %d entries", len(r.Lifecycle.Preconditions))
	}
}

// TestLoad_NonLiteralAttrs confirms that attributes the loader expects
// to be literal (output.description, output.sensitive, module.source,
// variable.description, etc.) surface a diagnostic instead of silently
// disappearing when a user writes an interpolation or reference.
func TestLoad_NonLiteralAttrs(t *testing.T) {
	t.Parallel()

	mod, err := Load(fixturePath(t, "invalid/non-literal-attrs"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !mod.Diagnostics.HasErrors() {
		t.Fatalf("expected error diagnostics for non-literal attributes, got none")
	}

	// Find a diagnostic for each of the three offending attributes via
	// the source line — HCL points the Subject at the offending
	// expression. The fixture is laid out so these lines are stable.
	wantLines := map[int]string{
		19: "output.bad_description.description",
		25: "output.bad_sensitive.sensitive",
		30: "module.bad_source.source",
	}
	gotLines := map[int]bool{}
	for _, d := range mod.Diagnostics {
		if d.Severity != model.SeverityError || d.Subject == nil {
			continue
		}
		gotLines[d.Subject.Start.Line] = true
	}
	for line, label := range wantLines {
		if !gotLines[line] {
			t.Errorf("missing error diagnostic for %s (expected at line %d); got lines %v",
				label, line, gotLines)
		}
	}

	// And the literal values must NOT have been retained.
	for _, o := range mod.Outputs {
		if o.Name == "bad_description" && o.Description != "" {
			t.Errorf("bad_description.Description = %q, want empty", o.Description)
		}
		if o.Name == "bad_sensitive" && o.Sensitive {
			t.Errorf("bad_sensitive.Sensitive = true, want false (literal extraction must have failed)")
		}
	}
	for _, m := range mod.ModuleCalls {
		if m.Name == "bad_source" && m.Source != "" {
			t.Errorf("bad_source.Source = %q, want empty", m.Source)
		}
	}
}

// TestLoad_ConfigurationAliases confirms Decision 3 from the design
// doc: configuration_aliases is a list of traversals (provider
// references), not values, and is captured via the traversal helpers.
func TestLoad_ConfigurationAliases(t *testing.T) {
	t.Parallel()

	mod, err := Load(fixturePath(t, "providers"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if mod.Diagnostics.HasErrors() {
		t.Fatalf("unexpected error diagnostics: %#v", mod.Diagnostics)
	}

	aws, ok := mod.RequiredProviders["aws"]
	if !ok {
		t.Fatalf("RequiredProviders[aws] missing")
	}
	want := []string{"aws.east", "aws.west"}
	if !reflect.DeepEqual(aws.ConfigurationAliases, want) {
		t.Errorf("aws.ConfigurationAliases:\n  got  %v\n  want %v", aws.ConfigurationAliases, want)
	}

	// Providers without configuration_aliases must come back as nil/empty,
	// not as a non-empty slice with garbage values.
	for _, name := range []string{"random", "http"} {
		got, ok := mod.RequiredProviders[name]
		if !ok {
			t.Errorf("RequiredProviders[%s] missing", name)
			continue
		}
		if len(got.ConfigurationAliases) != 0 {
			t.Errorf("RequiredProviders[%s].ConfigurationAliases = %v, want none", name, got.ConfigurationAliases)
		}
	}
}

// TestLoad_LifecycleAndMeta covers the lifecycle {} block (every
// sub-feature, including `ignore_changes = all`), the `provider =`
// meta-arg on a resource (Pitfall #8 — single traversal, not a list),
// the resource and module `depends_on` traversal lists, and the
// `providers = {…}` map on a module call.
func TestLoad_LifecycleAndMeta(t *testing.T) {
	t.Parallel()

	mod, err := Load(fixturePath(t, "resources-full"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if mod.Diagnostics.HasErrors() {
		t.Fatalf("unexpected error diagnostics: %#v", mod.Diagnostics)
	}

	primary := findResource(t, mod.ManagedResources, "aws_instance", "primary")
	if primary.Provider != "aws.east" {
		t.Errorf("primary.Provider = %q, want %q", primary.Provider, "aws.east")
	}
	if want := []string{"aws_security_group.web"}; !reflect.DeepEqual(primary.DependsOn, want) {
		t.Errorf("primary.DependsOn = %v, want %v", primary.DependsOn, want)
	}

	lc := primary.Lifecycle
	if lc == nil {
		t.Fatalf("primary.Lifecycle is nil")
	}
	if lc.CreateBeforeDestroy == nil || *lc.CreateBeforeDestroy != true {
		t.Errorf("CreateBeforeDestroy = %v, want *true", lc.CreateBeforeDestroy)
	}
	if lc.PreventDestroy == nil || *lc.PreventDestroy != false {
		t.Errorf("PreventDestroy = %v, want *false (pointer set, value false)", lc.PreventDestroy)
	}
	if want := []string{"tags", "ami"}; !reflect.DeepEqual(lc.IgnoreChanges, want) {
		t.Errorf("IgnoreChanges = %v, want %v", lc.IgnoreChanges, want)
	}
	if want := []string{"aws_security_group.web"}; !reflect.DeepEqual(lc.ReplaceTriggeredBy, want) {
		t.Errorf("ReplaceTriggeredBy = %v, want %v", lc.ReplaceTriggeredBy, want)
	}
	if len(lc.Preconditions) != 1 {
		t.Errorf("Preconditions: want 1, got %d", len(lc.Preconditions))
	} else if !strings.Contains(lc.Preconditions[0].ErrorMessage.Source, "ami must be set") {
		t.Errorf("Preconditions[0].ErrorMessage.Source = %q", lc.Preconditions[0].ErrorMessage.Source)
	}
	if len(lc.Postconditions) != 1 {
		t.Errorf("Postconditions: want 1, got %d", len(lc.Postconditions))
	}

	// `ignore_changes = all` (bare keyword) on a sibling resource.
	web := findResource(t, mod.ManagedResources, "aws_security_group", "web")
	if web.Lifecycle == nil {
		t.Fatalf("web.Lifecycle is nil")
	}
	if want := []string{"all"}; !reflect.DeepEqual(web.Lifecycle.IgnoreChanges, want) {
		t.Errorf("web.Lifecycle.IgnoreChanges = %v, want %v (bare-keyword fallback)", web.Lifecycle.IgnoreChanges, want)
	}

	// Module-level meta-args.
	if len(mod.ModuleCalls) != 1 {
		t.Fatalf("ModuleCalls: want 1, got %d", len(mod.ModuleCalls))
	}
	mc := mod.ModuleCalls[0]
	if mc.Name != "infra" {
		t.Errorf("ModuleCalls[0].Name = %q, want %q", mc.Name, "infra")
	}
	if want := []string{"aws_security_group.web"}; !reflect.DeepEqual(mc.DependsOn, want) {
		t.Errorf("ModuleCalls[0].DependsOn = %v, want %v", mc.DependsOn, want)
	}
	wantProviders := map[string]string{"aws": "aws.east"}
	if !reflect.DeepEqual(mc.Providers, wantProviders) {
		t.Errorf("ModuleCalls[0].Providers = %v, want %v", mc.Providers, wantProviders)
	}
}

// TestLoad_VariableDetails reads back the boolean flags, default
// expression source, and well-formed validation block on the existing
// variables-and-outputs fixture — fields the other tests touched only
// indirectly via snapshots.
func TestLoad_VariableDetails(t *testing.T) {
	t.Parallel()

	mod, err := Load(fixturePath(t, "variables-and-outputs"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if mod.Diagnostics.HasErrors() {
		t.Fatalf("unexpected error diagnostics: %#v", mod.Diagnostics)
	}
	byName := indexVariables(mod.Variables)

	region := byName["region"]
	if region.Nullable == nil || *region.Nullable != false {
		t.Errorf("region.Nullable = %v, want *false (pointer set, value false)", region.Nullable)
	}
	if region.Default == nil || region.Default.Source != `"us-east-1"` {
		t.Errorf("region.Default = %#v, want source %q", region.Default, `"us-east-1"`)
	}
	if len(region.Validations) != 1 {
		t.Fatalf("region.Validations: want 1, got %d", len(region.Validations))
	}
	v := region.Validations[0]
	if !strings.Contains(v.Condition.Source, "can(regex(") {
		t.Errorf("validation.Condition.Source = %q, want it to contain can(regex(", v.Condition.Source)
	}
	if !strings.Contains(v.ErrorMessage.Source, "Region must look like") {
		t.Errorf("validation.ErrorMessage.Source = %q", v.ErrorMessage.Source)
	}

	dbPassword := byName["db_password"]
	if !dbPassword.Sensitive {
		t.Errorf("db_password.Sensitive = false, want true")
	}
	if dbPassword.Nullable != nil {
		t.Errorf("db_password.Nullable = %v, want nil (attribute not set)", dbPassword.Nullable)
	}

	// Output sensitive flag round-trip.
	for _, o := range mod.Outputs {
		if o.Name == "db_password" && !o.Sensitive {
			t.Errorf("output db_password.Sensitive = false, want true")
		}
	}
}

// TestLoad_TofuExtension confirms that `.tofu` and `.tofu.json` files
// are picked up by the walker — a step-2 commitment per the design doc.
func TestLoad_TofuExtension(t *testing.T) {
	t.Parallel()

	mod, err := Load(fixturePath(t, "tofu-extension"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if mod.Diagnostics.HasErrors() {
		t.Fatalf("unexpected error diagnostics: %#v", mod.Diagnostics)
	}
	byName := indexVariables(mod.Variables)
	if _, ok := byName["greeting"]; !ok {
		t.Errorf("variable greeting (.tofu) missing; got %v", keysOf(byName))
	}
	if _, ok := byName["from_json"]; !ok {
		t.Errorf("variable from_json (.tofu.json) missing; got %v", keysOf(byName))
	}
}

// TestLoad_OverridesNotMerged confirms that step 2 collects override
// files but does NOT apply them — that's a step-3 promise. The
// original main.tf values must survive untouched.
func TestLoad_OverridesNotMerged(t *testing.T) {
	t.Parallel()

	mod, err := Load(fixturePath(t, "overrides"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if mod.Diagnostics.HasErrors() {
		t.Fatalf("unexpected error diagnostics: %#v", mod.Diagnostics)
	}
	region := indexVariables(mod.Variables)["region"]
	if region.Default == nil || region.Default.Source != `"us-east-1"` {
		t.Errorf("region.Default = %#v; override.tf must NOT be merged in step 2 (would set %q)",
			region.Default, `"eu-central-1"`)
	}
}

// TestLoad_NoPanic_StepThreeFixtures asserts the loader survives every
// fixture whose deep semantics are deferred to step 3 without panicking
// or returning a Go-level error. Diagnostics may be present — we don't
// care about their content here.
func TestLoad_NoPanic_StepThreeFixtures(t *testing.T) {
	t.Parallel()
	dirs := []string{
		"modern-blocks",
		"ephemeral",
		"opentofu-encryption",
		"opentofu-provider-foreach",
		"invalid/missing-required",
	}
	for _, dir := range dirs {
		t.Run(dir, func(t *testing.T) {
			t.Parallel()
			if _, err := Load(fixturePath(t, dir)); err != nil {
				t.Fatalf("Load(%s) returned Go error: %v", dir, err)
			}
		})
	}
}

// TestLoad_LegacyProviderForm covers `aws = "~> 4.0"` — the pre-0.13
// shorthand still accepted by the modern parser.
func TestLoad_LegacyProviderForm(t *testing.T) {
	t.Parallel()

	mod, err := Load(fixturePath(t, "providers-legacy-form"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if mod.Diagnostics.HasErrors() {
		t.Fatalf("unexpected error diagnostics: %#v", mod.Diagnostics)
	}
	aws, ok := mod.RequiredProviders["aws"]
	if !ok {
		t.Fatalf("RequiredProviders[aws] missing")
	}
	if aws.Source != "" {
		t.Errorf("legacy form sets no source; got %q", aws.Source)
	}
	if want := []string{"~> 4.0"}; !reflect.DeepEqual(aws.VersionConstraints, want) {
		t.Errorf("VersionConstraints = %v, want %v", aws.VersionConstraints, want)
	}
}

// TestLoad_MultipleTerraformBlocks confirms that terraform {} blocks
// from multiple files are aggregated: required_version values are
// concatenated, and required_providers entries merge by name.
func TestLoad_MultipleTerraformBlocks(t *testing.T) {
	t.Parallel()

	mod, err := Load(fixturePath(t, "multi-terraform-blocks"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if mod.Diagnostics.HasErrors() {
		t.Fatalf("unexpected error diagnostics: %#v", mod.Diagnostics)
	}
	if got, want := len(mod.RequiredCore), 2; got != want {
		t.Errorf("RequiredCore count = %d (%v), want %d (both files contribute)",
			got, mod.RequiredCore, want)
	}
	for _, name := range []string{"aws", "random"} {
		if _, ok := mod.RequiredProviders[name]; !ok {
			t.Errorf("RequiredProviders[%s] missing; got %v", name, keysOfProviders(mod.RequiredProviders))
		}
	}
}

// keysOf returns the keys of a string-keyed map of variables, in
// sorted order — useful for stable error messages.
func keysOf(m map[string]model.Variable) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// keysOfProviders is the same shape as keysOf but for required-providers.
func keysOfProviders(m map[string]model.ProviderRequirement) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// TestLoad_ExpressionCapture confirms Decision 1 from the design doc:
// expressions are captured as raw source bytes plus a source Range,
// never evaluated. The captured source must include the symbolic
// reference verbatim (e.g. var.replica_count) and the Range must
// point at the file/line where the expression appears.
func TestLoad_ExpressionCapture(t *testing.T) {
	t.Parallel()

	mod, err := Load(fixturePath(t, "resources-count-foreach"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if mod.Diagnostics.HasErrors() {
		t.Fatalf("unexpected error diagnostics: %#v", mod.Diagnostics)
	}

	r := findResource(t, mod.ManagedResources, "null_resource", "by_count")
	if r.Count == nil {
		t.Fatalf("null_resource.by_count: Count expression not captured")
	}
	if !strings.Contains(r.Count.Source, "var.replica_count") {
		t.Errorf("Count.Source does not contain var.replica_count: %q", r.Count.Source)
	}
	if r.Count.Range.Filename == "" || r.Count.Range.Start.Line == 0 {
		t.Errorf("Count.Range is missing source location: %#v", r.Count.Range)
	}

	r = findResource(t, mod.ManagedResources, "null_resource", "by_for_each")
	if r.ForEach == nil {
		t.Fatalf("null_resource.by_for_each: ForEach expression not captured")
	}
	if !strings.Contains(r.ForEach.Source, "var.names") {
		t.Errorf("ForEach.Source does not contain var.names: %q", r.ForEach.Source)
	}
	if r.ForEach.Range.Start.Line == 0 {
		t.Errorf("ForEach.Range missing line info: %#v", r.ForEach.Range)
	}
}

// indexVariables returns variables keyed by name for quick lookup.
func indexVariables(vars []model.Variable) map[string]model.Variable {
	out := make(map[string]model.Variable, len(vars))
	for _, v := range vars {
		out[v.Name] = v
	}
	return out
}

// findResource returns the resource matching (type, name) or fails t.
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
