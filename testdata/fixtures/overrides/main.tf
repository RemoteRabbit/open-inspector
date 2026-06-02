# Fixture: overrides - base file.
# See override.tf and main_override.tf for the overlays the loader must
# merge per Terraform's documented override semantics.

resource "null_resource" "configured" {
  triggers = {
    environment = "dev"
    owner       = "base"
  }
}

variable "region" {
  type    = string
  default = "us-east-1"
}
