// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/remoterabbit/open-inspector/pkg/model"
)

// traversalString renders an hcl.Traversal back into its source-like
// string form, e.g. {TraverseRoot{aws}, TraverseAttr{east}} -> "aws.east"
// and aws_subnet.public[0] or aws.by_region["us-east-1"] with their keys.
func traversalString(traversal hcl.Traversal) string {
	var builder strings.Builder
	for index, step := range traversal {
		switch step := step.(type) {
		case hcl.TraverseRoot:
			builder.WriteString(step.Name)
		case hcl.TraverseAttr:
			if index > 0 {
				builder.WriteByte('.')
			}
			builder.WriteString(step.Name)
		case hcl.TraverseIndex:
			builder.WriteByte('[')
			builder.Write(hclwrite.TokensForValue(step.Key).Bytes())
			builder.WriteByte(']')
		}
	}
	return builder.String()
}

// addressString renders an address expression, such as a provider
// reference or a moved/import target, in its source-like form. Besides
// static traversals it accepts a trailing index whose key is itself a
// reference, as in aws.by_region[each.key] (OpenTofu provider for_each) or
// aws_s3_bucket.this[each.key] (import for_each).
func addressString(expression hcl.Expression) (string, hcl.Diagnostics) {
	traversal, diags := hcl.AbsTraversalForExpr(expression)
	if !diags.HasErrors() {
		return traversalString(traversal), nil
	}
	index, ok := expression.(*hclsyntax.IndexExpr)
	if !ok {
		return "", diags
	}
	collection, cdiags := hcl.AbsTraversalForExpr(index.Collection)
	key, kdiags := hcl.AbsTraversalForExpr(index.Key)
	if cdiags.HasErrors() || kdiags.HasErrors() {
		return "", diags
	}
	return traversalString(collection) + "[" + traversalString(key) + "]", nil
}

// decodeTraversalList decodes a list expression of references (such as a
// depends_on value) into their source-like string forms, skipping and
// reporting entries that are not simple traversals.
func decodeTraversalList(expressions hcl.Expression) ([]string, model.Diagnostics) {
	list, hdiag := hcl.ExprList(expressions)
	diags := model.DiagnosticsFromHCL(hdiag)
	output := make([]string, 0, len(list))

	for _, expression := range list {
		traverse, tdiag := hcl.AbsTraversalForExpr(expression)
		if tdiag.HasErrors() {
			diags = append(diags, model.DiagnosticsFromHCL(tdiag)...)
			continue
		}
		output = append(output, traversalString(traverse))
	}
	return output, diags
}

// traversalStringFromAttr renders a single attribute's traversal (e.g. a
// moved block's from/to) as a source-like string. A nil attribute yields
// an empty string and no diagnostics.
func traversalStringFromAttr(attribute *hcl.Attribute) (string, model.Diagnostics) {
	if attribute == nil {
		return "", nil
	}

	address, diags := addressString(attribute.Expr)
	return address, model.DiagnosticsFromHCL(diags)
}
