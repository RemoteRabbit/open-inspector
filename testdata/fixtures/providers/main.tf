# Fixture: providers
# Multiple providers with version constraints, source addresses, aliases,
# and configuration_aliases (TF 0.15+) so a child module can declare
# required provider instances.

terraform {
  required_version = ">= 1.5.0"
  required_providers {
    aws = {
      source                = "hashicorp/aws"
      version               = "~> 5.0"
      configuration_aliases = [aws.east, aws.west]
    }
    random = {
      source  = "hashicorp/random"
      version = ">= 3.0, < 4.0"
    }
    http = {
      source  = "hashicorp/http"
      version = "~> 3.6"
    }
  }
}

provider "aws" {
  alias  = "east"
  region = "us-east-1"
}

provider "aws" {
  alias  = "west"
  region = "us-west-2"
}

provider "random" {}

data "http" "example" {
  url = "https://example.com"
}

resource "random_id" "example" {
  byte_length = 8
}
