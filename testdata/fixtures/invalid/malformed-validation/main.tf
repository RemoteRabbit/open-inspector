# Fixture: invalid/malformed-validation
# A variable validation {} block missing both required attributes
# (condition, error_message) and a resource lifecycle precondition {}
# missing the same. The loader must report diagnostics without panicking.

variable "name" {
  type = string

  validation {
    # forgot both condition and error_message
  }
}

resource "null_resource" "checked" {
  lifecycle {
    precondition {
      # forgot both condition and error_message
    }
  }
}
