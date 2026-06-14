// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package graph_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/remoterabbit/open-inspector/pkg/config"
	"github.com/remoterabbit/open-inspector/pkg/graph"
	"github.com/remoterabbit/open-inspector/pkg/model"
)

// loadModule writes src as main.tf into a temp dir and loads it.
func loadModule(t *testing.T, src string) *model.Module {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.tf"), []byte(src), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	module, err := config.Load(dir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	return module
}

// hasEdge reports whether the graph contains a From -> To edge.
func hasEdge(dg *model.DependencyGraph, from, to string) bool {
	for _, edge := range dg.Edges {
		if edge.From == from && edge.To == to {
			return true
		}
	}
	return false
}

// TestBuildDependencies_NestedBlockEdge is the headline case: a dependency
// expressed inside a nested block (logging { target_bucket = ... }) must
// produce an edge. This is exactly what capturing nested blocks unlocks.
func TestBuildDependencies_NestedBlockEdge(t *testing.T) {
	module := loadModule(t, `
resource "aws_s3_bucket" "logs" {
  bucket = "logs"
}

resource "aws_s3_bucket" "b" {
  bucket = "b"
  logging {
    target_bucket = aws_s3_bucket.logs.id
  }
}
`)

	graph.BuildDependencies(module)

	dg := module.DependencyGraph
	if dg == nil {
		t.Fatal("DependencyGraph = nil, want a graph")
	}
	if !hasEdge(dg, "aws_s3_bucket.b", "aws_s3_bucket.logs") {
		t.Errorf("missing nested-block edge b -> logs; edges = %v", dg.Edges)
	}
}

// TestBuildDependencies_AllSources exercises references from locals, outputs,
// count/for_each, depends_on, and lifecycle, plus the self-contained rule.
func TestBuildDependencies_AllSources(t *testing.T) {
	module := loadModule(t, `
variable "name" {
  type = string
}

locals {
  prefix = "${var.name}-x"
}

resource "aws_instance" "a" {
  count = 2
  ami   = local.prefix
}

resource "aws_instance" "b" {
  ami        = "static"
  depends_on = [aws_instance.a]
  lifecycle {
    replace_triggered_by = [aws_instance.a]
  }
}

output "id" {
  value = aws_instance.b.id
}
`)

	graph.BuildDependencies(module)
	dg := module.DependencyGraph

	want := [][2]string{
		{"local.prefix", "var.name"},
		{"aws_instance.a", "local.prefix"},
		{"aws_instance.b", "aws_instance.a"}, // from depends_on and replace_triggered_by, de-duplicated
		{"output.id", "aws_instance.b"},
	}
	for _, edge := range want {
		if !hasEdge(dg, edge[0], edge[1]) {
			t.Errorf("missing edge %s -> %s; edges = %v", edge[0], edge[1], dg.Edges)
		}
	}

	// depends_on + replace_triggered_by reference the same target: exactly
	// one edge, not two.
	occurrences := 0
	for _, edge := range dg.Edges {
		if edge.From == "aws_instance.b" && edge.To == "aws_instance.a" {
			occurrences++
		}
	}
	if occurrences != 1 {
		t.Errorf("b -> a edge appears %d times, want 1 (de-duplicated)", occurrences)
	}
}

// TestBuildDependencies_SelfContained drops self-edges and edges to
// declarations that do not exist in the module.
func TestBuildDependencies_SelfContained(t *testing.T) {
	module := loadModule(t, `
resource "aws_instance" "a" {
  ami        = "x"
  depends_on = [aws_instance.does_not_exist]
}
`)

	graph.BuildDependencies(module)
	dg := module.DependencyGraph

	for _, edge := range dg.Edges {
		if edge.To == "aws_instance.does_not_exist" {
			t.Errorf("edge to undeclared address should be dropped; edges = %v", dg.Edges)
		}
		if edge.From == edge.To {
			t.Errorf("self-edge should be dropped: %v", edge)
		}
	}
}

// TestBuildDependencies_EmptyModule leaves the graph nil when there is
// nothing to graph.
func TestBuildDependencies_EmptyModule(t *testing.T) {
	module := &model.Module{Path: t.TempDir()}
	graph.BuildDependencies(module)
	if module.DependencyGraph != nil {
		t.Errorf("DependencyGraph = %v, want nil for an empty module", module.DependencyGraph)
	}
}

// TestBuildDependencies_ModuleInputEdges turns module input arguments into
// edges from the module call node to the parent values it consumes.
func TestBuildDependencies_ModuleInputEdges(t *testing.T) {
	module := loadModule(t, `
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}

module "net" {
  source = "./modules/net"
  vpc_id = aws_vpc.main.id
}
`)

	graph.BuildDependencies(module)
	dg := module.DependencyGraph

	if !hasEdge(dg, "module.net", "aws_vpc.main") {
		t.Errorf("missing module-input edge module.net -> aws_vpc.main; edges = %v", dg.Edges)
	}
}

// TestFilterDependenciesByKind keeps only the requested kinds and drops edges
// touching removed nodes.
func TestFilterDependenciesByKind(t *testing.T) {
	module := loadModule(t, `
variable "name" { type = string }
locals { prefix = var.name }
resource "aws_s3_bucket" "b" {
  bucket = local.prefix
}
`)
	graph.BuildDependencies(module)

	graph.FilterDependenciesByKind(module, map[model.DependencyNodeKind]struct{}{
		model.DependencyNodeResource: {},
		model.DependencyNodeLocal:    {},
	})
	dg := module.DependencyGraph

	for _, node := range dg.Nodes {
		if node.Kind == model.DependencyNodeVariable {
			t.Errorf("variable node survived the filter: %v", node)
		}
	}
	// resource -> local survives (both kinds kept)...
	if !hasEdge(dg, "aws_s3_bucket.b", "local.prefix") {
		t.Errorf("expected aws_s3_bucket.b -> local.prefix to survive; edges = %v", dg.Edges)
	}
	// ...but local -> var.name is gone (var filtered out).
	if hasEdge(dg, "local.prefix", "var.name") {
		t.Errorf("edge to filtered-out variable should be dropped; edges = %v", dg.Edges)
	}
}

// TestContractDependenciesByKind contracts paths through removed nodes instead
// of breaking them: a -> local -> var -> b style chain collapses to a direct
// resource-to-resource edge when only resources are kept.
func TestContractDependenciesByKind(t *testing.T) {
	module := loadModule(t, `
variable "name" { type = string }
locals { prefix = var.name }
resource "aws_s3_bucket" "logs" {
  bucket = local.prefix
}
resource "aws_s3_bucket" "b" {
  bucket = "b"
  logging {
    target_bucket = aws_s3_bucket.logs.id
  }
}
`)
	graph.BuildDependencies(module)

	graph.ContractDependenciesByKind(module, map[model.DependencyNodeKind]struct{}{
		model.DependencyNodeResource: {},
	})
	dg := module.DependencyGraph

	for _, node := range dg.Nodes {
		if node.Kind != model.DependencyNodeResource {
			t.Errorf("non-resource node survived contraction: %v", node)
		}
	}
	// Direct resource edge is kept.
	if !hasEdge(dg, "aws_s3_bucket.b", "aws_s3_bucket.logs") {
		t.Errorf("expected b -> logs to survive; edges = %v", dg.Edges)
	}
	// aws_s3_bucket.logs depended on aws_s3_bucket.b? No; but the contracted
	// chain logs -> local.prefix -> var.name has no resource at the far end,
	// so it contributes no edge. Confirm no edge points at a removed node.
	for _, edge := range dg.Edges {
		if edge.To == "local.prefix" || edge.To == "var.name" {
			t.Errorf("edge to removed node survived contraction: %v", edge)
		}
	}
}

// TestContractDependenciesByKind_AcrossRemoved collapses a chain that runs
// through several removed nodes into a single direct edge.
func TestContractDependenciesByKind_AcrossRemoved(t *testing.T) {
	module := loadModule(t, `
variable "name" { type = string }
locals { prefix = var.name }
resource "aws_instance" "a" {
  ami = local.prefix
}
`)
	graph.BuildDependencies(module)

	// Keep resources and variables, drop locals: a -> local.prefix -> var.name
	// must contract to a -> var.name.
	graph.ContractDependenciesByKind(module, map[model.DependencyNodeKind]struct{}{
		model.DependencyNodeResource: {},
		model.DependencyNodeVariable: {},
	})
	dg := module.DependencyGraph

	if !hasEdge(dg, "aws_instance.a", "var.name") {
		t.Errorf("expected contracted edge a -> var.name; edges = %v", dg.Edges)
	}
	for _, node := range dg.Nodes {
		if node.Kind == model.DependencyNodeLocal {
			t.Errorf("local node survived contraction: %v", node)
		}
	}
}

// TestBuildDependencies_EphemeralAndData exercises dependency collection for
// ephemeral and data resources: both their inbound edges (something reads them)
// and outbound edges (they read other declarations).
func TestBuildDependencies_EphemeralAndData(t *testing.T) {
	module := loadModule(t, `
variable "owner" { type = string }
variable "length" { type = number }

data "aws_ami" "ubuntu" {
  owners = [var.owner]
}

ephemeral "random_password" "db" {
  length = var.length
}

resource "aws_instance" "web" {
  ami      = data.aws_ami.ubuntu.id
  password = ephemeral.random_password.db.result
}
`)
	graph.BuildDependencies(module)
	dg := module.DependencyGraph

	want := [][2]string{
		{"data.aws_ami.ubuntu", "var.owner"},                 // data -> var
		{"ephemeral.random_password.db", "var.length"},       // ephemeral -> var
		{"aws_instance.web", "data.aws_ami.ubuntu"},          // resource -> data
		{"aws_instance.web", "ephemeral.random_password.db"}, // resource -> ephemeral
	}
	for _, edge := range want {
		if !hasEdge(dg, edge[0], edge[1]) {
			t.Errorf("missing edge %s -> %s; edges = %v", edge[0], edge[1], dg.Edges)
		}
	}
}

// TestBuildDependencies_ModuleOutputAttribute records the specific output read
// on a module call as the edge's ToAttribute (without changing the coarse To).
func TestBuildDependencies_ModuleOutputAttribute(t *testing.T) {
	module := loadModule(t, `
module "net" {
  source = "./modules/net"
}
resource "aws_lb" "x" {
  subnet_id = module.net.subnet_id
}
`)
	graph.BuildDependencies(module)
	dg := module.DependencyGraph

	found := false
	for _, edge := range dg.Edges {
		if edge.From == "aws_lb.x" && edge.To == "module.net" {
			found = true
			if edge.ToAttribute != "subnet_id" {
				t.Errorf("ToAttribute = %q, want subnet_id; edge = %v", edge.ToAttribute, edge)
			}
		}
	}
	if !found {
		t.Errorf("missing edge aws_lb.x -> module.net; edges = %v", dg.Edges)
	}
}

// TestBuildDependencies_RecursesIntoChildren builds a graph for each loaded
// child module, not just the root.
func TestBuildDependencies_RecursesIntoChildren(t *testing.T) {
	child := loadModule(t, `
variable "n" { type = string }
locals { v = var.n }
`)
	root := &model.Module{
		Path: t.TempDir(),
		Children: map[string]*model.ChildModule{
			"c": {CallName: "c", Module: child},
		},
	}

	graph.BuildDependencies(root)

	if child.DependencyGraph == nil {
		t.Fatal("child module DependencyGraph = nil, want a graph")
	}
	if !hasEdge(child.DependencyGraph, "local.v", "var.n") {
		t.Errorf("missing child edge local.v -> var.n; edges = %v", child.DependencyGraph.Edges)
	}
}
