# Fixture: resources-count-foreach
# Resources, data sources, and module calls using both count and for_each
# meta-arguments. The loader should capture the expression (not evaluate
# it) and surface its source range.

variable "names" {
  type    = set(string)
  default = ["alpha", "beta", "gamma"]
}

variable "replica_count" {
  type    = number
  default = 3
}

resource "null_resource" "by_count" {
  count = var.replica_count

  triggers = {
    index = tostring(count.index)
  }
}

resource "null_resource" "by_for_each" {
  for_each = var.names

  triggers = {
    name = each.value
  }
}

data "null_data_source" "by_for_each" {
  for_each = toset(["one", "two"])

  inputs = {
    key = each.key
  }
}

module "fan_out" {
  source   = "./child"
  for_each = var.names
  name     = each.value
}

module "replicas" {
  source = "./child"
  count  = var.replica_count
  name   = "replica-${count.index}"
}
