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

// sortedKeys returns the child call names in stable, sorted order.
func sortedKeys(children map[string]*model.ChildModule) []string {
	names := make([]string, 0, len(children))
	for name := range children {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// nodeID builds a deterministic, mermaid-safe identifier from a node's parent path and call name.
func nodeID(name, parentPath string) string {
	sum := sha256.Sum256([]byte(parentPath + "\x00" + name))
	return "n" + hex.EncodeToString(sum[:4])
}

// childSourceSuffix renders the parenthetical annotation after a child call
// name in the tree view: its resolved kind (when known) and the raw source
// argument, e.g. "  (local: ./modules/net)" or "  (registry: foo/bar/aws)".
// It returns "" when there is nothing to show.
func childSourceSuffix(child *model.ChildModule) string {
	switch {
	case child.Resolved != nil && child.Source != "":
		return "  (" + child.Resolved.Kind + ": " + child.Source + ")"
	case child.Resolved != nil:
		return "  (" + child.Resolved.Kind + ")"
	case child.Source != "":
		return "  (" + child.Source + ")"
	default:
		return ""
	}
}

// RenderTree writes an indented, ASCII tree view of the module graph to w.
func RenderTree(w io.Writer, mod *model.Module) error {
	var walk func(*model.Module, string) error
	walk = func(m *model.Module, prefix string) error {
		if _, err := fmt.Fprintf(w, "%s%s\n", prefix, m.Path); err != nil {
			return err
		}
		names := sortedKeys(m.Children)
		for i, name := range names {
			child := m.Children[name]
			connector := "├── "
			nextPrefix := prefix + "│   "
			if i == len(names)-1 {
				connector = "└── "
				nextPrefix = prefix + "    "
			}
			label := name + childSourceSuffix(child)
			if _, err := fmt.Fprintf(w, "%s%s%s\n", prefix, connector, label); err != nil {
				return err
			}
			if child.Module != nil {
				if err := walk(child.Module, nextPrefix); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return walk(mod, "")
}

// RenderDot writes the module graph to w in Graphviz DOT format.
func RenderDot(w io.Writer, mod *model.Module) error {
	if _, err := fmt.Fprintln(w, "digraph G {"); err != nil {
		return err
	}
	var walk func(*model.Module) error
	walk = func(m *model.Module) error {
		for _, name := range sortedKeys(m.Children) {
			child := m.Children[name]
			if _, err := fmt.Fprintf(w, "  %q -> %q;\n", m.Path, name); err != nil {
				return err
			}
			if child.Module != nil {
				if err := walk(child.Module); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := walk(mod); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w, "}")
	return err
}

// RenderMermaid writes the module graph to w in Mermaid flowchart syntax.
func RenderMermaid(w io.Writer, mod *model.Module) error {
	if _, err := fmt.Fprintln(w, "graph TD"); err != nil {
		return err
	}
	var walk func(string, *model.Module) error
	walk = func(parentID string, m *model.Module) error {
		for _, name := range sortedKeys(m.Children) {
			child := m.Children[name]
			childID := nodeID(name, m.Path)
			if _, err := fmt.Fprintf(w, "  %s --> %s[%s]\n", parentID, childID, name); err != nil {
				return err
			}
			if child.Module != nil {
				if err := walk(childID, child.Module); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return walk(nodeID("root", mod.Path), mod)
}

// RenderJSON writes the module graph to w as indented JSON.
func RenderJSON(w io.Writer, mod *model.Module) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(mod); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}
