/*
* # ACME-sls Terraform Module
*
* Creates all the resources you need to start creating and renewing Let's
* Encrypt certificates.
*
* This is a work in progress...
*/

# This function solves the TLS-ALPN-01 challenge
resource "aws_lambda_function" "challenge" {
  function_name = "AcmeSLSCertificateCreator"
  description   = "See https://github.com/sjauld/acme-sls/ for details"

  # @TODO add s3 shit
  filename = "../../bin/lambda-tls.zip"

  role = aws_iam_role.lambda.arn

  runtime = "go1.x"
  handler = "lambda-tls"

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
    sid = "ACM"

    actions = [
      "acm:DescribeCertificate",
      "acm:ImportCertificate",
      "acm:ListCertificates",
      "acm:ListTagsForCertificate",
    ]

    resources = ["*"]
  }
}

# API Gateway to respond to the TLS-ALPN-01 challenge
resource "aws_apigatewayv2_api" "challenge" {
  name          = "AcmeSLSCertificateChallenge"
  protocol_type = "WEBSOCKET"

  description = "See https://github.com/sjauld/acme-sls/ for details"

  disable_execute_api_endpoint = true

  tags = local.tags
}

resource "aws_apigatewayv2_integration" "challenge" {
  api_id = aws_apigatewayv2_api.challenge.id

  description = "See https://github.com/sjauld/acme-sls/ for details"

  integration_type = "MOCK"
}

resource "aws_apigatewayv2_stage" "challenge" {
  api_id      = aws_apigatewayv2_api.challenge.id
  name        = "challenge"
  auto_deploy = true

  default_route_settings {
    throttling_burst_limit = 100
    throttling_rate_limit  = 20
  }

  tags = local.tags
}

resource "aws_apigatewayv2_route" "challenge" {
  api_id    = aws_apigatewayv2_api.challenge.id
  route_key = "$default"
  target    = "integrations/${aws_apigatewayv2_integration.challenge.id}"
}

resource "tls_private_key" "ca" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "tls_self_signed_cert" "ca" {
  key_algorithm     = tls_private_key.ca.algorithm
  private_key_pem   = tls_private_key.ca.private_key_pem
  is_ca_certificate = true
  # This cert is never used for trusted purposes so it can live for a while
  validity_period_hours = 5 * 365 * 24
  allowed_uses          = ["cert_signing"]

  subject {
    common_name  = "acme.sls"
    organization = "ACME SLS"
  }
}

module "certificate" {
  source = "./modules/certificate"

  for_each = var.certificates

  id      = each.key
  domains = each.value

  ca_cert_pem        = tls_self_signed_cert.ca.cert_pem
  ca_key_algorithm   = tls_private_key.ca.algorithm
  ca_private_key_pem = tls_private_key.ca.private_key_pem

  api_id    = aws_apigatewayv2_api.challenge.id
  api_stage = aws_apigatewayv2_stage.challenge.name

  tags = local.tags
}
