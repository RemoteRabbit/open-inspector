// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import "github.com/hashicorp/hcl/v2"

var rootSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{Type: "terraform"},
		{Type: "provider", LabelNames: []string{"name"}},
		{Type: "variable", LabelNames: []string{"name"}},
		{Type: "output", LabelNames: []string{"name"}},
		{Type: "locals"},
		{Type: "resource", LabelNames: []string{"type", "name"}},
		{Type: "data", LabelNames: []string{"type", "name"}},
		{Type: "module", LabelNames: []string{"name"}},
		{Type: "moved"},
		{Type: "import"},
		{Type: "removed"},
		{Type: "check", LabelNames: []string{"name"}},
		{Type: "ephemeral", LabelNames: []string{"type", "name"}},
	},
}
