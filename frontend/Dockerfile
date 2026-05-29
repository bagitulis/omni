# ============================================
# Multi-stage Dockerfile for React Frontend
# Build: node:20-alpine
# Serve: nginx:alpine
# ============================================

# Stage 1: Build React app
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm install --legacy-peer-deps

# Copy source code
COPY . .

# Build the app
RUN npm run build

# Stage 2: Serve with Nginx
FROM nginx:1.27-alpine

# Install runtime dependencies and security updates
RUN apk add --no-cache curl && \
    apk upgrade --no-cache

# Copy built dist from builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf
COPY nginx-default.conf /etc/nginx/conf.d/default.conf
COPY nginx-csp.conf /etc/nginx/conf.d/nginx-csp.conf

# Expose port
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost/ || exit 1

# Start Nginx
CMD ["nginx", "-g", "daemon off;"]
