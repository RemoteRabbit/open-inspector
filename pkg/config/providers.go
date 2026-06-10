// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"

	"github.com/remoterabbit/open-inspector/pkg/model"
)

// providerSchema describes the inside of a provider "name" {} block.
var providerSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "alias"},
		{Name: "for_each"},
	},
}

// decodeProviderBlock decodes a single provider "name" {} block and
// appends a ProviderConfig to mod.
func decodeProviderBlock(block *hcl.Block, source []byte, module *model.Module) model.Diagnostics {
	config := model.ProviderConfig{
		Name:     block.Labels[0],
		Position: model.PositionFromHCL(block.DefRange),
	}

	content, _, hdiag := block.Body.PartialContent(providerSchema)
	diags := model.DiagnosticsFromHCL(hdiag)

	if attribute, ok := content.Attributes["alias"]; ok {
		value, vdiag := attribute.Expr.Value(nil)
		diags = append(diags, model.DiagnosticsFromHCL(vdiag)...)
		if !value.IsNull() && value.IsKnown() && value.Type() == cty.String {
			config.Alias = value.AsString()
		}
	}

	if attribute, ok := content.Attributes["for_each"]; ok {
		expression := capture(attribute.Expr, source)
		config.ForEach = &expression
	}

	module.Providers = append(module.Providers, config)
	return diags
}
