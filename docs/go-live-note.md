# Northstar Support Deflection MVP — Go-Live Readiness Note

**Prepared by:** Group 109 · **Date:** [insert Day 5 date] · **For:** Northstar Retail Co. support ops

## What works

- **Order status lookup** — customer provides an order ID (or types it into any free-text message),
  bot returns current status (processing / shipped / delivered / delayed / cancelled), carrier,
  and tracking link where available. Handles unknown IDs and lowercase input gracefully.
- **Returns & refund eligibility** — checks delivery date against a 30-day return window, tells
  the customer whether they're eligible, how many days are left, and the refund processing
  timeframe. Correctly blocks returns on undelivered or cancelled orders.
- **Stock/size check (stretch)** — free-text item + size lookup against mock inventory, returns
  in-stock quantity or restock date.
- **12/12 automated test cases passing** (`prototype/tests/test-lookup.js` →
  `prototype/tests/test-results.csv`), covering happy paths and edge cases (unknown order ID,
  case-insensitive input, cancelled/undelivered orders).
- Runs entirely client-side — no backend server required for the demo; opens directly in a browser.

## What's known-broken / not done

- **Data source is mocked.** `prototype/data/orders.js` is a static 20-row sample. Before go-live
  this must be replaced with a live call to Northstar's actual order-management system (REST API
  or DB read) — the lookup functions (`lookup.js`) are already written to accept an `orders` array
  as a parameter, so swapping the source shouldn't require rewriting the logic, just the data-fetch
  layer.
- **No escalation-to-human path.** If the bot can't resolve a query (unmatched intent, no order ID
  found after one retry), it currently just repeats a prompt — there's no "hand off to a live agent"
  action wired in. This needs a real integration with Northstar's ticketing system before launch.
- **No authentication/verification.** Anyone who has an order ID can query its status — there's no
  check that the requester is the account owner. Northstar should decide whether to gate this
  behind login or add a secondary verification (email/zip match) before exposing it publicly.
- **Return window (30 days) and refund timing are hardcoded placeholders** — confirm against
  Northstar's actual policy and make configurable rather than a constant in `lookup.js`.
- **No analytics/logging of what gets deflected vs. escalated** — needed to actually measure
  ticket reduction; currently nothing is captured outside the browser session.
- **Not tested for concurrent/production load** — this was built and tested as a single-session
  demo only.

## What Northstar's team needs to pick this up without us in the room

1. **Swap the data layer:** replace `prototype/data/orders.js` with a fetch call to your order
   API. Keep the same object shape (see fields in that file) so `lookup.js` doesn't need changes.
2. **Add the escalation button:** wire a "Talk to a person" action in `index.html` that creates a
   ticket in your existing system, pre-filled with the order ID and conversation so far.
3. **Review the return-window and refund-timing constants** at the top of `lookup.js` and update
   to your real policy.
4. **Run `node prototype/tests/test-lookup.js`** after any change to `lookup.js` or the data shape —
   it's fast, deterministic, and will catch regressions before deploy.
5. **Decide on auth** before making this externally accessible — see "known-broken" above.
6. Full commit history and task-to-commit mapping are in `/audit` for context on why specific
   decisions were made.
