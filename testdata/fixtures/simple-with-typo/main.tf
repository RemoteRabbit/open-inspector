# Fixture: simple-with-typo
# Like the simple fixture, but the null_resource misspells `triggers` as
# `trigger_z`. Used to exercise schema enrichment's unknown-attribute path.

terraform {
  required_version = ">= 1.5.0"
  required_providers {
    null = {
      source  = "hashicorp/null"
      version = "~> 3.2"
    }
  }
}

resource "null_resource" "example" {
  trigger_z = {
    name = "oops"
  }
}
