FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
# Allow Go toolchain directive to be honoured without failing
ENV GOTOOLCHAIN=auto
# Use BuildKit inline cache for modules (speeds up rebuilds)
RUN --mount=type=cache,target=/root/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/root/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux go build -o gateway ./cmd/main.go

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/gateway .
EXPOSE 8088
ENTRYPOINT ["./gateway"]
