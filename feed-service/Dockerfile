FROM golang:1.26-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app

COPY go.mod go.sum ./
# Use only the toolchain baked into the builder image.
ENV GOTOOLCHAIN=local
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o feed-service ./cmd/main.go

FROM alpine:3.20
WORKDIR /app

COPY --from=builder /app/feed-service .

ENTRYPOINT ["./feed-service"]
