// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/remoterabbit/open-inspector/pkg/model"
)

// capture turns an hcl.Expression into a model.Expression by slicing the verbatim
// source bytes covered by the expression's range. The expression is never evaluated;
// downstream consumers can do that.
func capture(expression hcl.Expression, source []byte) model.Expression {
	rang := expression.Range()
	return model.Expression{
		Source: string(rang.SliceBytes(source)),
		Range:  model.RangeFromHcl(rang),
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
