// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package sources

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolve_Local(t *testing.T) {
	parent := t.TempDir()
	child := filepath.Join(parent, "sub")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	resolved, err := Resolve("./sub", "", parent, t.TempDir())
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if resolved.Kind != "local" {
		t.Errorf("Kind: got %q, want local", resolved.Kind)
	}
	if resolved.CachePath != child {
		t.Errorf("CachePath: got %q, want %q", resolved.CachePath, child)
	}
}

func TestResolve_LocalMissing(t *testing.T) {
	if _, err := Resolve("./does-not-exist", "", t.TempDir(), t.TempDir()); err == nil {
		t.Fatalf("expected error resolving a missing local path, got nil")
	}
}

func TestResolve_InvalidSource(t *testing.T) {
	if _, err := Resolve("", "", t.TempDir(), t.TempDir()); err == nil {
		t.Fatalf("expected error for empty source, got nil")
	}
}

func TestDefaultCacheDir(t *testing.T) {
	dir := DefaultCacheDir()
	if dir == "" {
		t.Fatal("DefaultCacheDir returned empty string")
	}
	want := filepath.Join("open-inspector", "modules")
	if !strings.HasSuffix(dir, want) {
		t.Errorf("DefaultCacheDir = %q, want suffix %q", dir, want)
	}
}
