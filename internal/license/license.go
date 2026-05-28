// Package license implements offline license verification using Ed25519.
//
// The app ships an embedded public key (PublicKeyB64) and verifies license keys
// locally — no server, works fully offline. Keys are issued by the seller with
// the matching private key via cmd/license-gen.
//
// A license "key" is: base64url(payloadJSON) + "." + base64url(signature).
package license

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// Product this build's licenses must be issued for.
const Product = "videoforge"

// PublicKeyB64 is the Ed25519 public key (std base64) used to verify licenses.
// Generate your production keypair with `go run ./cmd/license-gen keygen` and
// paste the public key here. The private key must stay secret, outside the repo.
// (var, not const, so tests can swap in a throwaway keypair.)
var PublicKeyB64 = "vfFMOU+iD0XNHnkPIR2fE+TnBWXQgSM1iKHOMWuR+D4="

// License is the signed payload a buyer receives.
type License struct {
	Name    string     `json:"name"`
	Email   string     `json:"email"`
	Product string     `json:"product"`
	Issued  time.Time  `json:"issued"`
	Expiry  *time.Time `json:"expiry,omitempty"` // nil = perpetual
}

var (
	ErrFormat    = errors.New("invalid license key format")
	ErrSignature = errors.New("license key is not valid")
	ErrProduct   = errors.New("license is for a different product")
	ErrExpired   = errors.New("license has expired")
	ErrNoPubKey  = errors.New("build is missing its license public key")
)

// Verify decodes and cryptographically validates a license key.
func Verify(key string) (*License, error) {
	payload, sig, err := split(key)
	if err != nil {
		return nil, err
	}
	pub, err := base64.StdEncoding.DecodeString(PublicKeyB64)
	if err != nil || len(pub) != ed25519.PublicKeySize {
		return nil, ErrNoPubKey
	}
	if !ed25519.Verify(ed25519.PublicKey(pub), payload, sig) {
		return nil, ErrSignature
	}
	var lic License
	if err := json.Unmarshal(payload, &lic); err != nil {
		return nil, ErrFormat
	}
	if lic.Product != Product {
		return nil, ErrProduct
	}
	if lic.Expiry != nil && time.Now().After(*lic.Expiry) {
		return nil, ErrExpired
	}
	return &lic, nil
}

// Sign produces a license key from a payload using the seller's private key.
// Used by cmd/license-gen, not by the app.
func Sign(priv ed25519.PrivateKey, lic License) (string, error) {
	lic.Product = Product
	payload, err := json.Marshal(lic)
	if err != nil {
		return "", err
	}
	sig := ed25519.Sign(priv, payload)
	return enc(payload) + "." + enc(sig), nil
}

func split(key string) (payload, sig []byte, err error) {
	parts := strings.Split(strings.TrimSpace(key), ".")
	if len(parts) != 2 {
		return nil, nil, ErrFormat
	}
	if payload, err = dec(parts[0]); err != nil {
		return nil, nil, ErrFormat
	}
	if sig, err = dec(parts[1]); err != nil {
		return nil, nil, ErrFormat
	}
	return payload, sig, nil
}

func enc(b []byte) string         { return base64.RawURLEncoding.EncodeToString(b) }
func dec(s string) ([]byte, error) { return base64.RawURLEncoding.DecodeString(s) }
