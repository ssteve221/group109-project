# Conversation Flows — Order Status & Returns

Issue: #4
Author: Stephen

Covers the two required flows.

## Two ways into any flow

The widget supports both a guided path and a free-text path:

1. **Chip-guided** — user taps a suggested button ("📦 Track my order" /
   "↩️ Start a return"), the bot asks a targeted follow-up, done.
2. **Free text** — user types straight into the box with no chip tapped.
   Intent is inferred from keywords in the message itself.

Both paths converge on the same lookup logic once an order ID is known.

---

## Flow 1 — Order Status

### Entry triggers
- Chip: "📦 Track my order"
- Free text containing an order ID pattern (`NS` followed by 4–6 digits) with no `return`/`refund` keyword present

### Decision tree
```
User enters flow
  → Bot: "Sure — what's your order ID? (e.g. NS10023)"
      [only shown on chip-guided entry; free-text entry skips straight
       to the ID check below]

  → Does the message contain an order ID (NS + 4-6 digits)?
      NO  → Bot: "I need an order ID to look that up — it looks like
                   NS10023. Where can I find it? It's in your
                   confirmation email."
             (flow stays open, waits for the ID)
      YES → look up the order:
            NOT FOUND     → "I couldn't find an order with ID '...'.
                              Double-check the order number..."
            processing    → prepared / estimated delivery date
            shipped       → ship date, carrier, estimated delivery,
                              + tracking link
            delivered     → delivered date
            delayed       → delay reason (if on file), new estimate
            cancelled     → cancellation reason (if on file)
```

### Sample dialogue — chip-guided, happy path
> **Bot:** Hi — I can check order status or return & refund questions right now. What do you need?
> **User:** *(taps "📦 Track my order")*
> **Bot:** Sure — what's your order ID? (e.g. NS10023)
> **User:** NS10025
> **Bot:** Order NS10025 (item) shipped on [date] via [carrier], estimated delivery [date]. **Track package →**

### Sample dialogue — free text, happy path
> **User:** where's my order NS10029
> **Bot:** *(no order-ID prompt needed — ID already present)* Order NS10029 (item) is being prepared. Estimated ship-by date puts delivery around [date].

### Sample dialogue — edge case, no ID provided
> **User:** *(taps "📦 Track my order")*
> **Bot:** Sure — what's your order ID? (e.g. NS10023)
> **User:** I don't have it on me
> **Bot:** I need an order ID to look that up — it looks like NS10023. Where can I find it? It's in your confirmation email.
> *(flow stays open — bot waits for a valid ID on the next message rather than giving up)*

---

## Flow 2 — Returns & Refunds

### Entry triggers
- Chip: "↩️ Start a return"
- Free text containing `return` or `refund`

Full eligibility branching is documented separately in
[`eligibility-logic.md`](./design/eligibility-logic.md) — this doc covers
the conversation shape, that one covers the decision rules.

### Decision tree
```
User enters flow
  → Bot: "No problem — what's the order ID for the item you'd like
          to return?"
      [chip-guided only, same as order-status]

  → Does the message contain an order ID?
      NO  → same "I need an order ID..." prompt as order-status
      YES → check eligibility (see eligibility-logic.md):
            NOT FOUND        → "couldn't find an order with ID '...'
                                 to check return eligibility"
            cancelled        → "cancelled, nothing to return..."
            not delivered    → "hasn't been delivered yet..."
            within window    → "eligible — X days left..." +
                                 Start your return link
            past window      → "past our 30-day return window..."
```

### Sample dialogue — eligible, happy path
> **User:** *(taps "↩️ Start a return")*
> **Bot:** No problem — what's the order ID for the item you'd like to return?
> **User:** NS10040
> **Bot:** Order NS10040 (item) is eligible for return — delivered 10 day(s) ago, 20 day(s) left in the 30-day window. Start a return from your Orders page; refunds are processed within 5–7 business days after we receive the returned item. **Start your return →**

### Sample dialogue — ineligible, past window
> **User:** I want to return NS10038, is that still possible?
> **Bot:** Order NS10038 (item) was delivered [N] days ago, which is past our 30-day return window. I can connect you with a support agent to check for exceptions.

### Sample dialogue — ineligible, not yet delivered
> **User:** *(taps "↩️ Start a return")*
> **Bot:** No problem — what's the order ID for the item you'd like to return?
> **User:** NS10024
> **Bot:** Order NS10024 (item) hasn't been delivered yet (current status: processing), so it isn't eligible for return until it arrives.

---

## Known gap (tracked separately, issue #11)

If free text matches *neither* the returns keywords nor a `stock`/`size`
keyword, the router currently assumes order-status intent by default rather
than asking. That means a genuinely unrelated message still gets routed as
if it were an order lookup instead of being flagged. Fix tracked in #11 —
not re-documented here since it changes the router, not these two flows.
