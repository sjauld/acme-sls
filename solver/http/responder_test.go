package http

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	petname "github.com/dustinkirkland/golang-petname"
	"github.com/gin-gonic/gin"
	"github.com/sjauld/acme-sls/helpers"
)

func TestValidateChallenge_valid(t *testing.T) {
	ch := &Challenge{
		Domain: "www.com",
	}

	url, err := url.Parse("http://www.com:80")
	if err != nil {
		t.Fatal(err)
	}

	req := &http.Request{
		URL: url,
	}

	if !validateChallenge(req, ch) {
		t.Errorf("Validation failed when it should have passed")
	}
}

func TestValidateChallenge_invalid(t *testing.T) {
	ch := &Challenge{
		Domain: "www.lolz",
	}

	url, err := url.Parse("http://www.com:80")
	if err != nil {
		t.Fatal(err)
	}

	req := &http.Request{
		URL: url,
	}

	if validateChallenge(req, ch) {
		t.Errorf("Validation passed when it should have failed")
	}
}

func TestNewGinHandlerFunc_valid(t *testing.T) {
	table := petname.Generate(2, "-")
	store := NewDynamoDBStore(dyn, table)

	f := NewGinHandlerFunc(store)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	requestURL, err := url.Parse("http://www.ginhandler.com:80/.well-known/acme-challenge/ginhandlerfunctesttoken")
	if err != nil {
		t.Fatal(err)
	}

	req := &http.Request{
		URL: requestURL,
	}

	c.Request = req
	c.Params = append(c.Params, gin.Param{
		Key:   "token",
		Value: "ginhandlerfunctesttoken",
	})

	expectedKey := map[string]*dynamodb.AttributeValue{
		"token": {
			S: aws.String("ginhandlerfunctesttoken"),
		},
	}

	output := dynamodb.GetItemOutput{
		Item: map[string]*dynamodb.AttributeValue{
			"domain": {
				S: aws.String("www.ginhandler.com"),
			},
			"keyAuth": {
				S: aws.String("ginhandlerfunckeyauth"),
			},
			"token": {
				S: aws.String("ginhandlerfunctesttoken"),
			},
		},
	}

	mock.ExpectGetItem().ToTable(table).WithKeys(expectedKey).WillReturns(output)
	f(c)

	helpers.ExpectIntMatch(t, http.StatusOK, w.Code)
	helpers.ExpectStringMatch(t, "ginhandlerfunckeyauth", w.Body.String())
}
