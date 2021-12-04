module github.com/sjauld/acme-sls

go 1.16

require (
	github.com/apex/gateway v1.1.2 // indirect
	github.com/aws/aws-lambda-go v1.27.0
	github.com/aws/aws-sdk-go v1.39.0
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/dustinkirkland/golang-petname v0.0.0-20191129215211-8e5a1ed0cff0 // indirect
	github.com/gin-gonic/gin v1.7.7
	github.com/go-acme/lego/v4 v4.5.3
	github.com/gusaul/go-dynamock v0.0.0-20210107061312-3e989056e1e6 // indirect
	golang.org/x/crypto v0.0.0-20211117183948-ae814b36b871 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
)

replace github.com/go-acme/lego/v4 => github.com/sjauld/lego/v4 v4.5.4
