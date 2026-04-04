# ADR-009: Search Engine (Meilisearch)

**Status:** Accepted  
**Date:** 2026-04-04

## Context

Users need full-text search across posts, comments, communities, and usernames with filtering and pagination. The search must respond at p95 ≤ 500 ms.

## Decision

Use **Meilisearch** as the dedicated search index.

**Indexed fields:** post title, post body, community name, username.

**Why Meilisearch over Elasticsearch:**
- Simpler to deploy and operate (single binary)
- Typo-tolerant and instant search out of the box
- Lower resource footprint for our expected data volume
- RESTful API with no query DSL learning curve

**Indexing flow:**
1. Upload/User services emit `post.created`, `post.deleted` events to Kafka
2. Search Service's Kafka consumer indexes/removes documents in Meilisearch
3. Search Controller queries Meilisearch with filters (type, date, community scope)

## Consequences

**Positive:**
- Fast setup, minimal configuration
- Typo tolerance improves user search experience
- Lightweight enough to run alongside other services

**Negative:**
- Less mature than Elasticsearch for complex aggregations
- Smaller community and plugin ecosystem
- May need to migrate to Elasticsearch if data volume grows beyond Meilisearch's capacity
