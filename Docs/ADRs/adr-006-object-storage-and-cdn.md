# ADR-006: Object Storage & CDN

**Status:** Accepted  
**Date:** 2026-04-04

## Context

The platform handles user-uploaded images (up to 20 MB) and videos (up to 100 MB). These files must be stored durably, processed, and served globally with low latency.

## Decision

- **Object Storage:** AWS S3 (or S3-compatible MinIO/Garage for self-hosted/dev)
  - Raw uploads stored in a staging bucket
  - Processed media (WebP images, HLS segments) stored in a serving bucket
- **CDN:** Cloudflare for global edge caching and delivery
  - Pre-signed URLs with long-lived cache headers
  - Target CDN cache hit ratio ≥ 95% for videos with > 100 plays

| Asset Type | Storage Path | CDN Cached |
|-----------|-------------|------------|
| Processed images (WebP) | `s3://serving/images/` | Yes |
| HLS video segments (.ts) | `s3://serving/video/` | Yes |
| Thumbnails | `s3://serving/thumbs/` | Yes |
| Raw uploads (staging) | `s3://staging/` | No |

## Consequences

**Positive:**
- CDN reduces origin load and delivers low-latency media globally
- Pre-signed URLs enable direct browser uploads without proxying through the API

**Negative:**
- Cache invalidation adds complexity for updated/deleted content
- MinIO requires self-managed storage infrastructure in non-AWS environments
