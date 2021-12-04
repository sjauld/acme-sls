resource "aws_apigatewayv2_domain_name" "challenge" {
  count = length(var.domains)

  domain_name = var.domains[count.index]
  domain_name_configuration {
    certificate_arn = aws_acm_certificate.cert.arn
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }

  tags = local.tags
}

resource "aws_apigatewayv2_api_mapping" "challenge" {
  count = length(var.domains)

  api_id      = var.api_id
  domain_name = aws_apigatewayv2_domain_name.challenge[count.index].domain_name
  stage       = var.api_stage
}

# The configuration of this key needs to match the key used by the lambda to
# sign the cert so that ACM will allow us to renew
resource "tls_private_key" "cert" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "tls_cert_request" "cert" {
  key_algorithm   = tls_private_key.cert.algorithm
  private_key_pem = tls_private_key.cert.private_key_pem

  dns_names = var.domains

  subject {
    common_name  = var.domains[0]
    organization = "ACME SLS"
  }
}

resource "tls_locally_signed_cert" "cert" {
  cert_request_pem = tls_cert_request.cert.cert_request_pem

  ca_key_algorithm   = var.ca_key_algorithm
  ca_private_key_pem = var.ca_private_key_pem
  ca_cert_pem        = var.ca_cert_pem

  # We only need this to create the domain name, so validity not important
  validity_period_hours = 24
  # lego adds the key_encipherment usage to the challenge cert, so we need this
  # use to be here or ACM will reject the challenge import
  allowed_uses = ["key_encipherment"]
}

resource "aws_acm_certificate" "cert" {
  private_key       = tls_private_key.cert.private_key_pem
  certificate_body  = tls_locally_signed_cert.cert.cert_pem
  certificate_chain = var.ca_cert_pem

  lifecycle {
    create_before_destroy = true
  }
}
