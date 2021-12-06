package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/apex/gateway"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gin-gonic/gin"
	solver "github.com/sjauld/acme-sls/solver/http"
)

var store *solver.DynamoDBStore

func init() {
	table, ok := os.LookupEnv("DYNAMODB_TABLE_NAME")
	if !ok {
		log.Fatal("Please specify a DYNAMODB_TABLE_NAME env")
	}

	c := dynamodb.New(session.Must(session.NewSession()))

	// Setup the DynamoDB store
	store = solver.NewDynamoDBStore(c, table)
}

// based on https://github.com/appleboy/gin-lambda
func routerEngine() *gin.Engine {
	// set server mode for production
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	// Add middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/hc", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	r.GET("/.well-known/acme-challenge/:token", solver.NewGinHandlerFunc(store))

	return r
}

func main() {
	addr := fmt.Sprintf(":%s", os.Getenv("PORT"))
	log.Fatal(gateway.ListenAndServe(addr, routerEngine()))
}
