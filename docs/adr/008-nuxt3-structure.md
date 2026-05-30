# ADR-008: Frontend — Nuxt 3 + Pinia + shadcn-vue

Status: Accepted

## Decision

Use Nuxt 3 with file-based routing for the frontend, Pinia for global state (auth, cart, notifications), composables + useAsyncData for server state (avoiding duplication in Pinia), shadcn-vue for UI components following atomic design, and Tailwind 4 for styling. The frontend is a pure API consumer — all business logic lives in the Go backend.
