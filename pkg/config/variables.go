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

// decodeVariableBlock decodes a variable {} block into module.Variables,
// capturing its type and default verbatim, its scalar flags (sensitive,
// nullable, ephemeral), and any validation {} blocks.
func decodeVariableBlock(block *hcl.Block, source []byte, module *model.Module) model.Diagnostics {
	variable := model.Variable{
		Name:  block.Labels[0],
		Range: model.RangeFromHCL(block.DefRange),
	}
	content, _, hdiag := block.Body.PartialContent(variableSchema)
	diags := model.DiagnosticsFromHCL(hdiag)

	if attribute, ok := content.Attributes["type"]; ok {
		// Parse with typeexpr purely for diagnostics (invalid types still
		// surface as errors). The serialized value is the verbatim source
		// of the type expression so that `optional(T, default)` markers
		// and any other user-authored detail survive a round-trip.
		_, _, tdiag := typeexpr.TypeConstraintWithDefaults(attribute.Expr)
		diags = append(diags, model.DiagnosticsFromHCL(tdiag)...)
		variable.Type = sliceSourceLF(source, attribute.Expr.Range())
	}

	if attribute, ok := content.Attributes["default"]; ok {
		expression := capture(attribute.Expr, source)
		variable.Default = &expression
	}

	if attribute, ok := content.Attributes["description"]; ok {
		str, ok, sdiag := literalString(attribute.Expr)
		diags = append(diags, model.DiagnosticsFromHCL(sdiag)...)
		if ok {
			variable.Description = str
		}
	}

	{
		value, set, bdiag := decodeBool(content.Attributes["sensitive"])
		diags = append(diags, model.DiagnosticsFromHCL(bdiag)...)
		if set {
			variable.Sensitive = value
		}
	}

	{
		value, set, bdiag := decodeBool(content.Attributes["nullable"])
		diags = append(diags, model.DiagnosticsFromHCL(bdiag)...)
		if set {
			variable.Nullable = &value
		}
	}

	{
		value, set, bdiag := decodeBool(content.Attributes["ephemeral"])
		diags = append(diags, model.DiagnosticsFromHCL(bdiag)...)
		if set {
			variable.Ephemeral = value
		}
	}

	for _, variableBlock := range content.Blocks.OfType("validation") {
		inner, _, vdiag := variableBlock.Body.PartialContent(validationSchema)
		diags = append(diags, model.DiagnosticsFromHCL(vdiag)...)

		// PartialContent records a diagnostic for missing Required
		// attributes but still returns; the entries are simply absent
		// from the map. Skip the block to avoid a nil dereference;
		// the diagnostic already tells the user what's wrong.
		condition, hasCondition := inner.Attributes["condition"]
		errorMessage, hasErrorMessage := inner.Attributes["error_message"]
		if !hasCondition || !hasErrorMessage {
			continue
		}

		variable.Validations = append(variable.Validations, model.Validation{
			Condition:    capture(condition.Expr, source),
			ErrorMessage: capture(errorMessage.Expr, source),
			Range:        model.RangeFromHCL(variableBlock.DefRange),
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

// literalString evaluates expression as a constant string. It returns
// the value and ok=true on success; on failure the third return carries
// any diagnostics produced by .Value(nil) so the caller can surface
// them (e.g. "Variables not allowed" when the user wrote an
// interpolation where a literal is required).
func literalString(expression hcl.Expression) (string, bool, hcl.Diagnostics) {
	value, diag := expression.Value(nil)
	if diag.HasErrors() || value.IsNull() || value.Type() != cty.String {
		return "", false, diag
	}
	return value.AsString(), true, diag
}
