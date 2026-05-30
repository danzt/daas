# ADR-010: API Design — REST + Cursor Pagination + RFC 7807

Status: Accepted

## Decision

Design the API as REST resource-oriented with URL prefix /api/v1/, cursor-based pagination (?cursor=X&limit=20) instead of offset-based for correctness on live data, RFC 7807 Problem Details for error responses, kebab-case URLs, camelCase JSON fields, and Content-Type: application/json throughout.
