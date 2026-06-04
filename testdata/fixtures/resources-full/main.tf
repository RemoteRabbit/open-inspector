# Fixture: resources-full
# Exercises every resource/module meta-argument and lifecycle sub-feature
# the step-2 loader is expected to capture:
#
#   - resource provider = aws.east           (single-traversal meta-arg)
#   - resource depends_on = [...]            (traversal list)
#   - lifecycle.create_before_destroy        (bool pointer)
#   - lifecycle.prevent_destroy              (bool pointer)
#   - lifecycle.ignore_changes = [foo, bar]  (traversal list)
#   - lifecycle.ignore_changes = all         (bare keyword fallback)
#   - lifecycle.replace_triggered_by         (traversal list)
#   - lifecycle.precondition / postcondition (validation blocks)
#   - module providers = { aws = aws.east }  (provider map)
#   - module depends_on = [...]              (traversal list)

terraform {
  required_providers {
    aws = {
      source                = "hashicorp/aws"
      version               = "~> 5.100"
      configuration_aliases = [aws.east, aws.west]
    }
  }
}

provider "aws" {
  alias  = "east"
  region = "us-east-1"
}

resource "aws_instance" "primary" {
  provider = aws.east

  depends_on = [aws_security_group.web]

  lifecycle {
    create_before_destroy = true
    prevent_destroy       = false
    ignore_changes        = [tags, ami]
    replace_triggered_by  = [aws_security_group.web]

    precondition {
      condition     = length("ami-12345") > 0
      error_message = "ami must be set"
    }

    postcondition {
      condition     = self.id != ""
      error_message = "instance must have an id"
    }
  }
}

resource "aws_security_group" "web" {
  lifecycle {
    ignore_changes = all
  }
}

module "infra" {
  source     = "./child"
  depends_on = [aws_security_group.web]

  providers = {
    aws = aws.east
  }
}
