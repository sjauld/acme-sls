package helpers

import (
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/aws/aws-sdk-go/service/acm/acmiface"
)

// CertificateDetails checks ACM for a certificate tagged with the ID
func CertificateDetails(acmClient acmiface.ACMAPI, domain, id, acmeSLSTagName string) (string, time.Duration, error) {
	var nextToken string
	for {
		in := &acm.ListCertificatesInput{}
		if nextToken != "" {
			in.NextToken = aws.String(nextToken)
		}

		resp, err := acmClient.ListCertificates(in)
		if err != nil {
			return "", 0, err
		}

		for _, cert := range resp.CertificateSummaryList {
			// Check that the domain matches
			if aws.StringValue(cert.DomainName) != domain {
				continue
			}
			// Now grab the tags and check if they match
			in := &acm.ListTagsForCertificateInput{
				CertificateArn: cert.CertificateArn,
			}
			tagResp, err := acmClient.ListTagsForCertificate(in)
			if err != nil {
				return "", 0, err
			}
			for _, tag := range tagResp.Tags {
				if aws.StringValue(tag.Key) != acmeSLSTagName {
					continue
				}
				if aws.StringValue(tag.Value) == id {
					// This is the right certificate! Now just need to retrieve the
					// validity
					return certificateValidity(acmClient, cert.CertificateArn)
				}
			}
		}

		// If we're on the last page of results, we didn't find a match
		if resp.NextToken == nil {
			return "", 0, nil
		}

		// Otherwise go to the next page
		nextToken = aws.StringValue(resp.NextToken)
	}
}

func certificateValidity(acmClient acmiface.ACMAPI, arn *string) (string, time.Duration, error) {
	resp, err := acmClient.DescribeCertificate(&acm.DescribeCertificateInput{
		CertificateArn: arn,
	})
	if err != nil {
		return "", 0, err
	}

	duration := resp.Certificate.NotAfter.Sub(time.Now())

	return aws.StringValue(arn), duration, nil
}

const endCertificate = "-----END CERTIFICATE-----"

// CertFromChain returns the first certificate from a bundled chain
func CertFromChain(chain []byte) []byte {
	certStr := string(chain)
	pos := strings.Index(certStr, endCertificate)
	return chain[:pos+len(endCertificate)]
}
