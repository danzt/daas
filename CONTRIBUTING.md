# Contributing to DaaS

## Commit Convention

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): description

Types: feat, fix, docs, style, refactor, test, chore
Scopes: api, web, infra, db, auth, products, inventory, invoices, reports
```

## Branch Strategy

```
main           <- always deployable
  └── s{n}/{description}   <- sprint feature branch -> PR -> merge
```

## PR Checklist

- [ ] Tests pass (`go test ./...`, `pnpm vitest run`)
- [ ] Linter clean (`golangci-lint run`, `pnpm lint`)
- [ ] No hardcoded secrets or credentials
- [ ] Acceptance criteria from the linked GitHub Issue are met
- [ ] `go mod tidy` run if Go deps changed

## Local Setup

See [README.md](README.md).
