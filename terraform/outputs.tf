output "bucket" {
  description = "The bucket, in case you want to serve some static content as well as using it for challenges"
  value       = aws_s3_bucket.challenge
}

output "cname_records" {
  description = "You should create CNAME records as follows"
  value       = [for _, v in aws_s3_bucket.challenge : {name = v.bucket, record = v.bucket_domain_name}]
}
