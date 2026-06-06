// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package sources

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGitResolver_LocalBareRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	bare := filepath.Join(t.TempDir(), "bare.git")
	seed := filepath.Join(t.TempDir(), "seed")

	// Init bare repo with an explicit default branch so the test does
	// not depend on the host's init.defaultBranch setting.
	run(t, "", "git", "init", "--bare", "-b", "main", bare)
	run(t, "", "git", "clone", bare, seed)
	writeFile(t, filepath.Join(seed, "main.tf"), `resource "null_resource" "x" {}`)
	run(t, seed, "git", "add", ".")
	run(t, seed, "git", "-c", "user.email=t@t", "-c", "user.name=t", "-c", "commit.gpgsign=false", "commit", "-m", "init")
	run(t, seed, "git", "push", "origin", "main")

	resolved, err := resolveGit(ParsedSource{
		Kind: SourceGit, URL: "file://" + bare, Ref: "main",
	}, t.TempDir())
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !exists(filepath.Join(resolved.CachePath, "main.tf")) {
		t.Errorf("main.tf missing from %s", resolved.CachePath)
	}
}

func TestGitResolver_SHARef(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	bare := filepath.Join(t.TempDir(), "bare.git")
	seed := filepath.Join(t.TempDir(), "seed")

	run(t, "", "git", "init", "--bare", "-b", "main", bare)
	run(t, "", "git", "clone", bare, seed)
	writeFile(t, filepath.Join(seed, "main.tf"), `resource "null_resource" "x" {}`)
	run(t, seed, "git", "add", ".")
	run(t, seed, "git", "-c", "user.email=t@t", "-c", "user.name=t", "-c", "commit.gpgsign=false", "commit", "-m", "init")
	run(t, seed, "git", "push", "origin", "main")

	sha := output(t, seed, "git", "rev-parse", "HEAD")

	resolved, err := resolveGit(ParsedSource{
		Kind: SourceGit, URL: "file://" + bare, Ref: sha,
	}, t.TempDir())
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if resolved.Ref != sha {
		t.Errorf("resolved ref: got %q, want %q", resolved.Ref, sha)
	}
	if !exists(filepath.Join(resolved.CachePath, "main.tf")) {
		t.Errorf("main.tf missing from %s", resolved.CachePath)
	}
}

func TestLooksLikeSHA(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"a1b2c3d", true}, // 7 hex chars (min)
		{"0123456789abcdef0123456789abcdef01234567", true}, // 40 hex chars (max)
		{"main", false},    // branch name
		{"v1.2.3", false},  // tag
		{"abc", false},     // too short
		{"g1b2c3d", false}, // non-hex char
		{"0123456789abcdef0123456789abcdef012345678", false}, // 41 chars (too long)
	}
	for _, tc := range cases {
		if got := looksLikeSHA(tc.in); got != tc.want {
			t.Errorf("looksLikeSHA(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

// run executes a command in dir (cwd when dir is "") and fails the test
// on a non-zero exit.
func run(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
}

// output runs a command in dir and returns its trimmed stdout, failing
// the test on a non-zero exit.
func output(t *testing.T, dir string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("%s %v: %v", name, args, err)
	}
	return strings.TrimSpace(string(out))
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
