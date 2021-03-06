/*
* # ACME-sls Terraform Module
*
* Creates all the resources you need to start creating and renewing Let's Encrypt certificates.
*/

data "aws_region" "current" {}

locals {
  domains = distinct(flatten([for k, v in var.certificates : v]))
}

# This function solves the HTTP-01 challenge
resource "aws_lambda_function" "challenge" {
  function_name = "${var.namespace}AcmeSLSCertificateCreator"
  description   = "See https://github.com/sjauld/acme-sls/ for details"

  # If a zipfile is not provided, then we assume that we're deploying in N. Virginia.
  # This could probably be improved
  s3_bucket = var.lambda_zipfile == null ? "viostream-mgmt-build-artifacts-us-east-1" : null
  s3_key    = var.lambda_zipfile == null ? "acme-sls/lambda-http-s3.latest.zip" : null

  filename = var.lambda_zipfile

  role = aws_iam_role.lambda.arn

  runtime = "go1.x"
  handler = var.lambda_handler

  environment {
    variables = {
      "RENEWAL_WINDOW" = "${var.renewal_window_hours}h"
      "S3_DELAY"       = "${var.s3_delay_seconds}s"
      "S3_REGION"      = coalesce(var.aws_s3_region, data.aws_region.current.name)
      "USER_EMAIL"     = var.user_email
    }
  }

  # The certificate negotiation process could take a while, so give the lambda
  # 5 minutes to run.
  timeout     = 300
  memory_size = 128

  tags = local.tags
}

# Client lambda permissions
resource "aws_iam_role" "lambda" {
  name               = "${var.namespace}AcmeSLSCertificateCreator"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume.json

  tags = local.tags
}

data "aws_iam_policy_document" "lambda_assume" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_policy_attachment" "lambda" {
  name       = aws_iam_role.lambda.name
  roles      = [aws_iam_role.lambda.name]
  policy_arn = aws_iam_policy.lambda.arn
}

resource "aws_iam_policy" "lambda" {
  name   = "${var.namespace}AcmeSLSCertificateCreator"
  policy = data.aws_iam_policy_document.lambda.json
}

data "aws_iam_policy_document" "lambda" {
  statement {
    sid = "Logging"

    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = ["*"]
  }

  statement {
    sid = "S3Management"

    actions = [
      "s3:DeleteObject",
      "s3:PutObject",
      "s3:PutObjectAcl",
    ]

    resources = formatlist("arn:aws:s3:::%v/.well-known/acme-challenge/*", local.domains)
  }

  statement {
    sid = "ACM"

    actions = [
      "acm:AddTagsToCertificate",
      "acm:DescribeCertificate",
      "acm:ImportCertificate",
      "acm:ListCertificates",
      "acm:ListTagsForCertificate",
    ]

    resources = ["*"]
  }
}

locals {
  # Set up the schedule so it commences after some specified delay (to give DNS a good chance to propogate)
  first_run = timeadd(timestamp(), var.first_run_delay)
}

# Cloudwatch events - we just want to trigger the lambda once per day for each
# certificate so that it can check if a renewal is required
resource "aws_cloudwatch_event_rule" "challenge" {
  name        = "${var.namespace}ACME-SLS-schedule"
  description = "See https://github.com/sjauld/acme-sls/ for details"

  schedule_expression = "cron(${formatdate("mm", local.first_run)} ${formatdate("hh", local.first_run)} * * ? *)"

  lifecycle {
    # This will get updated every time terraform runs so we should ignore it
    ignore_changes = [schedule_expression]
  }
}

resource "aws_cloudwatch_event_target" "challenge" {
  for_each = var.certificates

  arn   = aws_lambda_function.challenge.arn
  rule  = aws_cloudwatch_event_rule.challenge.id
  input = jsonencode({ "detail" = { "id" = each.key, "domains" = each.value } })
}

resource "aws_lambda_permission" "challenge" {
  statement_id  = "CloudWatchExecution"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.challenge.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.challenge.arn
}

resource "aws_s3_bucket" "challenge" {
  count = length(local.domains) * (var.create_buckets ? 1 : 0)

  bucket = local.domains[count.index]

  versioning {
    enabled = true
  }

  # Support to replicate everything back to a master bucket, in case you are
  # doing some multi-domain cloudfront magic.acceleration_status
  dynamic "replication_configuration" {
    # Don't try to replicate to yourself
    for_each = var.replication_target_bucket_arn == "" ? [] : var.replication_target_bucket_arn == "arn:aws:s3:::${local.domains[count.index]}" ? [] : [1]

    content {
      role = var.replication_role_arn
      rules {
        id     = "acme"
        prefix = ".well-known"
        status = "Enabled"

        destination {
          bucket        = var.replication_target_bucket_arn
          storage_class = "STANDARD"
        }
      }
    }
  }

  # In case any challenges failed to clean up properly, this allows us to nuke the bucket
  force_destroy = true

  tags = local.tags
}
