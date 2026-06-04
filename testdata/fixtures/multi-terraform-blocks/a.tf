# Fixture: multi-terraform-blocks - file A
# First terraform {} block: declares one required_version and the aws
# provider. Combined with b.tf, the loader must aggregate both files
# into a single Module (RequiredCore concatenated, RequiredProviders
# merged by name).

terraform {
  required_version = ">= 1.15.4"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}
