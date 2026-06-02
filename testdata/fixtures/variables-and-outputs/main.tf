# Fixture: variables-and-outputs
# Broad coverage of variable / output / local features the loader must
# extract: types, defaults, descriptions, sensitive, nullable, validation.

variable "region" {
  type        = string
  description = "AWS region."
  default     = "us-east-1"
  nullable    = false

  validation {
    condition     = can(regex("^[a-z]{2}-[a-z]+-[0-9]+$", var.region))
    error_message = "Region must look like us-east-1."
  }
}

variable "tags" {
  type        = map(string)
  description = "Tags applied to every resource."
  default     = {}
}

variable "instance_sizes" {
  type = list(object({
    name = string
    cpu  = number
    mem  = number
  }))
  default = []
}

variable "db_password" {
  type      = string
  sensitive = true
}

variable "feature_flags" {
  type = object({
    enable_logging = optional(bool, true)
    enable_metrics = optional(bool, false)
  })
  default = {}
}

locals {
  common_tags = merge(var.tags, {
    managed_by = "open-inspector-fixture"
  })

  region_short = substr(var.region, 0, 2)
}

output "region" {
  description = "Echoed region."
  value       = var.region
}

output "db_password" {
  description = "Re-emitted password (sensitive)."
  value       = var.db_password
  sensitive   = true
}

output "tag_count" {
  value = length(local.common_tags)
}
