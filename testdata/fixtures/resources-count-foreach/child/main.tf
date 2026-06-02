variable "name" {
  type = string
}

resource "null_resource" "child" {
  triggers = {
    name = var.name
  }
}
