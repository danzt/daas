import { test } from "@playwright/test";

// E2E stubs — full scenarios implemented in Sprint 5
// when the full app is running end-to-end with a real backend.
//
// Acceptance criteria: M3-e2e-entrada-stock-salida
// Covers: purchase-order entry → stock update → sale exit flow

test("inventory entry — receiving a purchase order increments product stock", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("inventory entry — stock row is auto-initialized at 0 when a product is created", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("inventory adjustment — owner can apply positive delta and notes are required", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("inventory adjustment — negative delta is blocked when it would leave stock below 0", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("inventory adjustment — stock badge updates immediately after a successful adjustment", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("inventory exit — creating a sale order decrements product stock", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("inventory exit — stock cannot go negative via sale (ErrInsufficientStock returned)", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("inventory movements — audit log shows entry, adjustment, and exit in chronological order", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("inventory movements — PATCH/PUT/DELETE on a movement return 405 (immutability)", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("inventory low-stock filter — products below threshold appear when filter is active", async () => {
  test.skip(true, "TODO Sprint 5");
});
