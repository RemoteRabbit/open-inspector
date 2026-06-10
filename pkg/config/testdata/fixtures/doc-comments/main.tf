# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at https://mozilla.org/MPL/2.0/.

# The AWS region to deploy into.
# Defaults to us-east-1.
variable "region" {
  type    = string
  default = "us-east-1"
}

// Bucket for application logs.
resource "aws_s3_bucket" "logs" {
  bucket = "my-logs"
}

variable "no_comment" {        # trailing comment, not leading: must NOT attach
  type = string
}

# This comment has a blank line under it, so it must NOT attach.

variable "detached" {
  type = string
}

/*
 * Multi-line block comment
 * for the instance type.
 */
variable "instance_type" {
  type = string
}
