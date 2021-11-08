package helpers

import "testing"

func TestNewUser(t *testing.T) {
	testUser, err := NewUser("asd")
	if err != nil {
		t.Fatal(err)
	}

	ExpectStringMatch(t, "asd", testUser.GetEmail())
}
