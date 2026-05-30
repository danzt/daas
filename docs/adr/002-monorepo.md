# ADR-002: Monorepo with pnpm workspaces

Status: Accepted

## Decision

Use a single monorepo with pnpm workspaces (apps/web, apps/landing, apps/api, packages/ui, packages/types, packages/config) to minimize operational overhead for a solo developer — enabling atomic commits across frontend and backend, a single CI workflow with path filters, and zero additional tooling compared to Nx/Turborepo.
