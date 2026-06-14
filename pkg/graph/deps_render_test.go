// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package graph_test

import (
	"strings"
	"testing"

	"github.com/remoterabbit/open-inspector/pkg/graph"
	"github.com/remoterabbit/open-inspector/pkg/model"
)

// depModule returns a module carrying a small, pre-built dependency graph so
// the render tests do not depend on the loader.
func depModule() *model.Module {
	return &model.Module{
		Path: "root",
		DependencyGraph: &model.DependencyGraph{
			Nodes: []model.DependencyNode{
				{Address: "aws_instance.web", Kind: model.DependencyNodeResource},
				{Address: "local.name", Kind: model.DependencyNodeLocal},
				{Address: "var.env", Kind: model.DependencyNodeVariable},
			},
			Edges: []model.DependencyEdge{
				{From: "aws_instance.web", To: "local.name"},
				{From: "local.name", To: "var.env"},
			},
		},
	}
}

func renderToString(t *testing.T, fn func(*strings.Builder, *model.Module) error, module *model.Module) string {
	t.Helper()
	var sb strings.Builder
	if err := fn(&sb, module); err != nil {
		t.Fatalf("render: %v", err)
	}
	return sb.String()
}

func TestRenderDepsDot(t *testing.T) {
	out := renderToString(t, func(w *strings.Builder, m *model.Module) error {
		return graph.RenderDepsDot(w, m)
	}, depModule())

	for _, want := range []string{
		"digraph deps {",
		`"aws_instance.web" [shape=box];`,
		`"var.env" [shape=ellipse];`,
		`"aws_instance.web" -> "local.name";`,
		`"local.name" -> "var.env";`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("DOT output missing %q\n---\n%s", want, out)
		}
	}
	// A single module must NOT be wrapped in a cluster.
	if strings.Contains(out, "subgraph") {
		t.Errorf("single-module DOT should have no subgraph cluster\n---\n%s", out)
	}
}

func TestRenderDepsMermaid(t *testing.T) {
	out := renderToString(t, func(w *strings.Builder, m *model.Module) error {
		return graph.RenderDepsMermaid(w, m)
	}, depModule())

	if !strings.HasPrefix(out, "graph LR") {
		t.Errorf("Mermaid output should start with 'graph LR'\n---\n%s", out)
	}
	if !strings.Contains(out, `["aws_instance.web"]`) {
		t.Errorf("Mermaid output missing node label for aws_instance.web\n---\n%s", out)
	}
	if strings.Count(out, "-->") != 2 {
		t.Errorf("Mermaid output should have 2 edges, got:\n%s", out)
	}
}

func TestRenderDepsTree(t *testing.T) {
	out := renderToString(t, func(w *strings.Builder, m *model.Module) error {
		return graph.RenderDepsTree(w, m)
	}, depModule())

	for _, want := range []string{"aws_instance.web", "└── local.name", "local.name", "└── var.env"} {
		if !strings.Contains(out, want) {
			t.Errorf("tree output missing %q\n---\n%s", want, out)
		}
	}
}

func TestRenderDepsJSON(t *testing.T) {
	out := renderToString(t, func(w *strings.Builder, m *model.Module) error {
		return graph.RenderDepsJSON(w, m)
	}, depModule())

	for _, want := range []string{
		`"modules"`,
		`"address_prefix": ""`,
		`"address": "aws_instance.web"`,
		`"from": "local.name"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("JSON output missing %q\n---\n%s", want, out)
		}
	}
}

// TestRenderDeps_RecursiveNamespacing verifies child-module graphs are
// rendered under their Terraform-style address prefix and wrapped in
// clusters when more than one module is present.
func TestRenderDeps_RecursiveNamespacing(t *testing.T) {
	root := &model.Module{
		Path: "root",
		DependencyGraph: &model.DependencyGraph{
			Nodes: []model.DependencyNode{{Address: "module.child", Kind: model.DependencyNodeModule}},
		},
		Children: map[string]*model.ChildModule{
			"child": {CallName: "child", Module: &model.Module{
				Path: "child",
				DependencyGraph: &model.DependencyGraph{
					Nodes: []model.DependencyNode{
						{Address: "null_resource.x", Kind: model.DependencyNodeResource},
						{Address: "var.n", Kind: model.DependencyNodeVariable},
					},
					Edges: []model.DependencyEdge{{From: "null_resource.x", To: "var.n"}},
				},
			}},
		},
	}

	out := renderToString(t, func(w *strings.Builder, m *model.Module) error {
		return graph.RenderDepsDot(w, m)
	}, root)

	for _, want := range []string{
		`subgraph "cluster_0"`,
		`label="root";`,
		`label="module.child";`,
		`"module.child.null_resource.x" -> "module.child.var.n";`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("recursive DOT output missing %q\n---\n%s", want, out)
		}
	}
}

// TestRenderDeps_CrossModuleBridge verifies that a module call's input
// argument produces a dashed bridge edge from the child's input variable to
// the parent value that feeds it.
func TestRenderDeps_CrossModuleBridge(t *testing.T) {
	root := &model.Module{
		Path: "root",
		ModuleCalls: []model.ModuleCall{{
			Name: "child",
			Inputs: map[string]model.Expression{
				"n": {References: []model.Reference{
					{Kind: model.ReferenceResource, Address: "aws_vpc.main"},
				}},
			},
		}},
		DependencyGraph: &model.DependencyGraph{
			Nodes: []model.DependencyNode{
				{Address: "aws_vpc.main", Kind: model.DependencyNodeResource},
				{Address: "module.child", Kind: model.DependencyNodeModule},
			},
		},
		Children: map[string]*model.ChildModule{
			"child": {CallName: "child", Module: &model.Module{
				Path: "child",
				DependencyGraph: &model.DependencyGraph{
					Nodes: []model.DependencyNode{{Address: "var.n", Kind: model.DependencyNodeVariable}},
				},
			}},
		},
	}

	out := renderToString(t, func(w *strings.Builder, m *model.Module) error {
		return graph.RenderDepsDot(w, m)
	}, root)

	want := `"module.child.var.n" -> "aws_vpc.main" [style=dashed];`
	if !strings.Contains(out, want) {
		t.Errorf("missing cross-module bridge %q\n---\n%s", want, out)
	}
}

// TestRenderDeps_PreciseModuleOutput verifies that a parent edge reading a
// specific module output (recorded as ToAttribute) is rendered as a precise
// bridge to the child's output node, not the coarse module-call node.
func TestRenderDeps_PreciseModuleOutput(t *testing.T) {
	root := &model.Module{
		Path: "root",
		DependencyGraph: &model.DependencyGraph{
			Nodes: []model.DependencyNode{
				{Address: "aws_lb.x", Kind: model.DependencyNodeResource},
				{Address: "module.net", Kind: model.DependencyNodeModule},
			},
			Edges: []model.DependencyEdge{
				{From: "aws_lb.x", To: "module.net", ToAttribute: "subnet_id"},
			},
		},
		Children: map[string]*model.ChildModule{
			"net": {CallName: "net", Module: &model.Module{
				Path: "net",
				DependencyGraph: &model.DependencyGraph{
					Nodes: []model.DependencyNode{
						{Address: "output.subnet_id", Kind: model.DependencyNodeOutput},
					},
				},
			}},
		},
	}

	out := renderToString(t, func(w *strings.Builder, m *model.Module) error {
		return graph.RenderDepsDot(w, m)
	}, root)

	want := `"aws_lb.x" -> "module.net.output.subnet_id" [style=dashed];`
	if !strings.Contains(out, want) {
		t.Errorf("missing precise module-output bridge %q\n---\n%s", want, out)
	}
	// The coarse intra-cluster edge must NOT also be drawn.
	if strings.Contains(out, `"aws_lb.x" -> "module.net";`) {
		t.Errorf("coarse edge should have been promoted to a precise bridge\n---\n%s", out)
	}
}

// TestRenderDeps_InputBridgeToModuleOutput verifies that a module input fed by
// another module's output produces a precise bridge to that output node.
func TestRenderDeps_InputBridgeToModuleOutput(t *testing.T) {
	root := &model.Module{
		Path: "root",
		ModuleCalls: []model.ModuleCall{{
			Name: "b",
			Inputs: map[string]model.Expression{
				"v": {References: []model.Reference{
					{Kind: model.ReferenceModule, Address: "module.a", Attribute: "x"},
				}},
			},
		}},
		DependencyGraph: &model.DependencyGraph{
			Nodes: []model.DependencyNode{
				{Address: "module.a", Kind: model.DependencyNodeModule},
				{Address: "module.b", Kind: model.DependencyNodeModule},
			},
		},
		Children: map[string]*model.ChildModule{
			"a": {CallName: "a", Module: &model.Module{
				Path: "a",
				DependencyGraph: &model.DependencyGraph{
					Nodes: []model.DependencyNode{{Address: "output.x", Kind: model.DependencyNodeOutput}},
				},
			}},
			"b": {CallName: "b", Module: &model.Module{
				Path: "b",
				DependencyGraph: &model.DependencyGraph{
					Nodes: []model.DependencyNode{{Address: "var.v", Kind: model.DependencyNodeVariable}},
				},
			}},
		},
	}

	out := renderToString(t, func(w *strings.Builder, m *model.Module) error {
		return graph.RenderDepsDot(w, m)
	}, root)

	want := `"module.b.var.v" -> "module.a.output.x" [style=dashed];`
	if !strings.Contains(out, want) {
		t.Errorf("missing precise input bridge %q\n---\n%s", want, out)
	}
}

// TestRenderDeps_CoarseWhenChildAbsent keeps the coarse module-call target
// when the child module (and thus its output node) was not loaded.
func TestRenderDeps_CoarseWhenChildAbsent(t *testing.T) {
	root := &model.Module{
		Path: "root",
		DependencyGraph: &model.DependencyGraph{
			Nodes: []model.DependencyNode{
				{Address: "aws_lb.x", Kind: model.DependencyNodeResource},
				{Address: "module.net", Kind: model.DependencyNodeModule},
			},
			Edges: []model.DependencyEdge{
				{From: "aws_lb.x", To: "module.net", ToAttribute: "subnet_id"},
			},
		},
	}

	out := renderToString(t, func(w *strings.Builder, m *model.Module) error {
		return graph.RenderDepsDot(w, m)
	}, root)

	if !strings.Contains(out, `"aws_lb.x" -> "module.net";`) {
		t.Errorf("expected coarse edge aws_lb.x -> module.net\n---\n%s", out)
	}
}
