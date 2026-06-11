// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"

	"github.com/remoterabbit/open-inspector/pkg/model"
)

func TestCtyTypeToSpec(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		typ  cty.Type
		want *model.TypeSpec
	}{
		{
			name: "string",
			typ:  cty.String,
			want: &model.TypeSpec{Kind: model.TypeString},
		},
		{
			name: "number",
			typ:  cty.Number,
			want: &model.TypeSpec{Kind: model.TypeNumber},
		},
		{
			name: "bool",
			typ:  cty.Bool,
			want: &model.TypeSpec{Kind: model.TypeBool},
		},
		{
			name: "any is dynamic",
			typ:  cty.DynamicPseudoType,
			want: &model.TypeSpec{Kind: model.TypeDynamic},
		},
		{
			name: "list",
			typ:  cty.List(cty.String),
			want: &model.TypeSpec{Kind: model.TypeList, Element: &model.TypeSpec{Kind: model.TypeString}},
		},
		{
			name: "set",
			typ:  cty.Set(cty.String),
			want: &model.TypeSpec{Kind: model.TypeSet, Element: &model.TypeSpec{Kind: model.TypeString}},
		},
		{
			name: "map",
			typ:  cty.Map(cty.String),
			want: &model.TypeSpec{Kind: model.TypeMap, Element: &model.TypeSpec{Kind: model.TypeString}},
		},
		{
			name: "tuple preserves element order",
			typ:  cty.Tuple([]cty.Type{cty.String, cty.Number}),
			want: &model.TypeSpec{Kind: model.TypeTuple, Elements: []*model.TypeSpec{
				{Kind: model.TypeString},
				{Kind: model.TypeNumber},
			}},
		},
		{
			name: "object",
			typ:  cty.Object(map[string]cty.Type{"name": cty.String}),
			want: &model.TypeSpec{Kind: model.TypeObject, Attributes: map[string]*model.ObjectAttr{
				"name": {Type: &model.TypeSpec{Kind: model.TypeString}},
			}},
		},
		{
			name: "nested collection of object",
			typ:  cty.List(cty.Object(map[string]cty.Type{"id": cty.Number})),
			want: &model.TypeSpec{Kind: model.TypeList, Element: &model.TypeSpec{
				Kind: model.TypeObject, Attributes: map[string]*model.ObjectAttr{
					"id": {Type: &model.TypeSpec{Kind: model.TypeNumber}},
				},
			}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := ctyTypeToSpec(tc.typ, nil)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("ctyTypeToSpec() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestCtyTypeToSpec_OptionalAttributes exercises the optional(...) markers
// and their defaults, which only exist via typeexpr.TypeConstraintWithDefaults.
func TestCtyTypeToSpec_OptionalAttributes(t *testing.T) {
	t.Parallel()

	constraint, defaults := parseTypeConstraint(t, `object({
		name   = string
		size   = optional(number, 10)
		nested = optional(object({ a = string }))
	})`)

	want := &model.TypeSpec{Kind: model.TypeObject, Attributes: map[string]*model.ObjectAttr{
		"name": {Type: &model.TypeSpec{Kind: model.TypeString}},
		"size": {
			Type:     &model.TypeSpec{Kind: model.TypeNumber},
			Optional: true,
			Default:  &model.Value{Kind: model.ValueNumber, Number: "10"},
		},
		"nested": {
			Type: &model.TypeSpec{Kind: model.TypeObject, Attributes: map[string]*model.ObjectAttr{
				"a": {Type: &model.TypeSpec{Kind: model.TypeString}},
			}},
			Optional: true,
		},
	}}

	got := ctyTypeToSpec(constraint, defaults)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ctyTypeToSpec() mismatch (-want +got):\n%s", diff)
	}
}

// parseTypeConstraint parses an HCL type-constraint expression and returns
// the resulting cty.Type and optional-attribute defaults.
func parseTypeConstraint(t *testing.T, src string) (cty.Type, *typeexpr.Defaults) {
	t.Helper()
	expr, diags := hclsyntax.ParseExpression([]byte(src), "test.tf", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatalf("parse expression: %v", diags)
	}
	constraint, defaults, tdiag := typeexpr.TypeConstraintWithDefaults(expr)
	if tdiag.HasErrors() {
		t.Fatalf("type constraint: %v", tdiag)
	}
	return constraint, defaults
}
