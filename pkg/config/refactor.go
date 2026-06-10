// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"github.com/hashicorp/hcl/v2"

	"github.com/remoterabbit/open-inspector/pkg/model"
)

var movedSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "from", Required: true},
		{Name: "to", Required: true},
	},
}

var importSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "to", Required: true},
		{Name: "id", Required: true},
		{Name: "provider"},
	},
}

var removedSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "from", Required: true},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{Type: "lifecycle"},
	},
}

var removedLifecycleSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{{Name: "destroy"}},
}

// decodeMovedBlock decodes a moved { from, to } block into module.Moved.
func decodeMovedBlock(block *hcl.Block, module *model.Module) model.Diagnostics {
	inner, _, hdiag := block.Body.PartialContent(movedSchema)
	diags := model.DiagnosticsFromHCL(hdiag)

	from, fdiag := traversalStringFromAttr(inner.Attributes["from"])
	diags = append(diags, fdiag...)

	to, tdiag := traversalStringFromAttr(inner.Attributes["to"])
	diags = append(diags, tdiag...)

	module.Moved = append(module.Moved, model.MovedBlock{
		From:     from,
		To:       to,
		Position: model.PositionFromHCL(block.DefRange),
	})
	return diags
}

// decodeImportBlock decodes an import { to, id, provider } block into
// module.Imports, capturing the id expression verbatim.
func decodeImportBlock(block *hcl.Block, source []byte, module *model.Module) model.Diagnostics {
	inner, _, hdiag := block.Body.PartialContent(importSchema)
	diags := model.DiagnosticsFromHCL(hdiag)

	to, tdiag := traversalStringFromAttr(inner.Attributes["to"])
	diags = append(diags, tdiag...)

	imp := model.ImportBlock{
		To:       to,
		Position: model.PositionFromHCL(block.DefRange),
	}

	// id can be a string (TF 1.5) or an object (TF 1.6+). Capture as
	// Expression source bytes; consumers decide how to interpret.
	if attr, ok := inner.Attributes["id"]; ok {
		imp.ID = capture(attr.Expr, source)
	}

	if attr, ok := inner.Attributes["provider"]; ok {
		provider, pdiag := traversalStringFromAttr(attr)
		diags = append(diags, pdiag...)
		imp.Provider = provider
	}

	module.Imports = append(module.Imports, imp)
	return diags
}

// decodeRemovedBlock decodes a removed { from, lifecycle { destroy } }
// block into module.Removed.
func decodeRemovedBlock(block *hcl.Block, module *model.Module) model.Diagnostics {
	inner, _, hdiag := block.Body.PartialContent(removedSchema)
	diags := model.DiagnosticsFromHCL(hdiag)

	from, fdiag := traversalStringFromAttr(inner.Attributes["from"])
	diags = append(diags, fdiag...)

	removed := model.RemovedBlock{
		From:     from,
		Position: model.PositionFromHCL(block.DefRange),
	}

	// removed.lifecycle is a meta-block scoped to this construct; do
	// NOT reuse the resource lifecycle decoder. The only attribute the
	// docs define here is `destroy`.
	for _, lifecycleBlock := range inner.Blocks.OfType("lifecycle") {
		lifeContent, _, lhdiag := lifecycleBlock.Body.PartialContent(removedLifecycleSchema)
		diags = append(diags, model.DiagnosticsFromHCL(lhdiag)...)

		value, set, bdiag := decodeBool(lifeContent.Attributes["destroy"])
		diags = append(diags, model.DiagnosticsFromHCL(bdiag)...)
		if set {
			v := value
			removed.DestroyOnDrop = &v
		}
	}

	module.Removed = append(module.Removed, removed)
	return diags
}
