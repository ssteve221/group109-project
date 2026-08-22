// Package warehouse — DEPRECATED
//
// This file contained the original Day 3 polling implementation: a background
// goroutine that called the Northstar Warehouse API every 5 minutes to refresh
// the stock cache.
//
// STATUS: REMOVED as part of the Day 4 pivot (see SCOPE_DELTA_ANALYSIS.md).
//
// Why it was killed:
//
//	The client (Northstar Retail Co.) announced that the warehouse API's
//	polling endpoint is being decommissioned in 48 hours. The architecture
//	was migrated to a webhook push model where the warehouse PUSHES updates
//	to our server instead of our server pulling.
//
// What replaced it:
//
//	POST /webhook in cmd/server/main.go receives signed payloads from the
//	warehouse and updates the stock cache directly.
//
// Regression note:
//
//	The GET /stock endpoint and StockCache remain. Only the data-ingestion
//	path changed (pull → push). All existing query functionality is intact.
//
// The code below is preserved for audit purposes to satisfy the assignment
// requirement that obsolete code be "visibly removed or marked deprecated".
// All function bodies are commented out — this package exports nothing.
package warehouse

// ─────────────────────────────────────────────────────────────────────────────
// ORIGINAL CODE (DO NOT USE — DEPRECATED — ALL COMMENTED OUT)
// ─────────────────────────────────────────────────────────────────────────────
//
// import (
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"time"
//
// 	"github.com/group109/northstar-webhook/internal/cache"
// )
//
// const (
// 	// warehouseAPIURL is the polling endpoint that is now decommissioned.
// 	// DEPRECATED: This endpoint will return 410 Gone from 2026-08-23 onward.
// 	warehouseAPIURL = "https://warehouse.northstar-internal.example.com/api/v1/stock"
//
// 	// pollInterval was the cadence for fetching stock data.
// 	pollInterval = 5 * time.Minute
// )
//
// // StartPoller launches a background goroutine that polls the warehouse API
// // every pollInterval and writes results into the stock cache.
// //
// // DEPRECATED: Replaced by POST /webhook handler in cmd/server/main.go.
// func StartPoller(stockCache *cache.StockCache, done <-chan struct{}) {
// 	ticker := time.NewTicker(pollInterval)
// 	defer ticker.Stop()
//
// 	log.Println("[poller] Starting warehouse API poll (interval: 5m)")
// 	poll(stockCache) // run immediately on startup
//
// 	for {
// 		select {
// 		case <-ticker.C:
// 			poll(stockCache)
// 		case <-done:
// 			log.Println("[poller] Shutting down.")
// 			return
// 		}
// 	}
// }
//
// // poll makes a single HTTP GET to the warehouse stock endpoint.
// func poll(stockCache *cache.StockCache) {
// 	resp, err := http.Get(warehouseAPIURL)
// 	if err != nil {
// 		log.Printf("[poller] ERROR fetching stock: %v", err)
// 		return
// 	}
// 	defer resp.Body.Close()
//
// 	if resp.StatusCode != http.StatusOK {
// 		log.Printf("[poller] ERROR: warehouse API returned %d", resp.StatusCode)
// 		return
// 	}
//
// 	var items []cache.StockItem
// 	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
// 		log.Printf("[poller] ERROR decoding response: %v", err)
// 		return
// 	}
//
// 	for _, item := range items {
// 		stockCache.Set(item)
// 	}
// 	log.Printf("[poller] Cache refreshed: %d items updated", len(items))
// }
//
// // simulateWarehouseResponse returns a mock API response — used in Day 3 dev.
// func simulateWarehouseResponse() ([]cache.StockItem, error) {
// 	raw := `[
// 		{"item":"Running Shoes","size":"10","qty":15,"inStock":true},
// 		{"item":"Trail Runner Jacket","size":"L","qty":3,"inStock":true}
// 	]`
// 	var items []cache.StockItem
// 	if err := json.Unmarshal([]byte(raw), &items); err != nil {
// 		return nil, fmt.Errorf("simulate: %w", err)
// 	}
// 	return items, nil
// }
