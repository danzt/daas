# ADR-012: CI/CD Pipeline — GitHub Actions + Hetzner VPS

Status: Accepted

## Decision

Use GitHub Actions with parallel jobs (Go tests + Vue tests + lint) triggered on PR and push to main, Docker images pushed to GHCR, staging auto-deployed to Hetzner VPS (CAX11 ARM, ~4 EUR/month) on every main push via SSH, and production deployed via manual workflow_dispatch with a confirmation guard. Docker Compose orchestrates services on the VPS with Traefik for TLS termination via Let's Encrypt.
