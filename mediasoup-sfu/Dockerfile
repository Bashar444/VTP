FROM node:22-alpine

# Install build dependencies for mediasoup
RUN apk add --no-cache \
    build-base \
    python3 \
    py3-pip \
    linux-headers \
    git

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --omit=dev

# Copy application code
COPY src ./src

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
  CMD node -e "require('http').get('http://localhost:3000/health', (r) => {if (r.statusCode !== 200) process.exit(1)})"

# Start application
CMD ["node", "src/index.js"]
