# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make protoc

# Set working directory
WORKDIR /build

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Generate Protocol Buffer code
RUN make proto

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux make build

# Final stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S hephaestus \
    && adduser -S hephaestus -G hephaestus

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/hephaestus /app/
COPY --from=builder /build/config.yaml /app/

# Set ownership
RUN chown -R hephaestus:hephaestus /app

# Switch to non-root user
USER hephaestus

# Expose gRPC port
EXPOSE 50051

# Set environment variables
ENV HEPHAESTUS_CONFIG_PATH=/app/config.yaml \
    HEPHAESTUS_MODE=production

# Command to run the binary
ENTRYPOINT ["/app/hephaestus"] 