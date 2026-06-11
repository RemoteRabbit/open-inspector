// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"bytes"

	"github.com/hashicorp/hcl/v2"
	"github.com/remoterabbit/open-inspector/pkg/model"
)

// sliceSourceLF returns the verbatim source bytes covered by rang, with
// CRLF line endings collapsed to LF. Files checked out on Windows with
// the default core.autocrlf=true contain \r\n on disk, but we want the
// captured source text (variable types, expression bodies, etc.) to be
// byte-identical across platforms.
func sliceSourceLF(source []byte, rang hcl.Range) string {
	b := rang.SliceBytes(source)
	if bytes.IndexByte(b, '\r') < 0 {
		return string(b)
	}
	return string(bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n")))
}

// capture turns an hcl.Expression into a model.Expression by slicing the verbatim
// source bytes covered by the expression's range. The expression is never evaluated;
// downstream consumers can do that.
func capture(expression hcl.Expression, source []byte) model.Expression {
	rang := expression.Range()
	return model.Expression{
		Source:     sliceSourceLF(source, rang),
		Position:   model.PositionFromHCL(rang),
		References: extractReferences(expression),
	}
}

// captureAttributeMap reads every attribute on body as an unevaluated
// Expression keyed by attribute name. Useful for free-form bodies whose
// schema is provider/plugin defined (e.g. encryption key_provider /
// method blocks).
func captureAttributeMap(body hcl.Body, source []byte) (map[string]model.Expression, model.Diagnostics) {
	attributes, hdiag := body.JustAttributes()
	diags := model.DiagnosticsFromHCL(hdiag)
	if len(attributes) == 0 {
		return nil, diags
	}
	out := make(map[string]model.Expression, len(attributes))
	for name, attribute := range attributes {
		out[name] = capture(attribute.Expr, source)
	}
	return out, diags
}
