variable "tags" {
  type = map(string)
}

variable "certificates" {
  description = "A list of the certificates to be created/managed by ACME SLS"
  type        = map(list(string))
}

variable "renewal_window_days" {
  description = "The minimum number of days validity left on a certificate before it is renewed"
  default     = 7
  type        = string
}

variable "user_email" {
  description = "An email address to use for registering certificates with Let's Encrypt"
  default     = "dev@null.com"
  type        = string
}
