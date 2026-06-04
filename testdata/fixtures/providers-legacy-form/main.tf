# Fixture: providers-legacy-form
# Pre-Terraform-0.13 shorthand: `name = "version constraint"` instead
# of the modern object form. The loader's decodeProviderReq must accept
# both.

terraform {
  required_providers {
    aws = "~> 4.67"
  }
}
