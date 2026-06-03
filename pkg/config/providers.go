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
	},
}

// decodeProviderBlock decodes a single provider "name" {} block and
// appends a ProviderConfig to mod.
func decodeProviderBlock(block *hcl.Block, mod *model.Module) model.Diagnostics {
	cfg := model.ProviderConfig{
		Name:  block.Labels[0],
		Range: model.RangeFromHcl(block.DefRange),
	}

	content, _, hd := block.Body.PartialContent(providerSchema)
	diags := model.DiagnosticsFromHCL(hd)

	if attr, ok := content.Attributes["alias"]; ok {
		val, vd := attr.Expr.Value(nil)
		diags = append(diags, model.DiagnosticsFromHCL(vd)...)
		if !val.IsNull() && val.Type() == cty.String {
			cfg.Alias = val.AsString()
		}
	}

	mod.Providers = append(mod.Providers, cfg)
	return diags
}
