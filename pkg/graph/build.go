// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package graph recursively resolves and loads a module's calls into a tree
// rooted at the requested module, populating Module.Children in place.
package graph

import (
	"fmt"

	"github.com/remoterabbit/open-inspector/pkg/config"
	"github.com/remoterabbit/open-inspector/pkg/model"
	"github.com/remoterabbit/open-inspector/pkg/sources"
)

// Options is the graph-owned configuration, mapped from inspector.Option values by
// the caller so pkg/graph never imports pkg/inspector.
type Options struct {
	MaxDepth int
	CacheDir string
}

// Build resolves and recursively loads every module call reachable from root, populating root.Children
// in place. Per-call failures attach to the relevant ChildModule.Error and do not abort the walk.
func Build(root *model.Module, opts Options) {
	state := newLoadState()
	state.visited[root.Path] = true
	for _, call := range root.ModuleCalls {
		resolveAndLoad(root, call, state, opts)
	}
}

func resolveAndLoad(parent *model.Module, call model.ModuleCall, state *loadState, opts Options) {
	if state.depth >= opts.MaxDepth {
		addDiagnostic(parent, fmt.Sprintf("module call %q exceeds max-depth=%d", call.Name, opts.MaxDepth))
		return
	}
	resolved, err := sources.Resolve(call.Source, call.Version, parent.Path, opts.CacheDir)
	if err != nil {
		attachError(parent, call, err)
		return
	}
	if state.visited[resolved.CachePath] {
		addDiagnostic(parent, fmt.Sprintf("module call %q forms a cycle at %s", call.Name, resolved.CachePath))
		return
	}
	state.visited[resolved.CachePath] = true
	defer delete(state.visited, resolved.CachePath)
	state.depth++
	defer func() {
		state.depth--
	}()

	child, err := config.Load(resolved.CachePath)
	if err != nil {
		attachError(parent, call, err)
		return
	}
	attachChild(parent, call, resolved, child)
	for _, childCall := range child.ModuleCalls {
		resolveAndLoad(child, childCall, state, opts)
	}
}

// attachChild records a successfully resolved + local child under its call name.
func attachChild(parent *model.Module, call model.ModuleCall, resolved model.ResolvedSource, child *model.Module) {
	if parent.Children == nil {
		parent.Children = map[string]*model.ChildModule{}
	}
	resolvedCopy := resolved
	parent.Children[call.Name] = &model.ChildModule{
		CallName: call.Name,
		Source:   call.Source,
		Version:  call.Version,
		Resolved: &resolvedCopy,
		Module:   child,
	}
}

// attachError records a per-call failure on the child entry, leaving the parent's own Diagnostics untouched.
func attachError(parent *model.Module, call model.ModuleCall, err error) {
	if parent.Children == nil {
		parent.Children = map[string]*model.ChildModule{}
	}
	parent.Children[call.Name] = &model.ChildModule{
		CallName: call.Name,
		Source:   call.Source,
		Version:  call.Version,
		Error: &model.Diagnostic{
			Severity: model.SeverityError,
			Summary:  fmt.Sprintf("module call %q: %v", call.Name, err),
		},
	}
}

// addDiagnostic appends a graph-level warning (cycle or depth-limit) to the module's own Diagnostics.
func addDiagnostic(module *model.Module, summary string) {
	module.Diagnostics = append(module.Diagnostics, model.Diagnostic{
		Severity: model.SeverityWarning,
		Summary:  summary,
	})
}
