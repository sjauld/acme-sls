// package solver solves the ACMEv2 HTTP-01 challenge. The workflow is as follows:
//
// 1. client requests a certificate from the remote CA, using the Solver as the HTTP-01 challenge
// 2. Solver populates the Challenge in the Store and notifies the CA that the challenge is ready
// 3. remote CA requests the keyauth from the well known path on the server
// 4. server retrieves the Challenge from the Store, validates the requests and presents the keyauth to the remote CA
package solver

import (
	"log"
)

// Solver implements lego's challenge.Provider
type Solver struct {
	store Store
}

// New returns a pointer to a Solver, initialised with a Store of your choice
func New(store Store) *Solver {
	return &Solver{
		store: store,
	}
}

// Present writes the challenge information into the Store so that the
// server can respond to HTTP queries with the correct value
func (s *Solver) Present(domain, token, keyAuth string) error {
	log.Printf("[INFO] Presenting domain: %v, token: %v, keyauth: %v", domain, token, keyAuth)
	ch := NewChallenge(domain, token, keyAuth)

	return s.store.PutChallenge(ch)
}

// CleanUp removes the challenge information from the Store
func (s *Solver) CleanUp(domain, token, keyAuth string) error {
	log.Printf("[INFO] CleaningUp domain: %v, token: %v, keyauth: %v", domain, token, keyAuth)

	return s.store.DeleteChallenge(token)
}
