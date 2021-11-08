package helpers

import "testing"

func ExpectStringMatch(t *testing.T, exp, act string) {
	if exp != act {
		t.Errorf("Expected %v, got %v", exp, act)
	}
}

func ExpectIntMatch(t *testing.T, exp, act int) {
	if exp != act {
		t.Errorf("Expected %d, got %d", exp, act)
	}
}
