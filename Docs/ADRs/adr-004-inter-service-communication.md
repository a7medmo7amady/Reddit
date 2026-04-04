# ADR-004: Inter-Service Communication (gRPC + Kafka)

**Status:** Accepted  
**Date:** 2026-04-04

## Context

Microservices need to communicate for operations like auth validation, feed updates, and notification delivery. Some calls are request/response (synchronous), while others are fire-and-forget events (asynchronous).

## Decision

Use a **dual communication model**:

| Pattern | Technology | Use Cases |
|---------|-----------|-----------|
| Synchronous | **gRPC** | Auth validation, profile lookups, ban-status checks |
| Asynchronous | **Apache Kafka** | `post.created`, `post.deleted`, `video.uploaded`, `comment.reply`, `mention`, `dm.received` |

**Why gRPC over REST for inter-service:**
- Binary serialization (Protobuf) — smaller payloads, faster parsing
- Strongly typed contracts via `.proto` files
- HTTP/2 multiplexing and streaming support

**Why Kafka:**
- Durable event log with replay capability
- Decouples producers from consumers
- Supports multiple consumer groups (Feed, Search, and Notifications all consume `post.created`)

> **Note:** The **API Gateway ↔ Client** communication remains **RESTful** (JSON over HTTP).

## Consequences

**Positive:**
- gRPC delivers lower latency for internal calls vs REST
- Kafka enables loose coupling — adding a new consumer requires no producer changes
- Event replay supports reindexing and recovery

**Negative:**
- gRPC is not browser-friendly — only used internally
- Kafka adds infrastructure overhead (brokers, ZooKeeper/KRaft)
- Eventual consistency for async event consumers
