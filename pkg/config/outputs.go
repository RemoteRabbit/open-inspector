// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"github.com/hashicorp/hcl/v2"

	"github.com/remoterabbit/open-inspector/pkg/model"
)

var outputSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "value", Required: true},
		{Name: "description"},
		{Name: "sensitive"},
		{Name: "ephemeral"},
		{Name: "depends_on"},
	},
}

// decodeOutputsBlock decodes an output {} block into module.Outputs:
// the value expression (captured verbatim), the description, the
// sensitive/ephemeral flags, and depends_on.
func decodeOutputsBlock(block *hcl.Block, source []byte, module *model.Module) model.Diagnostics {
	output := model.Output{
		Name:  block.Labels[0],
		Range: model.RangeFromHCL(block.DefRange),
	}

	content, _, hdiag := block.Body.PartialContent(outputSchema)
	diags := model.DiagnosticsFromHCL(hdiag)

	if attribute, ok := content.Attributes["value"]; ok {
		output.Value = capture(attribute.Expr, source)
	}

	if attribute, ok := content.Attributes["description"]; ok {
		str, ok, sdiag := literalString(attribute.Expr)
		diags = append(diags, model.DiagnosticsFromHCL(sdiag)...)
		if ok {
			output.Description = str
		}
	}

	{
		value, set, bdiag := decodeBool(content.Attributes["sensitive"])
		diags = append(diags, model.DiagnosticsFromHCL(bdiag)...)
		if set {
			output.Sensitive = value
		}
	}

	{
		value, set, bdiag := decodeBool(content.Attributes["ephemeral"])
		diags = append(diags, model.DiagnosticsFromHCL(bdiag)...)
		if set {
			output.Ephemeral = value
		}
	}

	if attribute, ok := content.Attributes["depends_on"]; ok {
		deps, ddiag := decodeTraversalList(attribute.Expr)
		diags = append(diags, ddiag...)
		output.DependsOn = deps
	}

	module.Outputs = append(module.Outputs, output)
	return diags
}
