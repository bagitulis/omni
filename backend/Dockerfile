# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies (git needed for go mod, libwebp-dev for WebP encoding)
RUN apk add --no-cache gcc musl-dev git libwebp-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Ensure all dependencies are resolved (git needed for this)
RUN go mod tidy

# Build the binary
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/server ./cmd/server

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies including timezone data and libwebp for image encoding
RUN apk add --no-cache ca-certificates sqlite tzdata libwebp

# Set timezone to Asia/Jakarta (WIB - UTC+7)
ENV TZ=Asia/Jakarta
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Copy binary from builder
COPY --from=builder /app/server /app/server

# Copy Google service account credentials from config
COPY config/static/google /app/config/static/google

# Expose port
EXPOSE 3000

# Environment variables
ENV GO_ENV=production
ENV PORT=3000
ENV DATABASE_PATH=/app/data

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000/api/health || exit 1

# Run the server
CMD ["/app/server"]
