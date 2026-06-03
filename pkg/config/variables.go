// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/zclconf/go-cty/cty"

	"github.com/remoterabbit/open-inspector/pkg/model"
)

var variableSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "type"},
		{Name: "default"},
		{Name: "description"},
		{Name: "sensitive"},
		{Name: "nullable"},
		{Name: "ephemeral"},
	},
	Blocks: []hcl.BlockHeaderSchema{{Type: "validation"}},
}

var validationSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "condition", Required: true},
		{Name: "error_message", Required: true},
	},
}

func decodeVariableBlock(block *hcl.Block, source []byte, module *model.Module) model.Diagnostics {
	variable := model.Variable{
		Name:  block.Labels[0],
		Range: model.RangeFromHcl(block.DefRange),
	}
	content, _, hdiag := block.Body.PartialContent(variableSchema)
	diags := model.DiagnosticsFromHCL(hdiag)

	if attribute, ok := content.Attributes["type"]; ok {
		typeConstraint, _, tdiag := typeexpr.TypeConstraintWithDefaults(attribute.Expr)
		diags = append(diags, model.DiagnosticsFromHCL(tdiag)...)
		if typeConstraint != cty.NilType {
			variable.Type = typeexpr.TypeString(typeConstraint)
		}
	}

	if attribute, ok := content.Attributes["default"]; ok {
		expression := capture(attribute.Expr, source)
		variable.Default = &expression
	}

	if attribute, ok := content.Attributes["description"]; ok {
		if str, ok := literalString(attribute.Expr); ok {
			variable.Description = str
		}
	}

	if value, set, bdiag := decodeBool(content.Attributes["sensitive"]); set {
		variable.Sensitive = value
		diags = append(diags, model.DiagnosticsFromHCL(bdiag)...)
	}

	if value, set, bdiag := decodeBool(content.Attributes["nullable"]); set {
		variable.Nullable = &value
		diags = append(diags, model.DiagnosticsFromHCL(bdiag)...)
	}

	if value, set, bdiag := decodeBool(content.Attributes["ephemeral"]); set {
		variable.Ephemeral = value
		diags = append(diags, model.DiagnosticsFromHCL(bdiag)...)
	}

	for _, variableBlock := range content.Blocks.OfType("validation") {
		inner, _, vdiag := variableBlock.Body.PartialContent(validationSchema)
		diags = append(diags, model.DiagnosticsFromHCL(vdiag)...)

		variable.Validations = append(variable.Validations, model.Validation{
			Condition:    capture(inner.Attributes["condition"].Expr, source),
			ErrorMessage: capture(inner.Attributes["error_message"].Expr, source),
			Range:        model.RangeFromHcl(variableBlock.DefRange),
		})
	}
	module.Variables = append(module.Variables, variable)
	return diags
}

// decodeBool returns (value, wasSet, diags). Use a *bool in the model when wasSet
// must be distinguishable from a literal false.
func decodeBool(attribute *hcl.Attribute) (bool, bool, hcl.Diagnostics) {
	if attribute == nil {
		return false, false, nil
	}

	value, diag := attribute.Expr.Value(nil)
	if diag.HasErrors() || value.IsNull() || value.Type() != cty.Bool {
		return false, false, diag
	}
	return value.True(), true, diag
}

func literalString(expression hcl.Expression) (string, bool) {
	value, diag := expression.Value(nil)
	if diag.HasErrors() || value.IsNull() || value.Type() != cty.String {
		return "", false
	}
	return value.AsString(), true
}
