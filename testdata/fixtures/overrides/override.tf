# Fixture: overrides - global overlay.
# Applies after main_override.tf; should change the default of var.region.

variable "region" {
  default = "eu-central-1"
}
