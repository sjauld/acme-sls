package main

import (
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/gin-gonic/gin"

	solver "github.com/sjauld/acme-sls/solver/http"
)

func testDynamoDBClient() dynamodbiface.DynamoDBAPI {
	os.Setenv("AWS_ACCESS_KEY_ID", "AKIABLAHBLAH")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "12345")

	config := &aws.Config{
		Region:   aws.String("ap-southeast-2"),
		Endpoint: aws.String("http://dynamodb:8000"),
	}

	sess := session.Must(session.NewSession(config))

	log.Printf("[INFO] dynamoDB initialising")

	return dynamodb.New(sess)
}

func main() {
	r := gin.Default()

	r.GET("/hc", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	r.GET("/.well-known/acme-challenge/:token", solver.NewGinHandlerFunc(solver.NewDynamoDBStore(testDynamoDBClient(), "challenges")))

	// listen and serve on 0.0.0.0:5002
	if err := r.Run(":5002"); err != nil {
		log.Printf("error starting server %+v", err)
	}
}
