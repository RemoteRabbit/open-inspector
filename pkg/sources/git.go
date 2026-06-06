// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package sources

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/remoterabbit/open-inspector/pkg/model"
)

func resolveGit(parsed ParsedSource, cacheDir string) (model.ResolvedSource, error) {
	if _, err := exec.LookPath("git"); err != nil {
		return model.ResolvedSource{}, fmt.Errorf("git binary not on PATH: %w", err)
	}

	hash := cacheKeyGit(parsed)
	lock := lockForHash(hash)
	lock.Lock()
	defer lock.Unlock()

	dest := filepath.Join(cacheDir, hash, "extracted")
	if exists(dest) {
		return gitResult(parsed, dest), nil
	}

	tmp := filepath.Join(cacheDir, hash, "tmp")
	_ = os.RemoveAll(tmp)
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return model.ResolvedSource{}, err
	}

	args := []string{"clone"}
	if parsed.Ref != "" && !looksLikeSHA(parsed.Ref) {
		args = append(args, "--depth=1", "--branch", parsed.Ref)
	}
	args = append(args, parsed.URL, tmp)

	if output, err := exec.Command("git", args...).CombinedOutput(); err != nil {
		return model.ResolvedSource{}, fmt.Errorf("git clone failed: %s: %w", output, err)
	}

	if parsed.Ref != "" && looksLikeSHA(parsed.Ref) {
		if output, err := exec.Command("git", "-C", tmp, "checkout", parsed.Ref).CombinedOutput(); err != nil {
			return model.ResolvedSource{}, fmt.Errorf("git checkout %s failed: %s: %w", parsed.Ref, output, err)
		}
	}
	if err := os.Rename(tmp, dest); err != nil {
		return model.ResolvedSource{}, err
	}
	return gitResult(parsed, dest), nil
}

// gitResult builds the ResolvedSource, applying //subdir if present.
func gitResult(parsed ParsedSource, dest string) model.ResolvedSource {
	cachePath := dest
	if parsed.Subdir != "" {
		cachePath = filepath.Join(dest, parsed.Subdir)
	}
	return model.ResolvedSource{
		Kind:      "git",
		Address:   parsed.URL,
		CachePath: cachePath,
		Ref:       parsed.Ref,
	}
}

func looksLikeSHA(s string) bool {
	if len(s) < 7 || len(s) > 40 {
		return false
	}
	for _, r := range s {
		isHexDigit := (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')
		if !isHexDigit {
			return false
		}
	}
	return true
}
