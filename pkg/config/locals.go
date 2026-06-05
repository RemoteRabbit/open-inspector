// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/remoterabbit/open-inspector/pkg/model"
)

// decodeLocalsBlock reads every name = value binding in a locals {} block
// and appends them to module.Locals, capturing each value verbatim.
func decodeLocalsBlock(block *hcl.Block, source []byte, module *model.Module) model.Diagnostics {
	attributes, hdiag := block.Body.JustAttributes()
	diags := model.DiagnosticsFromHCL(hdiag)

	for name, attribute := range attributes {
		module.Locals = append(module.Locals, model.Local{
			Name:  name,
			Value: capture(attribute.Expr, source),
			Range: model.RangeFromHCL(attribute.Range),
		})
	}
	return diags
}
