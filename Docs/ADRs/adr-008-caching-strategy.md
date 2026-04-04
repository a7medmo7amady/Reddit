# ADR-008: Caching Strategy (Redis)

**Status:** Accepted  
**Date:** 2026-04-04

## Context

The Feed Service must serve ≥10,000 concurrent requests/sec at p95 ≤ 300 ms. Hitting MongoDB on every feed request is infeasible at this scale. Notifications and chat also need a fast queue for offline delivery.

## Decision

Use **Redis 7+ (cluster mode)** as the centralized caching and pub/sub layer.

| Use Case | Key Pattern | TTL |
|----------|-----------|-----|
| Feed cache (Hot/Rising) | `feed:{sort}:{community_id}` | 60 s |
| Feed cache (New) | — | No cache (real-time) |
| User sessions | `session:{user_id}` | 30 days |
| Rate-limit counters | `rl:{ip}:{endpoint}` | 15 min |
| Offline notification queue | `notif:queue:{user_id}` | Until delivered |
| WebSocket pub/sub | Redis Pub/Sub channels | — |

**Cache invalidation:** Kafka consumers invalidate affected cache keys on `post.created`, `post.deleted`, and vote events.

## Consequences

**Positive:**
- Sub-millisecond reads keep feed latency within SLA
- Pub/Sub enables real-time WebSocket message fan-out
- Cluster mode provides horizontal scaling

**Negative:**
- Cache invalidation logic adds complexity
- Redis is in-memory — data loss on crash (mitigated by AOF persistence)
- Additional infrastructure to monitor and maintain
