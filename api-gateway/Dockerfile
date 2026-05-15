FROM golang:1.26-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app

COPY go.mod go.sum ./
# Use only the toolchain baked into the builder image.
ENV GOTOOLCHAIN=local
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o gateway ./cmd/main.go

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/gateway .
EXPOSE 8088
ENTRYPOINT ["./gateway"]
