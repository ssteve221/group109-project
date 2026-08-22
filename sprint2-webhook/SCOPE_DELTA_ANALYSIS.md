# Assignment 2: Scope Delta Analysis
**Project: Northstar Inventory Sync Service**  
**Sprint 2 — The Meridian Pivot | Group 109**  
**Learner: Steve | Tool: Webhook Verification (Go)**

---

## The Pivot

On Day 4, the client (Northstar Retail Co.) announced:

> *"The polling endpoint is being decommissioned in 48 hours. Switch to webhook push. Same deadline. No extensions."*

This document records exactly what changed, what was dropped, what was preserved, and what the trade-offs are.

---

## Backlog Reprioritization

| Item | Pre-Pivot Status | Post-Pivot Status | Action |
|------|-----------------|-------------------|--------|
| Background polling goroutine | ✅ Built (Day 3) | ❌ Deprecated | **DROPPED** — commented out in `internal/warehouse/poller.go` |
| `POST /webhook` endpoint | 🔲 Not planned | ✅ Built | **ADDED** — core of the pivot |
| HMAC-SHA256 signature verification | 🔲 Not planned | ✅ Built | **ADDED** — security requirement of push model |
| `GET /stock?item=...` query endpoint | ✅ Built (Day 3) | ✅ Unchanged | **PRESERVED** — cache serves same data, different ingestion path |
| In-memory stock cache | ✅ Built (Day 3) | ✅ Unchanged | **PRESERVED** — design was data-source-agnostic |
| 5-minute poll scheduler | ✅ Built (Day 3) | ❌ Deprecated | **DROPPED** — replaced by event-driven updates |
| Retry/backoff on poll failures | 🔲 Planned (Day 4) | ❌ Dropped | **DROPPED** — no longer relevant; push model doesn't poll |

---

## What Was Dropped

### 1. Background Polling Goroutine
**File:** `internal/warehouse/poller.go`  
**Status:** Code preserved in file but marked `// DEPRECATED`, package not imported in `cmd/server/main.go`.

The goroutine that launched `time.NewTicker(5 * time.Minute)` and called `http.Get(warehouseAPIURL)` is no longer started. The warehouse API endpoint it targeted is being decommissioned.

**Cost of dropping:**
- Automatic stock refresh on startup is lost (mitigation: cache seeded from Sprint 1 baseline on boot)
- If the webhook sender goes offline, the cache becomes stale (mitigation: future work — add `updatedAt` timestamp and staleness warning in `/stock` response)

---

### 2. Poll Retry Logic (was planned, never built)
Planned for Day 4 but deprioritized when the pivot was announced. Since polling is gone, retry/backoff for polling is no longer needed.

---

## What Was Modified

### `cmd/server/main.go`
- **Removed:** `StartPoller(stockCache, done)` goroutine launch (previously called in `main()`)
- **Added:** `POST /webhook` handler with signature verification and cache update
- **Added:** Comment block at the top explaining what changed and why

### `GET /stock` endpoint
- **Behavior:** Unchanged
- **Data source:** Changed from poll-refreshed cache → webhook-refreshed cache
- **Consumer impact:** Zero — the endpoint contract is identical

---

## What Was Added

### `POST /webhook` — New Endpoint
Receives signed JSON from the warehouse whenever a stock level changes. The payload schema:
```json
{
  "item": "Running Shoes",
  "size": "10",
  "qty": 15,
  "inStock": true
}
```

Verification flow:
1. Read raw body (before any parsing — raw bytes must match what was signed)
2. Extract `X-Hub-Signature-256` header
3. Compute `HMAC-SHA256(secret, body)` and compare with `hmac.Equal()` (constant-time)
4. If valid: parse JSON, update cache, return `200 OK`
5. If invalid: return `401 Unauthorized` (no payload detail leaked)

### `internal/webhook/verify.go` — New Package
Extracted signature logic into a testable package. Includes `VerifySignature()` and `ComputeSignature()`.

---

## Regression Check

| Original Feature | Status After Pivot |
|-----------------|-------------------|
| `GET /stock?item=Running+Shoes` returns results | ✅ Works — cache is same, path of population changed |
| `GET /health` returns 200 | ✅ Works — now reports `"mode": "webhook-push"` |
| Invalid requests get proper error codes | ✅ Works — added 401 for bad signatures, 400 for malformed payloads |
| Stock data is scoped to Northstar items | ✅ Works — only warehouse-pushed items enter the cache |

**No pre-pivot functionality was broken.** The `GET /stock` endpoint is functionally identical from the consumer's perspective.

---

## Trade-Off Documentation

| Trade-Off | Decision | Rationale |
|-----------|----------|-----------|
| Real-time vs. scheduled updates | Push (webhook) wins | Eliminates 5-min data lag; updates are immediate |
| Complexity: signing both sides | Accepted | The warehouse handles sending; we only verify |
| Cache staleness | Risk accepted for demo | Production would add staleness TTL + alerting |
| No persistent storage | Accepted | In-memory is sufficient for sprint; Redis for production |
| Replay attack protection | Not implemented | Production would add `X-Timestamp` + nonce checking |

---

## What Would Come Next (If Not Cut by Deadline)

1. **Timestamp verification** — reject webhooks older than 5 minutes to prevent replay attacks
2. **Idempotency key** — ignore duplicate payloads (same event pushed twice)
3. **Dead letter queue** — if cache update fails, log to retry queue
4. **Staleness indicator** — `/stock` response to include `"dataAgeSeconds"` so consumers know how fresh the data is

---

*This document satisfies the Assignment 2 Scope Delta Analysis requirement.*  
*Obsolete polling code is in `internal/warehouse/poller.go`, fully commented out and marked DEPRECATED.*
