// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package graph

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/remoterabbit/open-inspector/pkg/model"
)

// scopedGraph pairs a module's dependency graph with the Terraform-style
// address prefix that namespaces it. The root module has an empty prefix;
// a child module call "network" contributes "module.network.", and nested
// calls accumulate ("module.network.module.subnet.").
type scopedGraph struct {
	prefix string
	graph  *model.DependencyGraph
}

// label returns a human-readable name for the scope: "root" for the root
// module, otherwise the prefix without its trailing dot.
func (s scopedGraph) label() string {
	if s.prefix == "" {
		return "root"
	}
	return s.prefix[:len(s.prefix)-1]
}

// collectScopedGraphs walks root and its loaded children, returning each
// module's dependency graph paired with its address prefix. Modules whose
// graph was not built (nil) are skipped. Children are visited in sorted
// call-name order for deterministic output.
func collectScopedGraphs(root *model.Module) []scopedGraph {
	var out []scopedGraph
	var walk func(prefix string, module *model.Module)
	walk = func(prefix string, module *model.Module) {
		if module == nil {
			return
		}
		if module.DependencyGraph != nil {
			out = append(out, scopedGraph{prefix: prefix, graph: module.DependencyGraph})
		}
		for _, name := range sortedKeys(module.Children) {
			child := module.Children[name]
			if child != nil {
				walk(prefix+"module."+name+".", child.Module)
			}
		}
	}
	walk("", root)
	return out
}

// globalNodeSet returns the set of fully-qualified (prefixed) node addresses
// across every scope, used to validate cross-module bridge endpoints.
func globalNodeSet(scopes []scopedGraph) map[string]struct{} {
	set := make(map[string]struct{})
	for _, scope := range scopes {
		for _, node := range scope.graph.Nodes {
			set[scope.prefix+node.Address] = struct{}{}
		}
	}
	return set
}

// collectBridges derives cross-module edges from module call inputs: a child
// module's input variable (module.<call>.var.<name>) depends on whatever the
// parent passed into it. These edges connect the otherwise self-contained
// per-module graphs into one graph that spans module boundaries. Targets are
// fully qualified and carry the referenced module output (if any) in
// ToAttribute, for later refinement.
func collectBridges(root *model.Module) []model.DependencyEdge {
	var bridges []model.DependencyEdge
	var walk func(prefix string, module *model.Module)
	walk = func(prefix string, module *model.Module) {
		if module == nil {
			return
		}
		for _, call := range module.ModuleCalls {
			child := module.Children[call.Name]
			if child == nil || child.Module == nil {
				continue
			}
			childPrefix := prefix + "module." + call.Name + "."
			names := make([]string, 0, len(call.Inputs))
			for name := range call.Inputs {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				from := childPrefix + "var." + name
				for _, d := range referenceDeps(call.Inputs[name].References) {
					bridges = append(bridges, model.DependencyEdge{
						From: from, To: prefix + d.address, ToAttribute: d.attribute,
					})
				}
			}
		}
		for _, name := range sortedKeys(module.Children) {
			walk(prefix+"module."+name+".", module.Children[name].Module)
		}
	}
	walk("", root)
	return bridges
}

// outputNode returns the fully-qualified address of a module call's output
// node: "module.net" + "subnet_id" -> "module.net.output.subnet_id".
func outputNode(moduleAddress, attribute string) string {
	return moduleAddress + ".output." + attribute
}

// refineTarget upgrades a coarse module-call target to its specific output
// node when the reference named one and that output exists in nodeSet (i.e.
// the child module was loaded). Otherwise it returns the coarse target.
func refineTarget(target, attribute string, nodeSet map[string]struct{}) string {
	if attribute == "" {
		return target
	}
	if precise := outputNode(target, attribute); contains(nodeSet, precise) {
		return precise
	}
	return target
}

func contains(set map[string]struct{}, key string) bool {
	_, ok := set[key]
	return ok
}

// refineBridges resolves each bridge's target to a precise module output when
// possible, keeps only bridges whose endpoints both exist in nodeSet (so they
// respect any kind filtering), de-duplicates, and sorts them.
func refineBridges(bridges []model.DependencyEdge, nodeSet map[string]struct{}) []model.DependencyEdge {
	seen := make(map[[2]string]struct{})
	var out []model.DependencyEdge
	for _, bridge := range bridges {
		to := refineTarget(bridge.To, bridge.ToAttribute, nodeSet)
		if !contains(nodeSet, bridge.From) || !contains(nodeSet, to) {
			continue
		}
		key := [2]string{bridge.From, to}
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, model.DependencyEdge{From: bridge.From, To: to})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].From != out[j].From {
			return out[i].From < out[j].From
		}
		return out[i].To < out[j].To
	})
	return out
}

// refineScopes rewrites each scope's edges for rendering. Edges whose target
// is a module output that exists in another scope (nodeSet) are promoted to
// cross-module bridges (returned separately); the rest stay as coarse,
// attribute-less intra-scope edges, de-duplicated. The model graphs are not
// mutated: copies are returned.
func refineScopes(scopes []scopedGraph, nodeSet map[string]struct{}) ([]scopedGraph, []model.DependencyEdge) {
	refined := make([]scopedGraph, 0, len(scopes))
	var promoted []model.DependencyEdge
	for _, scope := range scopes {
		seen := make(map[[2]string]struct{})
		edges := make([]model.DependencyEdge, 0, len(scope.graph.Edges))
		for _, edge := range scope.graph.Edges {
			if edge.ToAttribute != "" {
				precise := outputNode(scope.prefix+edge.To, edge.ToAttribute)
				if contains(nodeSet, precise) {
					promoted = append(promoted, model.DependencyEdge{
						From: scope.prefix + edge.From, To: precise,
					})
					continue
				}
			}
			key := [2]string{edge.From, edge.To}
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			edges = append(edges, model.DependencyEdge{From: edge.From, To: edge.To})
		}
		refined = append(refined, scopedGraph{
			prefix: scope.prefix,
			graph:  &model.DependencyGraph{Nodes: scope.graph.Nodes, Edges: edges},
		})
	}
	return refined, promoted
}

// depsView returns the per-module scoped graphs plus the cross-module bridge
// edges that connect them. Edges to specific module outputs are resolved to
// the child's output node and emitted as bridges; coarse intra-scope edges
// remain on the scoped graphs. All endpoints are validated against the visible
// node set.
func depsView(root *model.Module) ([]scopedGraph, []model.DependencyEdge) {
	scopes := collectScopedGraphs(root)
	nodeSet := globalNodeSet(scopes)
	refined, promoted := refineScopes(scopes, nodeSet)
	bridges := append(collectBridges(root), promoted...)
	return refined, refineBridges(bridges, nodeSet)
}

// adjacency groups a graph's edges by their From address, preserving the
// graph's sorted edge order.
func adjacency(graph *model.DependencyGraph) ([]string, map[string][]string) {
	order := make([]string, 0)
	byFrom := make(map[string][]string)
	for _, edge := range graph.Edges {
		if _, seen := byFrom[edge.From]; !seen {
			order = append(order, edge.From)
		}
		byFrom[edge.From] = append(byFrom[edge.From], edge.To)
	}
	return order, byFrom
}

// RenderDepsTree writes an indented view of the dependency graph(s) to w,
// grouped by module and then by dependent node, followed by any cross-module
// edges.
func RenderDepsTree(w io.Writer, root *model.Module) error {
	scopes, bridges := depsView(root)
	multi := len(scopes) > 1
	for _, scope := range scopes {
		indent := ""
		if multi {
			if _, err := fmt.Fprintf(w, "%s\n", scope.label()); err != nil {
				return err
			}
			indent = "  "
		}
		order, byFrom := adjacency(scope.graph)
		for _, from := range order {
			if _, err := fmt.Fprintf(w, "%s%s\n", indent, from); err != nil {
				return err
			}
			targets := byFrom[from]
			for index, to := range targets {
				connector := "├── "
				if index == len(targets)-1 {
					connector = "└── "
				}
				if _, err := fmt.Fprintf(w, "%s%s%s\n", indent, connector, to); err != nil {
					return err
				}
			}
		}
	}
	if len(bridges) > 0 {
		if _, err := fmt.Fprintln(w, "cross-module"); err != nil {
			return err
		}
		for _, bridge := range bridges {
			if _, err := fmt.Fprintf(w, "  %s --> %s\n", bridge.From, bridge.To); err != nil {
				return err
			}
		}
	}
	return nil
}

// dotShape maps a node kind to a Graphviz shape: declarations that create
// objects are boxes; named values are ellipses.
func dotShape(kind model.DependencyNodeKind) string {
	switch kind {
	case model.DependencyNodeResource, model.DependencyNodeData,
		model.DependencyNodeEphemeral, model.DependencyNodeModule:
		return "box"
	default:
		return "ellipse"
	}
}

// RenderDepsDot writes the dependency graph(s) to w in Graphviz DOT format.
// With more than one module, each module's graph is wrapped in its own
// labeled cluster, and cross-module bridges are drawn dashed between them.
func RenderDepsDot(w io.Writer, root *model.Module) error {
	scopes, bridges := depsView(root)
	multi := len(scopes) > 1
	if _, err := fmt.Fprintln(w, "digraph deps {"); err != nil {
		return err
	}
	for index, scope := range scopes {
		indent := "  "
		if multi {
			if _, err := fmt.Fprintf(w, "  subgraph %q {\n", "cluster_"+fmt.Sprint(index)); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, "    label=%q;\n", scope.label()); err != nil {
				return err
			}
			indent = "    "
		}
		for _, node := range scope.graph.Nodes {
			if _, err := fmt.Fprintf(w, "%s%q [shape=%s];\n", indent, scope.prefix+node.Address, dotShape(node.Kind)); err != nil {
				return err
			}
		}
		for _, edge := range scope.graph.Edges {
			if _, err := fmt.Fprintf(w, "%s%q -> %q;\n", indent, scope.prefix+edge.From, scope.prefix+edge.To); err != nil {
				return err
			}
		}
		if multi {
			if _, err := fmt.Fprintln(w, "  }"); err != nil {
				return err
			}
		}
	}
	for _, bridge := range bridges {
		if _, err := fmt.Fprintf(w, "  %q -> %q [style=dashed];\n", bridge.From, bridge.To); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(w, "}")
	return err
}

// depNodeID derives a Mermaid-safe identifier from a fully-qualified node
// address.
func depNodeID(qualified string) string {
	sum := sha256.Sum256([]byte(qualified))
	return "n" + hex.EncodeToString(sum[:4])
}

// RenderDepsMermaid writes the dependency graph(s) to w in Mermaid flowchart
// syntax. With more than one module, each graph is wrapped in its own
// subgraph, and cross-module bridges are drawn dotted between them.
func RenderDepsMermaid(w io.Writer, root *model.Module) error {
	scopes, bridges := depsView(root)
	multi := len(scopes) > 1
	if _, err := fmt.Fprintln(w, "graph LR"); err != nil {
		return err
	}
	for index, scope := range scopes {
		indent := "  "
		if multi {
			if _, err := fmt.Fprintf(w, "  subgraph %s[%q]\n", "s"+fmt.Sprint(index), scope.label()); err != nil {
				return err
			}
			indent = "    "
		}
		for _, node := range scope.graph.Nodes {
			id := depNodeID(scope.prefix + node.Address)
			if _, err := fmt.Fprintf(w, "%s%s[%q]\n", indent, id, node.Address); err != nil {
				return err
			}
		}
		for _, edge := range scope.graph.Edges {
			from := depNodeID(scope.prefix + edge.From)
			to := depNodeID(scope.prefix + edge.To)
			if _, err := fmt.Fprintf(w, "%s%s --> %s\n", indent, from, to); err != nil {
				return err
			}
		}
		if multi {
			if _, err := fmt.Fprintln(w, "  end"); err != nil {
				return err
			}
		}
	}
	for _, bridge := range bridges {
		from := depNodeID(bridge.From)
		to := depNodeID(bridge.To)
		if _, err := fmt.Fprintf(w, "  %s -.-> %s\n", from, to); err != nil {
			return err
		}
	}
	return nil
}

// depsJSONModule is one module's graph in the JSON output of the deps
// command.
type depsJSONModule struct {
	AddressPrefix   string                 `json:"address_prefix"`
	DependencyGraph *model.DependencyGraph `json:"dependency_graph"`
}

// depsJSONOutput is the top-level JSON shape: one entry per module plus the
// cross-module bridge edges, so the output shape is the same whether or not
// recursion was requested.
type depsJSONOutput struct {
	Modules []depsJSONModule       `json:"modules"`
	Bridges []model.DependencyEdge `json:"bridges,omitempty"`
}

// RenderDepsJSON writes the dependency graph(s) to w as indented JSON,
// emitting only the graphs and cross-module bridges (not the full module
// model).
func RenderDepsJSON(w io.Writer, root *model.Module) error {
	scopes, bridges := depsView(root)
	output := depsJSONOutput{Modules: make([]depsJSONModule, 0, len(scopes)), Bridges: bridges}
	for _, scope := range scopes {
		output.Modules = append(output.Modules, depsJSONModule{
			AddressPrefix:   scope.prefix,
			DependencyGraph: scope.graph,
		})
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}
