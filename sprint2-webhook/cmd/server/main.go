// Assignment 2 — Northstar Inventory Sync Service (post-pivot)
//
// Receives stock updates pushed by the Northstar warehouse via signed webhooks
// and serves the cached stock data through GET /stock.
//
// The original polling goroutine (every 5 min) was removed in the Day 4 pivot.
// Deprecated polling code is preserved in internal/warehouse/poller.go.
//
// Run: WEBHOOK_SECRET=my-secret go run ./cmd/server/
package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/group109/northstar-webhook/internal/cache"
	"github.com/group109/northstar-webhook/internal/webhook"
)

type stockUpdatePayload struct {
	Item    string `json:"item"`
	Size    string `json:"size"`
	SKU     string `json:"sku,omitempty"`
	Qty     int    `json:"qty"`
	InStock bool   `json:"inStock"`
}

func main() {
	secret := os.Getenv("WEBHOOK_SECRET")
	if secret == "" {
		secret = "dev-secret-change-me"
		log.Println("[WARN] WEBHOOK_SECRET not set; using default dev secret")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	stockCache := cache.New()
	log.Printf("[server] Stock cache initialized with %d items", len(stockCache.All()))

	mux := http.NewServeMux()

	// POST /webhook — receive a signed stock update from the warehouse.
	mux.HandleFunc("POST /webhook", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "cannot read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if len(body) == 0 {
			http.Error(w, "empty body", http.StatusBadRequest)
			return
		}

		if !webhook.VerifySignature(secret, body, r.Header.Get("X-Hub-Signature-256")) {
			log.Printf("[webhook] REJECTED invalid signature (bytes=%d)", len(body))
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		var payload stockUpdatePayload
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "invalid JSON payload", http.StatusBadRequest)
			return
		}
		if payload.Item == "" || payload.Size == "" {
			http.Error(w, "payload missing required fields: item, size", http.StatusBadRequest)
			return
		}

		stockCache.Set(cache.StockItem{
			Item:    payload.Item,
			Size:    payload.Size,
			SKU:     payload.SKU,
			Qty:     payload.Qty,
			InStock: payload.InStock,
		})
		log.Printf("[webhook] ACCEPTED: %s (%s) qty=%d inStock=%v", payload.Item, payload.Size, payload.Qty, payload.InStock)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status":  "accepted",
			"item":    payload.Item,
			"size":    payload.Size,
			"qty":     payload.Qty,
			"inStock": payload.InStock,
		})
	})

	// GET /stock — query the cache. Use ?item=<name> for substring search.
	mux.HandleFunc("GET /stock", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("item")
		var results []cache.StockItem
		if query == "" {
			results = stockCache.All()
		} else {
			results = stockCache.Search(query)
		}
		if results == nil {
			results = []cache.StockItem{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"query":   query,
			"count":   len(results),
			"results": results,
		})
	})

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status":      "ok",
			"service":     "northstar-inventory-sync",
			"mode":        "webhook-push",
			"cachedItems": len(stockCache.All()),
			"time":        time.Now().UTC().Format(time.RFC3339),
		})
	})

	addr := ":" + port
	log.Printf("[server] Northstar Inventory Sync listening on %s (mode: webhook-push)", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("[server] Fatal: %v", err)
	}
}
