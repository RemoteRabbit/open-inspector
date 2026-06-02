# Fixture: opentofu-encryption  (OpenTofu only)
# OpenTofu 1.7+ state/plan encryption block. Terraform will reject this.
# The loader should recognize the nested key_provider / method / state /
# plan blocks and surface them on the model.

terraform {
  required_version = ">= 1.7.0"

  encryption {
    key_provider "pbkdf2" "mykey" {
      passphrase = "this-is-only-a-fixture-do-not-use"
    }

    method "aes_gcm" "new_method" {
      keys = key_provider.pbkdf2.mykey
    }

    state {
      method = method.aes_gcm.new_method
    }

    plan {
      method = method.aes_gcm.new_method
    }
  }
}

resource "null_resource" "example" {}
