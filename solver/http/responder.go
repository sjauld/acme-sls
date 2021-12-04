package http

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewGinHandlerFunc returns a gin.HandlerFunc that will parse the incoming challenge
// request from the remote CA, retrieve the challenge data from the Store, and return
// an appropriate response.
func NewGinHandlerFunc(store Store) func(*gin.Context) {
	return func(c *gin.Context) {
		log.Printf("[DEBUG] request %+v", c)

		ch, err := store.GetChallenge(c.Param("token"))
		if err != nil {
			log.Printf("[ERROR] could not GetChallenge: %v", err)
			perr := parseDynamoDBError(err)
			switch perr {
			case ErrStoreNotFound:
				c.String(http.StatusNotFound, "Challenge not found")
			case ErrStoreRateLimited:
				c.String(http.StatusTooManyRequests, "Please try again soon")
			default:
				c.String(http.StatusInternalServerError, "Unexpected error")
			}
			return
		}

		ok := validateChallenge(c.Request, ch)
		if !ok {
			c.String(http.StatusNotFound, "Challenge not found")
			return
		}

		c.String(http.StatusOK, ch.KeyAuth)
	}
}

// We just need to check that the hostname we stored for the token is the same as the
// hostname in the HTTP request
func validateChallenge(req *http.Request, ch *Challenge) bool {
	reqHost := req.URL.Hostname()
	expectedHost := ch.Domain
	log.Printf("[DEBUG] request host: %v, expected host: %v", reqHost, expectedHost)
	return reqHost == expectedHost
}
