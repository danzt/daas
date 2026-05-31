import { test } from "@playwright/test";

// E2E stubs — full scenarios implemented in Sprint 5
// when the full app is running end-to-end with a real backend.
//
// Covers: M7-suppliers acceptance criteria

test("supplier list — shows all active suppliers with name and RIF", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("supplier create — missing name shows validation error", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("supplier create — duplicate RIF within tenant returns conflict error", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("supplier detail — shows contact info and purchase order history", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("purchase order create — draft PO is saved with status 'draft'", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("purchase order order — transitioning to 'ordered' sets ordered_at timestamp", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("purchase order receive — receiving increments product stock for each line", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("purchase order receive — receiving creates inventory_movements with type 'entry'", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("purchase order cancel — cancelling a draft does not affect inventory", async () => {
  test.skip(true, "TODO Sprint 5");
});
