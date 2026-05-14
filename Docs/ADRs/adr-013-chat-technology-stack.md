# ADR-013: Use Go for Chat Service Implementation

**Status:** Accepted  
**Date:** 2026-05-07

## Context

The system architecture includes a custom API Gateway implemented in Go, a User Service, and a planned Chat Service responsible for:

- Real-time messaging
- WebSocket connection handling
- Typing indicators
- Delivery/read receipts
- Conversation management
- Notifications through asynchronous events
- Integration with MongoDB for message persistence

The initial design considered Node.js due to:
- Mature WebSocket ecosystem (`ws`, Socket.IO)
- Event-driven architecture
- Strong MongoDB ecosystem (Mongoose)

However, the Chat Service is expected to support large numbers of concurrent connections, frequent lightweight realtime events, and high-throughput message fan-out.

The service must also integrate efficiently with existing Go-based infrastructure.

## Decision

The Chat Service will be implemented in Go instead of Node.js.

Go was selected because:
- Goroutines provide lightweight concurrency for handling large numbers of simultaneous WebSocket connections
- Lower memory overhead compared to Node.js under high concurrent workloads
- Strong suitability for realtime systems with frequent lightweight event handling
- Native alignment with the existing Go API Gateway improves consistency in:
  - Middleware
  - Logging
  - Distributed tracing
  - Authentication forwarding
  - Service-to-service communication
- Single compiled binary simplifies deployment, scaling, and containerization
- Official MongoDB Go driver provides sufficient database compatibility without requiring Node.js-specific ODM tooling

MongoDB remains the primary persistence layer, with Go repository patterns replacing Mongoose-style schema abstractions.

## Consequences

**Positive:**

- Better scalability for high-concurrency WebSocket workloads
- Efficient handling of realtime chat events
- Improved consistency with Go API Gateway and infrastructure
- Simplified deployment through static binaries
- Reduced operational complexity from language standardization
- Explicit repository design enforces stronger service boundaries
- Suitable foundation for future scaling of:
  - Message queues
  - Event consumers
  - Notification systems
  - Presence systems

**Negative:**

- More boilerplate for validation, DTOs, and data access layers
- Less mature ecosystem for Socket.IO-style abstractions
- Slower initial development speed compared to Node.js
- Fewer rapid-prototyping conveniences compared to Mongoose and JavaScript tooling