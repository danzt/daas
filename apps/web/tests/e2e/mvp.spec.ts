import { test } from "@playwright/test";

// E2E stubs — full MVP day-of-operation flow (S5-T9)
// Implemented in Sprint 5 final pass when the full stack is running.
//
// Acceptance: login → product → inventory → mixed sale → report → CSV

test("MVP full day — login with valid credentials redirects to dashboard", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("MVP full day — create a product and verify it appears in inventory with stock 0", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("MVP full day — receive a purchase order and verify stock increases", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("MVP full day — mixed sale decrements stock and creates internal + fiscal invoices", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("MVP full day — sales report shows the day's revenue after the sale", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("MVP full day — CSV export of the sales report downloads a non-empty file", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("MVP full day — logout clears session and redirects to /auth/login", async () => {
  test.skip(true, "TODO Sprint 5");
});
