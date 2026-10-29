package authn

import (
	"testing"
	"time"
)

func TestPasswordAccessTokenAcceptsValidSignatureWithoutPIN(t *testing.T) {
	secret := []byte("test-secret-with-at-least-32-bytes-long")
	claims := Claims{Subject: "user-1", Role: "CUSTOMER", Issued: time.Now().Unix(), Expires: time.Now().Add(time.Minute).Unix()}
	token, err := Sign(claims, secret)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Verify(token, secret); err != nil {
		t.Fatal(err)
	}
	if _, err = Verify(token, []byte("different-secret-with-at-least-32-bytes")); err == nil {
		t.Fatal("incorrect signing key accepted")
	}
	claims.Expires = time.Now().Add(-time.Second).Unix()
	token, _ = Sign(claims, secret)
	if _, err = Verify(token, secret); err == nil {
		t.Fatal("expired token accepted")
	}
}
