# Fixture: multi-module - root.
# Two local child module calls with explicit dependency to exercise the
# module graph builder.

variable "name_prefix" {
  type    = string
  default = "fixture"
}

module "network" {
  source = "./modules/network"
  name   = "${var.name_prefix}-net"
}

module "compute" {
  source       = "./modules/compute"
  name         = "${var.name_prefix}-app"
  network_id   = module.network.id
  network_cidr = module.network.cidr
}

output "compute_id" {
  value = module.compute.id
}
