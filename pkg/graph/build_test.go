// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package graph_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/remoterabbit/open-inspector/pkg/graph"
	"github.com/remoterabbit/open-inspector/pkg/inspector"
	"github.com/remoterabbit/open-inspector/pkg/model"
)

// TestBuild_AttachError records a per-call failure on ChildModule.Error
// when a local source cannot be resolved, without touching the parent's
// own Diagnostics.
func TestBuild_AttachError(t *testing.T) {
	root := &model.Module{
		Path: t.TempDir(),
		ModuleCalls: []model.ModuleCall{
			{Name: "missing", Source: "./does-not-exist"},
		},
	}
	graph.Build(root, graph.Options{MaxDepth: 16, CacheDir: t.TempDir()})

	child, ok := root.Children["missing"]
	if !ok {
		t.Fatalf("expected a child entry for the failed call, got %#v", root.Children)
	}
	if child.Error == nil {
		t.Fatalf("expected ChildModule.Error to be set, got %#v", child)
	}
	if len(root.Diagnostics) != 0 {
		t.Errorf("per-call failure should not add to parent Diagnostics, got %#v", root.Diagnostics)
	}
}

// TestBuild_DepthLimit emits a warning diagnostic and stops recursing
// once the configured max depth is reached.
func TestBuild_DepthLimit(t *testing.T) {
	root := &model.Module{
		Path: t.TempDir(),
		ModuleCalls: []model.ModuleCall{
			{Name: "child", Source: "./child"},
		},
	}
	graph.Build(root, graph.Options{MaxDepth: 0, CacheDir: t.TempDir()})

	if len(root.Diagnostics) == 0 {
		t.Fatalf("expected a max-depth diagnostic, got none")
	}
	found := false
	for _, diag := range root.Diagnostics {
		if diag.Severity == model.SeverityWarning && strings.Contains(diag.Summary, "max-depth") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a max-depth warning, got %#v", root.Diagnostics)
	}
	if len(root.Children) != 0 {
		t.Errorf("depth limit should stop resolution before attaching children, got %#v", root.Children)
	}
}

// TestBuild_CycleDetection emits a warning when a module call resolves
// back to a directory already on the recursion stack.
func TestBuild_CycleDetection(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "main.tf"), `module "self" {
  source = "."
}`)

	module, err := inspector.Inspect(dir, inspector.WithModuleGraph())
	if err != nil {
		t.Fatalf("inspect: %v", err)
	}
	found := false
	for _, diag := range module.Diagnostics {
		if diag.Severity == model.SeverityWarning && strings.Contains(diag.Summary, "cycle") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a cycle warning, got %#v", module.Diagnostics)
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
