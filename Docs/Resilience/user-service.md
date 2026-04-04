# User Service — Resilience Strategy

**Stack:** Java / Spring Boot · **Library:** Resilience4j

## Circuit Breakers

| Downstream Call | Failure Threshold | Open Duration | Justification |
|----------------|-------------------|---------------|---------------|
| PostgreSQL queries | 5 consecutive failures | 30 s | DB outage shouldn't block all auth attempts; fail fast and return 503 |
| Email dispatch (SMTP) | 3 consecutive failures | 60 s | Email is non-critical to login flow; avoid holding threads on SMTP timeouts |

## Retries

| Operation | Max Attempts | Backoff | Justification |
|-----------|-------------|---------|---------------|
| DB read (profile, session lookup) | 3 | Exponential (100 ms base) + jitter | Transient connection resets are common; reads are idempotent |
| OAuth token exchange (Google) | 2 | Fixed 500 ms | Google OAuth can have brief 5xx blips; low retry count avoids token replay |

## Timeouts

| Call | Timeout | Justification |
|------|---------|---------------|
| gRPC inbound (from other services) | 3 s | Prevents slow consumers from holding goroutines |
| PostgreSQL query | 2 s | Long queries indicate a problem; fail fast |
| SMTP send | 5 s | Email providers can be slow; allow slightly more time |

## Fallbacks

- **Email dispatch failure** → Queue to Redis for retry by a background worker. Registration completes; verification email arrives later.
- **OAuth provider down** → Return clear error; no fallback (cannot fabricate third-party tokens).
