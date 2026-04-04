# Video Service — Resilience Strategy

**Stack:** Node.js · **Library:** `opossum`

## Circuit Breakers

| Downstream Call | Failure Threshold | Open Duration | Justification |
|----------------|-------------------|---------------|---------------|
| FFmpeg transcoding worker | 3 consecutive failures | 120 s | Transcoding is heavy; repeated failures indicate resource exhaustion — back off longer |
| S3 upload (HLS segments) | 3 consecutive failures | 60 s | Segment upload failure blocks the entire transcoding job |
| Kafka consumer (video.uploaded) | 5 consecutive failures | 60 s | Avoid crash-loop on persistent deserialization errors |

## Retries

| Operation | Max Attempts | Backoff | Justification |
|-----------|-------------|---------|---------------|
| S3 PUT (segment/manifest) | 3 | Exponential (500 ms base) | Each segment is a small file; transient S3 errors are retriable |
| FFmpeg job | 2 | Fixed 30 s | Transcoding can fail on resource contention; one retry after cooldown |
| MongoDB status update | 2 | Fixed 200 ms | Status writes are idempotent |

## Timeouts

| Call | Timeout | Justification |
|------|---------|---------------|
| FFmpeg transcoding (per job) | 10 min | SLA: 1080p ≤15 min source transcoded within 10 min |
| S3 segment upload | 10 s | Segments are small (~2 MB); shouldn't take longer |

## Fallbacks

- **FFmpeg crash** → Mark job status as `failed` in MongoDB. Expose via `GET /videos/{id}/status`. User can re-upload.
- **S3 down** → Pause transcoding queue. Resume when circuit closes. No partial uploads.
- **Partial transcoding failure** → Serve available lower-quality variants (e.g., 360p ready, 1080p failed). Player degrades gracefully.
