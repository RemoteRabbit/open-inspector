# Fixture: opentofu-provider-foreach  (OpenTofu only)
# OpenTofu 1.9+ multi-instance providers via for_each on the provider
# block. NOTE: confirm exact syntax against the OpenTofu 1.9 release
# notes before relying on this fixture in parser assertions.

terraform {
  required_version = ">= 1.9.0"
  required_providers {
    aws = {
      source                = "hashicorp/aws"
      version               = "~> 5.100"
      configuration_aliases = [aws.by_region]
    }
  }
}

variable "regions" {
  type    = set(string)
  default = ["us-east-1", "us-west-2", "eu-west-1"]
}

provider "aws" {
  alias    = "by_region"
  for_each = var.regions
  region   = each.key
}

resource "aws_s3_bucket" "regional" {
  for_each = var.regions
  provider = aws.by_region[each.key]
  bucket   = "open-inspector-fixture-${each.key}"
}
