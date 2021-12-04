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
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"github.com/sjauld/acme-sls/helpers"
	alpn "github.com/sjauld/acme-sls/solver/acm-tls-alpn"
)

const (
	acmeSLSTagName        = "ACME-SLS-Certificate-ID"
	fallbackDynamoDBTable = "acme-sls-certificates"
	fallbackEmail         = "dev@null.com"
	fallbackRenewalWindow = "7d"
)

var (
	renewalWindow time.Duration
	userEmail     string

	// AWS clients are instantiated during cold start
	acmClient *acm.ACM
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

	// Instantiate AWS clients
	sess := session.Must(session.NewSession())
	acmClient = acm.New(sess)
}

// certificateRequest contains the data we'll send via the CloudWatch event trigger
type certificateRequest struct {
	ID               string   `json:"id"`                      // Provide an ID for the certificate you're creating so we can manage certificate rotation in ACM
	ChallengeCertARN string   `json:"challengeCertificateARN"` // The ARN of a challenge certificate that is associated with the custom domains in API Gateway
	Domains          []string `json:"domains"`                 // A list of domains to request on the certificate - each of these needs a custom domain in API Gateway
}

func handler(ctx context.Context, event events.CloudWatchEvent) {
	log.Printf("[INFO] Processing certificate request: %v", string(event.Detail))
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
	legoClient, err := lego.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	// register our user
	reg, err := legoClient.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		log.Fatal(err)
	}
	user.SetRegistration(reg)

	// Set up the solver
	solver := alpn.New(acmClient, cr.ChallengeCertARN)
	legoClient.Challenge.SetTLSALPN01Provider(solver)

	// Now let's start the certificate request process with Let's Encrypt
	request := certificate.ObtainRequest{
		Domains: cr.Domains,
		Bundle:  true,
	}
	log.Printf("[INFO] Requesting certificate for: %v", cr.Domains)
	cert, err := legoClient.Certificate.Obtain(request)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("[INFO] Obtained certificate: %v", cert.CertURL)

	// And we'll persist the certificate to Amazon Certificate Manager
	req := &acm.ImportCertificateInput{
		Certificate: cert.Certificate,
		PrivateKey:  cert.PrivateKey,
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

	log.Printf("[INFO] ACM created/renewed: %v", resp.CertificateArn)
}

func main() {
	lambda.Start(handler)
}
