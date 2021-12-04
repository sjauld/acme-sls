# lambda-http-s3

This package is designed as a scheduled lambda, triggered by Cloudwatch Events.
It will kick off an HTTP-01 challenge with Let's Encrypt.

We rely on the fact that S3 will route HTTP traffic to a bucket name that
matches the HOST of a request, and so we are able to "prove" that we own a
domain name just by registering a matching bucket name and pointing our DNS
records at S3.
