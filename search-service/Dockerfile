# Stage 1: Build stage
FROM golang:1.26-alpine AS builder
ENV GOTOOLCHAIN=local

# Install protoc and go plugins
RUN apk add --no-cache git protobuf protobuf-dev
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /app

# Copy go mod and sum files from the service directory
COPY search-service/go.mod search-service/go.sum ./search-service/
WORKDIR /app/search-service
RUN go mod download

WORKDIR /app

# Copy the proto files
COPY proto ./proto

# Copy the service source code
COPY search-service ./search-service

# Generate proto code
RUN mkdir -p search-service/pkg/proto && \
    protoc --proto_path=./proto \
           --go_out=search-service/pkg/proto --go_opt=paths=source_relative \
           --go-grpc_out=search-service/pkg/proto --go-grpc_opt=paths=source_relative \
           ./proto/search/search.proto

# Build the application
WORKDIR /app/search-service
RUN CGO_ENABLED=0 GOOS=linux go build -o search-service ./cmd/main.go

# Stage 2: Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/search-service/search-service .

# Expose ports for HTTP and gRPC
EXPOSE 8082 50053

# Run the binary
CMD ["./search-service"]
