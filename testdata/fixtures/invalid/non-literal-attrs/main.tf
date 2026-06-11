# Fixture: invalid/non-literal-attrs
# Attributes that the loader requires to be literals receive
# interpolations / references instead. Each one must surface a
# diagnostic explaining the failure, NOT silently disappear.

variable "flag_source" {
  type    = bool
  default = false
}

variable "description_source" {
  type    = string
  default = "from variable"
}

# description must be a literal string; an interpolation should diagnose.
output "bad_description" {
  value       = "hello"
  description = "from ${var.description_source}"
}

# sensitive must be a literal bool; an interpolation should diagnose.
output "bad_sensitive" {
  value     = "hello"
  sensitive = var.flag_source
}
