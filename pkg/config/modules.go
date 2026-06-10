// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/remoterabbit/open-inspector/pkg/model"
	"github.com/zclconf/go-cty/cty"
)

var moduleSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "source", Required: true},
		{Name: "version"},
		{Name: "count"}, {Name: "for_each"},
		{Name: "providers"}, {Name: "depends_on"},
	},
}

// decodeModuleCallBlock decodes a module {} invocation into
// module.ModuleCalls: the literal source/version plus the count, for_each,
// providers, and depends_on meta-arguments.
func decodeModuleCallBlock(block *hcl.Block, source []byte, module *model.Module) model.Diagnostics {
	moduleCall := model.ModuleCall{
		Name:  block.Labels[0],
		Range: model.RangeFromHCL(block.DefRange),
	}

	inner, _, hdiag := block.Body.PartialContent(moduleSchema)
	diags := model.DiagnosticsFromHCL(hdiag)

	if attribute, ok := inner.Attributes["source"]; ok {
		str, ok, sdiag := literalString(attribute.Expr)
		diags = append(diags, model.DiagnosticsFromHCL(sdiag)...)
		if ok {
			moduleCall.Source = str
		}
	}

	if attribute, ok := inner.Attributes["version"]; ok {
		str, ok, sdiag := literalString(attribute.Expr)
		diags = append(diags, model.DiagnosticsFromHCL(sdiag)...)
		if ok {
			moduleCall.Version = str
		}
	}

	if attribute, ok := inner.Attributes["count"]; ok {
		expression := capture(attribute.Expr, source)
		moduleCall.Count = &expression
	}

	if attribute, ok := inner.Attributes["for_each"]; ok {
		expression := capture(attribute.Expr, source)
		moduleCall.ForEach = &expression
	}

	if attribute, ok := inner.Attributes["depends_on"]; ok {
		deps, ddiag := decodeTraversalList(attribute.Expr)
		diags = append(diags, ddiag...)
		moduleCall.DependsOn = deps
	}

	if attribute, ok := inner.Attributes["providers"]; ok {
		pMap, pdiag := decodeProviderMap(attribute.Expr)
		diags = append(diags, pdiag...)
		moduleCall.Providers = pMap
	}

	module.ModuleCalls = append(module.ModuleCalls, moduleCall)
	return diags
}

// decodeProviderMap handles `providers = { aws = aws.east }`. Keys are bare
// identifiers (decode as literal strings); values are traversals (use AbsTraversalForExpr,
// not Value()).
func decodeProviderMap(expression hcl.Expression) (map[string]string, model.Diagnostics) {
	pairs, hdiag := hcl.ExprMap(expression)
	diags := model.DiagnosticsFromHCL(hdiag)
	output := map[string]string{}

	for _, pair := range pairs {
		key, kdiag := pair.Key.Value(nil)
		diags = append(diags, model.DiagnosticsFromHCL(kdiag)...)
		if key.IsNull() || !key.IsKnown() || key.Type() != cty.String {
			continue
		}
		traversal, tdiag := hcl.AbsTraversalForExpr(pair.Value)
		if tdiag.HasErrors() {
			diags = append(diags, model.DiagnosticsFromHCL(tdiag)...)
			continue
		}
		output[key.AsString()] = traversalString(traversal)
	}
	return output, diags
}
