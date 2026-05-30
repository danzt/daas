# ADR-003: Backend Go Structure — Hexagonal Architecture

Status: Accepted

## Decision

Structure the Go backend using Hexagonal/Clean Architecture with ports and adapters: internal/domain/ contains pure business logic with no external imports, internal/adapter/ contains driving (HTTP/Echo) and driven (postgres, fiscal APIs) adapters, and internal/app/ contains application services that orchestrate the domain. The domain layer defines interfaces (ports); adapters implement them.
