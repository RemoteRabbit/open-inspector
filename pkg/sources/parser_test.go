// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package sources

import "testing"

func TestParseSource(t *testing.T) {
	// ParsedSource is all comparable fields, so == works. Every want
	// must set Raw, because ParseSource always records the original.
	cases := []struct {
		in   string
		want ParsedSource
	}{
		{"./modules/network", ParsedSource{Kind: SourceLocal, Raw: "./modules/network"}},
		{"../shared", ParsedSource{Kind: SourceLocal, Raw: "../shared"}},
		{"terraform-aws-modules/vpc/aws", ParsedSource{
			Kind: SourceRegistry, Raw: "terraform-aws-modules/vpc/aws",
			Host:      "registry.terraform.io",
			Namespace: "terraform-aws-modules", Name: "vpc", Provider: "aws",
		}},
		{"registry.opentofu.org/terraform-aws-modules/vpc/aws", ParsedSource{
			Kind: SourceRegistry, Raw: "registry.opentofu.org/terraform-aws-modules/vpc/aws",
			Host:      "registry.opentofu.org",
			Namespace: "terraform-aws-modules", Name: "vpc", Provider: "aws",
		}},
		{"git::https://example.com/repo.git?ref=v1", ParsedSource{
			Kind: SourceGit, Raw: "git::https://example.com/repo.git?ref=v1",
			URL: "https://example.com/repo.git", Ref: "v1",
		}},
		{"git::https://example.com/repo.git//modules/vpc?ref=v1.2.3", ParsedSource{
			Kind: SourceGit, Raw: "git::https://example.com/repo.git//modules/vpc?ref=v1.2.3",
			URL: "https://example.com/repo.git", Ref: "v1.2.3", Subdir: "modules/vpc",
		}},
		{"github.com/owner/repo", ParsedSource{
			Kind: SourceGit, Raw: "github.com/owner/repo",
			URL: "https://github.com/owner/repo.git",
		}},
		{"https://example.com/m.zip", ParsedSource{
			Kind: SourceHTTP, Raw: "https://example.com/m.zip",
			URL: "https://example.com/m.zip",
		}},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := ParseSource(tc.in)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %#v, want %#v", got, tc.want)
			}
		})
	}
}
