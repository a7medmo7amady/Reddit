# Stage 1: Build stage
FROM golang:1.25-alpine AS builder

# Install protoc and go plugins
RUN apk add --no-cache protobuf protobuf-dev
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /app

# Copy go mod and sum files from the service directory
COPY notification-service/go.mod notification-service/go.sum ./notification-service/
WORKDIR /app/notification-service
RUN go mod download

WORKDIR /app

# Copy the proto files
COPY proto ./proto

# Copy the service source code
COPY notification-service ./notification-service

# Generate proto code
RUN mkdir -p notification-service/pkg/proto && \
    protoc --proto_path=./proto \
           --go_out=notification-service/pkg/proto --go_opt=paths=source_relative \
           --go-grpc_out=notification-service/pkg/proto --go-grpc_opt=paths=source_relative \
           ./proto/notification/notification.proto

# Build the application
WORKDIR /app/notification-service
RUN CGO_ENABLED=0 GOOS=linux go build -o notification-service ./cmd/main.go

# Stage 2: Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/notification-service/notification-service .

# Expose ports for HTTP and gRPC
EXPOSE 8084 50055

# Run the binary
CMD ["./notification-service"]
