import { test } from "@playwright/test";

// E2E stubs — full scenarios implemented in Sprint 5 final pass
// when the full app is running end-to-end with a real backend.
//
// Covers: M6-mixed-sales acceptance criteria

test("sale order create — draft is saved with no stock deduction", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("sale order confirm — confirms draft and locks quantities", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("mixed sale — non-fiscal lines route to internal invoice", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("mixed sale — fiscal lines route to fiscal invoice queue", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("mixed sale — stock is atomically decremented for all lines on confirm", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("mixed sale — insufficient stock returns ErrInsufficientStock and rolls back", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("sale order cancel — cancellation restores inventory if previously confirmed", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("sales screen — POS-like cart shows fiscal/non-fiscal badge per product", async () => {
  test.skip(true, "TODO Sprint 5");
});
