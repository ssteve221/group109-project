# Northstar Support Deflection MVP — Group 109

A rule-based support widget that resolves **order status** and **returns/refund** questions
without a ticket. Built for the Northstar Sprint (Assignment 2).

## Quick start

Open `prototype/index.html` directly in a browser — no server or build step required.
Try order IDs `NS10023` (shipped), `NS10024` (processing), `NS10025` (delivered/eligible for
return), `NS10038` (delivered, past return window).

Run the test suite:
```
node prototype/tests/test-lookup.js
```

## Repo structure

```
prototype/
  index.html          — chat widget UI (T08, T09)
  lookup.js            — order-status + returns + stock lookup functions (T06, T07, T12)
  data/orders.js        — mock order dataset (T03)
  tests/
    test-lookup.js       — automated test script (T10)
    test-results.csv       — output of the last test run
docs/
  go-live-note.md      — 1-page readiness note (T13)
audit/
  audit-log-template.csv  — structure for the real contribution log
  generate-audit-log.sh   — pulls the REAL log from git history (run this on Day 4 and Day 5)
```

