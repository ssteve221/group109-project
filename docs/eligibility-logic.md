# Returns & Refund Eligibility — Logic Design

Issue: #8
Author: _fill in your name_

## Purpose

Decide whether an order qualifies for a return, and produce a clear
customer-facing message either way — without a human needing to look it up.

## Input / output shape

**In:** `orderId` — raw user text, may have extra whitespace or wrong case.

**Out:**
- `found` — was an order with this ID located at all?
- `eligible` — only meaningful when `found` is true
- `daysLeft` — only present when eligible
- `message` — the actual reply shown to the customer

## Decision tree (pseudocode)

```
FUNCTION check_return_eligibility(orderId):
    id ← normalize(orderId)              # trim whitespace, uppercase
    order ← find order in dataset where order.id == id

    IF order not found:
        RETURN { found: false,
                 message: "couldn't find that order — check the ID" }

    IF order.status == "cancelled":
        RETURN { found: true, eligible: false,
                 message: "cancelled, nothing to return — refund should
                           already be issued; offer to flag to an agent
                           if it's missing" }

    IF order.status != "delivered":
        # still processing, shipped, or delayed
        RETURN { found: true, eligible: false,
                 message: "not delivered yet — not eligible until it
                           arrives, current status shown" }

    # order IS delivered — this is the only branch where the return
    # window actually applies
    daysSinceDelivery ← today - order.deliveredDate
    daysLeft ← RETURN_WINDOW_DAYS - daysSinceDelivery

    IF daysSinceDelivery <= RETURN_WINDOW_DAYS:
        RETURN { found: true, eligible: true, daysLeft,
                 message: "eligible — X days left in the window, how
                           to start a return, refund timing" }
    ELSE:
        RETURN { found: true, eligible: false,
                 message: "past the window — offer to connect with an
                           agent in case of an exception" }
```

## Constants

| Name | Value | Notes |
|---|---|---|
| `RETURN_WINDOW_DAYS` | 30 | Counted from delivery date, inclusive |
| `REFUND_PROCESSING_DAYS` | "5–7 business days" | Starts after the returned item is received, not after the return is initiated |

## Branch coverage checklist

Every branch below should have at least one test case in #13:

- [ ] Unknown order ID
- [ ] Cancelled order
- [ ] Not yet delivered (processing)
- [ ] Not yet delivered (shipped)
- [ ] Not yet delivered (delayed)
- [ ] Delivered, comfortably within window
- [ ] Delivered, exactly at the 30-day boundary
- [ ] Delivered, just past the window

## Design decisions worth flagging in review

- **Boundary is inclusive** — day 30 exactly still counts as eligible, not day 29. Worth confirming this is the intended rule, not just what the code happens to do.
- **Cancelled orders get an escalation offer**, past-window orders also get one — undelivered orders currently don't. Intentional, or a gap?
- **Missing/malformed `deliveredDate`** isn't handled — not currently a real data case, but worth a one-line note in the go-live note as a known edge case if `TODO` isn't resolved before submission.
