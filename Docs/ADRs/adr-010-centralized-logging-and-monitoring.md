# ADR-010: Centralized Logging & Monitoring

**Status:** Accepted  
**Date:** 2026-04-04

## Context

A microservices system has logs scattered across multiple services and hosts. Debugging cross-service issues requires a centralized view. The requirements mandate centralized logging using OpenSearch and monitoring of both application-level and host machine metrics.

## Decision

- **Centralized Logging:** **OpenSearch** (+ OpenSearch Dashboards)
  - All services emit structured JSON logs
  - Logs are shipped via **Fluentd/Fluent Bit** to OpenSearch
  - Retention: 90 days for access logs, 30 days for debug logs

- **Monitoring & Observability:**
  - **Prometheus** for metrics collection (application + host)
  - **Grafana** for dashboards and alerting
  - Key metrics: request latency (p95/p99), error rates, CPU/memory usage, Kafka consumer lag

## Consequences

**Positive:**
- Single pane of glass for all service logs
- Structured logs enable fast filtering and correlation
- Prometheus + Grafana is the industry-standard, open-source monitoring stack

**Negative:**
- OpenSearch requires dedicated storage and compute resources
- Log ingestion pipeline adds infrastructure complexity
- Teams must adopt consistent structured-logging standards
