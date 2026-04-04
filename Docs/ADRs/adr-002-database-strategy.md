# ADR-002: Database Strategy (Polyglot Persistence)

**Status:** Accepted  
**Date:** 2026-04-04

## Context

Different services have different data access patterns. User relationships are highly relational while posts, chat messages, and notifications are document-oriented with high write throughput. A single database cannot optimally serve all patterns.

## Decision

Use **polyglot persistence** with three storage engines:

| Store | Technology | Owns |
|-------|-----------|------|
| Relational | PostgreSQL 15+ | Users, communities, follows, roles, sessions |
| Document | MongoDB 7+ | Posts, comments, chat messages, notifications, mod logs, votes, drafts |
| Cache / Pub-Sub | Redis 7+ (cluster) | Sessions, rate-limit counters, feed cache, WebSocket pub/sub, job queues |

**Rationale:**
- **PostgreSQL** — ACID transactions for user identity, auth, and community relationships
- **MongoDB** — Schema flexibility and high I/O throughput for posts, chat, and notifications that change frequently
- **Redis** — Sub-millisecond reads for caching, session storage, and real-time pub/sub

## Consequences

**Positive:**
- Each store is optimized for its workload
- MongoDB handles high-write chat/post traffic without burdening the relational DB
- Redis offloads read pressure from primary stores

**Negative:**
- No cross-store joins — must denormalize or join at the application layer
- Multiple backup and migration strategies to maintain
- Team must be proficient in three database technologies
