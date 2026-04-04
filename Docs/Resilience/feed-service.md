# Feed Service — Resilience Strategy

**Stack:** Go · **Libraries:** Custom middleware / `sony/gobreaker`

## Circuit Breakers

| Downstream Call | Failure Threshold | Open Duration | Justification |
|----------------|-------------------|---------------|---------------|
| User Service (gRPC — ban check, profile) | 5 consecutive failures | 30 s | Feed can render without full user details; fail fast instead of blocking |
| MongoDB queries | 5 consecutive failures | 30 s | If DB is down, serve from Redis cache; avoid saturating connection pool |
| Redis | 3 consecutive failures | 15 s | Redis failures are usually brief; short open window to recover quickly |

## Retries

| Operation | Max Attempts | Backoff | Justification |
|-----------|-------------|---------|---------------|
| MongoDB read (feed query) | 3 | Exponential (50 ms base) + jitter | Read-heavy service; transient replica lag can cause brief failures |
| Redis GET/SET | 2 | Fixed 50 ms | Redis reconnects fast; one quick retry usually succeeds |
| Kafka consume (commit offset) | 3 | Exponential (200 ms base) | Failed commits cause re-delivery; retry avoids duplicate processing |

## Kafka Dead-Letter Queue

| Event | DLQ Topic | Justification |
|-------|----------|---------------|
| `post.created` | `post.created.dlq` | Failed indexing must not block feed; replay from DLQ after fix |
| `post.deleted` | `post.deleted.dlq` | Deletion must eventually propagate to avoid stale feed entries |

Messages are retried 3 times with exponential backoff before routing to DLQ. Idempotent processing via event ID deduplication.

## Timeouts

| Call | Timeout | Justification |
|------|---------|---------------|
| MongoDB query | 1 s | Feed SLA is p95 ≤ 300 ms; DB must respond fast or be skipped |
| User Service gRPC | 2 s | Supplementary data; must not dominate response time |
| Redis | 500 ms | Cache miss should fallback to DB, not hang |

## Fallbacks

- **Redis down** → Bypass cache; query MongoDB directly. Latency increases but feed still loads.
- **User Service down** → Return feed posts with minimal user info (username from denormalized post document). Skip avatar/karma enrichment.
- **MongoDB down + Redis hit** → Serve stale cached feed (TTL may be expired). Better than empty page.
