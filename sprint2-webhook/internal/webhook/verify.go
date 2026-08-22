// Package webhook provides HMAC-SHA256 signature verification for incoming
// webhook payloads. The sender and receiver share a secret; the sender signs
// the raw body and sends the digest as X-Hub-Signature-256: sha256=<hex>.
package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// VerifySignature reports whether sigHeader is a valid HMAC-SHA256 signature
// of body using secret. sigHeader must have the form "sha256=<hex>".
// Uses constant-time comparison to prevent timing attacks.
func VerifySignature(secret string, body []byte, sigHeader string) bool {
	const prefix = "sha256="
	if !strings.HasPrefix(sigHeader, prefix) {
		return false
	}

	receivedMAC, err := hex.DecodeString(strings.TrimPrefix(sigHeader, prefix))
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)

	return hmac.Equal(receivedMAC, mac.Sum(nil))
}

// ComputeSignature returns the HMAC-SHA256 signature for body in the form
// "sha256=<hex>". Used in tests and the send_test_webhook script.
func ComputeSignature(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return fmt.Sprintf("sha256=%s", hex.EncodeToString(mac.Sum(nil)))
}
