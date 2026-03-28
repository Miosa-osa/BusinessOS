# Sandbox Failure Runbook

**Version:** 1.0.0
**Last Updated:** 2026-03-27
**Status:** ACTIVE

---

## Executive Summary

This runbook provides step-by-step instructions for handling sandbox container failures. The sandbox environment (Docker-in-Docker) provides isolated execution for untrusted code and agent workflows.

**Risk Level:** HIGH - Sandbox failure disables code execution, AI workflow automation, and agent skill execution.

**Business Impact:** No code can be executed safely, AI workflows halt, agent skills unavailable. Core BusinessOS features remain functional.

---

## Table of Contents

1. [Detection Symptoms](#detection-symptoms)
2. [Immediate Actions](#immediate-actions)
3. [Resolution Steps](#resolution-steps)
4. [Verification](#verification)
5. [Escalation Path](#escalation-path)

---

## Detection Symptoms

### Automatic Detection

**Symptoms:**
- Sandbox health check fails
- Code execution requests timeout
- Docker daemon errors in logs
- Container creation failures
- Resource exhaustion (CPU/memory)

**Detection Time:** <15 seconds

### Manual Detection

**Symptoms:**
- Agent workflows hang at code execution
- "Sandbox unavailable" errors in UI
- Skill execution failures
- Terminal sessions won't start

**Check Commands:**
```bash
# Check Docker daemon status
docker info

# Check sandbox containers
docker ps -a | grep sandbox

# Check sandbox logs
docker logs --tail 50 businessos-sandbox

# Test container creation
docker run --rm hello-world
# Expected: "Hello from Docker!"
```

---

## Immediate Actions

### 🚨 First 5 Minutes

- [ ] **Confirm Failure:** Verify sandbox container not responding
- [ ] **Check Docker Daemon:** Ensure Docker daemon is running
- [ ] **Check Resources:** CPU/memory not exhausted
- [ ] **Review Logs:** Identify failure pattern
- [ ] **Notify Team:** Alert backend team via Slack #backend-dev

### Resource Check Commands

```bash
# Check system resources
docker stats --all --no-stream

# Check disk space
df -h

# Check Docker daemon logs
sudo journalctl -u docker -n 50

# Check container limits
docker inspect businessos-sandbox | grep -A 10 "Memory"
```

---

## Resolution Steps

### Option A: Sandbox Container Restart

**Use when:** Container crashed but Docker daemon is healthy.

```bash
# 1. Navigate to BusinessOS directory
cd /Users/sac/chatmangpt/BusinessOS

# 2. Restart sandbox container
docker compose restart sandbox

# 3. Monitor startup
docker logs -f businessos-sandbox

# Expected output:
# [info] Sandbox initialized
# [info] Docker socket available
# [info] Ready to execute code
```

**Recovery Time:** 10-20 seconds

---

### Option B: Docker-in-Docker (DinD) Rebuild

**Use when:** DinD environment corrupted or Docker socket issues.

```bash
# 1. Stop sandbox container
docker compose stop sandbox

# 2. Remove sandbox volume (contains DinD state)
docker volume rm businessos_sandbox_dind

# 3. Rebuild sandbox image
docker compose build sandbox

# 4. Start sandbox container
docker compose up -d sandbox

# 5. Verify DinD is working
docker exec businessos-sandbox docker info
```

**Recovery Time:** 30-60 seconds

---

### Option C: Full Docker Daemon Recovery

**Use when:** Docker daemon itself is not responding.

```bash
# 1. Check Docker daemon status
sudo systemctl status docker

# 2. Restart Docker daemon
sudo systemctl restart docker

# 3. Wait for daemon to start (10-20 seconds)
docker info

# 4. Restart sandbox container
docker compose restart sandbox

# 5. Verify sandbox can create containers
docker exec businessos-sandbox docker run --rm hello-world
```

**Recovery Time:** 30-45 seconds

---

### Option D: Resource Limit Adjustment

**Use when:** Sandbox failing due to CPU/memory constraints.

```bash
# 1. Check current limits
docker inspect businessos-sandbox | grep -A 10 "Memory"

# 2. Edit docker-compose.yml to increase limits
# services:
#   sandbox:
#     deploy:
#       resources:
#         limits:
#           cpus: '2.0'
#           memory: 2G
#         reservations:
#           cpus: '1.0'
#           memory: 1G

# 3. Apply changes
docker compose up -d sandbox

# 4. Verify new limits
docker exec businessos-sandbox free -h
docker exec businessos-sandbox nproc
```

**Recovery Time:** 15-30 seconds

---

### Option E: Docker Socket Permission Fix

**Use when:** Sandbox can't access Docker socket due to permissions.

```bash
# 1. Check socket permissions
ls -l /var/run/docker.sock

# 2. Fix permissions if needed
sudo chmod 666 /var/run/docker.sock

# 3. Add sandbox user to docker group
docker exec businessos-sandbox usermod -aG docker appuser

# 4. Restart sandbox
docker compose restart sandbox

# 5. Verify socket access
docker exec businessos-sandbox docker ps
```

**Recovery Time:** 10-15 seconds

---

## Verification

### Step 1: Sandbox Container Health

```bash
# Check container is running
docker ps | grep businessos-sandbox

# Check container logs
docker logs --tail 20 businessos-sandbox

# Expected: "Sandbox ready", "DinD initialized"
```

### Step 2: Docker-in-Docker Functionality

```bash
# Test DinD can create containers
docker exec businessos-sandbox docker run --rm hello-world
# Expected: "Hello from Docker!"

# Test DinD can list containers
docker exec businessos-sandbox docker ps -a

# Test DinD info
docker exec businessos-sandbox docker info
# Expected: Docker version, containers, images
```

### Step 3: Code Execution via API

```bash
# Test code execution endpoint
curl -X POST http://localhost:8001/api/sandbox/execute \
  -H "Content-Type: application/json" \
  -d '{
    "language": "python",
    "code": "print(\"Hello from sandbox\")"
  }'

# Expected: {"output":"Hello from sandbox\n","status":"success"}
```

### Step 4: Agent Skill Execution

```bash
# Test skill execution through OSA
curl -X POST http://localhost:8089/api/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "python",
    "input": {"code": "print(\"test\")"}
  }'

# Expected: Skill execution result
```

### Step 5: Resource Availability

```bash
# Check sandbox resources
docker exec businessos-sandbox free -h
# Expected: Available memory >512MB

# Check CPU available
docker exec businessos-sandbox nproc
# Expected: >=1 CPU core

# Check disk space
docker exec businessos-sandbox df -h
# Expected: Available space >1GB
```

---

## Escalation Path

| Level | Contact | When to Escalate | Response Time |
|-------|---------|------------------|---------------|
| **L1** | Backend Team | Container restart, permission issues | 15 minutes |
| **L2** | Tech Lead | DinD corruption, resource limits | 30 minutes |
| **L3** | DevOps Engineer | Docker daemon issues, kernel problems | 1 hour |
| **L4** | CTO | Host system issues, extended outage | Immediate |

### Support Channels

- **Slack:** #backend-dev, #incident-response
- **Email:** backend-team@businessos.com
- **On-Call:** Check PagerDuty for current on-call engineer

---

## Post-Incident Actions

### Immediate (Within 1 Hour)

- [ ] **Verify Sandbox Working:** Code execution succeeds
- [ ] **Test Agent Skills:** All skills functional
- [ ] **Monitor Resources:** No exhaustion
- [ ] **Update Team:** Post resolution summary
- [ ] **Close Incident:** Mark as resolved

### Short-Term (Within 24 Hours)

- [ ] **Analyze Root Cause:** Why did sandbox fail?
- [ ] **Review Logs:** Check for patterns
- [ ] **Update Monitoring:** Add alerts for resource usage
- [ ] **Document Learnings:** Update runbook

### Long-Term (Within 1 Week)

- [ ] **Improve Isolation:** Add stricter resource limits
- [ ] **Implement Health Checks:** Continuous sandbox monitoring
- [ ] **Add Auto-Recovery:** Auto-restart on failure
- [ ] **Chaos Test:** Add sandbox failure scenario

---

## Common Issues and Solutions

### Issue 1: "Docker Socket Not Available"

**Symptoms:**
```
Error: Cannot connect to Docker daemon at unix:///var/run/docker.sock
```

**Solution:**
```bash
# Check if host Docker daemon is running
sudo systemctl status docker

# Start if stopped
sudo systemctl start docker

# Check socket exists
ls -l /var/run/docker.sock

# Restart sandbox
docker compose restart sandbox
```

---

### Issue 2: "Permission Denied" on Docker Socket

**Symptoms:**
```
Got permission denied while trying to connect to the Docker daemon socket
```

**Solution:**
```bash
# Fix socket permissions (temporary)
sudo chmod 666 /var/run/docker.sock

# Or add user to docker group (permanent)
sudo usermod -aG docker $USER

# Restart sandbox
docker compose restart sandbox
```

---

### Issue 3: "No Space Left on Device"

**Symptoms:**
```
Error: no space left on device
```

**Solution:**
```bash
# Check disk space
df -h

# Clean up unused Docker resources
docker system prune -a --volumes

# Clean up sandbox containers
docker exec businessos-sandbox docker system prune -a

# Remove old sandbox images
docker images | grep sandbox | awk '{print $3}' | xargs docker rmi

# Restart sandbox
docker compose restart sandbox
```

---

### Issue 4: "Container Creation Timeout"

**Symptoms:**
- Code execution requests timeout after 30 seconds
- Containers stuck in "created" state

**Solution:**
```bash
# Check for stuck containers
docker ps -a | grep sandbox

# Remove stuck containers
docker ps -a | grep sandbox | awk '{print $1}' | xargs docker rm -f

# Restart Docker daemon
sudo systemctl restart docker

# Restart sandbox
docker compose restart sandbox
```

---

### Issue 5: Memory Exhaustion in Sandbox

**Symptoms:**
```
Container killed due to OOM (Out of Memory)
```

**Solution:**
```bash
# Check current memory usage
docker stats businessos-sandbox --no-stream

# Increase memory limit in docker-compose.yml
# services:
#   sandbox:
#     deploy:
#       resources:
#         limits:
#           memory: 2G

# Apply changes
docker compose up -d sandbox

# Verify new limit
docker inspect businessos-sandbox | grep -A 5 "Memory"
```

---

## Quick Reference Commands

```bash
# Check sandbox status
docker ps | grep businessos-sandbox

# View sandbox logs
docker logs -f businessos-sandbox

# Restart sandbox
docker compose restart sandbox

# Rebuild sandbox
docker compose build sandbox && docker compose up -d sandbox

# Test DinD
docker exec businessos-sandbox docker run --rm hello-world

# Check DinD containers
docker exec businessos-sandbox docker ps -a

# Clean up DinD containers
docker exec businessos-sandbox docker system prune -f

# Check resources
docker stats businessos-sandbox --no-stream

# Fix Docker socket permissions
sudo chmod 666 /var/run/docker.sock

# Restart Docker daemon
sudo systemctl restart docker

# Full cleanup
docker system prune -a --volumes
```

---

## Prevention Measures

### Monitoring Alerts

Configure alerts for:
- Sandbox container down (threshold: immediate)
- Code execution timeout (threshold: >30s)
- Memory usage >80% (threshold: warning)
- Docker socket errors (threshold: any)
- Disk space <10% (threshold: critical)

### Resource Limits

**In `docker-compose.yml`:**
```yaml
services:
  sandbox:
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 2G
        reservations:
          cpus: '1.0'
          memory: 1G
    environment:
      - SANDBOX_TIMEOUT=30
      - SANDBOX_MAX_CONTAINERS=10
```

### Auto-Restart Policy

```yaml
services:
  sandbox:
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "docker", "ps"]
      interval: 10s
      timeout: 5s
      retries: 3
```

### Container Cleanup Job

**Add to crontab:**
```bash
# Clean up old sandbox containers every hour
0 * * * * docker exec businessos-sandbox docker system prune -f --filter "until=1h"
```

---

**Document Version:** 1.0.0
**Last Reviewed:** 2026-03-27
**Next Review:** 2026-04-27
**Maintained By:** Backend Team
