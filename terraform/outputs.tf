output "buckets" {
  description = "You should cname your domains to  the bucket_domain_name for each bucket"
  value       = aws_s3_bucket.challenge
}
