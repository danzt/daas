import { test } from "@playwright/test";

// E2E stubs — full scenarios implemented in Sprint 5 final pass
// when the full app is running end-to-end with a real backend.
//
// Covers: M8-reports acceptance criteria

test("sales report — shows total_revenue, total_invoices, internal_count, fiscal_count", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("sales report — by_day chart aggregates revenue per calendar day", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("sales report — top_products ranks by revenue descending", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("inventory report — shows quantity_on_hand per product with low-stock flag", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("inventory report — out_of_stock_count matches products with qty = 0", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("CSV export — sales report downloads a valid CSV file", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("reports date filter — changing date range updates all metrics", async () => {
  test.skip(true, "TODO Sprint 5");
});
