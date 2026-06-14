// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
)

// parseNativeBody parses src as native HCL and returns the first block's
// body, which is what the loader walks for a resource.
func parseNativeBody(t *testing.T, src string) (hcl.Body, []byte) {
	t.Helper()
	source := []byte(src)
	file, diags := hclsyntax.ParseConfig(source, "test.tf", hcl.InitialPos)
	if diags.HasErrors() {
		t.Fatalf("parse: %v", diags)
	}
	body := file.Body.(*hclsyntax.Body)
	if len(body.Blocks) != 1 {
		t.Fatalf("want exactly one block, got %d", len(body.Blocks))
	}
	return body.Blocks[0].Body, source
}

func TestDecodeBody_AttributesAndNestedBlocks(t *testing.T) {
	t.Parallel()

	body, source := parseNativeBody(t, `
resource "aws_s3_bucket" "b" {
  bucket = "x"
  versioning {
    enabled = true
  }
  dynamic "rule" {
    for_each = var.rules
    content {
      enabled = true
    }
  }
}
`)

	got := decodeBody(body, source)
	if got == nil {
		t.Fatal("decodeBody = nil, want a body")
	}

	if _, ok := got.Attributes["bucket"]; !ok {
		t.Errorf("missing top-level attribute bucket; got %v", got.Attributes)
	}
	if len(got.Blocks) != 2 {
		t.Fatalf("got %d nested blocks, want 2", len(got.Blocks))
	}

	versioning := got.Blocks[0]
	if versioning.Type != "versioning" {
		t.Errorf("Blocks[0].Type = %q, want versioning", versioning.Type)
	}
	if _, ok := versioning.Body.Attributes["enabled"]; !ok {
		t.Errorf("versioning block missing enabled attribute")
	}

	dynamic := got.Blocks[1]
	if dynamic.Type != "dynamic" {
		t.Errorf("Blocks[1].Type = %q, want dynamic", dynamic.Type)
	}
	if want := []string{"rule"}; len(dynamic.Labels) != 1 || dynamic.Labels[0] != want[0] {
		t.Errorf("dynamic.Labels = %v, want %v", dynamic.Labels, want)
	}
	if len(dynamic.Body.Blocks) != 1 || dynamic.Body.Blocks[0].Type != "content" {
		t.Errorf("dynamic block missing content sub-body; got %v", dynamic.Body.Blocks)
	}
}

func TestDecodeBodyFiltered_SkipsTopLevelMetaArgs(t *testing.T) {
	t.Parallel()

	body, source := parseNativeBody(t, `
resource "aws_instance" "web" {
  count      = 2
  depends_on = [aws_iam_role.r]
  ami        = "ami-123"
  lifecycle {
    prevent_destroy = true
  }
  ebs_block_device {
    volume_size = 10
  }
}
`)

	got := decodeBodyFiltered(body, source, resourceMetaArgs, resourceMetaBlocks)
	if got == nil {
		t.Fatal("decodeBodyFiltered = nil, want a body")
	}

	for _, skipped := range []string{"count", "depends_on"} {
		if _, ok := got.Attributes[skipped]; ok {
			t.Errorf("meta-arg %q leaked into NestedBody attributes", skipped)
		}
	}
	if _, ok := got.Attributes["ami"]; !ok {
		t.Errorf("real attribute ami was dropped; got %v", got.Attributes)
	}
	for _, block := range got.Blocks {
		if block.Type == "lifecycle" {
			t.Errorf("lifecycle block leaked into NestedBody blocks")
		}
	}
	if len(got.Blocks) != 1 || got.Blocks[0].Type != "ebs_block_device" {
		t.Errorf("Blocks = %v, want only ebs_block_device", got.Blocks)
	}
}

func TestDecodeBody_JSONBodyReturnsNil(t *testing.T) {
	t.Parallel()

	source := []byte(`{"resource": {"aws_s3_bucket": {"b": {"bucket": "x"}}}}`)
	file, diags := json.Parse(source, "test.tf.json")
	if diags.HasErrors() {
		t.Fatalf("parse: %v", diags)
	}
	if got := decodeBody(file.Body, source); got != nil {
		t.Errorf("decodeBody on JSON body = %v, want nil (JSON bodies are not walked)", got)
	}
}

func TestDecodeBody_EmptyBodyReturnsNil(t *testing.T) {
	t.Parallel()

	body, source := parseNativeBody(t, `
resource "null_resource" "empty" {
}
`)
	if got := decodeBody(body, source); got != nil {
		t.Errorf("decodeBody on empty body = %v, want nil", got)
	}
}
