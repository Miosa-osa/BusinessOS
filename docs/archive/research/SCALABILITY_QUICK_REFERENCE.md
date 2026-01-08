# Scalability Quick Reference

One-page guide for choosing the right architecture based on user scale.

---

## Decision Table

| Current Users | Action | Timeline | Cost Change |
|--------------|--------|----------|-------------|
| 0-100 | Deploy Phase 1 (MVP) | 1 week | $10-20/month |
| 100-500 | Optimize Phase 1 | 1-2 weeks | Same |
| 500-1000 | Add Redis + monitoring | 1 week | +$10/month |
| 1000-2000 | Deploy Phase 2 (Multi-server HA) | 2-3 weeks | $60-110/month |
| 2000-5000 | Add K3s OR migrate to managed K8s | 4-6 weeks | $200-400/month |
| 5000-10000 | Scale K8s, add read replicas | 2-3 weeks | $600-1200/month |
| 10000+ | Multi-region, global LB | 6-8 weeks | $1500+/month |

---

## What to Use When

### Session Management

| Users | Solution | Why |
|-------|----------|-----|
| 0-1000 | PostgreSQL | Simple, no extra cost, sufficient performance |
| 1000-5000 | PostgreSQL + Redis cache | 10x faster session lookups, still affordable |
| 5000+ | Managed Redis (Memorystore, ElastiCache) | Auto-scaling, high availability |

### Connection Pooling

| Servers | Solution | Pool Size | Why |
|---------|----------|-----------|-----|
| 1 server | pgx built-in | 50 max, 10 min | Zero overhead, direct connections |
| 2-3 servers | pgx built-in | 25 max each | Simple, no proxy needed |
| 3+ servers | PgBouncer | 25 DB conns, 1000 clients | Efficient connection reuse |
| 10+ servers | PgBouncer (transaction mode) | 50 DB conns, 5000+ clients | Maximum efficiency |

### Load Balancing

| Setup | Solution | Cost | Sticky Sessions |
|-------|----------|------|-----------------|
| Single server | None needed | $0 | N/A |
| 2-3 servers | Nginx (self-hosted) | $5/month (tiny VM) | IP hash OR cookie |
| 3-5 servers | HAProxy OR Nginx | $5/month | Cookie-based |
| Cloud deployment | Cloud LB (ALB, Azure LB, GCP LB) | $16-20/month | Built-in affinity |
| Budget option | Cloudflare Tunnel | $0 (free tier) | Automatic |

### Container Orchestration

| Containers | Solution | Cost | Complexity |
|------------|----------|------|------------|
| 0-100 | Docker on single VM | $10-20/month | Low |
| 100-500 | Docker with better resource limits | Same | Low |
| 500-2000 | Docker Swarm (3 nodes) | $40-60/month | Medium |
| 2000-5000 | K3s (self-managed) | $100-200/month | High |
| 5000+ | Managed K8s (GKE, EKS, AKS) | $200-600/month | Medium (managed) |

---

## Metrics to Monitor

### Critical Alerts (PagerDuty)

| Metric | Threshold | What to Do |
|--------|-----------|------------|
| P95 API Latency | >200ms for 5 min | Check DB connection pool, add server |
| Error Rate | >1% for 1 min | Check logs, rollback if recent deploy |
| DB Connection Pool | >90% for 2 min | Increase pool size OR add PgBouncer |
| CPU | >90% for 5 min | Scale horizontally (add server) |
| Memory | >90% for 5 min | Investigate memory leak, restart if needed |

### Warning Alerts (Slack)

| Metric | Threshold | What to Do |
|--------|-----------|------------|
| P95 API Latency | >100ms for 10 min | Investigate slow queries, optimize |
| DB Connection Pool | >80% for 5 min | Plan to add PgBouncer OR scale DB |
| WebSocket Errors | >5% for 5 min | Check container availability |
| Disk Usage | >80% | Clean old logs, expand disk |

---

## Phase 1 Optimization Checklist

Quick checklist for MVP optimization (0-500 users):

- [ ] Database indexes created (`idx_sessions_token_active`, etc.)
- [ ] Connection pool tuned (50 max, 10 min, 3s acquire timeout)
- [ ] Health check endpoint added (`/health`, `/health/deep`)
- [ ] Prometheus metrics exposed (`/metrics`)
- [ ] Grafana dashboards configured (request rate, latency, DB stats)
- [ ] Load tests run (k6 with 100 concurrent users)
- [ ] Monitoring alerts configured (Slack for warnings, PagerDuty for critical)
- [ ] Backup strategy implemented (daily PostgreSQL dumps)
- [ ] Documentation updated (architecture diagrams, runbooks)

---

## Phase 2 Migration Checklist

Multi-server HA migration (1000-2000 users):

- [ ] Provision 3 identical app servers (4 vCPU, 8GB RAM each)
- [ ] Set up Nginx load balancer with sticky sessions
- [ ] Deploy Redis for session caching
- [ ] Configure PgBouncer in transaction mode
- [ ] Update app to use Redis session store (with PostgreSQL fallback)
- [ ] Set up Docker Swarm across 3 nodes
- [ ] Implement graceful shutdown for zero-downtime deploys
- [ ] Configure health checks in load balancer
- [ ] Update monitoring for multi-server metrics
- [ ] Test failover (kill one server, ensure no downtime)
- [ ] Document new architecture and deploy process

---

## Phase 3 Migration Checklist

Kubernetes migration (2000-10000 users):

- [ ] Choose K8s platform (K3s, GKE, EKS, or AKS)
- [ ] Create Helm charts for backend service
- [ ] Implement HorizontalPodAutoscaler (min 3, max 20 pods)
- [ ] Migrate user containers to Kubernetes pods
- [ ] Set up namespace isolation (`user-workspaces`)
- [ ] Configure resource quotas and limits
- [ ] Deploy PgBouncer as Kubernetes service
- [ ] Migrate to managed database (Cloud SQL, RDS, or Azure SQL)
- [ ] Set up read replicas for database
- [ ] Configure network policies for security
- [ ] Implement CI/CD pipeline (GitOps with ArgoCD or Flux)
- [ ] Set up observability (Prometheus Operator, Grafana, Jaeger)
- [ ] Load test at scale (5000 concurrent users)
- [ ] Disaster recovery testing (simulate node failures)

---

## Cost Optimization Tips

### Immediate Savings (Phase 1)

1. **Use spot/preemptible instances** for user containers (70% savings)
   - Not for backend servers (need reliability)
   - Perfect for short-lived terminal sessions

2. **Right-size VMs**
   - Monitor actual CPU/memory usage for 1 week
   - Downsize if usage <50% consistently
   - Upsize only when >80% for sustained periods

3. **Cloudflare Tunnel (Free Tier)**
   - Replaces $20/month load balancer
   - Includes SSL, DDoS protection, CDN
   - Limit: 5 services (enough for MVP)

4. **Grafana Cloud Free Tier**
   - 10k metrics/month (enough for 1000 users)
   - 50GB logs/month
   - Upgrade to Pro ($29/month) only when needed

### Medium-Term Savings (Phase 2)

1. **Reserved Instances (AWS, Azure, GCP)**
   - 1-year commitment: 30-40% discount
   - 3-year commitment: 50-60% discount
   - Only for stable workloads (database, core backend)

2. **PgBouncer for Connection Pooling**
   - Reduces database size requirements
   - 25 connections serve 1000 clients
   - Saves $50-100/month on database costs

3. **Redis on Budget**
   - Upstash (serverless): Pay per request, free tier generous
   - Self-hosted Redis on tiny VM: $5/month vs $20/month managed

### Long-Term Savings (Phase 3)

1. **Kubernetes Spot Instances**
   - Use for stateless workloads (backend pods, user containers)
   - 70-90% savings vs on-demand
   - Combine with node autoscaling

2. **Database Read Replicas**
   - Cheaper than scaling primary vertically
   - Offload analytics, reports to replica
   - $200 read replica vs $500 larger primary

3. **Multi-Cloud Arbitrage**
   - Hetzner for compute: 50% cheaper than AWS/GCP
   - AWS/GCP for managed services (database, Redis)
   - CloudFlare for CDN/WAF (cheaper than cloud-native)

---

## Common Pitfalls

### Over-Engineering Too Early

- Don't deploy Kubernetes for <1000 users
- Don't use microservices if monolith works fine
- Don't add Redis if PostgreSQL session queries <50ms

### Under-Monitoring

- Always add Prometheus metrics from day 1
- Don't wait for production issues to add observability
- Free tier Grafana Cloud is enough for MVP

### Poor Connection Pool Tuning

- Default pgx pool (10 max) too small for production
- Acquire timeout too long (60s) causes request hangs
- Not monitoring pool exhaustion leads to mysterious errors

### Ignoring Security

- Always use TLS/HTTPS (Let's Encrypt is free)
- Never run containers as root
- Always implement rate limiting (even in MVP)

---

## Emergency Runbooks

### High Latency (P95 >200ms)

1. Check database connection pool: `curl localhost:8001/health/deep`
   - If pool usage >90%, increase `MaxConns` and restart
   - If pool usage <50%, investigate slow queries

2. Check database slow queries:
   ```sql
   SELECT query, calls, mean_time, max_time
   FROM pg_stat_statements
   ORDER BY mean_time DESC LIMIT 10;
   ```

3. Check CPU/memory:
   ```bash
   docker stats  # or kubectl top pods
   ```
   - If CPU >80%, scale horizontally (add server)
   - If memory >80%, investigate memory leak

### Database Connection Pool Exhausted

1. Temporary fix (restart backend):
   ```bash
   docker restart businessos-backend  # or kubectl rollout restart
   ```

2. Permanent fix (increase pool size):
   ```go
   poolConfig.MaxConns = 100  // Double current size
   ```

3. Long-term fix (add PgBouncer):
   - See Phase 2 migration checklist

### WebSocket Connections Failing

1. Check backend logs:
   ```bash
   docker logs businessos-backend --tail 100
   ```

2. Check container availability:
   ```bash
   docker ps | grep workspace  # or kubectl get pods -n user-workspaces
   ```

3. Check resource limits:
   ```bash
   # Docker
   docker stats

   # Kubernetes
   kubectl describe resourcequota -n user-workspaces
   ```

### Out of Memory (OOM) Kills

1. Identify culprit:
   ```bash
   dmesg | grep -i "killed process"
   ```

2. Analyze memory usage:
   ```bash
   # Backend
   docker stats businessos-backend

   # User containers
   docker stats | grep workspace
   ```

3. Fix:
   - If backend: Investigate memory leak, optimize, or add more RAM
   - If user containers: Lower per-container limits (512MB → 256MB)

---

## Support Contacts

### Cloud Provider Support

| Provider | MVP Support | Production Support | Cost |
|----------|------------|-------------------|------|
| Hetzner | Community forums (free) | Email support (free) | $0 |
| DigitalOcean | Ticket support (free) | Priority support | $0-100/month |
| AWS | Developer support | Business support | $29-100/month |
| GCP | Standard support (free) | Enhanced support | $150-400/month |
| Azure | Basic support (free) | Standard support | $100-300/month |

### Monitoring Tools

- Grafana Cloud: Community Slack, docs
- Prometheus: CNCF Slack (#prometheus)
- PostgreSQL: pgsql-general mailing list, Stack Overflow

---

## File Locations Reference

| Document | Path | Purpose |
|----------|------|---------|
| Full research | `/Users/ososerious/BusinessOS-1/docs/SCALABILITY_ARCHITECTURE_RESEARCH.md` | Detailed comparisons, cost analysis |
| Phase 1 guide | `/Users/ososerious/BusinessOS-1/docs/PHASE_1_IMPLEMENTATION_GUIDE.md` | Code snippets, configuration |
| Architecture diagrams | `/Users/ososerious/BusinessOS-1/docs/ARCHITECTURE_DIAGRAMS.md` | Visual representations |
| This quick reference | `/Users/ososerious/BusinessOS-1/docs/SCALABILITY_QUICK_REFERENCE.md` | One-page decisions |

---

## Next Steps

1. **This Week:** Implement Phase 1 optimizations
   - Database indexes, connection pool tuning, health checks
   - Estimated time: 4-8 hours

2. **Next 2 Weeks:** Set up monitoring
   - Prometheus metrics, Grafana dashboards, alerts
   - Estimated time: 4-6 hours

3. **Next Month:** Load testing
   - Run k6 tests, identify bottlenecks, optimize
   - Estimated time: 2-4 hours

4. **When Needed:** Scale to Phase 2
   - Triggered by metrics (P95 >100ms, users >1000, CPU >70%)
   - Estimated time: 1-2 weeks for migration

---

**Remember:** Don't optimize prematurely. Start with Phase 1, monitor closely, and scale based on actual metrics—not theoretical limits.
