provider "aws" {
  region = var.eg_aws_region
  secret_key = var.aws_secret_key
  access_key = var.aws_access_key
}

terraform {
  backend "s3" {
    bucket         = "insights-v2-terraform-state-dev"
    key            = "terraform/connector-ecs-tasks/terraform.tfstate"
    region         = "us-east-2" # this cant be replaced with the variable
    encrypt        = true
    kms_key_id     = "alias/terraform-bucket-key"
  }
}

resource "aws_kms_key" "terraform-bucket-key" {
  description             = "This key is used to encrypt bucket objects"
  deletion_window_in_days = 10
  enable_key_rotation     = true
}

resource "aws_kms_alias" "key-alias" {
  name          = "alias/terraform-bucket-key"
  target_key_id = aws_kms_key.terraform-bucket-key.key_id
}

resource "aws_s3_bucket" "terraform-state" {
  bucket = "insights-v2-terraform-state-dev"

  tags = {
    Name        = "Insights V2 cache Dev"
    Environment = "dev"
  }
}

resource "aws_s3_bucket_acl" "terraform-state-acl" {
  bucket = aws_s3_bucket.terraform-state.id
  acl    = "private"
}

resource "aws_s3_bucket_server_side_encryption_configuration" "terraform-state-encryption-configuration" {
  bucket = aws_s3_bucket.terraform-state.id

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.terraform-bucket-key.arn
      sse_algorithm     = "aws:kms"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "block" {
  bucket = aws_s3_bucket.terraform-state.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

data "aws_iam_policy_document" "kms_use" {
  statement {
    sid = ""
    effect = "Allow"
    actions = [
      "kms:Encrypt",
      "kms:Decrypt",
      "kms:ReEncrypt*",
      "kms:GenerateDataKey*",
      "kms:DescribeKey",
    ]
    resources = [
      "arn:aws:kms:${var.eg_aws_region}:${var.eg_account_id}:key/f36a45d3-9bce-4f10-bedc-5a20c2ff807e"
    ]
  }
}

resource "aws_iam_policy" "kms_use" {
  name        = "kmsuse"
  description = "Policy allows using KMS keys"
  policy      = data.aws_iam_policy_document.kms_use.json
}

