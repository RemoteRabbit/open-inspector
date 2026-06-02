variable "name" {
  type = string
}

variable "network_id" {
  type = string
}

variable "network_cidr" {
  type = string
}

resource "null_resource" "app" {
  triggers = {
    name    = var.name
    network = var.network_id
    cidr    = var.network_cidr
  }
}

output "id" {
  value = null_resource.app.id
}
