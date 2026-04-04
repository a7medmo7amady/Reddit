# ADR-005: API Gateway

**Status:** Accepted  
**Date:** 2026-04-04

## Context

The frontend (React SPA) needs a single entry point to access multiple backend microservices. Without a gateway, the client must know the address and API of every service, and cross-cutting concerns (auth, rate limiting, CORS) must be duplicated.

## Decision

Deploy a **centralized API Gateway** as the sole entry point for all client traffic.

**Responsibilities:**
- **Routing** — maps public REST endpoints to internal services
- **JWT validation** — verifies access tokens before forwarding
- **Rate limiting** — enforces per-IP and per-user request limits
- **CORS** — handles cross-origin headers centrally
- **TLS termination** — offloads HTTPS at the edge
- **Request logging** — emits structured access logs for observability

**Implementation options:** Kong, NGINX, AWS API Gateway, or a custom lightweight gateway.

## Consequences

**Positive:**
- Single entry point simplifies client integration
- Cross-cutting concerns (auth, rate limiting) handled once
- Internal service addresses remain private

**Negative:**
- Single point of failure — must be deployed with redundancy
- Adds a network hop to every request
- Gateway config must be updated when new services/routes are added
