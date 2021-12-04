/*
* # ACME-sls Terraform Module
*
* Creates all the resources you need to start creating and renewing Let's
* Encrypt certificates.
*/

locals {
  domains = distinct(flatten([for k, v in var.certificates : v]))
}

# This function solves the HTTP-01 challenge
resource "aws_lambda_function" "challenge" {
  function_name = "AcmeSLSCertificateCreator"
  description   = "See https://github.com/sjauld/acme-sls/ for details"

  # @TODO add s3 shit
  filename = "../../bin/lambda-http-s3.zip"

  role = aws_iam_role.lambda.arn

  runtime = "go1.x"
  handler = "lambda-http-s3"

  environment {
    variables = {
      "RENEWAL_WINDOW" = "${var.renewal_window_days}d"
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
  name               = "AcmeSLSCertificateCreator"
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
  name   = "AcmeSLSCertificateCreator"
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

resource "aws_s3_bucket" "challenge" {
  count = length(local.domains)

  bucket = local.domains[count.index]

  force_destroy = true

  tags = local.tags
}
