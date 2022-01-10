package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"

	"github.com/sjauld/acme-sls/helpers"
	solver "github.com/sjauld/acme-sls/solver/http-s3"
)

const (
	acmeSLSTagName        = "ACME-SLS-Certificate-ID"
	fallbackDynamoDBTable = "acme-sls-certificates"
	fallbackEmail         = "dev@null.com"
	fallbackRenewalWindow = "7d"
	fallbackS3Region      = "us-east-1"
)

var (
	dynamoDBTable string
	renewalWindow time.Duration
	userEmail     string

	// AWS clients are instantiated during cold start
	acmClient *acm.ACM
	s3Client  *s3.S3
)

func init() {
	// the Let's Encrypt user email, DynamoDB table name and renewal window can be
	// set via an env, otherwise we'll use some sensible defaults
	var ok bool
	userEmail, ok = os.LookupEnv("USER_EMAIL")
	if !ok {
		userEmail = fallbackEmail
	}

	rwstr := os.Getenv("RENEWAL_WINDOW")
	// Make sure the env variable is a valid duration
	if _, err := time.ParseDuration(rwstr); err != nil {
		rwstr = fallbackRenewalWindow
	}
	renewalWindow, _ = time.ParseDuration(rwstr)

	s3Region, ok := os.LookupEnv("S3_REGION")
	if !ok {
		s3Region = fallbackS3Region
	}

	// Instantiate AWS clients
	sess := session.Must(session.NewSession())
	acmClient = acm.New(sess)
	s3Sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(s3Region),
	}))
	s3Client = s3.New(s3Sess)
}

// certificateRequest contains the data we'll send via the CloudWatch event trigger
type certificateRequest struct {
	ID      string   `json:"id"`      // Provide an ID so we can manage certificate rotation in ACM
	Domains []string `json:"domains"` // A list of domains to request on the certificate
}

func handler(ctx context.Context, event events.CloudWatchEvent) {
	log.Printf("[INFO] Processing event: %v", string(event.Detail))
	// Unmarshal the request
	var cr certificateRequest
	err := json.Unmarshal(event.Detail, &cr)
	if err != nil {
		log.Fatal(err)
	}

	if len(cr.Domains) == 0 {
		log.Fatal("You need to provide at least one domain!")
	}

	certARN, validity, err := helpers.CertificateDetails(acmClient, cr.Domains[0], cr.ID, acmeSLSTagName)
	if err != nil {
		log.Fatal(err)
	}
	if cr.ID != "" && validity > renewalWindow {
		log.Printf("[INFO] Exiting because certificate still has %v remaining", validity)
		return
	}

	// Create the let's encrypt client
	user, err := helpers.NewUser(userEmail)
	if err != nil {
		log.Fatal(err)
	}
	config := lego.NewConfig(user)
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

	// Set up the solver
	solver := solver.New(s3Client)

	client.Challenge.SetHTTP01Provider(solver)

	// Now let's start the certificate request process with Let's Encrypt
	request := certificate.ObtainRequest{
		Domains: cr.Domains,
		Bundle:  false,
	}
	log.Printf("[INFO] Requesting certificate: %v", request)
	cert, err := client.Certificate.Obtain(request)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("[INFO] Obtained certificate: %v", cert.CertURL)

	// Even though we specify Bundle: false, we seem to get a bundled certificate
	// so we need to unbundle it first
	myCert := helpers.CertFromChain(cert.Certificate)

	// And we'll persist the certificate to Amazon Certificate Manager
	req := &acm.ImportCertificateInput{
		Certificate:      myCert,
		CertificateChain: cert.IssuerCertificate,
		PrivateKey:       cert.PrivateKey,
		Tags: []*acm.Tag{
			{
				Key:   aws.String(acmeSLSTagName),
				Value: aws.String(cr.ID),
			},
		},
	}
	if certARN != "" {
		log.Printf("[INFO] Renewing ACM certificate %v", certARN)
		req.CertificateArn = aws.String(certARN)
	}

	resp, err := acmClient.ImportCertificate(req)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("[INFO] ACM created/renewed: %v", aws.StringValue(resp.CertificateArn))
}

func main() {
	lambda.Start(handler)
}
