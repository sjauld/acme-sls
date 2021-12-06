package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"

	"github.com/sjauld/acme-sls/helpers"
	solver "github.com/sjauld/acme-sls/solver/http"
)

var domains = []string{"www.gin.com", "www.tonic.com"}

// Before testing, spin up the test environment with docker-compose up
func localPebbleClient(user *helpers.User) *lego.Client {
	// trust the Pebble root cert
	d, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	os.Setenv("SSL_CERT_FILE", filepath.Join(d, "pebble.minica.pem"))

	config := lego.NewConfig(user)
	config.CADirURL = "https://localhost:14000/dir"
	config.Certificate.KeyType = certcrypto.RSA2048

	client, err := lego.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	// register our user
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		log.Fatal(err)
	}

	user.SetRegistration(reg)

	return client
}

func testDynamodbClient() dynamodbiface.DynamoDBAPI {
	os.Setenv("AWS_ACCESS_KEY_ID", "AKIABLAHBLAH")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "12345")

	config := &aws.Config{
		Region:   aws.String("ap-southeast-2"),
		Endpoint: aws.String("http://localhost:8000"),
	}

	sess := session.Must(session.NewSession(config))

	svc := dynamodb.New(sess)

	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("token"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("token"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
		TableName: aws.String("challenges"),
	}

	_, err := svc.CreateTable(input)
	if err != nil {
		log.Println(err)
	}

	return svc
}

func main() {
	// test user
	user, err := helpers.NewUser("test@test.com")
	if err != nil {
		log.Fatal(err)
	}

	// test Pebble client
	client := localPebbleClient(user)
	store := solver.NewDynamoDBStore(testDynamodbClient(), "challenges")

	solver := solver.New(store)
	client.Challenge.SetHTTP01Provider(solver)

	request := certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true,
	}

	certificate, err := client.Certificate.Obtain(request)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Certificate:\n%+v", string(certificate.Certificate))
	log.Printf("Private key:\n%+v", string(certificate.PrivateKey))
}
