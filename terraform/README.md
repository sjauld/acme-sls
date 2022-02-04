# ACME-sls Terraform Module

Creates all the resources you need to start creating and renewing Let's Encrypt certificates.

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 0.13 |
| aws | ~> 3.66 |

## Providers

| Name | Version |
|------|---------|
| aws | ~> 3.66 |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| certificates | A list of the certificates to be created/managed by ACME SLS | `map(list(string))` | n/a | yes |
| aws\_s3\_region | Specify the region your buckets are in if it is different to the main region for this module | `string` | `""` | no |
| create\_buckets | Set this to false to BYO buckets | `bool` | `true` | no |
| first\_run\_delay | The delay between creating the terraform plan and firing the first lambda - increase this if you need more time to get DNS records in place | `string` | `"5m"` | no |
| lambda\_handler | This should match the filename of the binary contained in your zip file (if you provide one) | `string` | `"lambda-http-s3"` | no |
| lambda\_zipfile | Use this to feed in a zip of your own binary, otherwise we will use the public release | `string` | `null` | no |
| namespace | Use this if you have multiple ACME-SLS modules to avoid name clashes | `string` | `""` | no |
| renewal\_window\_days | The minimum number of days validity left on a certificate before it is renewed | `number` | `7` | no |
| replication\_role\_arn | An appropriate role if you need to replicate challenges | `string` | `""` | no |
| replication\_target\_bucket\_arn | Specify a master bucket that you'd like all challenges replicated to | `string` | `""` | no |
| s3\_delay\_seconds | Add a delay here if you are relying on S3 replication | `number` | `0` | no |
| tags | n/a | `map(string)` | `{}` | no |
| user\_email | An email address to use for registering certificates with Let's Encrypt - provide this if you want to get reminder emails when everything breaks | `string` | `"dev@null.com"` | no |

## Outputs

| Name | Description |
|------|-------------|
| cname\_records | You should create CNAME records as follows |

