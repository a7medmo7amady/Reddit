# ADR-012: Real-Time Communication (WebSocket)

**Status:** Accepted  
**Date:** 2026-04-04

## Context

The Chat module requires real-time text messaging (p95 ≤ 300 ms), and the Notification module needs instant in-app delivery (p95 ≤ 2 s). Both features need a persistent, bidirectional connection to the client.

## Decision

Use a **single WebSocket connection** (shared between Chat and Notifications) per authenticated client.

| Concern | Approach |
|---------|---------|
| Protocol | WebSocket (ws/wss) via **Socket.IO** or native `ws` library |
| Scaling | **Redis Pub/Sub** for fan-out across multiple server instances |
| Reconnection | Client auto-reconnects and fetches missed messages via `GET /chat/{room}/messages?since={timestamp}` |
| Offline delivery | Messages queued in Redis, delivered on next WebSocket connection |
| Concurrency target | ≥ 10,000 concurrent WebSocket connections for MVP |

**Why a shared WebSocket:**
- Avoids opening multiple persistent connections per client
- Reduces mobile battery and bandwidth consumption
- Simplifies connection lifecycle management

## Consequences

**Positive:**
- Single connection reduces client and server resource usage
- Redis Pub/Sub enables horizontal scaling of WebSocket servers
- Offline queue ensures no missed notifications

**Negative:**
- Shared connection means a bug in chat framing can affect notification delivery
- WebSocket servers are stateful — requires sticky sessions or Redis-backed state
- Scaling beyond 10K connections needs careful load balancing
