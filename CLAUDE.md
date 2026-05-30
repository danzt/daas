# DaaS — Multi-country Inventory Management SaaS

## Project Recovery (MANDATORY FIRST ACTION)

This is a new project. All planning artifacts are stored in Engram under project `conluc-soccer-frontend` with `scope: personal`. 

**On session start, recover context by searching these observation IDs:**

```
mem_get_observation(id: 178)  # Recovery guide (has all IDs and context)
mem_get_observation(id: 171)  # Proposal
mem_get_observation(id: 175)  # Spec (8 modules)
mem_get_observation(id: 176)  # Design (12 ADRs)
mem_get_observation(id: 174)  # Design system
mem_get_observation(id: 172)  # Stack decisions
mem_get_observation(id: 173)  # Design preferences
```

## What Is DaaS

SaaS multi-tenant de gestion de inventario para PyMEs con facturacion dual (fiscal/no-fiscal).

- **Markets**: Venezuela (first) -> Dominican Republic -> USA
- **Core differentiator**: Each product is flagged fiscal or non-fiscal, with different pricing and tax treatment
- **Solo dev**: Daniel

## Stack

- Frontend: Vue 3 + Nuxt 3 + shadcn-vue + Pinia
- Backend: Go (Echo) — hexagonal/clean architecture
- DB: Supabase (PostgreSQL managed + Auth) — Supabase as infra only, Go owns ALL logic
- DB tooling: sqlc + golang-migrate (zero ORM)
- Monorepo: pnpm workspaces
- Testing: go test + testcontainers + Vitest + Playwright
- CI/CD: GitHub Actions + Hetzner VPS + Docker Compose
- Mobile (future): Capacitor
- Desktop (future): Tauri

## Design System

See `design-system/MASTER.md` for full design system.

- Style: Flat Design, enterprise-level, mobile-first
- Colors: Primary #7C3AED (purple), CTA #F97316 (orange), BG #FAF5FF
- Fonts: Fira Code (headings) + Fira Sans (body)
- Components: shadcn-vue + Lucide icons
- Reference: https://github.com/arhamkhnz/next-shadcn-admin-dashboard-baseui

## UI/UX Rules

- Enterprise-level design ALWAYS. Nothing simple or generic.
- Mobile-first: Amazon/Alibaba quality standard
- Use ui-ux-pro-max skill for ALL design decisions
- Testing is MANDATORY

## SDD Status

- [x] Proposal (sdd/daas/proposal)
- [x] Spec (sdd/daas/spec) — 8 modules, 52 requirements, 36 test scenarios
- [x] Design (sdd/daas/design) — 12 ADRs
- [ ] Tasks — PENDING (next step: generate sdd-tasks)
- [ ] Implementation

## Next Steps

1. Generate sdd-tasks (task breakdown by sprint)
2. Task 0: Initialize monorepo (pnpm workspaces, Go module, Nuxt 3, Docker Compose, GitHub Actions)
3. Start Sprint 0 -> Sprint 1 (auth + multi-tenancy)
