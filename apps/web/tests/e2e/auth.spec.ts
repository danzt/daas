import { test } from "@playwright/test";

// E2E stubs — full scenarios implemented in Sprint 5
// when the full app is running end-to-end with a real backend.

test("login happy path — valid credentials redirect to /dashboard", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("login error — wrong password shows RFC 7807 detail message", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("register wizard step 1 — validation prevents advancing without required fields", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("register wizard step 2 — summary card shows step 1 data", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("register — successful registration auto-logs in and redirects to /dashboard", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("auth middleware — unauthenticated access to /dashboard redirects to /auth/login", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("auth middleware — authenticated user visiting /auth/login redirects to /dashboard", async () => {
  test.skip(true, "TODO Sprint 5");
});
test("logout — clears session and redirects to /auth/login", async () => {
  test.skip(true, "TODO Sprint 5");
});
