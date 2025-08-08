# Use the official Golang image as base
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Download dependencies (this will generate go.sum)
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/app/main.go

# Final stage
FROM alpine:latest

# Create non-root user
RUN adduser -D -u 1000 appuser

# Set working directory
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Change ownership to non-root user
RUN chown appuser:appuser main
USER appuser

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]