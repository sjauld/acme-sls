package alpn

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/aws/aws-sdk-go/service/acm/acmiface"
	"github.com/go-acme/lego/v4/challenge/tlsalpn01"
)

// Solver implements lego's challenge.Provider
type Solver struct {
	acmClient acmiface.ACMAPI
	certARN   string
}

// New returns a pointer to a Solver, initialised with an ACM client and a target
// certificateARN to update
func New(client acmiface.ACMAPI, certARN string) *Solver {
	return &Solver{
		acmClient: client,
		certARN:   certARN,
	}
}

// Present creates a certificate and imports it to ACM over the top of the
// pre-existing challenge certificate.
func (s *Solver) Present(domain, token, keyAuth string) error {
	log.Printf("[INFO] Presenting domain: %v, token: %v, keyauth: %v", domain, token, keyAuth)

	// use lego's library to create a certificate from the challenge issued by the CA
	cert, err := tlsalpn01.ChallengeCert(domain, keyAuth)
	if err != nil {
		return err
	}

	// tlsalpn01 creates an RSA2048 private key
	privateKey, ok := cert.PrivateKey.(*rsa.PrivateKey)
	if !ok {
		return fmt.Errorf("Expected an RSA key but got %v", cert.PrivateKey)
	}

	// Encode the certificate and key into PEM format for ACM
	pemCert := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Certificate[0],
	})

	pemKey := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// import the new cert into ACM
	req := &acm.ImportCertificateInput{
		Certificate: pemCert,
		PrivateKey:  pemKey,
	}
	if s.certARN != "" {
		req.CertificateArn = aws.String(s.certARN)
	}
	_, err = s.acmClient.ImportCertificate(req)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("[INFO] successfully imported certificate; having a nap so it can propogate")
	time.Sleep(60 * time.Second)

	return nil
}

// CleanUp is a no-op
func (s *Solver) CleanUp(domain, token, keyAuth string) error {
	log.Printf("[INFO] CleaningUp domain: %v, token: %v, keyauth: %v", domain, token, keyAuth)

	return nil
}
