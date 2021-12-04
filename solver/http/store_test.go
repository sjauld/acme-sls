package http

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	petname "github.com/dustinkirkland/golang-petname"
	dynamock "github.com/gusaul/go-dynamock"

	"github.com/sjauld/acme-sls/helpers"
)

var (
	dyn  dynamodbiface.DynamoDBAPI
	mock *dynamock.DynaMock
)

func init() {
	dyn, mock = dynamock.New()
}

func TestDynamoDBStoreDeleteChallenge(t *testing.T) {
	table := petname.Generate(2, "-")
	store := NewDynamoDBStore(dyn, table)

	expectedKey := map[string]*dynamodb.AttributeValue{
		"token": {
			S: aws.String("b"),
		},
	}

	mock.ExpectDeleteItem().ToTable(table).WithKeys(expectedKey)

	err := store.DeleteChallenge("b")
	if err != nil {
		t.Error(err)
	}
}

func TestDynamoDBStoreGetChallenge(t *testing.T) {
	table := petname.Generate(2, "-")
	store := NewDynamoDBStore(dyn, table)

	expectedKey := map[string]*dynamodb.AttributeValue{
		"token": {
			S: aws.String("b"),
		},
	}

	output := dynamodb.GetItemOutput{
		Item: map[string]*dynamodb.AttributeValue{
			"domain": {
				S: aws.String("a"),
			},
			"keyAuth": {
				S: aws.String("c"),
			},
			"token": {
				S: aws.String("b"),
			},
		},
	}

	mock.ExpectGetItem().ToTable(table).WithKeys(expectedKey).WillReturns(output)
	ch, err := store.GetChallenge("b")
	if err != nil {
		t.Fatal(err)
	}

	helpers.ExpectStringMatch(t, "a", ch.Domain)
	helpers.ExpectStringMatch(t, "b", ch.Token)
	helpers.ExpectStringMatch(t, "c", ch.KeyAuth)
}

func TestDynamoDBStorePutChallenge(t *testing.T) {
	table := petname.Generate(2, "-")
	store := NewDynamoDBStore(dyn, table)
	mock.ExpectPutItem().ToTable(table).WithItems(map[string]*dynamodb.AttributeValue{
		"domain": {
			S: aws.String("a"),
		},
		"keyAuth": {
			S: aws.String("c"),
		},
		"token": {
			S: aws.String("b"),
		},
	})

	err := store.PutChallenge(NewChallenge("a", "b", "c"))
	if err != nil {
		t.Error(err)
	}
}
