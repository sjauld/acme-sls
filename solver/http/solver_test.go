package http

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	petname "github.com/dustinkirkland/golang-petname"
)

func TestPresent(t *testing.T) {
	table := petname.Generate(2, "-")
	store := NewDynamoDBStore(dyn, table)

	solver := New(store)

	mock.ExpectPutItem().ToTable(table).WithItems(map[string]*dynamodb.AttributeValue{
		"domain": {
			S: aws.String("testing.com"),
		},
		"keyAuth": {
			S: aws.String("keyauth"),
		},
		"token": {
			S: aws.String("token"),
		},
	})
	err := solver.Present("testing.com", "token", "keyauth")
	if err != nil {
		t.Error(err)
	}
}

func TestCleanUp(t *testing.T) {
	table := petname.Generate(2, "-")
	store := NewDynamoDBStore(dyn, table)

	solver := New(store)

	expectedKey := map[string]*dynamodb.AttributeValue{
		"token": {
			S: aws.String("token"),
		},
	}

	mock.ExpectDeleteItem().ToTable(table).WithKeys(expectedKey)

	err := solver.CleanUp("testing.com", "token", "keyauth")
	if err != nil {
		t.Error(err)
	}
}
