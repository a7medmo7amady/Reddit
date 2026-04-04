# ADR-003: Authentication & Authorization

**Status:** Accepted  
**Date:** 2026-04-04

## Context

The platform needs secure user authentication supporting both traditional and social login. Sessions must work across microservices without each service hitting the database on every request.

## Decision

- **OAuth 2.0 via Google** for social login (additional providers like GitHub/Apple can be added)
- **JWT access tokens** (15-min TTL) for stateless inter-service auth
- **Refresh tokens** (30-day TTL) stored in **httpOnly cookies** with rotation on every use
- **Token family reuse detection** — reuse of a consumed refresh token invalidates the entire session
- **Passwords** hashed with **bcrypt** (cost factor ≥ 12)
- Optional **2FA via TOTP** (Google Authenticator compatible)
- **Rate limiting**: max 10 failed login attempts per IP per 15 min, then 15-min lockout

## Consequences

**Positive:**
- JWT enables stateless validation at the API Gateway without per-request DB calls
- Refresh token rotation mitigates token theft
- OAuth reduces registration friction

**Negative:**
- JWT revocation requires a deny-list (stored in Redis) until token expiry
- Refresh token rotation adds complexity to the token lifecycle
- OAuth dependency on Google's availability
