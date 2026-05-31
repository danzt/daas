import { test } from "@playwright/test";

// E2E stubs — full scenarios implemented in Sprint 5 final pass
// when the full app is running end-to-end with a real backend.
//
// Covers: M5-fiscal-invoicing acceptance criteria

test("fiscal invoice create — only fiscal products are selectable", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("fiscal invoice create — non-fiscal product returns ErrNonFiscalProductForbidden", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("fiscal invoice totals — subtotal_base, tax_amount, and total computed correctly", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("fiscal invoice send — successful SENIAT response sets status to issued", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("fiscal invoice send — SENIAT timeout queues invoice for retry", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("fiscal invoice retry — manual retry resets status to pending_fiscal", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("fiscal invoice retry — exceeding max_attempts sets status to failed", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("fiscal invoice cancel — cancellation is blocked once status is issued", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("fiscal invoices list — pending tray shows only pending_fiscal and failed", async () => {
  test.skip(true, "TODO Sprint 5");
});
