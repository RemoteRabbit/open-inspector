// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package graph_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/remoterabbit/open-inspector/pkg/graph"
	"github.com/remoterabbit/open-inspector/pkg/inspector"
)

func TestRenderTree_MultiModule(t *testing.T) {
	module, _ := inspector.Inspect("../../testdata/fixtures/multi-module",
		inspector.WithModuleGraph())
	var buf bytes.Buffer
	if err := graph.RenderTree(&buf, module); err != nil {
		t.Fatalf("RenderTree: %v", err)
	}
	for _, want := range []string{"network", "compute"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("tree output missing %q:\n%s", want, buf.String())
		}
	}
}

func TestRenderDot_MultiModule(t *testing.T) {
	module, _ := inspector.Inspect("../../testdata/fixtures/multi-module",
		inspector.WithModuleGraph())
	var buf bytes.Buffer
	if err := graph.RenderDot(&buf, module); err != nil {
		t.Fatalf("RenderDot: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"digraph G {", "-> \"network\"", "-> \"compute\"", "}"} {
		if !strings.Contains(out, want) {
			t.Errorf("dot output missing %q:\n%s", want, out)
		}
	}
}

func TestRenderMermaid_MultiModule(t *testing.T) {
	module, _ := inspector.Inspect("../../testdata/fixtures/multi-module",
		inspector.WithModuleGraph())
	var buf bytes.Buffer
	if err := graph.RenderMermaid(&buf, module); err != nil {
		t.Fatalf("RenderMermaid: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"graph TD", "[network]", "[compute]", "-->"} {
		if !strings.Contains(out, want) {
			t.Errorf("mermaid output missing %q:\n%s", want, out)
		}
	}
}

func TestRenderJSON_MultiModule(t *testing.T) {
	module, _ := inspector.Inspect("../../testdata/fixtures/multi-module",
		inspector.WithModuleGraph())
	var buf bytes.Buffer
	if err := graph.RenderJSON(&buf, module); err != nil {
		t.Fatalf("RenderJSON: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("RenderJSON produced invalid JSON: %v\n%s", err, buf.String())
	}
	children, ok := decoded["children"].(map[string]any)
	if !ok {
		t.Fatalf("children missing or wrong type in JSON output:\n%s", buf.String())
	}
	for _, want := range []string{"network", "compute"} {
		if _, ok := children[want]; !ok {
			t.Errorf("JSON children missing %q:\n%s", want, buf.String())
		}
	}
}
