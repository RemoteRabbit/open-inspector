# Fixture: tofu-extension  (OpenTofu only)
# .tofu and .tofu.json files are loaded by OpenTofu but ignored by
# Terraform. The loader must walk both extensions.

terraform {
  required_version = ">= 1.8.0"
}

variable "greeting" {
  type    = string
  default = "hello from .tofu"
}

output "greeting" {
  value = var.greeting
}
