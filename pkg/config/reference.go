// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/remoterabbit/open-inspector/pkg/model"
)

// model.References. It de-duplicates by address while preserving first-seen
// order, so `"${var.x}-${var.x}"` yields a single var.x reference.
func extractReferences(expression hcl.Expression) []model.Reference {
	traversals := expression.Variables()
	if len(traversals) == 0 {
		return nil
	}

	references := make([]model.Reference, 0, len(traversals))
	seen := make(map[string]struct{}, len(traversals))

	for _, traversal := range traversals {
		reference, ok := referenceFromTraversal(traversal)
		if !ok {
			continue
		}
		if _, dup := seen[reference.Address]; dup {
			continue
		}
		seen[reference.Address] = struct{}{}
		references = append(references, reference)
	}
	if len(references) == 0 {
		return nil
	}
	return references
}

// referenceFromTraversal classifies a single absolute traversal and builds
// its canonical address. It returns ok=false for traversals that carry no
// meaningful root name (should not happen for well-formed HCL).
func referenceFromTraversal(traversal hcl.Traversal) (model.Reference, bool) {
	if traversal.IsRelative() || len(traversal) == 0 {
		return model.Reference{}, false
	}

	root := traversal.RootName()
	reference := model.Reference{
		Position: model.PositionFromHCL(traversal.SourceRange()),
	}

	switch root {
	case "var":
		reference.Kind = model.ReferenceVar
		reference.Address = address(traversal, 2) // var.NAME
	case "local":
		reference.Kind = model.ReferenceLocal
		reference.Address = address(traversal, 2) // local.NAME
	case "module":
		reference.Kind = model.ReferenceModule
		reference.Address = address(traversal, 2) // module.NAME
	case "data":
		reference.Kind = model.ReferenceData
		reference.Address = address(traversal, 3) // data.TYPE.NAME
	case "each", "self", "count", "path", "terraform":
		reference.Kind = model.ReferenceOther
		reference.Address = address(traversal, 2)
	default:
		// Anything else with TYPE.NAME shape is a managed resource reference (aws_s3_bucket.this).
		// A bare root with no attribute is ambiguous; treat it as "other".
		if len(traversal) >= 2 {
			reference.Kind = model.ReferenceResource
			reference.Address = address(traversal, 2) // TYPE.NAME
		} else {
			reference.Kind = model.ReferenceOther
			reference.Address = root
		}
	}
	return reference, true
}

// address renders the first n steps of traversal as a dotted string using
// the existing traversalString helper, clamping n to the traversal length.
func address(traversal hcl.Traversal, n int) string {
	if n > len(traversal) {
		n = len(traversal)
	}
	return traversalString(traversal[:n])
}
