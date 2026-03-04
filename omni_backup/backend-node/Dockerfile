# ============================================
# Multi-stage Dockerfile for Backend
# ============================================

# Stage 1: Build
FROM node:22-alpine3.19 AS builder

WORKDIR /app

# Install dependencies for native modules
RUN apk add --no-cache python3 make g++ 

# Copy package files
COPY package*.json ./
COPY prisma ./prisma/

# Install ALL dependencies (including devDependencies for build)
RUN npm install --legacy-peer-deps && \
    npm cache clean --force

# Copy source code
COPY . .

# Generate Prisma Client
RUN npx prisma generate

# Build TypeScript
RUN npm run build

# ============================================
# Stage 2: Production
FROM node:22-alpine3.19

WORKDIR /app

# Install runtime dependencies for native modules and security updates
RUN apk add --no-cache \
    python3 make g++ \
    openssl \
    ca-certificates \
    dumb-init && \
    apk upgrade --no-cache

# Create non-root user
RUN addgroup -g 1001 -S nodejs && \
    adduser -S nodejs -u 1001

# Copy package files and install production dependencies only
COPY --from=builder --chown=nodejs:nodejs /app/package*.json ./
RUN npm install --only=production --legacy-peer-deps && npm cache clean --force

# Copy built assets from builder
COPY --from=builder --chown=nodejs:nodejs /app/dist ./dist
COPY --from=builder --chown=nodejs:nodejs /app/prisma ./prisma

# Copy Prisma client from builder (already generated with correct version)
COPY --from=builder --chown=nodejs:nodejs /app/node_modules/.prisma ./node_modules/.prisma
COPY --from=builder --chown=nodejs:nodejs /app/node_modules/@prisma ./node_modules/@prisma

# Copy config and required files (including database files)
COPY --chown=nodejs:nodejs config ./config

# Create necessary directories
RUN mkdir -p /app/logs/audit && \
    chown -R nodejs:nodejs /app/config /app/logs

# Switch to non-root user
USER nodejs

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=40s --retries=3 \
    CMD node -e "require('http').get('http://localhost:3000/api/health', (r) => {process.exit(r.statusCode === 200 ? 0 : 1)})"

# Use dumb-init to handle signals properly
ENTRYPOINT ["dumb-init", "--"]

# Start application (path matches tsconfig outDir + source structure)
CMD ["node", "-r", "tsconfig-paths/register", "dist/server.js"]
