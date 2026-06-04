# Fixture: ephemeral
# Ephemeral resources, variables, and outputs (Terraform 1.10+ /
# OpenTofu 1.10+). Values exist only during a single graph walk and are
# never persisted to state.

terraform {
  required_version = ">= 1.10.0"
  required_providers {
    random = {
      source  = "hashicorp/random"
      version = "~> 3.9"
    }
  }
}

variable "ephemeral_input" {
  type      = string
  ephemeral = true
}

ephemeral "random_password" "db" {
  length  = 24
  special = true
}

output "ephemeral_token" {
  value     = ephemeral.random_password.db.result
  ephemeral = true
  sensitive = true
}
