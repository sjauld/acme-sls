variable "tags" {
  type = map(string)
}

variable "api_id" {
  type = string
}

variable "api_stage" {
  type = string
}

variable "ca_cert_pem" {
  type = string
}

variable "ca_key_algorithm" {
  type = string
}

variable "ca_private_key_pem" {
  type = string
}

variable "id" {
  type = string
}

variable "domains" {
  type = list(string)
}
