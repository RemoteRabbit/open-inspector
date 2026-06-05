// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"github.com/hashicorp/hcl/v2"

	"github.com/remoterabbit/open-inspector/pkg/model"
)

var encryptionSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{Type: "key_provider", LabelNames: []string{"type", "name"}},
		{Type: "method", LabelNames: []string{"type", "name"}},
		{Type: "state"},
		{Type: "plan"},
		{Type: "remote_state_data_sources"},
	},
}

// encryptionTargetSchema describes the inside of a state {} or plan {}
// block: a required `method` traversal plus an optional `fallback`.
var encryptionTargetSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "method"},
		{Name: "fallback"},
	},
}

// decodeEncryptionBlock decodes the OpenTofu terraform.encryption {} tree
// (key providers, methods, state/plan targets, and remote state sources)
// into module.Encryption. Provider/method bodies are schema-less, so each
// attribute is captured verbatim.
func decodeEncryptionBlock(block *hcl.Block, source []byte, module *model.Module) model.Diagnostics {
	inner, _, hdiag := block.Body.PartialContent(encryptionSchema)
	diags := model.DiagnosticsFromHCL(hdiag)

	encryption := &model.Encryption{Range: model.RangeFromHCL(block.DefRange)}

	for _, keyProvider := range inner.Blocks.OfType("key_provider") {
		body, bdiag := captureAttributeMap(keyProvider.Body, source)
		diags = append(diags, bdiag...)
		encryption.KeyProviders = append(encryption.KeyProviders, model.EncryptionKeyProvider{
			Type:  keyProvider.Labels[0],
			Name:  keyProvider.Labels[1],
			Body:  body,
			Range: model.RangeFromHCL(keyProvider.DefRange),
		})
	}

	for _, method := range inner.Blocks.OfType("method") {
		body, bdiag := captureAttributeMap(method.Body, source)
		diags = append(diags, bdiag...)
		encryption.Methods = append(encryption.Methods, model.EncryptionMethod{
			Type:  method.Labels[0],
			Name:  method.Labels[1],
			Body:  body,
			Range: model.RangeFromHCL(method.DefRange),
		})
	}

	for _, state := range inner.Blocks.OfType("state") {
		target, tdiag := decodeEncryptionTarget(state, source)
		diags = append(diags, tdiag...)
		encryption.State = target
	}

	for _, plan := range inner.Blocks.OfType("plan") {
		target, tdiag := decodeEncryptionTarget(plan, source)
		diags = append(diags, tdiag...)
		encryption.Plan = target
	}

	for _, remote := range inner.Blocks.OfType("remote_state_data_sources") {
		body, bdiag := captureAttributeMap(remote.Body, source)
		diags = append(diags, bdiag...)
		encryption.RemoteStateSources = append(encryption.RemoteStateSources, model.EncryptionRemoteState{
			Body:  body,
			Range: model.RangeFromHCL(remote.DefRange),
		})
	}

	module.Encryption = encryption
	return diags
}

// decodeEncryptionTarget decodes a state {} or plan {} block: the
// required `method` traversal plus an optional `fallback` block (whose
// shape mirrors the parent).
func decodeEncryptionTarget(block *hcl.Block, source []byte) (*model.EncryptionTarget, model.Diagnostics) {
	content, _, hdiag := block.Body.PartialContent(encryptionTargetSchema)
	diags := model.DiagnosticsFromHCL(hdiag)

	target := &model.EncryptionTarget{Range: model.RangeFromHCL(block.DefRange)}
	if attribute, ok := content.Attributes["method"]; ok {
		target.Method = capture(attribute.Expr, source)
	}
	if attribute, ok := content.Attributes["fallback"]; ok {
		target.Fallback = capture(attribute.Expr, source)
	}
	return target, diags
}
