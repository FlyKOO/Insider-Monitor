# Build stage
FROM golang:1.23.2-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /insider-monitor ./cmd/monitor

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /insider-monitor .
# Copy example config
COPY config.example.json .

# Create volume for data persistence
VOLUME ["/app/data"]

# Run the binary
ENTRYPOINT ["./insider-monitor"] 