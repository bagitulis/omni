# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies (gcc/musl-dev for CGO, libwebp-dev for WebP encoding)
RUN apk add --no-cache gcc musl-dev libwebp-dev

# Copy go mod files first so dependency download is cached independently of source changes
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
# NOTE: CGO is required for libwebp. CGO_ENABLED=1 with alpine musl is intentional.
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/server ./cmd/server
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/credential-migration ./cmd/credential-migration

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install runtime deps. libwebp is required at runtime for image encoding.
# sqlite is NOT required: runtime database is PostgreSQL-only.
RUN apk add --no-cache ca-certificates tzdata libwebp

# Set timezone to Asia/Jakarta (WIB - UTC+7)
ENV TZ=Asia/Jakarta
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Copy binary from builder
COPY --from=builder /app/server /app/server
COPY --from=builder /app/credential-migration /app/credential-migration

# Copy Google service account credentials from config
COPY config/static/google /app/config/static/google

# Expose port
EXPOSE 3000

# Environment variables
ENV GO_ENV=production
ENV PORT=3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000/api/health || exit 1

# Run the server
CMD ["/app/server"]
