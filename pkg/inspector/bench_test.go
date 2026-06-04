// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package inspector

import (
	"path/filepath"
	"testing"
)

// fixtureDir resolves a directory under testdata/fixtures/ relative to
// this package. It fails the benchmark if the path cannot be resolved.
func fixtureDir(b *testing.B, name string) string {
	b.Helper()
	p, err := filepath.Abs(filepath.Join("..", "..", "testdata", "fixtures", name))
	if err != nil {
		b.Fatalf("resolve fixture path: %v", err)
	}
	return p
}

// BenchmarkInspect measures the cost of a full Inspect (file walk + HCL
// partial decode + model assembly) across representative fixtures. These
// establish a baseline; treat regressions as a prompt to investigate,
// not a hard CI gate.
func BenchmarkInspect(b *testing.B) {
	cases := []string{
		"simple",
		"resources-full",
		"modern-blocks",
		"multi-terraform-blocks",
		"multi-module",
	}

	for _, name := range cases {
		dir := fixtureDir(b, name)
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				mod, err := Inspect(dir)
				if err != nil {
					b.Fatalf("Inspect(%s) returned error: %v", name, err)
				}
				if mod == nil {
					b.Fatalf("Inspect(%s) returned nil module", name)
				}
			}
		})
	}
}
