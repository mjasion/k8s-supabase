# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /secret-extractor \
    ./cmd/job

# Runtime stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy the binary from builder
COPY --from=builder /secret-extractor /usr/local/bin/secret-extractor

# Create non-root user
RUN addgroup -g 1000 extractor && \
    adduser -u 1000 -G extractor -s /bin/sh -D extractor

USER extractor

ENTRYPOINT ["/usr/local/bin/secret-extractor"]
