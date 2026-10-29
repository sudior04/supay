package transport

import (
	"bytes"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestPINFormatPreservesLeadingZeroAndRejectsNonASCIIDigits(t *testing.T) {
	for _, v := range []struct {
		pin string
		ok  bool
	}{{"012345", true}, {"123456", true}, {"12345", false}, {"1234567", false}, {"12 456", false}, {"１２３４５６", false}, {"12a456", false}, {"", false}} {
		if validPIN(v.pin) != v.ok {
			t.Errorf("invalid format result for %q", v.pin)
		}
	}
}
func TestPINHashBoundToUserAndServerPepper(t *testing.T) {
	a := &Auth{pinPepper: []byte("pin-pepper-with-at-least-32-bytes")}
	hash, err := a.hashPIN("user-a", "012345")
	if err != nil {
		t.Fatal(err)
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), a.pinMaterial("user-a", "012345")) != nil {
		t.Fatal("valid PIN rejected")
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), a.pinMaterial("user-b", "012345")) == nil {
		t.Fatal("PIN hash reusable across users")
	}
	b := &Auth{pinPepper: []byte("different-pin-pepper-at-least-32-bytes")}
	if bytes.Equal(a.pinMaterial("user-a", "012345"), b.pinMaterial("user-a", "012345")) {
		t.Fatal("server pepper did not affect hash material")
	}
}
