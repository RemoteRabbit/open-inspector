// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package graph

import (
	"sort"

	"github.com/remoterabbit/open-inspector/pkg/model"
)

// BuildDependencies derives the intra-module dependency graph of module and
// stores it on module.DependencyGraph, in place. When module.Children is
// populated (WithModuleGraph), it recurses so every loaded child module gets
// its own graph too.
//
// Edges come from the references captured on expressions (including those
// nested inside resource bodies via NestedBody) plus explicit depends_on and
// replace_triggered_by entries. Only edges between declarations that exist in
// the same module are kept; references to undeclared addresses are dropped so
// the graph stays self-contained. Self-edges are dropped.
func BuildDependencies(module *model.Module) {
	if module == nil {
		return
	}
	module.DependencyGraph = buildDependencyGraph(module)
	for _, child := range module.Children {
		if child != nil && child.Module != nil {
			BuildDependencies(child.Module)
		}
	}
}

func buildDependencyGraph(module *model.Module) *model.DependencyGraph {
	nodes := collectNodes(module)
	if len(nodes) == 0 {
		return nil
	}

	known := make(map[string]struct{}, len(nodes))
	for _, node := range nodes {
		known[node.Address] = struct{}{}
	}

	type edgeKey struct{ from, to, attribute string }
	seen := make(map[edgeKey]struct{})
	var edges []model.DependencyEdge
	addEdge := func(from string, to dep) {
		if from == to.address {
			return
		}
		if _, ok := known[to.address]; !ok {
			return
		}
		key := edgeKey{from, to.address, to.attribute}
		if _, dup := seen[key]; dup {
			return
		}
		seen[key] = struct{}{}
		edges = append(edges, model.DependencyEdge{From: from, To: to.address, ToAttribute: to.attribute})
	}

	for _, resource := range module.ManagedResources {
		from := resourceAddress(resource.Type, resource.Name)
		for _, to := range dependenciesOfResource(&resource) {
			addEdge(from, to)
		}
	}
	for _, resource := range module.DataResources {
		from := dataAddress(resource.Type, resource.Name)
		for _, to := range dependenciesOfResource(&resource) {
			addEdge(from, to)
		}
	}
	for _, resource := range module.EphemeralResources {
		from := ephemeralAddress(resource.Type, resource.Name)
		for _, to := range dependenciesOfEphemeral(&resource) {
			addEdge(from, to)
		}
	}
	for _, local := range module.Locals {
		from := "local." + local.Name
		for _, to := range referenceDeps(local.Value.References) {
			addEdge(from, to)
		}
	}
	for _, output := range module.Outputs {
		from := "output." + output.Name
		for _, to := range referenceDeps(output.Value.References) {
			addEdge(from, to)
		}
		for _, to := range output.DependsOn {
			addEdge(from, dep{address: to})
		}
	}
	for _, call := range module.ModuleCalls {
		from := "module." + call.Name
		for _, to := range dependenciesOfModuleCall(&call) {
			addEdge(from, to)
		}
	}

	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From != edges[j].From {
			return edges[i].From < edges[j].From
		}
		if edges[i].To != edges[j].To {
			return edges[i].To < edges[j].To
		}
		return edges[i].ToAttribute < edges[j].ToAttribute
	})

	return &model.DependencyGraph{Nodes: nodes, Edges: edges}
}

// collectNodes returns every declaration in module as a graph node, sorted
// by address for deterministic output.
func collectNodes(module *model.Module) []model.DependencyNode {
	var nodes []model.DependencyNode
	for _, variable := range module.Variables {
		nodes = append(nodes, model.DependencyNode{Address: "var." + variable.Name, Kind: model.DependencyNodeVariable})
	}
	for _, local := range module.Locals {
		nodes = append(nodes, model.DependencyNode{Address: "local." + local.Name, Kind: model.DependencyNodeLocal})
	}
	for _, output := range module.Outputs {
		nodes = append(nodes, model.DependencyNode{Address: "output." + output.Name, Kind: model.DependencyNodeOutput})
	}
	for _, resource := range module.ManagedResources {
		nodes = append(nodes, model.DependencyNode{Address: resourceAddress(resource.Type, resource.Name), Kind: model.DependencyNodeResource})
	}
	for _, resource := range module.DataResources {
		nodes = append(nodes, model.DependencyNode{Address: dataAddress(resource.Type, resource.Name), Kind: model.DependencyNodeData})
	}
	for _, resource := range module.EphemeralResources {
		nodes = append(nodes, model.DependencyNode{Address: ephemeralAddress(resource.Type, resource.Name), Kind: model.DependencyNodeEphemeral})
	}
	for _, call := range module.ModuleCalls {
		nodes = append(nodes, model.DependencyNode{Address: "module." + call.Name, Kind: model.DependencyNodeModule})
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Address < nodes[j].Address })
	return nodes
}

// dep is one dependency target: the canonical address of the referenced
// declaration plus, for module references, the specific output read on it (so
// "module.net.subnet_id" becomes {address: "module.net", attribute:
// "subnet_id"}). attribute is empty for every other kind.
type dep struct {
	address   string
	attribute string
}

// dependenciesOfResource gathers every declaration a managed or data resource
// depends on: meta-argument and nested-body references, explicit depends_on,
// and lifecycle references.
func dependenciesOfResource(resource *model.Resource) []dep {
	var deps []dep
	deps = appendExprRefs(deps, resource.Count)
	deps = appendExprRefs(deps, resource.ForEach)
	deps = appendStringDeps(deps, resource.DependsOn)
	deps = appendBodyRefs(deps, resource.NestedBody)
	deps = appendLifecycleRefs(deps, resource.Lifecycle)
	return deps
}

// dependenciesOfEphemeral mirrors dependenciesOfResource for the
// EphemeralResource shape.
func dependenciesOfEphemeral(resource *model.EphemeralResource) []dep {
	var deps []dep
	deps = appendExprRefs(deps, resource.Count)
	deps = appendExprRefs(deps, resource.ForEach)
	deps = appendStringDeps(deps, resource.DependsOn)
	deps = appendBodyRefs(deps, resource.NestedBody)
	deps = appendLifecycleRefs(deps, resource.Lifecycle)
	return deps
}

// dependenciesOfModuleCall gathers the declarations a module call depends on
// from its meta-arguments, source/version expressions, and depends_on.
func dependenciesOfModuleCall(call *model.ModuleCall) []dep {
	var deps []dep
	deps = appendExprRefs(deps, call.Count)
	deps = appendExprRefs(deps, call.ForEach)
	deps = appendExprRefs(deps, call.SourceExpression)
	deps = appendExprRefs(deps, call.VersionExpression)
	deps = appendStringDeps(deps, call.DependsOn)
	for _, input := range call.Inputs {
		deps = append(deps, referenceDeps(input.References)...)
	}
	return deps
}

// appendLifecycleRefs appends the references a lifecycle block contributes:
// replace_triggered_by addresses and precondition/postcondition expression
// references. ignore_changes lists attribute names, not addresses, so it is
// skipped.
func appendLifecycleRefs(out []dep, lifecycle *model.Lifecycle) []dep {
	if lifecycle == nil {
		return out
	}
	out = appendStringDeps(out, lifecycle.ReplaceTriggeredBy)
	for _, condition := range lifecycle.Preconditions {
		out = appendExprRefs(out, &condition.Condition)
		out = appendExprRefs(out, &condition.ErrorMessage)
	}
	for _, condition := range lifecycle.Postconditions {
		out = appendExprRefs(out, &condition.Condition)
		out = appendExprRefs(out, &condition.ErrorMessage)
	}
	return out
}

// appendBodyRefs appends every dependency reference found in a captured body,
// recursing into nested blocks.
func appendBodyRefs(out []dep, body *model.Body) []dep {
	if body == nil {
		return out
	}
	for _, attribute := range body.Attributes {
		out = append(out, referenceDeps(attribute.References)...)
	}
	for index := range body.Blocks {
		out = appendBodyRefs(out, &body.Blocks[index].Body)
	}
	return out
}

// appendExprRefs appends the dependency references carried by expr.
func appendExprRefs(out []dep, expr *model.Expression) []dep {
	if expr == nil {
		return out
	}
	return append(out, referenceDeps(expr.References)...)
}

// appendStringDeps wraps plain address strings (depends_on,
// replace_triggered_by) as attribute-less deps.
func appendStringDeps(out []dep, addresses []string) []dep {
	for _, address := range addresses {
		out = append(out, dep{address: address})
	}
	return out
}

// referenceDeps returns the references whose kind can resolve to another
// declaration in the same module. Iteration/meta symbols (each, self, count,
// path, terraform) are excluded. Module references carry their output name in
// the dep's attribute.
func referenceDeps(references []model.Reference) []dep {
	var out []dep
	for _, reference := range references {
		switch reference.Kind {
		case model.ReferenceResource, model.ReferenceData, model.ReferenceEphemeral,
			model.ReferenceVar, model.ReferenceLocal:
			out = append(out, dep{address: reference.Address})
		case model.ReferenceModule:
			out = append(out, dep{address: reference.Address, attribute: reference.Attribute})
		}
	}
	return out
}

// FilterDependenciesByKind reduces every dependency graph in the module tree
// (root and its loaded children) to the induced subgraph of the given node
// kinds, in place. Nodes whose kind is not in kinds are removed, along with
// any edge that touched them. A nil or empty kinds set is a no-op.
//
// Because the filter drops whole nodes, paths that ran through a removed kind
// are broken rather than contracted: filtering to only "resource" nodes, for
// example, hides resource-to-resource links that flowed through a local or a
// variable.
func FilterDependenciesByKind(root *model.Module, kinds map[model.DependencyNodeKind]struct{}) {
	if root == nil || len(kinds) == 0 {
		return
	}
	root.DependencyGraph = filterGraphByKind(root.DependencyGraph, kinds)
	for _, child := range root.Children {
		if child != nil && child.Module != nil {
			FilterDependenciesByKind(child.Module, kinds)
		}
	}
}

func filterGraphByKind(graph *model.DependencyGraph, kinds map[model.DependencyNodeKind]struct{}) *model.DependencyGraph {
	if graph == nil {
		return nil
	}
	keep := make(map[string]struct{})
	var nodes []model.DependencyNode
	for _, node := range graph.Nodes {
		if _, ok := kinds[node.Kind]; ok {
			nodes = append(nodes, node)
			keep[node.Address] = struct{}{}
		}
	}
	var edges []model.DependencyEdge
	for _, edge := range graph.Edges {
		if _, ok := keep[edge.From]; !ok {
			continue
		}
		if _, ok := keep[edge.To]; !ok {
			continue
		}
		edges = append(edges, edge)
	}
	return &model.DependencyGraph{Nodes: nodes, Edges: edges}
}

// ContractDependenciesByKind reduces every dependency graph in the module tree
// to the given node kinds like FilterDependenciesByKind, but instead of
// breaking paths that ran through a removed node it contracts them: if a kept
// node reaches another kept node through a chain of removed nodes, a direct
// edge is added. So contracting to "resource" turns
// resource.a -> local.x -> resource.b into resource.a -> resource.b. A nil or
// empty kinds set is a no-op.
//
// Contraction is per module: chains that cross a module boundary (via a bridge)
// are not contracted. Contracted edges are coarse and carry no ToAttribute.
func ContractDependenciesByKind(root *model.Module, kinds map[model.DependencyNodeKind]struct{}) {
	if root == nil || len(kinds) == 0 {
		return
	}
	root.DependencyGraph = contractGraphByKind(root.DependencyGraph, kinds)
	for _, child := range root.Children {
		if child != nil && child.Module != nil {
			ContractDependenciesByKind(child.Module, kinds)
		}
	}
}

func contractGraphByKind(graph *model.DependencyGraph, kinds map[model.DependencyNodeKind]struct{}) *model.DependencyGraph {
	if graph == nil {
		return nil
	}

	keep := make(map[string]struct{})
	var nodes []model.DependencyNode
	for _, node := range graph.Nodes {
		if _, ok := kinds[node.Kind]; ok {
			nodes = append(nodes, node)
			keep[node.Address] = struct{}{}
		}
	}

	successors := make(map[string][]string)
	for _, edge := range graph.Edges {
		successors[edge.From] = append(successors[edge.From], edge.To)
	}

	type edgeKey struct{ from, to string }
	seen := make(map[edgeKey]struct{})
	var edges []model.DependencyEdge
	for _, node := range nodes {
		// Walk out from this kept node, passing through removed nodes, and
		// connect it directly to every kept node reached. visited guards
		// against cycles among the removed nodes traversed for this source.
		visited := make(map[string]struct{})
		var walk func(current string)
		walk = func(current string) {
			for _, next := range successors[current] {
				if _, kept := keep[next]; kept {
					key := edgeKey{node.Address, next}
					if node.Address != next {
						if _, dup := seen[key]; !dup {
							seen[key] = struct{}{}
							edges = append(edges, model.DependencyEdge{From: node.Address, To: next})
						}
					}
					continue // stop at kept nodes; do not expand past them
				}
				if _, been := visited[next]; been {
					continue
				}
				visited[next] = struct{}{}
				walk(next)
			}
		}
		walk(node.Address)
	}

	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From != edges[j].From {
			return edges[i].From < edges[j].From
		}
		return edges[i].To < edges[j].To
	})

	return &model.DependencyGraph{Nodes: nodes, Edges: edges}
}

func resourceAddress(typ, name string) string  { return typ + "." + name }
func dataAddress(typ, name string) string      { return "data." + typ + "." + name }
func ephemeralAddress(typ, name string) string { return "ephemeral." + typ + "." + name }
