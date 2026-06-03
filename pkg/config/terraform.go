// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"

	"github.com/remoterabbit/open-inspector/pkg/model"
)

// terraformSchema describes the inside of a terraform {} block.
var terraformSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "required_version"},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{Type: "required_providers"},
	},
}

// decodeTerraformBlock decodes a single terraform {} block, appending
// any findings to mod. Multiple terraform blocks across files are
// aggregated: required_version values are appended in encounter order
// and required_providers entries are merged keyed by provider name
// (last definition wins on conflict).
func decodeTerraformBlock(block *hcl.Block, mod *model.Module) model.Diagnostics {
	content, _, hd := block.Body.PartialContent(terraformSchema)
	diags := model.DiagnosticsFromHCL(hd)

	if attr, ok := content.Attributes["required_version"]; ok {
		val, vd := attr.Expr.Value(nil)
		diags = append(diags, model.DiagnosticsFromHCL(vd)...)
		if !val.IsNull() && val.Type() == cty.String {
			mod.RequiredCore = append(mod.RequiredCore, val.AsString())
		}
	}

	for _, rp := range content.Blocks.OfType("required_providers") {
		diags = append(diags, decodeRequiredProviders(rp, mod)...)
	}

	return diags
}

// decodeRequiredProviders decodes a required_providers {} block and
// stores entries on mod.RequiredProviders.
func decodeRequiredProviders(block *hcl.Block, mod *model.Module) model.Diagnostics {
	attrs, hd := block.Body.JustAttributes()
	diags := model.DiagnosticsFromHCL(hd)

	for name, attr := range attrs {
		req, rd := decodeProviderReq(attr)
		diags = append(diags, rd...)
		if mod.RequiredProviders == nil {
			mod.RequiredProviders = map[string]model.ProviderRequirement{}
		}
		mod.RequiredProviders[name] = req
	}
	return diags
}

// decodeProviderReq handles both the legacy string form
// (`aws = "~> 4.0"`) and the modern object form with source, version,
// and configuration_aliases.
func decodeProviderReq(attr *hcl.Attribute) (model.ProviderRequirement, model.Diagnostics) {
	req := model.ProviderRequirement{Range: model.RangeFromHcl(attr.Range)}
	var diags model.Diagnostics

	// Legacy form: a single string version constraint.
	if val, vd := attr.Expr.Value(nil); !vd.HasErrors() &&
		!val.IsNull() && val.Type() == cty.String {
		req.VersionConstraints = []string{val.AsString()}
		return req, diags
	}

	// Modern form: an object expression. Walk key/value pairs without
	// evaluating, so traversals like `aws.east` don't trip up the
	// evaluator.
	pairs, hd := hcl.ExprMap(attr.Expr)
	diags = append(diags, model.DiagnosticsFromHCL(hd)...)

	for _, p := range pairs {
		keyVal, kd := p.Key.Value(nil)
		diags = append(diags, model.DiagnosticsFromHCL(kd)...)
		if keyVal.IsNull() || keyVal.Type() != cty.String {
			continue
		}
		switch keyVal.AsString() {
		case "source":
			v, vd := p.Value.Value(nil)
			diags = append(diags, model.DiagnosticsFromHCL(vd)...)
			if !v.IsNull() && v.Type() == cty.String {
				req.Source = v.AsString()
			}
		case "version":
			v, vd := p.Value.Value(nil)
			diags = append(diags, model.DiagnosticsFromHCL(vd)...)
			if !v.IsNull() && v.Type() == cty.String {
				req.VersionConstraints = append(req.VersionConstraints, v.AsString())
			}
		case "configuration_aliases":
			aliases, ad := decodeConfigurationAliases(p.Value)
			diags = append(diags, ad...)
			req.ConfigurationAliases = aliases
		}
	}
	return req, diags
}

// decodeConfigurationAliases turns `[aws.east, aws.west]` into
// `["aws.east", "aws.west"]`. The list elements are traversals to
// not-yet-declared provider aliases, so plain expression evaluation is
// not an option.
func decodeConfigurationAliases(expr hcl.Expression) ([]string, model.Diagnostics) {
	list, hd := hcl.ExprList(expr)
	diags := model.DiagnosticsFromHCL(hd)

	out := make([]string, 0, len(list))
	for _, e := range list {
		trav, td := hcl.AbsTraversalForExpr(e)
		if td.HasErrors() {
			diags = append(diags, model.DiagnosticsFromHCL(td)...)
			continue
		}
		out = append(out, traversalString(trav))
	}
	return out, diags
}
