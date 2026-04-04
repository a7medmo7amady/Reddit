# Notification Service — Resilience Strategy

**Stack:** Node.js · **Library:** `opossum`

## Circuit Breakers

| Downstream Call | Failure Threshold | Open Duration | Justification |
|----------------|-------------------|---------------|---------------|
| SMTP (email delivery) | 3 consecutive failures | 120 s | Email providers can have extended outages; long open window to avoid retry storms |
| Redis (offline queue) | 3 consecutive failures | 15 s | Redis recovers fast; short window |
| WebSocket push | 5 consecutive failures | 30 s | Connection-specific; isolate per-user failures |

## Retries

| Operation | Max Attempts | Backoff | Justification |
|-----------|-------------|---------|---------------|
| Email send (SMTP) | 3 | Exponential (1 s base) | Transient SMTP errors are common; emails are non-urgent |
| Redis queue write | 2 | Fixed 100 ms | Must persist notification for offline users |
| MongoDB notification write | 2 | Fixed 200 ms | Record must be stored for the notification bell |

## Kafka Dead-Letter Queue

| Event | DLQ Topic | Justification |
|-------|----------|---------------|
| `comment.reply` | `notification.dlq` | User must eventually be notified of replies |
| `mention` | `notification.dlq` | Mentions are time-sensitive but must not be lost |
| `dm.received` | `notification.dlq` | DM notification triggers email; must be retried |

Messages retried 3 times before routing to DLQ. Idempotent via event ID deduplication.

## WebSocket Resilience

| Pattern | Implementation | Justification |
|---------|---------------|---------------|
| Client reconnection | Exponential backoff (1 s → max 30 s) + jitter | Prevents thundering herd on server recovery |
| Heartbeat | Server ping every 30 s; disconnect after 2 missed pongs | Detects dead connections promptly |
| Offline queue drain | Redis queue drained on reconnect | Ensures no missed notifications |

## Timeouts

| Call | Timeout | Justification |
|------|---------|---------------|
| SMTP send | 10 s | Email providers can be slow; allow generous timeout |
| Redis write | 500 ms | Queue write must be fast or skip |
| WebSocket push | 1 s | Connected user should receive within SLA (p95 ≤ 2 s) |

## Fallbacks

- **SMTP down** → Queue email to Redis. Background worker retries every 5 min. Notification still delivered in-app.
- **Redis down** → Notifications delivered only to currently connected WebSocket clients. Offline users receive notifications on next login via MongoDB query.
- **WebSocket disconnected** → Notification stored in MongoDB + queued in Redis. Delivered on reconnect.
