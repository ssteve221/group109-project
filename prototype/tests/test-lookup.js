// T10 — Test script for order-status and returns lookup functions.
// Run with: node prototype/tests/test-lookup.js
// Writes a pass/fail log to prototype/tests/test-results.csv (used as evidence in the audit log).

const path = require("path");
const fs = require("fs");
const { ORDERS, STOCK } = require("../data/orders.js");
const { getOrderStatus, getReturnEligibility, getStockStatus } = require("../lookup.js");

const cases = [
  // Order Status — happy paths
  { flow: "order_status", input: "NS10023", expectContains: "shipped", desc: "shipped order returns carrier + ETA" },
  { flow: "order_status", input: "NS10024", expectContains: "being prepared", desc: "processing order returns prep status + estimated delivery" },
  { flow: "order_status", input: "NS10025", expectContains: "delivered", desc: "delivered order returns delivery date" },
  { flow: "order_status", input: "NS10026", expectContains: "delayed", desc: "delayed order surfaces delay reason" },
  { flow: "order_status", input: "NS10032", expectContains: "cancelled", desc: "cancelled order under order-status flow" },
  { flow: "returns", input: "NS99999", expectContains: "couldn't find", desc: "unknown order ID under returns flow" },
  { flow: "returns", input: "NS10035", expectContains: "past our", desc: "boundary: delivered 33 days ago, just past window" },
  // Order Status — edge cases
  { flow: "order_status", input: "ns10023", expectContains: "shipped", desc: "lowercase order ID still resolves (case-insensitive)" },
  { flow: "order_status", input: "NS99999", expectContains: "couldn't find", desc: "unknown order ID handled gracefully, no crash" },

  // Returns — happy paths
  { flow: "returns", input: "NS10025", expectContains: "eligible for return", desc: "delivered order within window is return-eligible" },
  { flow: "returns", input: "NS10038", expectContains: "past our", desc: "delivered order past window is correctly denied" },
  // Returns — edge cases
  { flow: "returns", input: "NS10024", expectContains: "hasn't been delivered", desc: "processing order correctly blocked from return" },
  { flow: "returns", input: "NS10032", expectContains: "cancelled", desc: "cancelled order returns cancellation message, not a return flow" },

  // Stock (stretch) — spot check
  { flow: "stock", input: "Running Shoes size 10", expectContains: "out of stock", desc: "out-of-stock size correctly flagged with restock date" },
  { flow: "stock", input: "Trail Runner Jacket size M", expectContains: "in stock", desc: "in-stock size returns quantity" },
];

const results = cases.map((c, i) => {
  let actual;
  if (c.flow === "order_status") actual = getOrderStatus(c.input, ORDERS).message;
  else if (c.flow === "returns") actual = getReturnEligibility(c.input, ORDERS).message;
  else actual = getStockStatus(c.input, STOCK).message;

  const pass = actual.toLowerCase().includes(c.expectContains.toLowerCase());
  return {
    id: `TC${String(i + 1).padStart(2, "0")}`,
    flow: c.flow,
    input: c.input,
    description: c.desc,
    expectContains: c.expectContains,
    actual,
    result: pass ? "PASS" : "FAIL",
  };
});

const failCount = results.filter((r) => r.result === "FAIL").length;

// Print to console
results.forEach((r) =>
  console.log(`[${r.result}] ${r.id} (${r.flow}) — ${r.description}`)
);
console.log(`\n${results.length - failCount}/${results.length} passed.`);

// Write CSV log
const header = "TestID,Flow,Input,Description,ExpectedContains,ActualResponse,Result\n";
const rows = results
  .map(
    (r) =>
      `${r.id},${r.flow},"${r.input}","${r.description}","${r.expectContains}","${r.actual.replace(/"/g, '""')}",${r.result}`
  )
  .join("\n");

fs.writeFileSync(path.join(__dirname, "test-results.csv"), header + rows + "\n");
console.log("\nWrote prototype/tests/test-results.csv");

process.exit(failCount > 0 ? 1 : 0);
