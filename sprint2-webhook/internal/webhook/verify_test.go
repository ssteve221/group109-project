package webhook

import (
	"testing"
)

func TestVerifySignature(t *testing.T) {
	secret := "my-test-secret"
	body := []byte(`{"item":"Running Shoes","size":"10","qty":15,"inStock":true}`)

	// Compute a valid signature
	validSig := ComputeSignature(secret, body)

	tests := []struct {
		name      string
		secret    string
		body      []byte
		sigHeader string
		wantValid bool
	}{
		{
			name:      "valid signature",
			secret:    secret,
			body:      body,
			sigHeader: validSig,
			wantValid: true,
		},
		{
			name:      "wrong secret",
			secret:    "wrong-secret",
			body:      body,
			sigHeader: validSig,
			wantValid: false,
		},
		{
			name:      "tampered body",
			secret:    secret,
			body:      []byte(`{"item":"Running Shoes","size":"10","qty":999,"inStock":true}`),
			sigHeader: validSig,
			wantValid: false,
		},
		{
			name:      "missing sha256 prefix",
			secret:    secret,
			body:      body,
			sigHeader: "abcdef1234567890",
			wantValid: false,
		},
		{
			name:      "empty signature header",
			secret:    secret,
			body:      body,
			sigHeader: "",
			wantValid: false,
		},
		{
			name:      "invalid hex in header",
			secret:    secret,
			body:      body,
			sigHeader: "sha256=not-valid-hex!!",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VerifySignature(tt.secret, tt.body, tt.sigHeader)
			if got != tt.wantValid {
				t.Errorf("VerifySignature() = %v, want %v", got, tt.wantValid)
			}
		})
	}
}

func TestComputeSignature(t *testing.T) {
	secret := "test-secret"
	body := []byte(`hello`)
	sig := ComputeSignature(secret, body)

	// Should start with sha256=
	if len(sig) < 7 || sig[:7] != "sha256=" {
		t.Errorf("ComputeSignature() output %q does not start with 'sha256='", sig)
	}

	// Computing twice with same inputs should give same result (deterministic)
	sig2 := ComputeSignature(secret, body)
	if sig != sig2 {
		t.Errorf("ComputeSignature is not deterministic: %q != %q", sig, sig2)
	}

	// Different secret should produce different signature
	sig3 := ComputeSignature("different-secret", body)
	if sig == sig3 {
		t.Errorf("Different secrets should produce different signatures")
	}
}
