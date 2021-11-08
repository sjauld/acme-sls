package helpers

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	"github.com/go-acme/lego/registration"
)

// User represents a user of the CA server
type User struct {
	email        string
	key          crypto.PrivateKey
	registration *registration.Resource
}

func NewUser(email string) (*User, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	return &User{
		email: email,
		key:   privateKey,
	}, nil
}

// GetEmail implements registration.User
func (u *User) GetEmail() string {
	return u.email
}

// GetPrivateKey implements registration.User
func (u *User) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

// GetRegistration implements regsitration.User
func (u *User) GetRegistration() *registration.Resource {
	return u.registration
}

// SetRegistration setter for registration
func (u *User) SetRegistration(r *registration.Resource) {
	u.registration = r
}
