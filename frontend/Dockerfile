# Build stage
FROM node:20-alpine AS builder
WORKDIR /app

# Copy package files
COPY package*.json ./
RUN npm ci

# Copy source and build
COPY . .

# Build args for Vite (baked at build time)
ARG PUBLIC_API_URL
ARG PUBLIC_ENVIRONMENT=production
ENV PUBLIC_API_URL=$PUBLIC_API_URL
ENV PUBLIC_ENVIRONMENT=$PUBLIC_ENVIRONMENT

# Build the app
RUN npm run build

# Production stage
FROM node:20-alpine
WORKDIR /app

# Copy built app and production deps only
COPY --from=builder /app/build ./build
COPY --from=builder /app/package*.json ./
RUN npm ci --omit=dev

# Create non-root user for security
RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 sveltekit
USER sveltekit

# Set environment
ENV NODE_ENV=production
ENV PORT=3000
# ORIGIN must be set at deploy time to match the actual domain (e.g., https://app.example.com)
# SvelteKit adapter-node uses ORIGIN for CSRF protection
# This default is for local/container use only -- override at runtime
ENV ORIGIN=http://localhost:3000

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000/ || exit 1

EXPOSE 3000

CMD ["node", "build"]
