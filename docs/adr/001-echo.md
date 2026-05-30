# ADR-001: Go Framework — Echo

Status: Accepted

## Decision

Use Echo as the Go HTTP framework over Fiber or Chi, because Echo is net/http compatible (unlike Fiber which uses fasthttp), which ensures full compatibility with the Go standard library ecosystem, Supabase Go client, testing tools, and third-party middleware.
