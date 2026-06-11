// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/remoterabbit/open-inspector/pkg/model"
)

func TestExtractReferences(t *testing.T) {
	cases := []struct {
		src  string
		want []model.Reference // kind+address only; ignore Range
	}{
		{`var.region`, []model.Reference{{Kind: model.ReferenceVar, Address: "var.region"}}},
		{`local.name`, []model.Reference{{Kind: model.ReferenceLocal, Address: "local.name"}}},
		{`module.net.vpc_id`, []model.Reference{{Kind: model.ReferenceModule, Address: "module.net"}}},
		{`data.aws_ami.a.id`, []model.Reference{{Kind: model.ReferenceData, Address: "data.aws_ami.a"}}},
		{`aws_s3_bucket.b.arn`, []model.Reference{{Kind: model.ReferenceResource, Address: "aws_s3_bucket.b"}}},
		{`each.key`, []model.Reference{{Kind: model.ReferenceOther, Address: "each.key"}}},
		{`"${var.x}-${var.x}"`, []model.Reference{{Kind: model.ReferenceVar, Address: "var.x"}}}, // dedup
		{`"literal"`, nil},
	}
	for _, tc := range cases {
		expr, diags := hclsyntax.ParseExpression([]byte(tc.src), "test.tf", hcl.InitialPos)
		if diags.HasErrors() {
			t.Fatalf("parse %q: %v", tc.src, diags)
		}
		got := extractReferences(expr)
		if len(got) != len(tc.want) {
			t.Errorf("extractReferences(%q): got %d references, want %d\n got: %+v\nwant: %+v",
				tc.src, len(got), len(tc.want), got, tc.want)
			continue
		}
		for i := range tc.want {
			if got[i].Kind != tc.want[i].Kind || got[i].Address != tc.want[i].Address {
				t.Errorf("extractReferences(%q)[%d]: got {Kind:%q Address:%q}, want {Kind:%q Address:%q}",
					tc.src, i, got[i].Kind, got[i].Address, tc.want[i].Kind, tc.want[i].Address)
			}
		}
	}
}
