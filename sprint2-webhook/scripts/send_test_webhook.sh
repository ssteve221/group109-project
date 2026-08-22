#!/bin/bash
# scripts/send_test_webhook.sh
#
# Test script for the Northstar Inventory Sync webhook server.
# Simulates the Northstar warehouse pushing stock updates.
#
# Usage:
#   ./scripts/send_test_webhook.sh                  — run all tests
#   WEBHOOK_SECRET=my-secret ./scripts/send_test_webhook.sh
#
# Requires: curl, openssl

set -euo pipefail

SECRET="${WEBHOOK_SECRET:-dev-secret-change-me}"
BASE_URL="${SERVER_URL:-http://localhost:9090}"
PROTOTYPE_URL="${PROTOTYPE_URL:-http://localhost:8080}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}  Northstar Webhook Test Script${NC}"
echo -e "${CYAN}  SECRET: ${SECRET:0:4}****${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Helper: compute HMAC-SHA256 signature
sign_payload() {
    local secret="$1"
    local payload="$2"
    local hex
    hex=$(echo -n "$payload" | openssl dgst -sha256 -hmac "$secret" | awk '{print $2}')
    echo "sha256=${hex}"
}

# Helper: test result printer
pass() { echo -e "  ${GREEN}✓ PASS${NC} — $1"; }
fail() { echo -e "  ${RED}✗ FAIL${NC} — $1"; }

# ─────────────────────────────────────────────────────────────────────────────
echo -e "${YELLOW}[1] Health Check — Full Server (port 9090)${NC}"
RESP=$(curl -sf "${BASE_URL}/health" 2>/dev/null) && \
    pass "Health check: $(echo $RESP | python3 -c 'import sys,json; d=json.load(sys.stdin); print(d["status"], "| mode:", d["mode"])')" || \
    fail "Server not running. Start with: WEBHOOK_SECRET=${SECRET} go run ./cmd/server/"

# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo -e "${YELLOW}[2] Test: Valid webhook payload${NC}"
PAYLOAD='{"item":"Running Shoes","size":"10","qty":15,"inStock":true}'
SIG=$(sign_payload "$SECRET" "$PAYLOAD")
echo "   Payload : ${PAYLOAD}"
echo "   Signature: ${SIG}"
RESP=$(curl -sf -X POST "${BASE_URL}/webhook" \
    -H "Content-Type: application/json" \
    -H "X-Hub-Signature-256: ${SIG}" \
    -d "$PAYLOAD" 2>/dev/null) && \
    pass "Valid payload accepted — response: $(echo $RESP | python3 -c 'import sys,json; d=json.load(sys.stdin); print(d["status"])')" || \
    fail "Valid payload was rejected"

# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo -e "${YELLOW}[3] Test: Invalid signature (tampered payload)${NC}"
TAMPERED='{"item":"Running Shoes","size":"10","qty":999,"inStock":true}'
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/webhook" \
    -H "Content-Type: application/json" \
    -H "X-Hub-Signature-256: ${SIG}" \
    -d "$TAMPERED" 2>/dev/null)
if [ "$HTTP_CODE" = "401" ]; then
    pass "Tampered payload rejected with HTTP 401 (expected)"
else
    fail "Expected HTTP 401, got ${HTTP_CODE}"
fi

# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo -e "${YELLOW}[4] Test: Missing signature header${NC}"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/webhook" \
    -H "Content-Type: application/json" \
    -d "$PAYLOAD" 2>/dev/null)
if [ "$HTTP_CODE" = "401" ]; then
    pass "Missing header rejected with HTTP 401 (expected)"
else
    fail "Expected HTTP 401, got ${HTTP_CODE}"
fi

# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo -e "${YELLOW}[5] Test: Stock query after webhook push${NC}"
RESP=$(curl -sf "${BASE_URL}/stock?item=Running+Shoes" 2>/dev/null) && \
    pass "Stock query: $(echo $RESP | python3 -c 'import sys,json; d=json.load(sys.stdin); print(d["count"], "result(s) for", repr(d["query"]))')" || \
    fail "Stock query failed"

# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo -e "${YELLOW}[6] Test: Multi-item push (batch simulation)${NC}"
ITEMS=(
    '{"item":"Trail Runner Jacket","size":"L","qty":5,"inStock":true}'
    '{"item":"Yoga Mat Pro","size":"Standard","qty":0,"inStock":false}'
    '{"item":"Ergonomic Office Chair","size":"Standard","qty":7,"inStock":true}'
)
ALL_PASS=true
for ITEM in "${ITEMS[@]}"; do
    SIG=$(sign_payload "$SECRET" "$ITEM")
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/webhook" \
        -H "Content-Type: application/json" \
        -H "X-Hub-Signature-256: ${SIG}" \
        -d "$ITEM" 2>/dev/null)
    if [ "$HTTP_CODE" != "200" ]; then
        ALL_PASS=false
        fail "Push failed for: ${ITEM}"
    fi
done
$ALL_PASS && pass "All 3 batch pushes accepted (HTTP 200)"

# ─────────────────────────────────────────────────────────────────────────────
echo ""
echo -e "${YELLOW}[7] Prototype Health Check (port 8080)${NC}"
RESP=$(curl -sf "${PROTOTYPE_URL}/health" 2>/dev/null) && \
    pass "Prototype healthy: $(echo $RESP | python3 -c 'import sys,json; d=json.load(sys.stdin); print(d["service"])')" || \
    echo -e "  ${YELLOW}⚠ Prototype not running (optional). Start with: WEBHOOK_SECRET=${SECRET} go run ./cmd/prototype/${NC}"

echo ""
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}  Tests complete.${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
