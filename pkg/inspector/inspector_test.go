// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package inspector

import (
	"path/filepath"
	"testing"
)

func TestInspectReturnsAbsolutePath(t *testing.T) {
	t.Parallel()

	mod, err := Inspect(".")
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
