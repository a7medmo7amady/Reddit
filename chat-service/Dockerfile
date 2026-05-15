FROM golang:1.26-alpine AS builder
RUN apk add --no-cache git
ENV GOTOOLCHAIN=local

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o chat-service ./cmd/api


FROM alpine:3.20

WORKDIR /app

RUN adduser -D -g '' appuser

COPY --from=builder /app/chat-service ./chat-service

USER appuser

EXPOSE 8085

CMD ["./chat-service"]
