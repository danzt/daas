# DaaS — Multi-country Inventory Management SaaS

Multi-tenant inventory SaaS for PyMEs with dual fiscal/non-fiscal invoicing.

**Markets**: Venezuela -> Dominican Republic -> USA

## Stack

| Layer | Tech |
|-------|------|
| Frontend | Vue 3 + Nuxt 3 + shadcn-vue + Pinia + Tailwind 4 |
| Backend | Go (Echo) — hexagonal architecture |
| Database | Supabase (PostgreSQL 17 + Auth) |
| DB Tooling | sqlc + golang-migrate |
| Monorepo | pnpm workspaces |
| Testing | go test + testcontainers + Vitest + Playwright |
| CI/CD | GitHub Actions + Hetzner VPS + Docker Compose |

## Quick Start

```bash
# 1. Install dependencies
pnpm install

# 2. Copy env files
cp .env.example .env
cp apps/api/.env.example apps/api/.env
# Fill in Supabase credentials

# 3. Start dev services
docker compose up postgres -d

# 4. Run API
cd apps/api && make run

# 5. Run web
pnpm --filter @daas/web dev
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `SUPABASE_URL` | Supabase project URL |
| `SUPABASE_ANON_KEY` | Supabase anon/publishable key |
| `SUPABASE_SERVICE_KEY` | Supabase service role key (server-side only) |
| `JWT_SECRET` | JWT signing secret |
| `DATABASE_URL` | PostgreSQL connection string |
| `INTEGRATION_ENCRYPTION_KEY` | AES-256-GCM key for fiscal API credentials |

## Project Structure

```
apps/
  api/          # Go backend (hexagonal architecture)
  web/          # Nuxt 3 frontend
  landing/      # Marketing landing page
packages/
  ui/           # Shared Vue components + cn() utility
  types/        # Shared TypeScript types
  config/       # Shared ESLint + Prettier + TypeScript config
docs/
  adr/          # Architecture Decision Records
```

## ADRs

See [docs/adr/000-index.md](docs/adr/000-index.md) for all architecture decisions.
