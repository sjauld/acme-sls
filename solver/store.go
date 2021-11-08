package solver

import (
	"errors"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

var (
	ErrStoreRateLimited = errors.New("We were rate limited, try again later")
	ErrStoreNotFound    = errors.New("Challenge not found in the store")
)

// Store represents a backend storage system that is used to persist challenge
// information between the client and server
type Store interface {
	DeleteChallenge(string) error
	GetChallenge(string) (*Challenge, error)
	PutChallenge(*Challenge) error
}

// Challenge represents the information required for an ACMEv2 HTTP-01 challenge
type Challenge struct {
	Domain  string
	Token   string
	KeyAuth string
}

// NewChallenge returns a pointer to a Challenge
func NewChallenge(domain, token, keyAuth string) *Challenge {
	return &Challenge{
		Domain:  domain,
		Token:   token,
		KeyAuth: keyAuth,
	}
}

// DynamoDBStore is an implementation of Store using AWS DynamoDB to persist Challenges
type DynamoDBStore struct {
	c     dynamodbiface.DynamoDBAPI
	table string
}

// NewDynamoDBStore returns a pointer to a DynamoDBStore
func NewDynamoDBStore(c dynamodbiface.DynamoDBAPI, table string) *DynamoDBStore {
	return &DynamoDBStore{
		c:     c,
		table: table,
	}
}

const (
	dynamoDBColumnDomain  = "domain"
	dynamoDBColumnKeyAuth = "keyAuth"
	dynamoDBColumnToken   = "token"
)

// DeleteChallenge deletes the relevant row from DynamoDB
func (ds *DynamoDBStore) DeleteChallenge(token string) error {
	in := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			dynamoDBColumnToken: {
				S: aws.String(token),
			},
		},
		TableName: aws.String(ds.table),
	}
	_, err := ds.c.DeleteItem(in)

	return parseDynamoDBError(err)
}

// GetChallenge retrieves the relevant row from DynamoDB and returns it as a pointer
// to a Challenge
func (ds *DynamoDBStore) GetChallenge(token string) (*Challenge, error) {
	in := &dynamodb.GetItemInput{
		AttributesToGet: []*string{
			aws.String(dynamoDBColumnDomain),
			aws.String(dynamoDBColumnKeyAuth),
		},
		ConsistentRead: aws.Bool(true),
		Key: map[string]*dynamodb.AttributeValue{
			dynamoDBColumnToken: {
				S: aws.String(token),
			},
		},
		TableName: aws.String(ds.table),
	}

	resp, err := ds.c.GetItem(in)
	if err != nil {
		return nil, parseDynamoDBError(err)
	}

	return NewChallenge(aws.StringValue(resp.Item[dynamoDBColumnDomain].S), token, aws.StringValue(resp.Item[dynamoDBColumnKeyAuth].S)), nil
}

// PutChallenge serialises a Challenge and puts it in a row in DynamoDB
func (ds *DynamoDBStore) PutChallenge(ch *Challenge) error {
	in := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"domain": {
				S: aws.String(ch.Domain),
			},
			"token": {
				S: aws.String(ch.Token),
			},
			"keyAuth": {
				S: aws.String(ch.KeyAuth),
			},
		},
		TableName: aws.String(ds.table),
	}

	_, err := ds.c.PutItem(in)
	return parseDynamoDBError(err)
}

// parseDynamoDBError checks for known DynamoDB response codes to see if we can return a meaningful error
func parseDynamoDBError(err error) error {
	log.Printf("[ERROR] %v", err)
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case dynamodb.ErrCodeProvisionedThroughputExceededException, dynamodb.ErrCodeRequestLimitExceeded:
			// We exceeded our AWS limits
			return ErrStoreRateLimited
		case dynamodb.ErrCodeResourceNotFoundException:
			// Challenge didn't exist
			return ErrStoreNotFound
		}
	}

	// Some other unexpected error condition
	return err
}
