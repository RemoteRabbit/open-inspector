// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/remoterabbit/open-inspector/pkg/model"
)

// maxBodyDepth bounds recursion into nested blocks to guard against
// pathological inputs. Real Terraform/OpenTofu configs never approach this.
const maxBodyDepth = 32

// decodeBody walks a native HCL body generically (without a schema),
// capturing every attribute as an unevaluated Expression and recursing into
// every nested block. It returns nil for non-native (JSON) bodies, which
// cannot be walked without a schema, and nil for an empty body so callers
// can omit it.
func decodeBody(body hcl.Body, source []byte) *model.Body {
	return decodeBodyAt(body, source, 0, nil, nil)
}

// decodeBodyFiltered is like decodeBody but drops the named attributes and
// block types at the top level only; nested bodies keep everything. It is
// used for resource bodies, whose meta-arguments and lifecycle block are
// modeled explicitly elsewhere and must not be duplicated here.
func decodeBodyFiltered(body hcl.Body, source []byte, skipAttrs, skipBlocks map[string]struct{}) *model.Body {
	return decodeBodyAt(body, source, 0, skipAttrs, skipBlocks)
}

func decodeBodyAt(body hcl.Body, source []byte, depth int, skipAttrs, skipBlocks map[string]struct{}) *model.Body {
	native, ok := body.(*hclsyntax.Body)
	if !ok || depth > maxBodyDepth {
		return nil
	}

	out := &model.Body{}

	for name, attribute := range native.Attributes {
		if _, skip := skipAttrs[name]; skip {
			continue
		}
		if out.Attributes == nil {
			out.Attributes = make(map[string]model.Expression, len(native.Attributes))
		}
		out.Attributes[name] = capture(attribute.Expr, source)
	}

	for _, block := range native.Blocks {
		if _, skip := skipBlocks[block.Type]; skip {
			continue
		}
		out.Blocks = append(out.Blocks, model.NestedBlock{
			Type:     block.Type,
			Labels:   block.Labels,
			Body:     derefBody(decodeBodyAt(block.Body, source, depth+1, nil, nil)),
			Position: model.PositionFromHCL(block.DefRange()),
		})
	}

	if len(out.Attributes) == 0 && len(out.Blocks) == 0 {
		return nil
	}
	return out
}

// derefBody turns a possibly-nil *Body into a value Body so NestedBlock.Body
// is always present (an empty body for leaf blocks).
func derefBody(body *model.Body) model.Body {
	if body == nil {
		return model.Body{}
	}
	return *body
}
