FROM nginx:alpine

# Install curl for healthcheck
RUN apk add --no-cache curl && \
    apk upgrade --no-cache

# Copy nginx configs directly into image (no bind mount)
COPY nginx.conf /etc/nginx/nginx.conf
COPY conf.d/tunnel.conf /etc/nginx/conf.d/default.conf

# Create log directory
RUN mkdir -p /var/log/nginx && \
    chmod 755 /var/log/nginx

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
