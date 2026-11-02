# Fixture: modern-blocks
# Blocks the original terraform-config-inspect never learned about:
#   - moved   (TF 1.1+), including indexed addresses
#   - import  (TF 1.5+), including for_each targets (TF 1.7+)
#   - check   (TF 1.5+)
#   - removed (TF 1.7+)

terraform {
  required_providers {
    null = {
      source  = "hashicorp/null"
      version = "~> 3.2"
    }
    http = {
      source  = "hashicorp/http"
      version = "~> 3.4"
    }
  }
}

resource "null_resource" "renamed" {
  triggers = {
    note = "I used to be called 'old'."
  }
}

moved {
  from = null_resource.old
  to   = null_resource.renamed
}

# Index keys are part of the address.
moved {
  from = null_resource.keyed[0]
  to   = null_resource.keyed["first"]
}

resource "null_resource" "keyed" {
  for_each = toset(["first"])
}

import {
  to = null_resource.imported
  id = "static-id-123"
}

resource "null_resource" "imported" {}

# TF 1.7+ import for_each: the target index is a reference, not a constant.
import {
  for_each = toset(["a", "b"])
  to       = null_resource.imported_each[each.key]
  id       = each.key
}

resource "null_resource" "imported_each" {
  for_each = toset(["a", "b"])
}

removed {
  from = null_resource.gone
  lifecycle {
    destroy = false
  }
}

check "site_is_up" {
  data "http" "site" {
    url = "https://example.com"
  }

  assert {
    condition     = data.http.site.status_code == 200
    error_message = "example.com returned ${data.http.site.status_code}."
  }
}
