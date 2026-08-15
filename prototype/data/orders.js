// Mock order dataset — Task T03
// In production this would be replaced by a call to Northstar's real order-management API.
// Dates are relative to "today" = 2026-08-15 for demo purposes.

const ORDERS = [
  { id: "NS10023", item: "Trail Runner Jacket (M)", status: "shipped", shippedDate: "2026-08-12", estDelivery: "2026-08-17", carrier: "DHL", trackingUrl: "https://track.example.com/DHL10023", deliveredDate: null, orderedDate: "2026-08-10" },
  { id: "NS10024", item: "Ceramic Pour-Over Set", status: "processing", shippedDate: null, estDelivery: "2026-08-20", carrier: null, trackingUrl: null, deliveredDate: null, orderedDate: "2026-08-14" },
  { id: "NS10025", item: "Merino Wool Socks (3-pack)", status: "delivered", shippedDate: "2026-08-05", estDelivery: "2026-08-09", carrier: "FedEx", trackingUrl: "https://track.example.com/FDX10025", deliveredDate: "2026-08-08", orderedDate: "2026-08-02" },
  { id: "NS10026", item: "Standing Desk Converter", status: "delayed", shippedDate: null, estDelivery: "2026-08-25", carrier: null, trackingUrl: null, deliveredDate: null, orderedDate: "2026-08-01", delayReason: "Supplier restock delay" },
  { id: "NS10027", item: "Wireless Charging Pad", status: "shipped", shippedDate: "2026-08-13", estDelivery: "2026-08-16", carrier: "UPS", trackingUrl: "https://track.example.com/UPS10027", deliveredDate: null, orderedDate: "2026-08-11" },
  { id: "NS10028", item: "Cast Iron Skillet 10in", status: "delivered", shippedDate: "2026-07-28", estDelivery: "2026-08-01", carrier: "USPS", trackingUrl: "https://track.example.com/USPS10028", deliveredDate: "2026-07-31", orderedDate: "2026-07-25" },
  { id: "NS10029", item: "Running Shoes (Size 9)", status: "processing", shippedDate: null, estDelivery: "2026-08-19", carrier: null, trackingUrl: null, deliveredDate: null, orderedDate: "2026-08-14" },
  { id: "NS10030", item: "Leather Weekender Bag", status: "delivered", shippedDate: "2026-06-20", estDelivery: "2026-06-24", carrier: "DHL", trackingUrl: "https://track.example.com/DHL10030", deliveredDate: "2026-06-23", orderedDate: "2026-06-17" },
  { id: "NS10031", item: "Bluetooth Speaker Mini", status: "shipped", shippedDate: "2026-08-14", estDelivery: "2026-08-18", carrier: "FedEx", trackingUrl: "https://track.example.com/FDX10031", deliveredDate: null, orderedDate: "2026-08-12" },
  { id: "NS10032", item: "Yoga Mat Pro", status: "cancelled", shippedDate: null, estDelivery: null, carrier: null, trackingUrl: null, deliveredDate: null, orderedDate: "2026-08-09", cancelReason: "Customer requested cancellation" },
  { id: "NS10033", item: "Insulated Water Bottle 32oz", status: "delivered", shippedDate: "2026-08-01", estDelivery: "2026-08-05", carrier: "UPS", trackingUrl: "https://track.example.com/UPS10033", deliveredDate: "2026-08-04", orderedDate: "2026-07-29" },
  { id: "NS10034", item: "Ergonomic Office Chair", status: "processing", shippedDate: null, estDelivery: "2026-08-22", carrier: null, trackingUrl: null, deliveredDate: null, orderedDate: "2026-08-13" },
  { id: "NS10035", item: "Trail Runner Jacket (L)", status: "delivered", shippedDate: "2026-07-10", estDelivery: "2026-07-14", carrier: "DHL", trackingUrl: "https://track.example.com/DHL10035", deliveredDate: "2026-07-13", orderedDate: "2026-07-07" },
  { id: "NS10036", item: "Espresso Grinder", status: "delayed", shippedDate: null, estDelivery: "2026-08-28", carrier: null, trackingUrl: null, deliveredDate: null, orderedDate: "2026-08-03", delayReason: "Customs hold" },
  { id: "NS10037", item: "Running Shoes (Size 10)", status: "shipped", shippedDate: "2026-08-13", estDelivery: "2026-08-17", carrier: "USPS", trackingUrl: "https://track.example.com/USPS10037", deliveredDate: null, orderedDate: "2026-08-10" },
  { id: "NS10038", item: "Wool Beanie", status: "delivered", shippedDate: "2026-05-01", estDelivery: "2026-05-05", carrier: "FedEx", trackingUrl: "https://track.example.com/FDX10038", deliveredDate: "2026-05-04", orderedDate: "2026-04-28" },
  { id: "NS10039", item: "Camping Hammock", status: "processing", shippedDate: null, estDelivery: "2026-08-21", carrier: null, trackingUrl: null, deliveredDate: null, orderedDate: "2026-08-15" },
  { id: "NS10040", item: "Noise-Cancelling Headphones", status: "delivered", shippedDate: "2026-08-02", estDelivery: "2026-08-06", carrier: "UPS", trackingUrl: "https://track.example.com/UPS10040", deliveredDate: "2026-08-05", orderedDate: "2026-07-30" },
  { id: "NS10041", item: "Stainless Steel Water Flask", status: "shipped", shippedDate: "2026-08-14", estDelivery: "2026-08-19", carrier: "DHL", trackingUrl: "https://track.example.com/DHL10041", deliveredDate: null, orderedDate: "2026-08-12" },
  { id: "NS10042", item: "Compact Tripod", status: "delivered", shippedDate: "2026-04-10", estDelivery: "2026-04-14", carrier: "FedEx", trackingUrl: "https://track.example.com/FDX10042", deliveredDate: "2026-04-13", orderedDate: "2026-04-07" },
];

// Stock/size lookup — supports the stretch T12 flow
const STOCK = [
  { sku: "Trail Runner Jacket", size: "S", inStock: true, qty: 14 },
  { sku: "Trail Runner Jacket", size: "M", inStock: true, qty: 6 },
  { sku: "Trail Runner Jacket", size: "L", inStock: false, qty: 0, restockDate: "2026-08-29" },
  { sku: "Running Shoes", size: "9", inStock: true, qty: 22 },
  { sku: "Running Shoes", size: "10", inStock: false, qty: 0, restockDate: "2026-08-24" },
  { sku: "Running Shoes", size: "11", inStock: true, qty: 9 },
  { sku: "Ergonomic Office Chair", size: "Standard", inStock: true, qty: 3 },
  { sku: "Yoga Mat Pro", size: "Standard", inStock: true, qty: 40 },
];

if (typeof module !== "undefined") {
  module.exports = { ORDERS, STOCK };
}
