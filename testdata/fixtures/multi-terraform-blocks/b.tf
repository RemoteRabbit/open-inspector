# Fixture: multi-terraform-blocks - file B
# Second terraform {} block in a separate file. Adds another
# required_version constraint and a different provider.

terraform {
  required_version = "< 2.0.0"

  required_providers {
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }
}
