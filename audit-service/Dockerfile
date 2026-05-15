FROM golang:1.26-alpine AS builder
RUN apk add --no-cache git
ENV GOTOOLCHAIN=local
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o audit-service ./cmd/main.go

FROM alpine:3.20
WORKDIR /app

COPY --from=builder /app/audit-service .
COPY --from=builder /app/config ./config

ENTRYPOINT ["./audit-service"]
