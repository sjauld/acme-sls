module github.com/sjauld/acme-sls

go 1.16

require (
	github.com/apex/gateway v1.1.2
	github.com/aws/aws-lambda-go v1.27.0
	github.com/aws/aws-sdk-go v1.39.0
	github.com/dustinkirkland/golang-petname v0.0.0-20191129215211-8e5a1ed0cff0
	github.com/gin-gonic/gin v1.9.0
	github.com/go-acme/lego/v4 v4.5.3
	github.com/gusaul/go-dynamock v0.0.0-20210107061312-3e989056e1e6
)

replace github.com/go-acme/lego/v4 => github.com/sjauld/lego/v4 v4.5.4
