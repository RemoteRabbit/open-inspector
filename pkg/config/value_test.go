// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/zclconf/go-cty/cty"

	"github.com/remoterabbit/open-inspector/pkg/model"
)

func TestCtyValueToValue(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		value cty.Value
		want  *model.Value
	}{
		{
			name:  "string",
			value: cty.StringVal("hello"),
			want:  &model.Value{Kind: model.ValueString, String: "hello"},
		},
		{
			name:  "number",
			value: cty.NumberIntVal(42),
			want:  &model.Value{Kind: model.ValueNumber, Number: "42"},
		},
		{
			name:  "big number keeps precision",
			value: cty.MustParseNumberVal("1234567890123456789"),
			want:  &model.Value{Kind: model.ValueNumber, Number: "1234567890123456789"},
		},
		{
			name:  "bool",
			value: cty.True,
			want:  &model.Value{Kind: model.ValueBool, Bool: true},
		},
		{
			name:  "list",
			value: cty.ListVal([]cty.Value{cty.StringVal("a"), cty.StringVal("b")}),
			want: &model.Value{Kind: model.ValueList, List: []*model.Value{
				{Kind: model.ValueString, String: "a"},
				{Kind: model.ValueString, String: "b"},
			}},
		},
		{
			name:  "set renders as list",
			value: cty.SetVal([]cty.Value{cty.StringVal("a")}),
			want: &model.Value{Kind: model.ValueList, List: []*model.Value{
				{Kind: model.ValueString, String: "a"},
			}},
		},
		{
			name:  "tuple",
			value: cty.TupleVal([]cty.Value{cty.StringVal("a"), cty.NumberIntVal(1)}),
			want: &model.Value{Kind: model.ValueTuple, Tuple: []*model.Value{
				{Kind: model.ValueString, String: "a"},
				{Kind: model.ValueNumber, Number: "1"},
			}},
		},
		{
			name:  "map",
			value: cty.MapVal(map[string]cty.Value{"env": cty.StringVal("prod")}),
			want: &model.Value{Kind: model.ValueMap, Map: map[string]*model.Value{
				"env": {Kind: model.ValueString, String: "prod"},
			}},
		},
		{
			name:  "object",
			value: cty.ObjectVal(map[string]cty.Value{"name": cty.StringVal("web"), "port": cty.NumberIntVal(8080)}),
			want: &model.Value{Kind: model.ValueObject, Object: map[string]*model.Value{
				"name": {Kind: model.ValueString, String: "web"},
				"port": {Kind: model.ValueNumber, Number: "8080"},
			}},
		},
		{
			name:  "typed null",
			value: cty.NullVal(cty.String),
			want:  &model.Value{Kind: model.ValueNull},
		},
		{
			name:  "unknown is nil",
			value: cty.UnknownVal(cty.String),
			want:  nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := ctyValueToValue(tc.value)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("ctyValueToValue() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
