// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/remoterabbit/open-inspector/pkg/model"
)

var checkSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{Type: "data", LabelNames: []string{"type", "name"}},
		{Type: "assert"},
	},
}

// decodeCheckBlock decodes a check "<name>" {} block into module.Checks,
// including its optional scoped data source and its assert {} blocks.
func decodeCheckBlock(block *hcl.Block, source []byte, module *model.Module) model.Diagnostics {
	inner, _, hdiag := block.Body.PartialContent(checkSchema)
	diags := model.DiagnosticsFromHCL(hdiag)

	check := model.CheckBlock{
		Name:  block.Labels[0],
		Range: model.RangeFromHCL(block.DefRange),
	}

	// NOTE: a check has at most one data block in practice. If a malformed
	// config declares several, only the last is captured and the extras are
	// silently dropped (no diagnostic).
	for _, dataBlock := range inner.Blocks.OfType("data") {
		// Decode into a throwaway module so we can capture the result without polluting
		// Module.DataResources.
		var tmp model.Module
		ddiag := decodeResourceBlock(dataBlock, source, model.DataResourceMode, &tmp)
		diags = append(diags, ddiag...)

		if len(tmp.DataResources) > 0 {
			resource := tmp.DataResources[0]
			check.DataSource = &resource
		}
	}

	for _, assertBlock := range inner.Blocks.OfType("assert") {
		assertion, ok, adiag := decodeAssertion(assertBlock, source)
		diags = append(diags, adiag...)

		if ok {
			check.Assertions = append(check.Assertions, assertion)
		}
	}
	module.Checks = append(module.Checks, check)
	return diags
}

// decodeAssertion decodes a single assert { condition, error_message }
// block. The bool result is false when the block is malformed and should
// be skipped.
func decodeAssertion(block *hcl.Block, source []byte) (model.Assertion, bool, model.Diagnostics) {
	inner, _, adiag := block.Body.PartialContent(validationSchema)
	diags := model.DiagnosticsFromHCL(adiag)

	condition, hasCondition := inner.Attributes["condition"]
	errorMessage, hasErrorMessage := inner.Attributes["error_message"]
	if !hasCondition || !hasErrorMessage {
		return model.Assertion{}, false, diags
	}

	return model.Assertion{
		Condition:    capture(condition.Expr, source),
		ErrorMessage: capture(errorMessage.Expr, source),
		Range:        model.RangeFromHCL(block.DefRange),
	}, true, diags
}
