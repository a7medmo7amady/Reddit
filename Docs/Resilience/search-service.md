# Search Service — Resilience Strategy

**Stack:** Node.js · **Library:** `opossum`

## Circuit Breakers

| Downstream Call | Failure Threshold | Open Duration | Justification |
|----------------|-------------------|---------------|---------------|
| Meilisearch query | 5 consecutive failures | 30 s | Search is non-critical to core browsing; fail fast and show "Search unavailable" |
| Kafka consumer (indexing) | 5 consecutive failures | 60 s | Indexing can lag temporarily; avoid crash loops |

## Retries

| Operation | Max Attempts | Backoff | Justification |
|-----------|-------------|---------|---------------|
| Meilisearch query | 2 | Fixed 200 ms | Queries are idempotent; brief Meilisearch restarts are common |
| Meilisearch index write | 3 | Exponential (300 ms base) | Index writes are idempotent (upsert); must eventually succeed |

## Kafka Dead-Letter Queue

| Event | DLQ Topic | Justification |
|-------|----------|---------------|
| `post.created` (index) | `search.index.dlq` | Failed indexing means missing search results; must be replayed |
| `post.deleted` (de-index) | `search.index.dlq` | Deleted content must be removed from index eventually |

Messages retried 3 times before routing to DLQ. Index writes are idempotent (upsert by post ID), so replay is safe.

## Timeouts

| Call | Timeout | Justification |
|------|---------|---------------|
| Meilisearch query | 2 s | Search SLA is p95 ≤ 500 ms; allow headroom but don't hang |
| Kafka message processing | 5 s | Indexing a single document shouldn't take longer |

## Fallbacks

- **Meilisearch down** → Return HTTP 503 with "Search temporarily unavailable" message. No degraded search mode.
- **Kafka consumer lag** → Meilisearch serves slightly stale index. Acceptable trade-off for availability.
