# Fixture: module-sources
# Declarations of every module source type the graph builder must
# recognize. These are NOT fetched by unit tests; integration tests opt
# in explicitly.

module "local_relative" {
  source = "../multi-module/modules/network"
  name   = "from-local"
}

module "registry" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.21"

  name = "open-inspector-fixture"
  cidr = "10.0.0.0/16"
}

module "git_https" {
  source = "git::https://example.com/example/network.git//modules/vpc?ref=v1.2.3"
}

module "git_ssh" {
  source = "git::ssh://git@example.com/example/network.git?ref=main"
}

module "github_shorthand" {
  source = "github.com/example/network//modules/vpc?ref=v1.2.3"
}

module "http_archive" {
  source = "https://example.com/modules/network-1.0.0.zip"
}
