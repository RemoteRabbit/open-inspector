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
		{Type: "encryption"},
	},
}

// decodeTerraformBlock decodes a single terraform {} block, appending
// any findings to mod. Multiple terraform blocks across files are
// aggregated: required_version values are appended in encounter order
// and required_providers entries are merged keyed by provider name
// (last definition wins on conflict).
func decodeTerraformBlock(block *hcl.Block, source []byte, module *model.Module) model.Diagnostics {
	content, _, hdiag := block.Body.PartialContent(terraformSchema)
	diags := model.DiagnosticsFromHCL(hdiag)

	if attribute, ok := content.Attributes["required_version"]; ok {
		value, vdiag := attribute.Expr.Value(nil)
		diags = append(diags, model.DiagnosticsFromHCL(vdiag)...)
		if !value.IsNull() && value.Type() == cty.String {
			module.RequiredCore = append(module.RequiredCore, value.AsString())
		}
	}

	for _, requiredProvider := range content.Blocks.OfType("required_providers") {
		diags = append(diags, decodeRequiredProviders(requiredProvider, module)...)
	}

	for _, encryption := range content.Blocks.OfType("encryption") {
		diags = append(diags, decodeEncryptionBlock(encryption, source, module)...)
	}

	return diags
}

// decodeRequiredProviders decodes a required_providers {} block and
// stores entries on mod.RequiredProviders.
func decodeRequiredProviders(block *hcl.Block, module *model.Module) model.Diagnostics {
	attributes, hdiag := block.Body.JustAttributes()
	diags := model.DiagnosticsFromHCL(hdiag)

	for name, attribute := range attributes {
		req, rdiag := decodeProviderReq(attribute)
		diags = append(diags, rdiag...)
		if module.RequiredProviders == nil {
			module.RequiredProviders = map[string]model.ProviderRequirement{}
		}
		module.RequiredProviders[name] = req
	}
	return diags
}

// decodeProviderReq handles both the legacy string form
// (`aws = "~> 4.0"`) and the modern object form with source, version,
// and configuration_aliases.
func decodeProviderReq(attribute *hcl.Attribute) (model.ProviderRequirement, model.Diagnostics) {
	req := model.ProviderRequirement{Range: model.RangeFromHcl(attribute.Range)}
	var diags model.Diagnostics

	// Legacy form: a single string version constraint.
	if value, vdiag := attribute.Expr.Value(nil); !vdiag.HasErrors() &&
		!value.IsNull() && value.Type() == cty.String {
		req.VersionConstraints = []string{value.AsString()}
		return req, diags
	}

	// Modern form: an object expression. Walk key/value pairs without
	// evaluating, so traversals like `aws.east` don't trip up the
	// evaluator.
	pairs, hdiag := hcl.ExprMap(attribute.Expr)
	diags = append(diags, model.DiagnosticsFromHCL(hdiag)...)

	for _, pair := range pairs {
		keyVal, kdiag := pair.Key.Value(nil)
		diags = append(diags, model.DiagnosticsFromHCL(kdiag)...)
		if keyVal.IsNull() || keyVal.Type() != cty.String {
			continue
		}
		switch keyVal.AsString() {
		case "source":
			value, vdiag := pair.Value.Value(nil)
			diags = append(diags, model.DiagnosticsFromHCL(vdiag)...)
			if !value.IsNull() && value.Type() == cty.String {
				req.Source = value.AsString()
			}
		case "version":
			value, vdiag := pair.Value.Value(nil)
			diags = append(diags, model.DiagnosticsFromHCL(vdiag)...)
			if !value.IsNull() && value.Type() == cty.String {
				req.VersionConstraints = append(req.VersionConstraints, value.AsString())
			}
		case "configuration_aliases":
			aliases, adiag := decodeTraversalList(pair.Value)
			diags = append(diags, adiag...)
			req.ConfigurationAliases = aliases
		}
	}
	return req, diags
}
