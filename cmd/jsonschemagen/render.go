// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/invopop/jsonschema"
)

// renderMarkdown produces a human-readable description of the reflected
// schema: a table for the root object followed by one table per named
// type definition. It is intentionally deterministic so the committed
// docs/schema/v1.md can be diff-checked in CI.
func renderMarkdown(schema *jsonschema.Schema) string {
	var b strings.Builder
	b.WriteString("# open-inspector JSON output - schema v1\n\n")
	b.WriteString("Generated from `pkg/model.Module` by `make jsonschema`; do not edit by hand.\n\n")
	b.WriteString("This documents the object carried under the `module` key of the ")
	b.WriteString("`--json` envelope. The envelope wraps it with three scalar fields: ")
	b.WriteString("`schema_version` (int), `tool` (string), and `version` (string).\n\n")

	b.WriteString("## Module (root)\n\n")
	renderObject(&b, schema)

	names := make([]string, 0, len(schema.Definitions))
	for name := range schema.Definitions {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		fmt.Fprintf(&b, "## %s\n\n", name)
		renderObject(&b, schema.Definitions[name])
	}

	// renderObject ends every property table with a blank line as a
	// separator before the next heading; on the final type that leaves a
	// trailing blank line. Collapse it so the file ends with exactly one
	// newline and stays clean under the end-of-file pre-commit hook.
	return strings.TrimRight(b.String(), "\n") + "\n"
}

// renderObject writes the description and a property table for a single
// object schema. Objects with no properties (e.g. free-form maps) render
// just their description.
func renderObject(b *strings.Builder, schema *jsonschema.Schema) {
	if schema.Description != "" {
		b.WriteString(schema.Description)
		b.WriteString("\n\n")
	}

	if schema.Properties == nil || schema.Properties.Len() == 0 {
		return
	}

	required := make(map[string]bool, len(schema.Required))
	for _, name := range schema.Required {
		required[name] = true
	}

	b.WriteString("| Field | Type | Required | Description |\n")
	b.WriteString("| --- | --- | --- | --- |\n")
	for name, property := range schema.Properties.FromOldest() {
		req := "no"
		if required[name] {
			req = "yes"
		}
		fmt.Fprintf(b, "| `%s` | %s | %s | %s |\n",
			name, typeName(property), req, tableCell(property.Description))
	}
	b.WriteString("\n")
}

// typeName renders a compact, human-readable type for a property schema:
// scalars by name, references as a Markdown link to the type's section,
// and arrays/maps by their (linked) element type.
func typeName(schema *jsonschema.Schema) string {
	switch {
	case schema.Ref != "":
		name := refName(schema.Ref)
		return fmt.Sprintf("[%s](#%s)", name, anchor(name))
	case schema.Type == "array" && schema.Items != nil:
		// Escape the literal brackets so a following type link (e.g.
		// "[]" + "[Variable](#variable)") is not misparsed as a Markdown
		// reference link.
		return `\[\]` + typeName(schema.Items)
	case schema.Type == "object" && schema.AdditionalProperties != nil:
		return `map\[string\]` + typeName(schema.AdditionalProperties)
	case schema.Type != "":
		return schema.Type
	default:
		return "object"
	}
}

// anchor returns the heading anchor for a referenced type. Type names are
// single Go identifiers, so lowercasing matches the slug GitHub and mkdocs
// derive from "## TypeName". The root is rendered under "## Module (root)",
// so a reference to Module (e.g. ChildModule.module) points there instead.
func anchor(name string) string {
	if name == "Module" {
		return "module-root"
	}
	return strings.ToLower(name)
}

// tableCell makes a description safe to drop into a Markdown table cell:
// newlines collapse to spaces and pipes are escaped so they don't start a
// new column (model field comments like `"git" | "http"` contain them).
func tableCell(text string) string {
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "|", "\\|")
	return text
}

// refName strips the "#/$defs/" prefix from a JSON Schema $ref.
func refName(ref string) string {
	if index := strings.LastIndex(ref, "/"); index >= 0 {
		return ref[index+1:]
	}
	return ref
}
