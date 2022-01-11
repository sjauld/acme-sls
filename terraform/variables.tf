variable "tags" {
  type    = map(string)
  default = {}
}

variable "aws_s3_region" {
  description = "Specify the region your buckets are in if it is different to the main region for this module"
  type        = string
  default     = ""
}

variable "certificates" {
  description = "A list of the certificates to be created/managed by ACME SLS"
  type        = map(list(string))
}

variable "create_buckets" {
  description = "Set this to false to BYO buckets"
  default     = true
  type        = bool
}

variable "first_run_delay" {
  description = "The delay between creating the terraform plan and firing the first lambda - increase this if you need more time to get DNS records in place"
  default     = "5m"
  type        = string
}

variable "lambda_handler" {
  description = "This should match the filename of the binary contained in your zip file (if you provide one)"
  default     = "lambda-http-s3"
}

variable "lambda_zipfile" {
  description = "Use this to feed in a zip of your own binary, otherwise we will use the public release"
  default     = null
  type        = string
}

variable "renewal_window_days" {
  description = "The minimum number of days validity left on a certificate before it is renewed"
  default     = 7
  type        = string
}

variable "user_email" {
  description = "An email address to use for registering certificates with Let's Encrypt - provide this if you want to get reminder emails when everything breaks"
  default     = "dev@null.com"
  type        = string
}
