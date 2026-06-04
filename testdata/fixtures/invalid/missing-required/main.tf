# Fixture: invalid/missing-required
# Syntactically valid, but the resource omits arguments its provider
# schema marks as required. Only meaningful once provider schema
# enrichment (step 6) is wired in; until then the config loader should
# still parse this cleanly.

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.46"
    }
  }
}

# aws_s3_bucket requires no arguments today but aws_instance does.
# ami and instance_type are both required by the provider schema.
resource "aws_instance" "missing_required" {
  tags = {
    Name = "missing-required-fixture"
  }
}
