// package s3 solves the ACMEv2 HTTP-01 challenge. The workflow is as follows:
//
// 1. client requests a certificate from the remote CA, using the Solver as the HTTP-01 challenge
// 2. Solver populates the Challenge in S3 and notifies the CA that the challenge is ready
// 3. remote CA requests the keyauth from the well known path in S3
// 4. s3 presents the keyauth to the remote CA
//
// In order for s3 to route to your bucket using http, the bucket name will need to match
// the domain name for which you are creating a certificate
package s3

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

// Solver implements lego's challenge.Provider
type Solver struct {
	s3Client s3iface.S3API
	delay    time.Duration
}

// New returns a pointer to a Solver, initialised with an s3 client
func New(client s3iface.S3API) *Solver {
	return &Solver{
		s3Client: client,
	}
}

// WithDelay allows you to introduce a delay between uploading to S3 and
// continuing with the validation process. This is helpful if you are doing
// something with S3 replication, for example
func (s *Solver) WithDelay(t time.Duration) *Solver {
	s.delay = t
	return s
}

// Present writes the challenge information into S3 so that we
// can respond to HTTP queries with the correct value
func (s *Solver) Present(domain, token, keyAuth string) error {
	log.Printf("[INFO] Presenting domain: %v, token: %v, keyauth: %v", domain, token, keyAuth)

	in := &s3.PutObjectInput{
		Bucket: aws.String(domain),
		ACL:    aws.String(s3.ObjectCannedACLPublicRead),
		Body:   strings.NewReader(keyAuth),
		Key:    keyFromToken(token),
	}

	_, err := s.s3Client.PutObject(in)
	if err != nil {
		return err
	}

	time.Sleep(s.delay)
	return nil
}

// CleanUp removes the challenge information from S3
func (s *Solver) CleanUp(domain, token, keyAuth string) error {
	log.Printf("[INFO] CleaningUp domain: %v, token: %v, keyauth: %v", domain, token, keyAuth)

	in := &s3.DeleteObjectInput{
		Bucket: aws.String(domain),
		Key:    keyFromToken(token),
	}

	_, err := s.s3Client.DeleteObject(in)
	return err
}

// The request from ACME will be to http://<domain>/.well-known/acme-challenge/<token>
func keyFromToken(t string) *string {
	k := fmt.Sprintf(".well-known/acme-challenge/%s", t)
	return &k
}
