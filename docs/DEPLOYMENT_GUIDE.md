# BusinessOS Deployment Guide

Production deployment guide with security hardening checklist.

---

## Prerequisites

| Requirement | Version | Purpose |
|-------------|---------|---------|
| Go | 1.21+ | Backend compilation |
| Node.js | 18+ | Frontend build |
| PostgreSQL | 15+ | Database |
| Docker | 24+ | Container isolation |
| Docker Compose | 2.x | Service orchestration |

---

## Quick Start (Development)

```bash
# 1. Clone repository
git clone https://github.com/robertohluna/BusinessOS.git
cd BusinessOS

# 2. Start all services
./startup.sh

# 3. Access application
open http://localhost:5173
```

---

## Production Deployment

### 1. Database Setup

```bash
# Create database
createdb business_os

# Apply schema
psql business_os < desktop/backend-go/internal/database/schema.sql

# Create production user (do not use superuser)
psql -c "CREATE USER businessos WITH PASSWORD 'secure-password';"
psql -c "GRANT CONNECT ON DATABASE business_os TO businessos;"
psql -c "GRANT USAGE ON SCHEMA public TO businessos;"
psql -c "GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO businessos;"
```

### 2. Environment Configuration

Create `.env` file:

```env
# Server
SERVER_PORT=8001
GIN_MODE=release

# Database (use connection pooling in production)
DATABASE_URL=postgresql://businessos:secure-password@localhost:5432/business_os?sslmode=require

# Security
SECRET_KEY=<generate-256-bit-key>
ALLOWED_ORIGINS=https://yourdomain.com

# AI Providers (configure as needed)
OLLAMA_LOCAL_URL=http://localhost:11434
GROQ_API_KEY=<your-key>
ANTHROPIC_API_KEY=<your-key>

# Google Calendar (optional)
GOOGLE_CLIENT_ID=<your-client-id>
GOOGLE_CLIENT_SECRET=<your-client-secret>
GOOGLE_REDIRECT_URL=https://yourdomain.com/api/integrations/google/callback
```

### 3. Build Backend

```bash
cd desktop/backend-go

# Build with optimizations
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-s -w" \
  -o server \
  ./cmd/server

# Verify build
./server --version
```

### 4. Build Frontend

```bash
cd frontend

# Install dependencies
npm ci --production

# Build for production
npm run build

# Output in build/ directory
```

### 5. Docker Deployment

```yaml
# docker-compose.prod.yml
version: '3.8'
services:
  backend:
    build: ./desktop/backend-go
    ports:
      - "8001:8001"
    environment:
      - GIN_MODE=release
      - DATABASE_URL=${DATABASE_URL}
      - SECRET_KEY=${SECRET_KEY}
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    restart: unless-stopped

  frontend:
    build: ./frontend
    ports:
      - "80:80"
      - "443:443"
    restart: unless-stopped

  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: business_os
      POSTGRES_USER: businessos
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - pgdata:/var/lib/postgresql/data
    restart: unless-stopped

volumes:
  pgdata:
```

---

## Security Hardening Checklist

### Pre-Deployment

| Item | Status | Notes |
|------|--------|-------|
| Generate strong SECRET_KEY (256-bit) | [ ] | `openssl rand -hex 32` |
| Configure ALLOWED_ORIGINS | [ ] | Exact domains, no wildcards |
| Enable PostgreSQL SSL | [ ] | `sslmode=require` in connection |
| Create dedicated database user | [ ] | No superuser access |
| Review firewall rules | [ ] | Only expose 80/443 |
| Configure rate limiting | [ ] | Built-in: 100 msg/sec |
| Enable HTTPS | [ ] | TLS 1.2+ required |

### Container Security (Built-in)

| Feature | Implementation | Verified |
|---------|----------------|----------|
| Read-only root filesystem | `ReadonlyRootfs: true` | [x] |
| Tmpfs for writable paths | `/tmp`, `/var/tmp`, `/run` | [x] |
| Capability dropping | `CapDrop: ALL` | [x] |
| Minimal capabilities | Only `CHOWN`, `FOWNER` | [x] |
| Custom Seccomp profile | Blocks 15+ syscalls | [x] |
| No new privileges | `no-new-privileges:true` | [x] |
| Network isolation | `NetworkMode: none` | [x] |
| Resource limits | 512MB RAM, 50% CPU, 100 PIDs | [x] |

### Input Validation (Built-in)

| Protection | Patterns Blocked | Verified |
|------------|------------------|----------|
| Fork bombs | `:(){ :|:& };:` | [x] |
| Dangerous rm | `rm -rf /`, `rm -rf /*` | [x] |
| Container escape | `nsenter`, `unshare` | [x] |
| Privilege escalation | `sudo`, `su root` | [x] |
| Reverse shells | `bash -i`, `/dev/tcp` | [x] |
| ANSI injection | OSC 8 hyperlinks, clipboard | [x] |

### Session Security (Built-in)

| Feature | Configuration | Verified |
|---------|---------------|----------|
| IP binding | Session locked to IP | [x] |
| Max session time | 8 hours | [x] |
| Idle timeout | 30 minutes | [x] |
| WebSocket origin validation | Configured origins only | [x] |
| Connection limit | 5 per user | [x] |

### Post-Deployment

| Item | Status | Notes |
|------|--------|-------|
| Enable application logging | [ ] | JSON format to aggregator |
| Set up monitoring | [ ] | Prometheus/Grafana |
| Configure alerting | [ ] | Rate limit breaches, errors |
| Enable backup automation | [ ] | Database + volumes |
| Document incident response | [ ] | Runbooks ready |
| Schedule security reviews | [ ] | Quarterly minimum |

---

## Monitoring

### Health Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/mcp/health` | GET | MCP service health |
| `/api/ai/system` | GET | AI system status |
| `/api/settings/system` | GET | System settings |

### Key Metrics to Monitor

| Metric | Threshold | Action |
|--------|-----------|--------|
| Terminal connections/user | > 5 | Rate limit triggered |
| Message rate | > 100/sec | Investigate abuse |
| Container memory | > 400MB | Review processes |
| Session duration | > 8h | Auto-terminate |
| Failed auth attempts | > 10/min | Block IP |

### Log Locations

| Component | Path | Format |
|-----------|------|--------|
| Backend | stdout/stderr | JSON |
| Container logs | Docker json-file | JSON |
| PostgreSQL | `/var/log/postgresql/` | Standard |

---

## Backup and Recovery

### Database Backup

```bash
# Daily backup
pg_dump -Fc business_os > backup_$(date +%Y%m%d).dump

# Restore
pg_restore -d business_os backup_20241222.dump
```

### Volume Backup

```bash
# Backup user workspaces
docker run --rm \
  -v workspace_user1:/data \
  -v $(pwd)/backups:/backup \
  alpine tar czf /backup/workspace_user1.tar.gz /data
```

### Recovery Procedure

1. Stop services: `docker-compose down`
2. Restore database: `pg_restore -c -d business_os backup.dump`
3. Restore volumes: Extract tar archives
4. Start services: `docker-compose up -d`
5. Verify health endpoints

---

## Troubleshooting

### Common Issues

| Symptom | Cause | Solution |
|---------|-------|----------|
| WebSocket fails | Origin mismatch | Check ALLOWED_ORIGINS |
| Container won't start | Docker socket permission | Add user to docker group |
| Database connection fails | SSL config | Use `sslmode=require` |
| Terminal unresponsive | Rate limit hit | Wait for token refill |
| Session drops | Idle timeout | Reconnect; increase timeout if needed |

### Debug Commands

```bash
# Check container security
docker inspect terminal-<user>-<session> | jq '.HostConfig.SecurityOpt'

# View rate limit status
curl -H "Cookie: session=..." http://localhost:8001/api/terminal/sessions

# Check database connections
psql -c "SELECT * FROM pg_stat_activity WHERE datname='business_os';"

# View terminal container logs
docker logs terminal-<user>-<session>
```

---

## Performance Tuning

### Benchmark Results (Phase 2)

| Operation | Time | Memory | Allocations |
|-----------|------|--------|-------------|
| AllowMessage | 285 ns | 0 B | 0 |
| QuickValidate | 468 ns | 0 B | 0 |
| FullValidation | 6.6 us | 725 B | 41 |
| ValidateIP | 162 ns | 80 B | 2 |

### Optimization Tips

| Area | Recommendation |
|------|----------------|
| Database | Enable connection pooling (pgbouncer) |
| Sessions | Use Redis for horizontal scaling |
| Containers | Pre-warm base images |
| Static assets | CDN for frontend |
| WebSocket | Load balancer with sticky sessions |

---

## Compliance

### OWASP Top 10 Coverage

| Risk | Mitigation | Status |
|------|-----------|--------|
| A01 Broken Access Control | Auth middleware, session validation | Implemented |
| A02 Cryptographic Failures | HTTPS required, secure secrets | Configured |
| A03 Injection | Input sanitization, parameterized queries | Implemented |
| A04 Insecure Design | Threat modeling, security reviews | Documented |
| A05 Security Misconfiguration | Hardened defaults, checklist | Provided |
| A06 Vulnerable Components | Dependency audit | Scheduled |
| A07 Auth Failures | Rate limiting, session security | Implemented |
| A08 Software Integrity | Container signing | Available |
| A09 Logging Failures | Sanitized logging | Implemented |
| A10 SSRF | Network isolation | Implemented |

---

## Support

| Resource | Location |
|----------|----------|
| Documentation | `/docs/` |
| API Reference | `/docs/API_REFERENCE.md` |
| Security Docs | `/docs/BUSINESSOS_TERMINAL_SYSTEM_DOCUMENTATION.md` |
| Issues | GitHub Issues |
