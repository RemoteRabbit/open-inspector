// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package schema

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// autoBinaries lists the CLIs Auto tries, in preference order.
var autoBinaries = []string{"tofu", "terraform"}

// Auto runs `<binary> -chdir=<dir> providers schema -json` using `tofu`
// (preferred) or `terraform`, whichever is found on PATH first, and
// returns the parsed schema along with the binary name that produced it.
//
// It returns a helpful error when neither binary is available or when the
// module has not been initialized (the schema command requires a prior
// `init`).
func Auto(dir string) (*Schema, string, error) {
	var binary string
	for _, candidate := range autoBinaries {
		if _, err := exec.LookPath(candidate); err == nil {
			binary = candidate
			break
		}
	}
	if binary == "" {
		return nil, "", fmt.Errorf("no `tofu` or `terraform` binary on PATH; install one or pass an explicit schema path")
	}

	// -chdir must precede the subcommand: `tofu -chdir=<dir> providers ...`.
	command := exec.Command(binary, "-chdir="+dir, "providers", "schema", "-json")
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		message := stderr.String()
		if needsInit(message) {
			return nil, binary, fmt.Errorf("module not initialized; run `%s -chdir=%s init` first", binary, dir)
		}
		return nil, binary, fmt.Errorf("%s providers schema: %w: %s", binary, err, strings.TrimSpace(message))
	}

	loaded, err := Load(&stdout)
	if err != nil {
		return nil, binary, fmt.Errorf("parse %s schema output: %w", binary, err)
	}
	return loaded, binary, nil
}

// needsInitMarkers are substrings tofu/terraform emit (across versions)
// when `providers schema -json` is run before `init`.
var needsInitMarkers = []string{
	"not been initialized",
	"Initialization required",
	"Inconsistent dependency lock file",
	"please run \"terraform init\"",
	"run: terraform init",
	"run: tofu init",
}

// needsInit reports whether a command's stderr indicates the module must
// be initialized first.
func needsInit(message string) bool {
	for _, marker := range needsInitMarkers {
		if strings.Contains(message, marker) {
			return true
		}
	}
	return false
}
