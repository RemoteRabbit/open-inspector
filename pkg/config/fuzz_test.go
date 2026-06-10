// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// This file fuzz-tests the config loader. Fuzzing throws large volumes of
// randomly mutated input at a function to discover inputs that make it
// crash. The contract under test is simple: config.Load must handle ANY
// file content gracefully (return diagnostics or a Go error) and must
// never panic. A panic on attacker or typo-controlled input would be a
// denial-of-service bug, so these targets guard against it.
//
// How Go fuzzing works here:
//   - A fuzz target is a function named Fuzz* taking *testing.F.
//   - Seeds added with f.Add give the fuzzer realistic starting inputs.
//     The corpus also includes any crash inputs previously saved under
//     testdata/fuzz/<TargetName>/ (committed as regression cases).
//   - f.Fuzz registers the body run for each input. Its only failure
//     condition is a panic (or hang/OOM); a returned error is fine.
//
// Two run modes:
//   - Plain `go test ./pkg/config` runs the body ONCE per seed (no
//     mutation). This makes the committed crash corpus act as permanent
//     regression tests, cheap enough for every CI run.
//   - `go test -fuzz=FuzzConfigLoader_HCL` runs the coverage-guided
//     mutation loop until it finds a crash or you stop it. See the
//     `make fuzz-config-*` targets and the weekly fuzz workflow.
//
// When the fuzzer finds a crash it writes the minimized input to
// testdata/fuzz/<TargetName>/<hash> and prints it; commit that file so
// the bug can never silently return.
package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/remoterabbit/open-inspector/pkg/config"
)

// FuzzConfigLoader_HCL fuzzes the loader with native HCL syntax. Each
// generated input is written to a main.tf in a throwaway directory and
// handed to config.Load; the test fails only if Load panics.
func FuzzConfigLoader_HCL(f *testing.F) {
	// Seed the corpus from real fixtures so the mutator starts from valid
	// HCL and explores nearby (often more interesting) malformed variants.
	seedFromGlob(f, "../../testdata/fixtures/*/*.tf")
	seedFromGlob(f, "../../testdata/fixtures/*/*.tofu")

	// The fuzz body runs once per seed under `go test`, and repeatedly with
	// mutated data under `go test -fuzz`. data is the (possibly mutated)
	// file content.
	f.Fuzz(func(t *testing.T, data []byte) {
		// t.TempDir is auto-removed when the sub-test ends, so each input
		// gets a clean directory and nothing leaks between runs.
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "main.tf"), data, 0o644); err != nil {
			// A filesystem failure is unrelated to the loader; skip rather
			// than report a spurious fuzz failure.
			t.Skip()
		}
		// Diagnostics or a Go error are both acceptable outcomes; we only
		// care that Load returns instead of panicking, so the results are
		// intentionally discarded.
		_, _ = config.Load(dir)
	})
}

// FuzzConfigLoader_JSON fuzzes the loader with the JSON configuration
// variant (.tf.json), which goes through a different decode path than
// native HCL and so needs its own target and seeds.
func FuzzConfigLoader_JSON(f *testing.F) {
	seedFromGlob(f, "../../testdata/fixtures/*/*.tf.json")
	seedFromGlob(f, "../../testdata/fixtures/*/*.tofu.json")

	f.Fuzz(func(t *testing.T, data []byte) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "main.tf.json"), data, 0o644); err != nil {
			t.Skip()
		}
		_, _ = config.Load(dir)
	})
}

// seedFromGlob adds every file matching pattern to the fuzz seed corpus.
// More seeds give the coverage-guided fuzzer more code paths to start
// from, so it reaches deep loader behavior faster. Unreadable files and
// bad patterns are skipped silently: seeding is best-effort and must
// never fail the test.
func seedFromGlob(f *testing.F, pattern string) {
	f.Helper()
	paths, err := filepath.Glob(pattern)
	if err != nil {
		return
	}
	for _, path := range paths {
		bytes, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		// f.Add registers one seed input. Its argument type must match the
		// fuzz body's parameter type ([]byte here).
		f.Add(bytes)
	}
}
