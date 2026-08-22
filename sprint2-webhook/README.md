# Northstar Inventory Sync Service

**Sprint 2 — The Meridian Pivot | Group 109**  
**Language: Go | Tool: Webhook Verification**

A live inventory sync service for Northstar Retail Co.'s support tool.  
Provides a query endpoint (`GET /stock`) backed by a webhook-fed cache,  
so "is this in stock?" answers are always accurate and up-to-date.

---

## Architecture (Post-Pivot)

```
Northstar Warehouse
        │
        │  POST /webhook
        │  X-Hub-Signature-256: sha256=<hmac>
        ▼
┌──────────────────────────┐
│  Signature Verification  │ ← HMAC-SHA256 (internal/webhook)
│  (reject if invalid)     │
└───────────┬──────────────┘
            │ valid payload
            ▼
┌──────────────────────────┐
│    In-Memory Stock Cache │ ← sync.RWMutex map (internal/cache)
└───────────┬──────────────┘
            │
            ▼
┌──────────────────────────┐
│   GET /stock?item=...    │ ← Support tool queries this
└──────────────────────────┘
```

> **Note on the pivot:** This service originally (Day 3) used a polling goroutine to fetch stock every 5 minutes. The client decommissioned that endpoint on Day 4, forcing an immediate switch to the webhook push model. See `SCOPE_DELTA_ANALYSIS.md` for full trade-off documentation.

---

## Project Structure

```
northstar-webhook/
├── cmd/
│   ├── prototype/main.go   ← Assignment 1: standalone webhook verify prototype
│   └── server/main.go      ← Assignment 2: full inventory sync service
├── internal/
│   ├── webhook/
│   │   ├── verify.go       ← HMAC-SHA256 verification logic
│   │   └── verify_test.go  ← Unit tests
│   ├── cache/
│   │   └── stock.go        ← Thread-safe in-memory stock cache
│   └── warehouse/
│       └── poller.go       ← DEPRECATED: polling code (kept for audit)
├── scripts/
│   └── send_test_webhook.sh ← Simulates warehouse pushes
├── BLOCKER_JOURNAL.md       ← Assignment 1 deliverable
├── SCOPE_DELTA_ANALYSIS.md  ← Assignment 2 deliverable
└── README.md
```

---

## Requirements

- Go 1.22+ (uses method-based routing: `"POST /webhook"`)
- No external dependencies (stdlib only)

---

## Running

### Assignment 1 — Mini-Prototype (port 8080)
```bash
WEBHOOK_SECRET=my-secret go run ./cmd/prototype/
```

### Assignment 2 — Full Northstar Inventory Server (port 9090)
```bash
WEBHOOK_SECRET=my-secret go run ./cmd/server/
```

### Day 4 Pivot — Solstice Events Kiosk (port 7070) ⭐
```bash
go run ./cmd/kiosk/
# Then open http://localhost:7070 in your browser
```

### Run Tests
```bash
go test ./...
```

### Run Test Script
```bash
chmod +x scripts/send_test_webhook.sh
WEBHOOK_SECRET=my-secret ./scripts/send_test_webhook.sh
```

---

## API Reference

### POST /webhook
Receive a signed stock update from the warehouse.

**Request headers:**
- `Content-Type: application/json`
- `X-Hub-Signature-256: sha256=<hmac-sha256-hex>`

**Request body:**
```json
{
  "item": "Running Shoes",
  "size": "10",
  "qty": 15,
  "inStock": true
}
```

**Responses:**
- `200 OK` — signature valid, cache updated
- `401 Unauthorized` — invalid or missing signature
- `400 Bad Request` — malformed payload

### GET /stock
Query current stock from the cache.

**Query params:**
- `?item=<name>` — substring search (case-insensitive)
- No params — returns all items

**Response:**
```json
{
  "query": "Running Shoes",
  "count": 3,
  "results": [
    { "item": "Running Shoes", "size": "9", "qty": 22, "inStock": true, "updatedAt": "..." },
    { "item": "Running Shoes", "size": "10", "qty": 15, "inStock": true, "updatedAt": "..." },
    { "item": "Running Shoes", "size": "11", "qty": 9, "inStock": true, "updatedAt": "..." }
  ]
}
```

### GET /health
```json
{ "status": "ok", "service": "northstar-inventory-sync", "mode": "webhook-push", "cachedItems": 8 }
```

---

## Generating a Test Signature

Using `openssl`:
```bash
PAYLOAD='{"item":"Running Shoes","size":"10","qty":15,"inStock":true}'
SECRET=my-secret
echo -n "$PAYLOAD" | openssl dgst -sha256 -hmac "$SECRET"
```

Using the prototype's `/sign` helper:
```bash
curl "http://localhost:8080/sign?payload=%7B%22item%22%3A%22Running+Shoes%22%7D"
```

---

## Deliverables

| Assignment | File |
|---|---|
| Assignment 1 — mini-prototype | `cmd/prototype/main.go` |
| Assignment 1 — blocker journal | `BLOCKER_JOURNAL.md` |
| Assignment 2 — full deliverable | `cmd/server/main.go` |
| Assignment 2 — scope delta | `SCOPE_DELTA_ANALYSIS.md` |
| Deprecated polling code | `internal/warehouse/poller.go` |
