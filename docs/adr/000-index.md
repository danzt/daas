# Architecture Decision Records — Index

All architecture decisions for DaaS are recorded here.

| ADR                                    | Title                                                                 | Status   |
| -------------------------------------- | --------------------------------------------------------------------- | -------- |
| [ADR-001](001-echo.md)                 | Go Framework — Echo over Fiber/Chi                                    | Accepted |
| [ADR-002](002-monorepo.md)             | Monorepo with pnpm workspaces                                         | Accepted |
| [ADR-003](003-hexagonal-backend.md)    | Backend Go structure — Hexagonal/Clean Architecture                   | Accepted |
| [ADR-004](004-multi-tenant-rls.md)     | Multi-Tenant — shared-schema + PostgreSQL RLS + Go middleware         | Accepted |
| [ADR-005](005-invoicing-adapter.md)    | Invoicing Adapter Pattern — FiscalGateway port per country            | Accepted |
| [ADR-006](006-mixed-sale-splitting.md) | Mixed Sale Splitting — two separate documents per sale                | Accepted |
| [ADR-007](007-database-sqlc.md)        | Database — Supabase PostgreSQL + sqlc + golang-migrate (no ORM)       | Accepted |
| [ADR-008](008-nuxt3-structure.md)      | Frontend — Nuxt 3 structure + Pinia + shadcn-vue                      | Accepted |
| [ADR-009](009-auth-flow.md)            | Auth Flow — Supabase JWT + Go JWKS verification + custom tenant claim | Accepted |
| [ADR-010](010-api-design.md)           | API Design — REST + cursor pagination + RFC 7807 errors               | Accepted |
| [ADR-011](011-testing-strategy.md)     | Testing Strategy — unit + integration + E2E + RLS isolation tests     | Accepted |
| [ADR-012](012-cicd.md)                 | CI/CD Pipeline — GitHub Actions + Hetzner VPS + Docker Compose        | Accepted |
