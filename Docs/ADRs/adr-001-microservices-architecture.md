# ADR-001: Microservices Architecture

**Status:** Accepted  
**Date:** 2026-04-04

## Context

The platform spans multiple domains (Users, Feed, Search, Upload, Video, Notifications, Chat) with varying load profiles and scaling needs. A monolith would couple all domains and limit independent scaling.

## Decision

Adopt a **microservices architecture** with each domain as an independently deployable service:

| Service | Responsibility |
|---------|---------------|
| User Service | Auth, profiles, roles, moderation |
| Feed Service | Home/community/discovery feeds, voting, personalization |
| Search Service | Full-text search indexing and querying |
| Upload Service | Post creation, image processing, media pipeline |
| Video Service | Video transcoding, HLS delivery |
| Notification Service | In-app and email notification delivery |
| Chat Service | Real-time DMs and community chat rooms |

All services sit behind a single **API Gateway** and communicate via **gRPC** (sync) and **Kafka** (async).

## Consequences

**Positive:**
- Independent deployment and scaling per service
- Team autonomy — each team owns a service end-to-end
- Technology flexibility per service

**Negative:**
- Increased operational complexity (networking, deployment, monitoring)
- Distributed data management requires eventual consistency patterns
- Cross-service debugging is harder
