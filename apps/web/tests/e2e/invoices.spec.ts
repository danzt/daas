import { test } from "@playwright/test";

// E2E stubs — full scenarios implemented in Sprint 5
// when the full app is running end-to-end with a real backend.
//
// Covers: M4-internal-invoicing acceptance criteria

test("invoice list — only non-fiscal products are selectable in the line form", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("invoice create — adding a fiscal product returns ErrFiscalProductForbidden error", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("invoice create — draft invoice is saved with status 'draft' and no correlative", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("invoice issue — issuing a draft generates INT-YYYY-NNNNN correlative", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("invoice issue — issuing decrements product stock by the line quantities", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("invoice cancel — cancelling an issued invoice restores product stock", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("invoice cancel — cancelling a draft does not touch stock", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("invoice detail — shows customer info, lines, subtotal, and correlative", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("invoice list — filter by status shows only matching invoices", async () => {
  test.skip(true, "TODO Sprint 5");
});
