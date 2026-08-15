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

## ⚠️ Important note on this starter commit

This repository was scaffolded in one working session to get the team a running starting point
(mock data, lookup functions, widget UI, tests, go-live note draft) so Day 2 can start from
working code instead of a blank folder. **That scaffolding is reflected as a single starter
commit below — it is not, and should not be mistaken for, five people's individual work.**

For Assignment 2's audit trail to mean anything, **every team member needs to commit their own
subsequent changes under their own git identity**, referencing their board task ID. Before you
start work:

```bash
git config user.name "Your Full Name"
git config user.email "your.email@example.com"
```

Then commit your own work in small, task-linked increments using the convention below — don't
have one person commit on everyone's behalf.

## Commit convention (per the Team Charter)

```
<type>: <what changed> — <why it matters>
```
`<type>` ∈ `feat | fix | docs | test | chore | refactor`. No `wip` / `updates`.
Example: `feat: add stock lookup for size queries — covers T12 stretch flow`

Branch naming: `<initials>/<task-id>-<short-desc>` (e.g. `sw/T13-golive-note`)

## Generating the real audit log

Once the team has been committing for a few days, run:
```bash
bash audit/generate-audit-log.sh
```
This reads actual `git log` output — author, timestamp, message — and writes
`audit/audit-log-export.csv`. That's the file that goes into Assignment 2's submission, alongside
board-movement timestamps exported from your project board tool. See `audit/audit-log-template.csv`
for the expected shape if you're also logging non-code contributions (deck edits, board updates).
