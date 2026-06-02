variable "name" {
  type = string
}

resource "null_resource" "net" {
  triggers = {
    name = var.name
  }
}

output "id" {
  value = null_resource.net.id
}

output "cidr" {
  value = "10.0.0.0/16"
}
