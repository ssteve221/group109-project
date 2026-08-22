# Assignment 1: Learning & Blocker Journal
**Tool: Webhook Verification | Language: Go**  
**Sprint 2 — The Meridian Pivot | Group 109**  
**Learner: Steve**

---

## Day 1–2 Log (Solo Recon Phase)

### What I Was Assigned

Webhook Verification — a system for proving that an HTTP request from a third party
is authentic and hasn't been tampered with in transit. I had not previously implemented
this myself (I had consumed webhooks from services like Stripe passively, but never
written the verification layer).

---

## Hour-by-Hour Log

### Hour 1 — Understanding the problem space

**Goal:** Know what webhook verification is and why it matters.

**Resources consulted:**
- GitHub Webhooks documentation: https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries
- Stripe webhook signing docs: https://stripe.com/docs/webhooks#verify-events
- HMAC Wikipedia article

**Key insight I had to learn from scratch:**
The sender and receiver share a secret string. The sender uses HMAC-SHA256 to create
a digest of the request body using that secret, then sends it as a header
(`X-Hub-Signature-256: sha256=<hex>`). The receiver does the same computation and
compares. If they match, the payload is authentic.

**First blocker:** I initially thought the signature covered the headers too. Wrong — 
it only covers the raw request body. Reading GitHub's docs carefully cleared this up.

---

### Hour 2 — First attempt in Go

**Goal:** Write a minimal HTTP handler that reads the body and checks the header.

**First error I hit:**
```
cannot use r.Body (variable of type io.ReadCloser) as type []byte
```
**Cause:** I tried to pass `r.Body` directly to `hmac.New().Write()`. 
**Fix:** Used `io.ReadAll(r.Body)` to read the body into a `[]byte` first.

**Second blocker — critical mistake:**
I initially read `r.Body` twice: once to check content-type, once to verify.
The second read returned empty bytes because `r.Body` is a streaming reader —
reading it once exhausts it.

**Fix:** Read the body once at the top of the handler and store it in a variable.
This is a fundamental Go HTTP pattern I didn't know before this sprint.

---

### Hour 3 — HMAC implementation

**Goal:** Implement HMAC-SHA256 in Go's standard library.

**Resources:**
- `crypto/hmac` package docs: https://pkg.go.dev/crypto/hmac
- `crypto/sha256` package docs: https://pkg.go.dev/crypto/sha256
- Go standard library source

**Code that didn't work first:**
```go
// Wrong — sha256.Sum256 returns [32]byte, not []byte
h := hmac.New(sha256.New, secret)
h.Write(body)
result := sha256.Sum256(h.Sum(nil)) // this is wrong
```

**Fix:** Use `mac.Sum(nil)` to get the HMAC digest, then `hex.EncodeToString()`:
```go
mac := hmac.New(sha256.New, []byte(secret))
mac.Write(body)
digest := hex.EncodeToString(mac.Sum(nil))
```

**Third blocker — timing attack:**
I initially wrote `receivedSig == computedSig` for comparison.
A staff member's resource (not help — just pointed me to crypto docs) led me to discover
`hmac.Equal()` which uses constant-time comparison to prevent timing attacks.
Without this, an attacker could measure response times to guess the secret byte by byte.

---

### Hour 4–5 — Building the prototype server

**Goal:** Make a working HTTP server with the verify endpoint.

**Error hit:**
```
pattern "POST /webhook": invalid pattern
```
**Cause:** Go 1.22+ introduced method-based routing (`"POST /path"`), but I initially
tested with Go 1.21 on the second machine. 

**Fix:** Confirmed `go version` — I'm running Go 1.26.6 which supports the new routing.
Added a version note to the README.

**Error hit:**
```
undefined: webhook.VerifySignature
```
**Cause:** My `internal/webhook/verify.go` had a lowercase function name `verifySignature`.
**Fix:** Exported it by capitalizing: `VerifySignature`.

---

### Hour 6 — Writing tests

**Goal:** Unit test the verify logic independently of the HTTP server.

**Blocker:** I wasn't sure how to test both `true` and `false` cases efficiently.
**Resource consulted:** Go testing documentation, table-driven test patterns.
**Fix:** Learned the table-driven test pattern — define a slice of test cases, loop over them.
This is idiomatic Go and makes adding edge cases trivial.

---

### Hour 7 — Integration and cleanup

**Goal:** Make the prototype clean, documented, and runnable.

**Mistakes corrected:**
1. Forgot to set `Content-Type: application/json` on responses — added it.
2. My error responses were plain text, but the spec examples use JSON — standardized.
3. Missing `defer r.Body.Close()` — added to prevent resource leaks.

---

## Summary: What I Learned

| Concept | Prior knowledge | After this sprint |
|---------|----------------|-------------------|
| HMAC-SHA256 theory | Knew it existed | Can implement from scratch |
| Go `crypto/hmac` package | Zero | Comfortable |
| Constant-time comparison (`hmac.Equal`) | Unaware | Essential for security |
| `io.ReadAll` + single-read pattern | Partial | Solid |
| Go table-driven tests | Heard of it | Can write them |
| HTTP method routing in Go 1.22+ | Unaware | Working |

## Dead Ends (What I Tried That Didn't Work)

1. **Tried `crypto/md5` first** — looked at old StackOverflow posts that used MD5 for webhook verification. Realized from GitHub/Stripe docs that SHA-256 is the current standard. Discarded MD5 approach.

2. **Tried a third-party JWT library** — briefly considered using JWT for payload signing. Realized webhook verification doesn't use JWT — it's simpler raw HMAC. No external dependency needed.

3. **Tried hex-encoding the HMAC key** — I misread a doc and thought the secret needed to be hex-encoded before passing to `hmac.New()`. Tests failed. Realized the secret is used as raw bytes.

---

*Total time (Days 1–2): approximately 7 hours solo*  
*No teammate or instructor how-to help was received during this phase.*
