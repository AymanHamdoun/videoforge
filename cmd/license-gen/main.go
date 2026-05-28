// Command license-gen is the seller-side tool for the offline license system.
//
//	go run ./cmd/license-gen keygen
//	    Generate an Ed25519 keypair. Paste the public key into
//	    internal/license/license.go (PublicKeyB64); keep the private key secret.
//
//	LICENSE_PRIVATE_KEY=<b64> go run ./cmd/license-gen issue --name "Jane" --email j@x.com [--days 365]
//	    Print a signed license key to give to a buyer.
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"videoforge/internal/license"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	switch os.Args[1] {
	case "keygen":
		keygen()
	case "issue":
		issue(os.Args[2:])
	default:
		usage()
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `license-gen — VideoForge license tool

  keygen
      Generate an Ed25519 keypair.

  issue --name NAME --email EMAIL [--days N] [--key FILE]
      Issue a signed license key. Private key from --key FILE or $LICENSE_PRIVATE_KEY.`)
	os.Exit(1)
}

func keygen() {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	must(err)
	fmt.Println("Public key  → paste into internal/license/license.go (PublicKeyB64):")
	fmt.Println(base64.StdEncoding.EncodeToString(pub))
	fmt.Println()
	fmt.Println("Private key → KEEP SECRET, store outside the repo (e.g. a password manager):")
	fmt.Println(base64.StdEncoding.EncodeToString(priv))
}

func issue(args []string) {
	fs := flag.NewFlagSet("issue", flag.ExitOnError)
	name := fs.String("name", "", "licensee name")
	email := fs.String("email", "", "licensee email")
	days := fs.Int("days", 0, "validity in days (0 = perpetual)")
	keyFile := fs.String("key", "", "private key file (base64); else $LICENSE_PRIVATE_KEY")
	_ = fs.Parse(args)

	privB64 := os.Getenv("LICENSE_PRIVATE_KEY")
	if *keyFile != "" {
		b, err := os.ReadFile(*keyFile)
		must(err)
		privB64 = strings.TrimSpace(string(b))
	}
	if privB64 == "" {
		fmt.Fprintln(os.Stderr, "no private key: pass --key FILE or set LICENSE_PRIVATE_KEY")
		os.Exit(1)
	}
	priv, err := base64.StdEncoding.DecodeString(privB64)
	must(err)
	if len(priv) != ed25519.PrivateKeySize {
		fmt.Fprintln(os.Stderr, "invalid private key length")
		os.Exit(1)
	}

	lic := license.License{Name: *name, Email: *email, Issued: time.Now().UTC()}
	if *days > 0 {
		exp := time.Now().UTC().AddDate(0, 0, *days)
		lic.Expiry = &exp
	}
	key, err := license.Sign(ed25519.PrivateKey(priv), lic)
	must(err)
	fmt.Println(key)
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
