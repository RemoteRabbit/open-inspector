// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package inspector_test

import (
	"path/filepath"
	"testing"

	"github.com/remoterabbit/open-inspector/pkg/inspector"
)

func TestInspectReturnsAbsolutePath(t *testing.T) {
	t.Parallel()

	mod, err := inspector.Inspect(".")
	if err != nil {
		t.Fatalf("Inspect returned error: %v", err)
	}
	if mod == nil {
		t.Fatal("Inspect returned nil module")
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
