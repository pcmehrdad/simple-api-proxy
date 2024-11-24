FROM golang:1.22.5-alpine AS builder

# Install required system packages and update certificates
RUN apk update && \
    apk upgrade && \
    apk add --no-cache ca-certificates && \
    update-ca-certificates

# Add Maintainer Info to the Image
LABEL maintainer="Mehrdad Amini <pcmehrdad@gmail.com>"
LABEL description="API Proxy Service"

# Set the Current Working Directory inside the container
WORKDIR /build/api-proxy

# Copy go mod files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/api-proxy ./cmd/processor/main.go

# Start a new stage from scratch for a smaller final image
FROM scratch

WORKDIR /app

# Copy binary and configuration
COPY --from=builder /app/api-proxy .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY config.json .

# Expose port
EXPOSE 3003

# Command to run
CMD ["./api-proxy"]