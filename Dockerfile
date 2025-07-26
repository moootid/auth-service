# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git for Go modules
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Expose port
EXPOSE 8080

# Set environment variables with defaults
ENV PORT=8080
ENV DB_HOST=localhost
ENV DB_PORT=26257
ENV DB_USER=root
ENV DB_PASSWORD=""
ENV DB_NAME=microservices
ENV JWT_SECRET=your-secret-key

# Run the application
CMD ["./main"]