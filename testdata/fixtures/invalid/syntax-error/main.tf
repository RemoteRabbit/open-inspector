# Fixture: invalid/syntax-error
# Deliberately broken HCL. The loader must return diagnostics with
# source ranges instead of panicking.

resource "null_resource" "broken" {
  triggers = {
    name = "missing closing brace"
}
