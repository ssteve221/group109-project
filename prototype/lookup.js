// Lookup functions — Tasks T06 (order status) and T07 (returns/refund eligibility)
// Pure functions, no UI/DOM dependency, so they're independently testable (see /prototype/tests).

const RETURN_WINDOW_DAYS = 30;
const REFUND_PROCESSING_DAYS = "5–7 business days after we receive the returned item";
const TODAY = new Date("2026-08-15"); // fixed "now" for demo determinism

function daysBetween(dateStr, ref = TODAY) {
  const d = new Date(dateStr);
  return Math.floor((ref - d) / (1000 * 60 * 60 * 24));
}

/**
 * T06 — Order Status lookup
 * @param {string} orderId
 * @returns {object} { found, status, message, trackingUrl }
 */
function getOrderStatus(orderId, orders = ORDERS) {
  const id = String(orderId || "").trim().toUpperCase();
  const order = orders.find((o) => o.id === id);

  if (!order) {
    return {
      found: false,
      message: `I couldn't find an order with ID "${orderId}". Double-check the order number in your confirmation email — it usually starts with "NS".`,
    };
  }

  switch (order.status) {
    case "processing":
      return {
        found: true,
        status: "processing",
        message: `Order ${order.id} (${order.item}) is being prepared. Estimated ship-by date puts delivery around ${order.estDelivery}.`,
      };
    case "shipped":
      return {
        found: true,
        status: "shipped",
        message: `Order ${order.id} (${order.item}) shipped on ${order.shippedDate} via ${order.carrier}, estimated delivery ${order.estDelivery}.`,
        trackingUrl: order.trackingUrl,
      };
    case "delivered":
      return {
        found: true,
        status: "delivered",
        message: `Order ${order.id} (${order.item}) was delivered on ${order.deliveredDate}.`,
      };
    case "delayed":
      return {
        found: true,
        status: "delayed",
        message: `Order ${order.id} (${order.item}) is delayed — ${order.delayReason || "no reason on file"}. New estimated delivery: ${order.estDelivery}.`,
      };
    case "cancelled":
      return {
        found: true,
        status: "cancelled",
        message: `Order ${order.id} (${order.item}) was cancelled. Reason on file: ${order.cancelReason || "not specified"}.`,
      };
    default:
      return {
        found: true,
        status: order.status,
        message: `Order ${order.id} (${order.item}) status: ${order.status}.`,
      };
  }
}

/**
 * T07 — Returns / refund eligibility lookup
 * @param {string} orderId
 * @returns {object} { found, eligible, message }
 */
function getReturnEligibility(orderId, orders = ORDERS) {
  const id = (orderId || "").trim().toUpperCase();
  const order = orders.find((o) => o.id === id);

  if (!order) {
    return {
      found: false,
      message: `I couldn't find an order with ID "${orderId}" to check return eligibility.`,
    };
  }

  if (order.status === "cancelled") {
    return {
      found: true,
      eligible: false,
      message: `Order ${order.id} was cancelled, so there's nothing to return — any payment should already be refunded. If you don't see that refund, I can flag this to a support agent.`,
    };
  }

  if (order.status !== "delivered") {
    return {
      found: true,
      eligible: false,
      message: `Order ${order.id} (${order.item}) hasn't been delivered yet (current status: ${order.status}), so it isn't eligible for return until it arrives.`,
    };
  }

  const daysSinceDelivery = daysBetween(order.deliveredDate);
  const withinWindow = daysSinceDelivery <= RETURN_WINDOW_DAYS;
  const daysLeft = RETURN_WINDOW_DAYS - daysSinceDelivery;

  if (withinWindow) {
    return {
      found: true,
      eligible: true,
      daysLeft,
      message: `Order ${order.id} (${order.item}) is eligible for return — delivered ${daysSinceDelivery} day(s) ago, ${daysLeft} day(s) left in the ${RETURN_WINDOW_DAYS}-day window. Start a return from your Orders page; refunds are processed within ${REFUND_PROCESSING_DAYS}.`,
    };
  }

  return {
    found: true,
    eligible: false,
    message: `Order ${order.id} (${order.item}) was delivered ${daysSinceDelivery} days ago, which is past our ${RETURN_WINDOW_DAYS}-day return window. I can connect you with a support agent to check for exceptions.`,
  };
}

/**
 * T12 (stretch) — Stock / size availability lookup
 */
function getStockStatus(query, stock = STOCK) {
  const q = (query || "").toLowerCase();
  const matches = stock.filter(
    (s) => q.includes(s.sku.toLowerCase()) || q.includes(s.size.toLowerCase())
  );

  if (matches.length === 0) {
    return {
      found: false,
      message: `I couldn't match "${query}" to a product in our catalog. Try naming the item and size, e.g. "Running Shoes size 10".`,
    };
  }

  const lines = matches.map((s) =>
    s.inStock
      ? `${s.sku} (${s.size}): in stock — ${s.qty} available.`
      : `${s.sku} (${s.size}): out of stock — expected back around ${s.restockDate}.`
  );

  return { found: true, message: lines.join(" ") };
}

if (typeof module !== "undefined") {
  module.exports = { getOrderStatus, getReturnEligibility, getStockStatus, RETURN_WINDOW_DAYS };
}
