FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
# Allow the Go toolchain directive in go.mod to download a newer toolchain if needed
ENV GOTOOLCHAIN=auto
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o feed-service ./cmd/main.go

FROM alpine:3.20
WORKDIR /app

COPY --from=builder /app/feed-service .

ENTRYPOINT ["./feed-service"]
