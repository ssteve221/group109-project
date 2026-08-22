// Assignment 1 — Webhook Verification Mini-Prototype
//
// A minimal HTTP server demonstrating HMAC-SHA256 webhook verification.
//
// Run:   WEBHOOK_SECRET=my-secret go run ./cmd/prototype/
// Test:  curl http://localhost:8080/sign   (generates a test signature)
//        Then POST that signature to /webhook
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/group109/northstar-webhook/internal/webhook"
)

func main() {
	secret := os.Getenv("WEBHOOK_SECRET")
	if secret == "" {
		secret = "dev-secret-change-me"
		log.Println("[WARN] WEBHOOK_SECRET not set; using default dev secret")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /webhook", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "could not read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		sigHeader := r.Header.Get("X-Hub-Signature-256")
		if sigHeader == "" {
			log.Println("[webhook] REJECTED: missing X-Hub-Signature-256 header")
			http.Error(w, "missing signature header", http.StatusUnauthorized)
			return
		}

		if !webhook.VerifySignature(secret, body, sigHeader) {
			log.Printf("[webhook] REJECTED: invalid signature (header=%s)", sigHeader)
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		log.Printf("[webhook] ACCEPTED: %d bytes, sig=%s", len(body), sigHeader)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status":  "accepted",
			"message": "Webhook signature verified successfully",
			"bytes":   len(body),
		})
	})

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "webhook-prototype"})
	})

	// Helper: compute and display a valid test signature for a given payload.
	mux.HandleFunc("GET /sign", func(w http.ResponseWriter, r *http.Request) {
		payload := r.URL.Query().Get("payload")
		if payload == "" {
			payload = `{"item":"Running Shoes","size":"10","qty":15,"inStock":true}`
		}
		sig := webhook.ComputeSignature(secret, []byte(payload))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"payload":   payload,
			"signature": sig,
			"curl_example": fmt.Sprintf(
				`curl -X POST http://localhost:8080/webhook -H "Content-Type: application/json" -H "X-Hub-Signature-256: %s" -d '%s'`,
				sig, payload,
			),
		})
	})

	log.Printf("[prototype] Listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("[prototype] %v", err)
	}
}
