# ADR-011: Resilience Patterns

**Status:** Accepted  
**Date:** 2026-04-04

## Context

In a distributed system, any inter-service call can fail due to network issues, timeouts, or downstream outages. Without resilience mechanisms, a single failing service can cascade failures across the platform.

## Decision

Implement the following resilience patterns:

| Pattern | Implementation | Applied To |
|---------|---------------|-----------|
| **Circuit Breaker** | Library-based (e.g., Resilience4j for Spring Boot, `opossum` for Node.js) | All outbound gRPC and HTTP calls |
| **Retries** | Exponential backoff with jitter, max 3 attempts | Idempotent requests (reads, search queries) |
| **Timeouts** | Hard timeout per call (default 3 s for gRPC, 5 s for HTTP) | All inter-service calls |
| **Bulkhead** | Thread/connection pool isolation per downstream | Feed → User Service, Upload → User Service |

**Circuit breaker states:**
- **Closed** → normal operation
- **Open** → requests fail fast for 30 s after 5 consecutive failures
- **Half-Open** → allows 1 probe request to test recovery

**Fallback strategies (optional):**
- Feed Service returns cached feed if User Service is unavailable
- Notification Service queues events to Redis if email delivery fails

## Consequences

**Positive:**
- Prevents cascading failures across the service mesh
- Fail-fast behavior keeps latency predictable under partial outages
- Retries with backoff handle transient network issues transparently

**Negative:**
- Adds complexity to every inter-service call path
- Misconfigured thresholds can cause unnecessary circuit opens
- Retry storms can amplify load if not properly bounded
