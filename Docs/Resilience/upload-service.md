# Upload Service — Resilience Strategy

**Stack:** Node.js · **Library:** `opossum`

## Circuit Breakers

| Downstream Call | Failure Threshold | Open Duration | Justification |
|----------------|-------------------|---------------|---------------|
| User Service (gRPC — JWT + ban check) | 5 consecutive failures | 30 s | Cannot allow uploads without auth; fail fast with 503 |
| S3 upload | 3 consecutive failures | 60 s | S3 outages are rare but impactful; stop attempting uploads quickly |
| Kafka producer | 5 consecutive failures | 30 s | Events can be retried; avoid blocking the upload response |

## Retries

| Operation | Max Attempts | Backoff | Justification |
|-----------|-------------|---------|---------------|
| S3 PUT (media upload) | 3 | Exponential (500 ms base) + jitter | S3 transient 5xx errors are documented by AWS; retry is safe |
| MongoDB write (post metadata) | 2 | Fixed 300 ms | Idempotent upsert; brief replica failovers cause transient errors |
| Kafka produce | 3 | Exponential (200 ms base) | Event delivery must eventually succeed for downstream indexing |

## Kafka Dead-Letter Queue

| Event | DLQ Topic | Justification |
|-------|----------|---------------|
| `post.created` | `post.created.dlq` | Downstream consumers (Feed, Search) depend on this event |
| `post.deleted` | `post.deleted.dlq` | Deletion must propagate to purge content from feeds and search |
| `video.uploaded` | `video.uploaded.dlq` | Transcoding must eventually be triggered |

If Kafka producer fails after exhausting retries, events are written to a local MongoDB dead-letter collection and replayed by a background worker.

## Timeouts

| Call | Timeout | Justification |
|------|---------|---------------|
| Image processing (Sharp) | 10 s | Large images (20 MB) need processing time; but bound it |
| S3 upload | 15 s | Large files over network; generous but finite |
| User Service gRPC | 3 s | Auth check must be fast |

## Fallbacks

- **S3 down** → Return HTTP 503. Upload cannot proceed without storage. Client retries.
- **Kafka down** → Write event to a local dead-letter queue (MongoDB collection). Background worker replays on recovery.
- **Image processing failure** → Return 422 with error details. Don't store a corrupt file.
