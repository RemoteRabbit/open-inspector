resource "aws_s3_bucket" "b" {
  bucket = "x"
  tags   = { env = "prod" }

  versioning {
    enabled = true
  }

  logging {
    target_bucket = aws_s3_bucket.logs.id
    target_prefix = "log/"
  }

  dynamic "lifecycle_rule" {
    for_each = var.rules
    content {
      enabled = true
    }
  }
}
