// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

// DependencyNodeKind classifies a node in a DependencyGraph by the kind of
// declaration it represents.
type DependencyNodeKind string

// Dependency node kinds.
const (
	DependencyNodeResource  DependencyNodeKind = "resource"  // managed resource (TYPE.NAME)
	DependencyNodeData      DependencyNodeKind = "data"      // data source (data.TYPE.NAME)
	DependencyNodeEphemeral DependencyNodeKind = "ephemeral" // ephemeral resource (ephemeral.TYPE.NAME)
	DependencyNodeVariable  DependencyNodeKind = "variable"  // input variable (var.NAME)
	DependencyNodeLocal     DependencyNodeKind = "local"     // local value (local.NAME)
	DependencyNodeOutput    DependencyNodeKind = "output"    // output value (output.NAME)
	DependencyNodeModule    DependencyNodeKind = "module"    // module call (module.NAME)
)

// DependencyGraph is a directed, intra-module dependency graph of the
// declarations within a single module. An edge points from a declaration to
// each other declaration it references.
//
// It is derived statically from the references captured on expressions
// (including those inside nested blocks via Resource.NestedBody) plus
// explicit depends_on and replace_triggered_by entries; nothing is
// evaluated. Edges are kept only between declarations that exist in the same
// module: references to undeclared addresses (typos, cross-module values
// not modeled here) are dropped, so the graph is self-contained.
//
// It is populated only when inspection opts in (WithDependencyGraph);
// otherwise it is nil and omitted from JSON.
type DependencyGraph struct {
	Nodes []DependencyNode `json:"nodes,omitempty"` // declarations, sorted by address
	Edges []DependencyEdge `json:"edges,omitempty"` // dependencies, sorted by (from, to)
}

// DependencyNode is one declaration in a DependencyGraph.
type DependencyNode struct {
	Address string             `json:"address"` // canonical address, e.g. "aws_s3_bucket.b" or "var.region"
	Kind    DependencyNodeKind `json:"kind"`    // kind of declaration
}

// DependencyEdge is a directed "From depends on To" relationship between two
// declarations, identified by their canonical addresses.
type DependencyEdge struct {
	From string `json:"from"` // address of the dependent declaration
	To   string `json:"to"`   // address of the declaration it depends on
	// ToAttribute records which member of To the dependency is on, when that
	// is known and useful. It is set only when To is a module call and the
	// reference reads a specific output (the OUTPUT in "module.NAME.OUTPUT");
	// To stays "module.NAME" so the edge still points at a real node. Empty
	// otherwise.
	ToAttribute string `json:"to_attribute,omitempty"`
}
