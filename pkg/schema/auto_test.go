// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package schema

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuto_RequiresBinary(t *testing.T) {
	for _, binary := range autoBinaries {
		if _, err := exec.LookPath(binary); err == nil {
			t.Skipf("%s present; this test exercises the absent-binary path", binary)
		}
	}

	_, _, err := Auto(t.TempDir())
	if err == nil {
		t.Fatalf("expected an error when no tofu/terraform binary is present")
	}
	if !strings.Contains(err.Error(), "PATH") {
		t.Errorf("error should mention PATH: %v", err)
	}
}

func TestAuto_RequiresInit(t *testing.T) {
	haveBinary := false
	for _, binary := range autoBinaries {
		if _, err := exec.LookPath(binary); err == nil {
			haveBinary = true
			break
		}
	}
	if !haveBinary {
		t.Skip("no tofu/terraform binary present; cannot exercise the needs-init path")
	}

	// A config that requires a provider but has not been initialized makes
	// `providers schema -json` fail with a needs-init style error.
	dir := t.TempDir()
	config := `terraform {
  required_providers {
    null = {
      source = "hashicorp/null"
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(dir, "main.tf"), []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, _, err := Auto(dir)
	if err == nil {
		t.Fatalf("expected an error for an uninitialized directory")
	}
	if !strings.Contains(err.Error(), "init") {
		t.Errorf("error should guide the user to run init: %v", err)
	}
}
