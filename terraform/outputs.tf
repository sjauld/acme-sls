output "cname_records" {
  description = "You should create CNAME records as follows"
  value       = [for _, v in aws_s3_bucket.challenge : {name = v.bucket, record = v.bucket_domain_name}]
}
