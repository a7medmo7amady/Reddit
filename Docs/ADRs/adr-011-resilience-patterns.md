# ADR-011: Resilience Patterns

**Status:** Accepted  
**Date:** 2026-04-04

## Context

In a distributed system, any inter-service call can fail due to network issues, timeouts, or downstream outages. Without resilience mechanisms, a single failing service can cascade failures across the platform. The system uses three communication channels — **gRPC** (synchronous inter-service), **Kafka** (asynchronous events), and **WebSocket** (real-time client connections) — each requiring distinct resilience strategies.

## Decision

### Synchronous Calls (gRPC)

gRPC is used **only** for direct, synchronous inter-service calls (e.g., auth validation, profile lookups, ban checks). It is **not** used for broker-mediated communication.

| Pattern | Implementation | Applied To |
|---------|---------------|-----------|
| **Circuit Breaker** | Resilience4j (Spring Boot), `opossum` (Node.js), `gobreaker` (Go) | All outbound gRPC calls |
| **Retries** | Exponential backoff with jitter, max 3 attempts | Idempotent gRPC calls (reads, lookups) |
| **Timeouts** | Hard timeout 3 s per call | All gRPC calls |

**Circuit breaker states:**
- **Closed** → normal operation
- **Open** → requests fail fast for 30 s after 5 consecutive failures
- **Half-Open** → allows 1 probe request to test recovery

### Asynchronous Events (Kafka)

Kafka consumers process events like `post.created`, `post.deleted`, `video.uploaded`, `comment.reply`, and `mention`. These require different resilience handling than request/response calls.

| Pattern | Implementation | Applied To |
|---------|---------------|-----------|
| **Consumer Retries** | Per-message retry (max 3 attempts, exponential backoff) before sending to DLQ | All Kafka consumers |
| **Dead-Letter Queue (DLQ)** | Failed messages routed to a `*.dlq` topic after exhausting retries | All event types |
| **Idempotent Processing** | Deduplication via event ID stored in DB | Indexing, karma updates, notification dispatch |
| **Consumer Lag Monitoring** | Alert when lag exceeds threshold (e.g., >1,000 messages) | All consumer groups |

**DLQ handling:** Ops team reviews and replays DLQ messages manually or via automated retry job.

### WebSocket (Client Connections)

WebSocket is used for real-time chat and notification delivery. Resilience is handled at both client and server level.

| Pattern | Implementation | Applied To |
|---------|---------------|-----------|
| **Client Reconnection** | Exponential backoff (1 s → 2 s → 4 s → … max 30 s) with jitter | All WebSocket clients |
| **Missed Message Recovery** | `GET /chat/{room}/messages?since={timestamp}` on reconnect | Chat and Notification services |
| **Offline Queue** | Redis queue per user; drained on next WebSocket connection | Notifications |
| **Heartbeat / Ping-Pong** | Server pings every 30 s; client disconnected after 2 missed pongs | Connection health detection |

### Fallback Strategies

- Feed Service returns cached feed if User Service is unavailable
- Notification Service queues events to Redis if email delivery fails
- Chat Service buffers messages in-memory if MongoDB is temporarily down

## Consequences

**Positive:**
- Prevents cascading failures across the service mesh
- DLQ ensures no event is silently lost — failed messages are recoverable
- WebSocket reconnection with backoff prevents thundering herd on server recovery

**Negative:**
- Adds complexity to every communication path (gRPC, Kafka, WebSocket)
- DLQ requires operational tooling for monitoring and replay
- Misconfigured retry/backoff thresholds can cause unnecessary circuit opens or retry storms
