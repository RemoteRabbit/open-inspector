# Fixture: simple
# The smallest plausibly-real module: one variable, one resource, one output.

terraform {
  required_version = ">= 1.15.4"
  required_providers {
    null = {
      source  = "hashicorp/null"
      version = "~> 3.2"
    }
  }
}

variable "name" {
  type        = string
  description = "Name applied as a trigger to the null resource."
}

resource "null_resource" "example" {
  triggers = {
    name = var.name
  }
}

output "id" {
  description = "The id assigned to the null_resource."
  value       = null_resource.example.id
}
