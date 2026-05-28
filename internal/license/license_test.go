package license

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"testing"
	"time"
)

// withKeypair temporarily swaps in a fresh keypair so the test doesn't depend on
// the embedded production key, and returns the private key for signing.
func withKeypair(t *testing.T) ed25519.PrivateKey {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	orig := PublicKeyB64
	PublicKeyB64 = base64.StdEncoding.EncodeToString(pub)
	t.Cleanup(func() { PublicKeyB64 = orig })
	return priv
}

func TestVerifyValid(t *testing.T) {
	priv := withKeypair(t)
	key, err := Sign(priv, License{Name: "Jane", Email: "j@x.com"})
	if err != nil {
		t.Fatal(err)
	}
	lic, err := Verify(key)
	if err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
	if lic.Name != "Jane" || lic.Product != Product {
		t.Errorf("unexpected payload: %+v", lic)
	}
}

func TestVerifyTampered(t *testing.T) {
	priv := withKeypair(t)
	key, _ := Sign(priv, License{Name: "Jane", Email: "j@x.com"})
	// Flip a character in the payload portion.
	bad := []byte(key)
	bad[0] = bad[0] ^ 0x01
	if _, err := Verify(string(bad)); err == nil {
		t.Fatal("tampered key should not verify")
	}
}

func TestVerifyWrongKey(t *testing.T) {
	withKeypair(t)
	_, otherPriv, _ := ed25519.GenerateKey(rand.Reader)
	key, _ := Sign(otherPriv, License{Name: "Mallory"})
	if _, err := Verify(key); err != ErrSignature {
		t.Fatalf("expected ErrSignature, got %v", err)
	}
}

func TestVerifyExpired(t *testing.T) {
	priv := withKeypair(t)
	past := time.Now().Add(-time.Hour)
	key, _ := Sign(priv, License{Name: "Jane", Expiry: &past})
	if _, err := Verify(key); err != ErrExpired {
		t.Fatalf("expected ErrExpired, got %v", err)
	}
}

func TestVerifyBadFormat(t *testing.T) {
	if _, err := Verify("not-a-license"); err != ErrFormat {
		t.Fatalf("expected ErrFormat, got %v", err)
	}
}
