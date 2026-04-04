# Chat Service — Resilience Strategy

**Stack:** Node.js · **Library:** `opossum`

## Circuit Breakers

| Downstream Call | Failure Threshold | Open Duration | Justification |
|----------------|-------------------|---------------|---------------|
| MongoDB (message persistence) | 5 consecutive failures | 30 s | Chat can buffer messages in-memory briefly; fail fast on DB |
| Redis Pub/Sub (fan-out) | 3 consecutive failures | 15 s | Without pub/sub, messages only reach the local server instance |
| User Service (gRPC — block check) | 5 consecutive failures | 30 s | Blocking is a safety feature; degrade to allowing delivery if unavailable |

## Retries

| Operation | Max Attempts | Backoff | Justification |
|-----------|-------------|---------|---------------|
| MongoDB write (message) | 3 | Exponential (100 ms base) | Messages must be persisted; retries cover brief replica failovers |
| Redis publish | 2 | Fixed 50 ms | Pub/sub is best-effort; one retry is sufficient |
| Missed-message fetch (`GET /chat/{room}/messages?since=`) | 2 | Fixed 300 ms | Client reconnection flow; idempotent read |

## WebSocket Resilience

| Pattern | Implementation | Justification |
|---------|---------------|---------------|
| Client reconnection | Exponential backoff (1 s → max 30 s) + jitter | Prevents thundering herd on server recovery |
| Heartbeat | Server ping every 30 s; disconnect after 2 missed pings | Detects dead connections and frees resources |
| Missed message recovery | `GET /chat/{room}/messages?since={timestamp}` on reconnect | Ensures no messages lost during disconnection |
| Single active player | Only one video/chat stream active at a time; scroll-away pauses | Prevents resource contention on client |

## Timeouts

| Call | Timeout | Justification |
|------|---------|---------------|
| MongoDB write | 1 s | Chat SLA is p95 ≤ 300 ms; DB must not dominate latency |
| Redis publish | 500 ms | Fan-out must be near-instant |
| User Service gRPC (block check) | 2 s | Non-blocking path; don't delay message delivery |

## Fallbacks

- **MongoDB down** → Buffer messages in-memory (bounded queue, max 1,000 messages). Flush to DB when circuit closes. Warn user of "messages may be delayed."
- **Redis Pub/Sub down** → Messages only delivered to users connected to the same server instance. Multi-instance fan-out paused until recovery.
- **User Service down** → Skip block check; deliver message. Log for post-hoc moderation review.
