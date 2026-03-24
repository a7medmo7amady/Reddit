# Project Requirements
## Web Application Features
### The web application must include:
- User authentication and authorization using OAuth2 (Google authentication).
- A core feature with CRUD operations (e.g., chat system, social networking, or event management).
- A dashboard with real-time updates or data visualization. System Design Requirements
- Each service must have an Architectural Decision Record (ADR) documenting the reasoning behind chosen technologies and patterns.
- Resilience mechanisms including circuit breakers and retries. Fallback strategies are optional.
- Token-based authentication (OAuth2 via Google for user login, any other providers can be used with justification in the ADR).
- Use an API Gateway and design at least two RESTful APIs.
- Scalable database design (SQL or NoSQL) with sharding considerations.
- Mode of integration and communication between services (synchronous/asynchronous) should be clearly justified in the ADR.
- Memory-store implementation (Redis or equivalent) for caching and session management.
- Centralized logging using OpenSearch.
- Monitoring and observability setup
- System monitoring must include both application-level and host machine metrics.
- Not all microservices need to be implemented in the same programming language. Justify yourchoices in the ADR.


# Phase 1 (Week 7): High-Level System DesignObjective

Develop a complete high-level design for the web application, integrating architectural principles covered in the course.
## Submission Guidelines
### You must submit a design document that includes:
1. Requirements Specification Document:
	- Functional and non-functional requirements 
	- User stories/use cases
	- Architectural drivers (quality attributes like scalability, resilience)
2. C4 Model Diagrams (Must include at least these levels:)
	- Context Diagram – High-level system overview, including users and external dependencies
	- Container Diagram – Breakdown of the system into microservices, databases, caches, etc.
	- Component Diagram – Internal structure of key services (at least one service in detail)
	- (Optional: Code Diagram if needed for specific technical decisions)
3. Architecture Decision Records (ADRs)
	- Justifications for chosen tools, languages, architectural decisions (e.g., why microservices, why Redis, etc.)
	- At least one ADR per major decision (e.g., database choice, authentication method, message broker)
4. API Specifications
	- RESTful API endpoints documented using OpenAPI (Swagger)
	- Synchronous and asynchronous communication clearly explained
5. Resilience Strategies
	- Explanation of implemented circuit breakers, retries, or fallbacks
	- Justification for selected mechanisms
6. Presentation
	- 5-minute video walkthrough of the system design

## Submission Format:
- PDF document containing all diagrams, ADRs, and explanations
- C4 diagrams in PNG/SVG format or a link to an online C4 tool (Structurizr, Draw.io, etc.)
- GitHub Repository containing design artifacts (Markdown files for ADRs, OpenAPI spec, diagram sources if applicable)
- Video Presentation (link to YouTube or cloud storage)