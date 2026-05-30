# ADR-011: Testing Strategy

Status: Accepted

## Decision

Layer tests across: Go unit tests (domain services, 80%+ coverage target), Go integration tests with testcontainers-go against real PostgreSQL (all repository queries), dedicated RLS isolation tests (two tenants, verify tenant A cannot see tenant B data — mandatory in CI), Vue unit tests with Vitest (composables 80%+, components selective), and Playwright E2E for the critical happy path (login -> create product -> sell -> invoice). CI fails on any test failure; coverage is reported but does not block to avoid gaming.
