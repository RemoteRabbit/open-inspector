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

func decodeOuputsBlock(block *hcl.Block, source []byte, module *model.Module) model.Diagnostics {
	output := model.Output{
		Name:  block.Labels[0],
		Range: model.RangeFromHcl(block.DefRange),
	}

	content, _, hdiag := block.Body.PartialContent(outputSchema)
	diags := model.DiagnosticsFromHCL(hdiag)

	if attribute, ok := content.Attributes["value"]; ok {
		output.Value = capture(attribute.Expr, source)
	}

	if attribute, ok := content.Attributes["description"]; ok {
		if str, ok := literalString(attribute.Expr); ok {
			output.Description = str
		}
	}

	if value, set, _ := decodeBool(content.Attributes["sensitive"]); set {
		output.Sensitive = value
	}

	if value, set, _ := decodeBool(content.Attributes["ephemeral"]); set {
		output.Ephemeral = value
	}

	if attribute, ok := content.Attributes["depends_on"]; ok {
		deps, ddiag := decodeTraversalList(attribute.Expr)
		diags = append(diags, ddiag...)
		output.DependsOn = deps
	}

	module.Outputs = append(module.Outputs, output)
	return diags
}
