# Multi-stage Dockerfile for OSA Voice Agent
# Optimized for Google Cloud Run deployment

FROM python:3.11-slim as builder

# Install build dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy requirements first (for layer caching)
COPY requirements.txt .

# Install Python dependencies
RUN pip install --no-cache-dir --user -r requirements.txt

# Production image
FROM python:3.11-slim

# Install runtime dependencies only
RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user for security
RUN useradd -m -u 1000 osaagent && \
    mkdir -p /app && \
    chown -R osaagent:osaagent /app

# Set working directory
WORKDIR /app

# Copy Python packages from builder
COPY --from=builder --chown=osaagent:osaagent /root/.local /home/osaagent/.local

# Copy application code
COPY --chown=osaagent:osaagent agent.py .
COPY --chown=osaagent:osaagent README.md .

# Switch to non-root user
USER osaagent

# Add local Python packages to PATH
ENV PATH=/home/osaagent/.local/bin:$PATH

# Expose port (Cloud Run will set PORT env var)
ENV PORT=8080
EXPOSE 8080

# Health check (optional, for debugging)
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s \
  CMD python -c "import sys; sys.exit(0)"

# Run the agent
# Cloud Run will provide environment variables via Secret Manager
CMD ["python", "agent.py", "start"]
