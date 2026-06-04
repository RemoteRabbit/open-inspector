# Fixture: modern-blocks
# Blocks the original terraform-config-inspect never learned about:
#   - moved   (TF 1.1+)
#   - import  (TF 1.5+)
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
      version = "~> 3.6"
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

import {
  to = null_resource.imported
  id = "static-id-123"
}

resource "null_resource" "imported" {}

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
