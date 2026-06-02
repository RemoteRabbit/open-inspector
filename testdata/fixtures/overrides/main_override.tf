# Fixture: overrides - file-specific overlay.
# Arguments listed here replace the matching arguments in main.tf.

resource "null_resource" "configured" {
  triggers = {
    owner = "override"
  }
}
