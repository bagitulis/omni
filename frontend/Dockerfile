# ============================================
# Simple Dockerfile for Frontend (Pre-built)
# Build locally first: npm run build
# Then: docker build -t omni-frontend .
# ============================================

FROM nginx:1.27-alpine

# Install runtime dependencies and security updates
RUN apk add --no-cache curl && \
    apk upgrade --no-cache

# Copy pre-built dist from local machine
COPY dist /usr/share/nginx/html

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf
COPY nginx-default.conf /etc/nginx/conf.d/default.conf

# Expose port
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost/ || exit 1


# Start Nginx
CMD ["nginx", "-g", "daemon off;"]
