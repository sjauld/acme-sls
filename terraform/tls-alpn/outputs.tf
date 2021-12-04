output "challenge_fqdn" {
  description = "You should configure your environment so that all requests to http://<domain name>/.well-known/acme-challenge get directed here"
  value       = split("/", aws_apigatewayv2_stage.challenge.invoke_url)[2]
}

output "certificates" {
  value = module.certificate
}
