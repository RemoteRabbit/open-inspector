// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"sort"

	"github.com/hashicorp/hcl/v2"

	"github.com/remoterabbit/open-inspector/pkg/model"
)

var resourceSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "count"},
		{Name: "for_each"},
		{Name: "provider"},
		{Name: "depends_on"},
	},
	Blocks: []hcl.BlockHeaderSchema{{Type: "lifecycle"}},
}

// decodeResourceBlock decodes a resource {} or data {} block into the
// model, appending it to the managed or data slice according to mode. It
// captures the meta-arguments (count, for_each, provider, depends_on) and
// the lifecycle {} block.
func decodeResourceBlock(block *hcl.Block, source []byte, mode model.ResourceMode, module *model.Module) model.Diagnostics {
	resource := model.Resource{
		Mode:     mode,
		Type:     block.Labels[0],
		Name:     block.Labels[1],
		Position: model.PositionFromHCL(block.DefRange),
	}

	inner, remain, hdiag := block.Body.PartialContent(resourceSchema)
	diags := model.DiagnosticsFromHCL(hdiag)

	// Capture the user-set attribute names (everything left after the
	// meta-args and lifecycle block were consumed) for downstream schema
	// enrichment. JustAttributes reports a diagnostic when the remainder
	// still holds nested blocks (provisioner, dynamic, etc.); we ignore
	// that and keep the attributes it does return. Nested-block contents
	// are intentionally not captured here.
	if attributes, _ := remain.JustAttributes(); len(attributes) > 0 {
		names := make([]string, 0, len(attributes))
		for name := range attributes {
			names = append(names, name)
		}
		sort.Strings(names)
		resource.AttrNames = names
	}

	if attribute, ok := inner.Attributes["count"]; ok {
		expression := capture(attribute.Expr, source)
		resource.Count = &expression
	}

	if attribute, ok := inner.Attributes["for_each"]; ok {
		expression := capture(attribute.Expr, source)
		resource.ForEach = &expression
	}

	if attribute, ok := inner.Attributes["provider"]; ok {
		if traversal, tdiag := hcl.AbsTraversalForExpr(attribute.Expr); !tdiag.HasErrors() {
			resource.Provider = traversalString(traversal)
		} else {
			diags = append(diags, model.DiagnosticsFromHCL(tdiag)...)
		}
	}

	if attribute, ok := inner.Attributes["depends_on"]; ok {
		deps, ddiag := decodeTraversalList(attribute.Expr)
		diags = append(diags, ddiag...)
		resource.DependsOn = deps
	}

	for _, lifeBlock := range inner.Blocks.OfType("lifecycle") {
		lifeCycle, ldiag := decodeLifecycle(lifeBlock, source)
		diags = append(diags, ldiag...)
		resource.Lifecycle = lifeCycle
	}

	switch mode {
	case model.ManagedResourceMode:
		module.ManagedResources = append(module.ManagedResources, resource)
	case model.DataResourceMode:
		module.DataResources = append(module.DataResources, resource)
	case model.EphemeralResourceMode:
		module.EphemeralResources = append(module.EphemeralResources,
			model.EphemeralResource{
				Type: resource.Type, Name: resource.Name,
				Provider: resource.Provider,
				Count:    resource.Count, ForEach: resource.ForEach,
				DependsOn: resource.DependsOn,
				AttrNames: resource.AttrNames,
				Lifecycle: resource.Lifecycle,
				Position:  resource.Position,
			})
	}
	return diags
}

var lifecycleSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "create_before_destroy"},
		{Name: "prevent_destroy"},
		{Name: "ignore_changes"},
		{Name: "replace_triggered_by"},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{Type: "precondition"},
		{Type: "postcondition"},
	},
}

// decodeLifecycle decodes a resource lifecycle {} block: the scalar flags
// (create_before_destroy, prevent_destroy), the traversal lists
// (ignore_changes, replace_triggered_by), and the precondition/
// postcondition blocks.
func decodeLifecycle(block *hcl.Block, source []byte) (*model.Lifecycle, model.Diagnostics) {
	inner, _, hdiag := block.Body.PartialContent(lifecycleSchema)
	diags := model.DiagnosticsFromHCL(hdiag)
	lifeCycle := &model.Lifecycle{}

	if block, set, _ := decodeBool(inner.Attributes["create_before_destroy"]); set {
		lifeCycle.CreateBeforeDestroy = &block
	}

	if block, set, _ := decodeBool(inner.Attributes["prevent_destroy"]); set {
		lifeCycle.PreventDestroy = &block
	}

	if attribute, ok := inner.Attributes["ignore_changes"]; ok {
		// ignore_changes accepts either a list (`[id, tags]`) OR the bare keyword
		// `all`. AbsTraversalForExpr handles `all` as a single TraverseRoot; fall
		// back to list decoding otherwise.
		if traversal, tdiag := hcl.AbsTraversalForExpr(attribute.Expr); !tdiag.HasErrors() {
			lifeCycle.IgnoreChanges = []string{traversalString(traversal)}
		} else {
			list, ldiag := decodeTraversalList(attribute.Expr)
			diags = append(diags, ldiag...)
			lifeCycle.IgnoreChanges = list
		}
	}

	if attribute, ok := inner.Attributes["replace_triggered_by"]; ok {
		list, ldiag := decodeTraversalList(attribute.Expr)
		diags = append(diags, ldiag...)
		lifeCycle.ReplaceTriggeredBy = list
	}

	for _, preBlock := range inner.Blocks.OfType("precondition") {
		value, ok, vdiag := decodeConditionBlock(preBlock, source)
		diags = append(diags, vdiag...)
		if ok {
			lifeCycle.Preconditions = append(lifeCycle.Preconditions, value)
		}
	}

	for _, postBlock := range inner.Blocks.OfType("postcondition") {
		value, ok, vdiag := decodeConditionBlock(postBlock, source)
		diags = append(diags, vdiag...)
		if ok {
			lifeCycle.Postconditions = append(lifeCycle.Postconditions, value)
		}
	}

	return lifeCycle, diags
}

// decodeConditionBlock is the same shape as the variable validation block.
// Pull it out so variables.go and resources.go share one implementation.
// Returns ok=false when the block is missing one of its required
// attributes, so the caller can skip it; the diagnostic from
// PartialContent already explains the failure to the user.
func decodeConditionBlock(block *hcl.Block, source []byte) (model.Validation, bool, model.Diagnostics) {
	inner, _, hdiag := block.Body.PartialContent(validationSchema)
	diags := model.DiagnosticsFromHCL(hdiag)

	condition, hasCondition := inner.Attributes["condition"]
	errorMessage, hasErrorMessage := inner.Attributes["error_message"]
	if !hasCondition || !hasErrorMessage {
		return model.Validation{}, false, diags
	}

	return model.Validation{
		Condition:    capture(condition.Expr, source),
		ErrorMessage: capture(errorMessage.Expr, source),
		Position:     model.PositionFromHCL(block.DefRange),
	}, true, diags
}
