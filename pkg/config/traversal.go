// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
)

// traversalString renders an hcl.Traversal back into its source-like
// string form, e.g. {TraverseRoot{aws}, TraverseAttr{east}} -> "aws.east".
func traversalString(trav hcl.Traversal) string {
	var b strings.Builder
	for i, step := range trav {
		switch s := step.(type) {
		case hcl.TraverseRoot:
			b.WriteString(s.Name)
		case hcl.TraverseAttr:
			if i > 0 {
				b.WriteByte('.')
			}
			b.WriteString(s.Name)
		}
	}
	return b.String()
}
