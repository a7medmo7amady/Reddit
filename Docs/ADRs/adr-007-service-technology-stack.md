# ADR-007: Service Technology Stack

**Status:** Accepted  
**Date:** 2026-04-04

## Context

The requirements state that not all microservices need to use the same programming language. Each service has different performance and development speed requirements. The team's expertise and preferences also factor in.

## Decision

| Service | Language / Framework | Justification |
|---------|---------------------|---------------|
| User Service | Java / **Spring Boot** | Mature ecosystem for auth, security, and relational DB access. Team familiarity. |
| Feed Service | **Go** | High concurrency via goroutines. Low memory footprint for handling ≥10K concurrent requests. |
| Upload Service | **Node.js** | Excellent async I/O for file uploads. Rich ecosystem for image processing (Sharp). |
| Video Service | **Node.js** | Async job dispatching to FFmpeg workers. Shared tooling with Upload Service. |
| Notification Service | **Node.js** | Event-driven architecture aligns with Kafka consumer pattern. Nodemailer for email. |
| Search Service | **Node.js** | Lightweight HTTP proxy to Meilisearch. Minimal business logic. |
| Chat Service | **Node.js** | Native WebSocket support (ws/Socket.IO). Event-driven model fits real-time messaging. |
| Frontend | **React** (SPA) | Component-based UI. Large ecosystem. Team familiarity. |

## Consequences

**Positive:**
- Each service uses the best tool for its workload
- Node.js reuse across 4 services reduces tooling fragmentation
- Go delivers the raw throughput needed for the feed hot path

**Negative:**
- Team must maintain expertise in Java, Go, and Node.js
- CI/CD pipelines must support multiple build toolchains
- Shared libraries (e.g., Protobuf stubs) must be generated for each language
