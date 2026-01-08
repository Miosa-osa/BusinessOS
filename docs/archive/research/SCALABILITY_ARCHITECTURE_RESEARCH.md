# BusinessOS Go Backend - Scalability Architecture Research

**Date:** 2025-12-22
**Current State:** MVP with Go 1.25, Gin, PostgreSQL, Docker containers, WebSocket terminals
**Focus:** Startup-friendly ($5-20/month), horizontally scalable architecture

---

## Table of Contents

1. [Session Management: Redis vs PostgreSQL](#1-session-management-redis-vs-postgresql)
2. [Connection Pooling: PgBouncer vs pgx Built-in](#2-connection-pooling-pgbouncer-vs-pgx-built-in)
3. [Load Balancing for WebSocket Sticky Sessions](#3-load-balancing-for-websocket-sticky-sessions)
4. [Container Orchestration: Docker vs Kubernetes](#4-container-orchestration-docker-vs-kubernetes)
5. [Cost Analysis by User Scale](#5-cost-analysis-by-user-scale)
6. [Implementation Roadmap](#6-implementation-roadmap)

---

## 1. Session Management: Redis vs PostgreSQL

### Current Implementation
- **BetterAuth** session-based authentication
- Sessions likely stored in PostgreSQL (default)
- No distributed caching layer

### Comparison Table

| Criteria | PostgreSQL Sessions | Redis Sessions | **Recommendation** |
|----------|-------------------|----------------|-------------------|
| **Read Latency** | 5-20ms (disk I/O) | 0.5-2ms (in-memory) | Redis wins 10x faster |
| **Write Latency** | 10-50ms (ACID overhead) | 1-3ms (async replication) | Redis wins |
| **Horizontal Scaling** | Requires read replicas | Native multi-node support | Redis easier |
| **Data Persistence** | Full ACID guarantees | AOF/RDB (eventual consistency) | PostgreSQL safer |
| **Cost (100 users)** | $0 (piggyback on DB) | $5/month (shared Redis) | PostgreSQL cheaper |
| **Cost (1000 users)** | $0 (same DB) | $10/month (dedicated Redis) | PostgreSQL cheaper |
| **Cost (10k users)** | Read replicas needed (+$20) | $20-40/month (Redis cluster) | Similar |
| **Complexity** | Zero (already using DB) | Add Redis dependency | PostgreSQL simpler |
| **Session Expiry** | Manual cleanup queries | Native TTL (automatic) | Redis wins |
| **Concurrent Sessions** | Locks can bottleneck | Lock-free data structures | Redis wins |
| **Cold Start** | Always available | Risk of cache miss | PostgreSQL safer |

### Analysis

**For Startup Mode (100-1000 users):**
- **Use PostgreSQL** for sessions to minimize costs and complexity
- pgx connection pooling handles 1000 concurrent sessions easily
- Session reads are ~5% of total queries (mostly auth checks)
- Add proper indexes: `CREATE INDEX idx_sessions_token ON sessions(token) WHERE expires_at > NOW();`

**Migration Trigger (1000+ users):**
- Monitor session query latency with Prometheus/Grafana
- If P95 latency > 50ms OR CPU > 70%, migrate to Redis
- Use hybrid approach: Redis for active sessions, PostgreSQL for long-term storage

### Implementation Pattern (Future Redis Migration)

```go
// Session store interface for easy swapping
type SessionStore interface {
    Get(ctx context.Context, token string) (*Session, error)
    Set(ctx context.Context, session *Session, ttl time.Duration) error
    Delete(ctx context.Context, token string) error
}

// PostgreSQL implementation (current)
type PgSessionStore struct {
    pool *pgxpool.Pool
}

// Redis implementation (future)
type RedisSessionStore struct {
    client *redis.Client
    fallback SessionStore // PostgreSQL as fallback
}
```

**Recommendation:** Start with PostgreSQL, design for Redis migration at 1000+ users.

---

## 2. Connection Pooling: PgBouncer vs pgx Built-in

### Current Implementation
```go
poolConfig.MaxConns = 25
poolConfig.MinConns = 5
poolConfig.MaxConnLifetime = time.Hour
poolConfig.MaxConnIdleTime = 30 * time.Minute
```

### Comparison Table

| Criteria | pgx Built-in Pool | PgBouncer | **Recommendation** |
|----------|------------------|-----------|-------------------|
| **Connection Overhead** | 25 conns × N servers | 25 conns total (shared) | PgBouncer wins |
| **Setup Complexity** | Zero (in-process) | Separate service/container | pgx simpler |
| **Transaction Pooling** | No (session-level only) | Yes (transaction mode) | PgBouncer advanced |
| **Prepared Statements** | Fully supported | Limited (session mode) | pgx better |
| **Latency Overhead** | 0ms (direct) | 0.5-1ms (proxy hop) | pgx slightly faster |
| **Max Connections** | Limited by Go memory | 10,000+ with low RAM | PgBouncer scales better |
| **Cost (100 users)** | $0 | $0 (same VM) | Equal |
| **Cost (1000 users)** | 3 servers × 25 = 75 conns | 25 conns shared | PgBouncer $10/month savings |
| **Cost (10k users)** | 20 servers × 25 = 500 conns | 100 conns shared | PgBouncer $100/month savings |
| **Health Checks** | Per-pool monitoring | Centralized metrics | PgBouncer easier |
| **Connection Reuse** | Per-server pool | Global pool | PgBouncer wins |

### Analysis

**Current Settings Are Good for MVP:**
- 25 max connections handles ~250 concurrent requests (10 req/sec per conn)
- 5 min connections reduce cold-start latency
- 1-hour max lifetime prevents stale connections

**Optimization for 100-1000 Users:**
```go
// Tuned for single-server deployment
poolConfig.MaxConns = 50  // 2x increase for burst traffic
poolConfig.MinConns = 10  // Keep warm connections ready
poolConfig.MaxConnLifetime = 2 * time.Hour
poolConfig.MaxConnIdleTime = 15 * time.Minute  // Faster cleanup
poolConfig.HealthCheckPeriod = 30 * time.Second  // More frequent checks

// Add connection acquisition timeout
poolConfig.ConnectTimeout = 5 * time.Second
poolConfig.AcquireTimeout = 3 * time.Second  // Fail fast under load
```

**When to Add PgBouncer (1000+ users, 3+ servers):**

```yaml
# docker-compose.yml addition
pgbouncer:
  image: pgbouncer/pgbouncer:latest
  environment:
    - DATABASES_HOST=postgres
    - DATABASES_PORT=5432
    - DATABASES_USER=rhl
    - DATABASES_PASSWORD=password
    - DATABASES_DBNAME=business_os
    - POOL_MODE=transaction  # Best for stateless APIs
    - MAX_CLIENT_CONN=1000
    - DEFAULT_POOL_SIZE=25
    - RESERVE_POOL_SIZE=5
  ports:
    - "6432:6432"
```

**PgBouncer Transaction Mode Benefits:**
- Each HTTP request = 1 transaction
- Connection returned to pool immediately after query
- 25 DB connections serve 1000+ concurrent requests
- PostgreSQL limit (100 conns) supports 4+ app servers

**Recommendation:** Use pgx built-in pool for MVP, add PgBouncer at 3+ servers (1000+ users).

---

## 3. Load Balancing for WebSocket Sticky Sessions

### Challenge
WebSocket terminals require persistent connections to the same server instance.

### Comparison Table

| Solution | Sticky Sessions | Cost | Complexity | Cloud Support | **Best For** |
|----------|----------------|------|------------|---------------|--------------|
| **Nginx** | IP hash / Cookie | $0 (self-hosted) | Low | All clouds | Startup MVP |
| **HAProxy** | Cookie / Source IP | $0 (self-hosted) | Medium | All clouds | Advanced routing |
| **AWS ALB** | Target group stickiness | $16/month + data | Low | AWS only | AWS-native |
| **Azure Load Balancer** | Session affinity | $18/month + data | Low | Azure only | Azure-native |
| **GCP Load Balancer** | Session affinity | $18/month + data | Low | GCP only | GCP-native |
| **Cloudflare Tunnel** | Free tier (5 services) | $0-7/month | Low | Multi-cloud | Cost-effective |
| **Traefik** | Cookie-based | $0 (self-hosted) | Medium | Kubernetes | Container-native |

### Nginx Configuration (Recommended for MVP)

```nginx
# /etc/nginx/nginx.conf
upstream businessos_backend {
    # IP hash for sticky sessions (simple but limited)
    ip_hash;

    # OR cookie-based (better for mobile/dynamic IPs)
    # sticky cookie srv_id expires=1h domain=.yourdomain.com path=/;

    server backend1:8001 max_fails=3 fail_timeout=30s;
    server backend2:8001 max_fails=3 fail_timeout=30s;
    server backend3:8001 max_fails=3 fail_timeout=30s;
}

server {
    listen 80;
    server_name api.yourdomain.com;

    # WebSocket upgrade headers
    location /api/terminal/ws {
        proxy_pass http://businessos_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Timeouts for long-lived connections
        proxy_connect_timeout 7d;
        proxy_send_timeout 7d;
        proxy_read_timeout 7d;
    }

    # Regular HTTP endpoints
    location / {
        proxy_pass http://businessos_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Advanced: Cookie-Based Sticky Sessions

```go
// internal/middleware/loadbalancer.go
func LoadBalancerCookie() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Set server ID cookie for sticky sessions
        serverID := os.Getenv("SERVER_ID") // e.g., "server-1"
        if serverID == "" {
            serverID = "server-1"
        }

        c.SetCookie(
            "srv_id",           // name
            serverID,           // value
            3600,               // max age (1 hour)
            "/",                // path
            "",                 // domain
            false,              // secure (true for HTTPS)
            true,               // httpOnly
        )

        c.Next()
    }
}
```

### Cost Breakdown (1000 concurrent users)

| Solution | Monthly Cost | Data Transfer | Setup Time | Notes |
|----------|-------------|---------------|------------|-------|
| Nginx (self-hosted) | $5 (tiny VM) | $0 (internal) | 2 hours | Best ROI |
| HAProxy (self-hosted) | $5 (tiny VM) | $0 (internal) | 3 hours | More features |
| AWS ALB | $16 base + $8/GB | $50 (500GB out) | 30 mins | Managed service |
| Cloudflare Tunnel | $0 (free tier) | $0 (no egress) | 1 hour | Hidden gem |

**Cloudflare Tunnel Setup (Free Tier):**

```bash
# Install cloudflared
curl -L https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 -o cloudflared
sudo mv cloudflared /usr/local/bin/
sudo chmod +x /usr/local/bin/cloudflared

# Authenticate and create tunnel
cloudflared tunnel login
cloudflared tunnel create businessos
cloudflared tunnel route dns businessos api.yourdomain.com

# Run tunnel (supports WebSockets natively)
cloudflared tunnel run --url http://localhost:8001 businessos
```

**Recommendation:** Start with Nginx (simplest), migrate to Cloudflare Tunnel (free + WebSocket support) or cloud LB (managed) at scale.

---

## 4. Container Orchestration: Docker vs Kubernetes

### Current Implementation
- User terminals run in isolated Docker containers
- Docker API via `/var/run/docker.sock`
- Manual container lifecycle management

### Comparison Table

| Criteria | Single VM Docker | Docker Swarm | Kubernetes (K3s) | Managed K8s | **Best For** |
|----------|-----------------|--------------|------------------|-------------|--------------|
| **Setup Time** | 30 mins | 2 hours | 4 hours | 1 hour | Docker wins |
| **Cost (100 users)** | $5/month (1 VM) | $10/month (3 VMs) | $20/month (control + workers) | $75/month (managed) | Docker cheapest |
| **Cost (1000 users)** | $20/month (1 large VM) | $40/month (3 VMs) | $60/month (self-managed) | $150/month (managed) | Swarm/K3s competitive |
| **Cost (10k users)** | Not feasible | $200/month (10 VMs) | $300/month (cluster) | $500/month (managed) | Swarm/K3s cheaper |
| **Max Containers** | 100-200 (resource limits) | 1000+ (multi-node) | 10,000+ (scales horizontally) | Unlimited | K8s scales best |
| **Auto-scaling** | Manual scripts | Basic swarm scaling | HPA + cluster autoscaler | Fully managed | K8s best |
| **Health Checks** | Manual monitoring | Built-in health checks | Liveness/readiness probes | Managed monitoring | K8s most robust |
| **Networking** | Docker bridge | Overlay network | CNI plugins (Calico, Flannel) | Managed VPC | K8s most flexible |
| **Security** | AppArmor/Seccomp | Same + Swarm secrets | Pod security policies + RBAC | Managed security | K8s most secure |
| **Learning Curve** | 1 week | 2 weeks | 6 weeks | 2 weeks | Docker easiest |
| **Operational Overhead** | Low (cron jobs) | Medium (3-node cluster) | High (cluster management) | Low (cloud-managed) | Docker/Managed easiest |

### Analysis by Scale

#### 100 Users (~50 concurrent containers)
**Recommendation: Single VM Docker**

```yaml
# Optimized docker-compose.yml for user containers
version: '3.8'
services:
  workspace-pool:
    image: businessos-workspace:latest
    deploy:
      replicas: 50
      resources:
        limits:
          cpus: '0.5'
          memory: 512M
        reservations:
          cpus: '0.1'
          memory: 128M
```

**VM Specs:**
- 4 vCPUs, 8GB RAM ($10-20/month on Hetzner, DigitalOcean, Linode)
- 50 containers × 512MB = 25GB max (with overcommit)
- Docker resource limits prevent noisy neighbors

#### 1000 Users (~500 concurrent containers)
**Recommendation: Docker Swarm OR K3s**

**Docker Swarm (Simpler):**

```yaml
# Deploy stack across 3 nodes
docker stack deploy -c docker-compose.yml businessos

# docker-compose.yml
version: '3.8'
services:
  backend:
    image: businessos-backend:latest
    deploy:
      replicas: 3
      placement:
        max_replicas_per_node: 1
      resources:
        limits:
          cpus: '2'
          memory: 2G

  workspace-template:
    image: businessos-workspace:latest
    deploy:
      mode: replicated
      replicas: 500
      resources:
        limits:
          cpus: '0.5'
          memory: 512M
```

**K3s (Better long-term):**

```yaml
# namespace for user workspaces
apiVersion: v1
kind: Namespace
metadata:
  name: user-workspaces

---
# ResourceQuota to limit total resources
apiVersion: v1
kind: ResourceQuota
metadata:
  name: workspace-quota
  namespace: user-workspaces
spec:
  hard:
    requests.cpu: "200"
    requests.memory: 200Gi
    limits.cpu: "400"
    limits.memory: 400Gi
    pods: "1000"

---
# Pod template for user containers
apiVersion: v1
kind: Pod
metadata:
  name: user-{{ .UserID }}-workspace
  namespace: user-workspaces
spec:
  containers:
  - name: terminal
    image: businessos-workspace:latest
    resources:
      requests:
        memory: "128Mi"
        cpu: "100m"
      limits:
        memory: "512Mi"
        cpu: "500m"
    securityContext:
      runAsNonRoot: true
      runAsUser: 1000
      capabilities:
        drop: [ALL]
```

**Cost Comparison (1000 users):**

| Setup | Monthly Cost | Pros | Cons |
|-------|-------------|------|------|
| Single VM (8 vCPU, 32GB) | $40 | Simple, cheap | No HA, limited scale |
| Docker Swarm (3×4 vCPU, 16GB) | $60 | HA, easier than K8s | Less ecosystem |
| K3s (3×4 vCPU, 16GB) | $60 | Full K8s, extensible | Steeper learning |
| GKE/EKS/AKS (managed) | $150 | Managed, autoscale | 3x cost |

#### 10,000 Users (~5000 concurrent containers)
**Recommendation: Managed Kubernetes (GKE, EKS, AKS)**

**Why Managed K8s:**
- Cluster autoscaling handles traffic spikes
- Node pools isolate user workloads from backend services
- Integrated monitoring (Prometheus, Grafana via cloud)
- Security compliance (SOC2, HIPAA) built-in

**Example Architecture:**

```yaml
# Node pool for backend services (stable)
gcloud container node-pools create backend-pool \
  --cluster=businessos \
  --machine-type=n2-standard-4 \
  --num-nodes=3 \
  --enable-autoscaling \
  --min-nodes=3 \
  --max-nodes=10

# Node pool for user containers (autoscale)
gcloud container node-pools create workspace-pool \
  --cluster=businessos \
  --machine-type=n2-standard-8 \
  --num-nodes=5 \
  --enable-autoscaling \
  --min-nodes=5 \
  --max-nodes=50 \
  --spot  # Use spot instances for 70% cost savings
```

**Recommendation Ladder:**
1. **0-500 users:** Single VM Docker ($10-20/month)
2. **500-2000 users:** Docker Swarm OR K3s ($40-80/month)
3. **2000+ users:** Managed Kubernetes ($200-500/month)

---

## 5. Cost Analysis by User Scale

### Infrastructure Cost Breakdown

#### 100 Concurrent Users

| Component | Service | Specs | Monthly Cost |
|-----------|---------|-------|--------------|
| **App Server** | Hetzner CPX21 | 3 vCPU, 4GB RAM | $6 |
| **Database** | Same VM (PostgreSQL) | Shared resources | $0 |
| **Redis** | Not needed yet | - | $0 |
| **Load Balancer** | Cloudflare Tunnel | Free tier | $0 |
| **Container Host** | Same VM (Docker) | 50 containers | $0 |
| **Monitoring** | Grafana Cloud Free | 10k metrics | $0 |
| **Total** | | | **$6/month** |

**Per-user cost:** $0.06/month
**Revenue target:** $5/user/month (83x margin)

---

#### 1,000 Concurrent Users

| Component | Service | Specs | Monthly Cost |
|-----------|---------|-------|--------------|
| **App Servers (3)** | Hetzner CPX31 | 4 vCPU, 8GB RAM × 3 | $36 |
| **Database** | Hetzner CPX21 | 3 vCPU, 4GB RAM + 100GB storage | $10 |
| **Redis** | Upstash (serverless) | 10GB data, 10M commands | $10 |
| **PgBouncer** | Same as app server | No extra cost | $0 |
| **Load Balancer** | Nginx on app servers | No extra cost | $0 |
| **Container Host** | Docker Swarm (same servers) | 500 containers | $0 |
| **Monitoring** | Grafana Cloud Pro | 50k metrics | $29 |
| **CDN/SSL** | Cloudflare Pro | DDoS protection | $20 |
| **Backups** | Hetzner Backups | Daily snapshots | $5 |
| **Total** | | | **$110/month** |

**Per-user cost:** $0.11/month
**Revenue target:** $5/user/month (45x margin)

---

#### 10,000 Concurrent Users

| Component | Service | Specs | Monthly Cost |
|-----------|---------|-------|--------------|
| **App Servers** | GKE Autopilot (10 pods) | 2 vCPU, 4GB RAM each | $200 |
| **Database** | Cloud SQL (GCP) | 4 vCPU, 16GB RAM, 500GB SSD | $250 |
| **Read Replica** | Cloud SQL (GCP) | 4 vCPU, 16GB RAM (read-only) | $200 |
| **Redis** | GCP Memorystore | 5GB, standard tier | $50 |
| **PgBouncer** | GKE sidecar | No extra cost | $0 |
| **Load Balancer** | GCP Cloud Load Balancer | 1TB data processed | $40 |
| **Container Host** | GKE (separate node pool) | 20 nodes × n2-standard-4 (spot) | $600 |
| **Monitoring** | GCP Cloud Monitoring | Included with GCP | $50 |
| **CDN** | Cloudflare Enterprise | 10TB bandwidth | $200 |
| **Backups** | GCS snapshots | Daily full + incremental | $30 |
| **Total** | | | **$1,620/month** |

**Per-user cost:** $0.16/month
**Revenue target:** $5/user/month (31x margin)

---

### Cost Scaling Summary

```text
Users       Monthly Cost    Per-User Cost    Revenue (@ $5/user)    Margin
-----       ------------    -------------    -------------------    ------
100         $6              $0.06            $500                   98.8%
500         $50             $0.10            $2,500                 98.0%
1,000       $110            $0.11            $5,000                 97.8%
5,000       $400            $0.08            $25,000                98.4%
10,000      $1,620          $0.16            $50,000                96.8%
50,000      $6,000          $0.12            $250,000               97.6%
```
**Key Insights:**
- Economies of scale kick in at 5,000+ users (per-user cost drops)
- Infrastructure costs stay <5% of revenue at all scales
- Biggest cost jumps: 100→1000 (need Redis + HA), 5k→10k (managed services)

---

## 6. Implementation Roadmap

### Phase 1: MVP Optimization (0-500 users)
**Timeline:** 1-2 weeks
**Cost:** $10-20/month
**Goal:** Maximize performance on single VM

#### Tasks

1. **Optimize pgx Connection Pool**
```go
poolConfig.MaxConns = 50
poolConfig.MinConns = 10
poolConfig.AcquireTimeout = 3 * time.Second
poolConfig.ConnectTimeout = 5 * time.Second
```

2. **Add Database Indexes**
```sql
-- Session lookup optimization
CREATE INDEX CONCURRENTLY idx_sessions_token ON sessions(token)
WHERE expires_at > NOW();

-- User lookup optimization
CREATE INDEX CONCURRENTLY idx_users_email ON users(email);

-- WebSocket session tracking
CREATE INDEX CONCURRENTLY idx_terminal_sessions_user_active
ON terminal_sessions(user_id, created_at)
WHERE closed_at IS NULL;
```

3. **Implement Health Checks**
```go
// internal/handlers/health.go
func (h *Handlers) HealthCheck(c *gin.Context) {
    health := gin.H{
        "status": "healthy",
        "timestamp": time.Now().Unix(),
    }

    // Check database
    ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
    defer cancel()

    if err := h.pool.Ping(ctx); err != nil {
        health["database"] = "unhealthy"
        health["status"] = "degraded"
        c.JSON(503, health)
        return
    }
    health["database"] = "healthy"

    // Check Docker (if available)
    if h.containerMgr != nil {
        health["containers"] = "healthy"
    }

    c.JSON(200, health)
}
```

4. **Add Metrics Endpoint (Prometheus)**
```go
import "github.com/gin-gonic/gin"
import "github.com/prometheus/client_golang/prometheus/promhttp"

// Register metrics
router.GET("/metrics", gin.WrapH(promhttp.Handler()))

// Add custom metrics
var (
    activeTerminals = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "businessos_active_terminals",
        Help: "Number of active terminal sessions",
    })

    dbConnections = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "businessos_db_connections",
        Help: "Number of active database connections",
    })
)
```

**Expected Results:**
- P95 latency <100ms for all endpoints
- Support 500 concurrent WebSocket sessions
- Database CPU <50% under normal load

---

### Phase 2: Multi-Server HA (500-2000 users)
**Timeline:** 2-3 weeks
**Cost:** $60-100/month
**Goal:** High availability with 3 app servers

#### Tasks

1. **Deploy Nginx Load Balancer**
```nginx
# /etc/nginx/conf.d/businessos.conf
upstream backend {
    ip_hash;  # Sticky sessions
    server 10.0.1.10:8001 max_fails=3 fail_timeout=30s;
    server 10.0.1.11:8001 max_fails=3 fail_timeout=30s;
    server 10.0.1.12:8001 max_fails=3 fail_timeout=30s;
}

server {
    listen 443 ssl http2;
    server_name api.businessos.com;

    ssl_certificate /etc/ssl/certs/businessos.crt;
    ssl_certificate_key /etc/ssl/private/businessos.key;

    location /api/terminal/ws {
        proxy_pass http://backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_read_timeout 7d;
    }

    location / {
        proxy_pass http://backend;
        proxy_set_header Host $host;
    }
}
```

2. **Add Redis for Sessions**
```go
// internal/session/redis.go
import "github.com/redis/go-redis/v9"

type RedisStore struct {
    client *redis.Client
    fallback SessionStore  // PostgreSQL fallback
}

func (r *RedisStore) Get(ctx context.Context, token string) (*Session, error) {
    val, err := r.client.Get(ctx, "session:"+token).Result()
    if err == redis.Nil {
        // Cache miss - load from PostgreSQL
        return r.fallback.Get(ctx, token)
    }
    if err != nil {
        return nil, err
    }

    var session Session
    json.Unmarshal([]byte(val), &session)
    return &session, nil
}
```

3. **Add PgBouncer**
```yaml
# docker-compose.yml
pgbouncer:
  image: pgbouncer/pgbouncer:latest
  environment:
    POOL_MODE: transaction
    MAX_CLIENT_CONN: 1000
    DEFAULT_POOL_SIZE: 25
```

4. **Implement Graceful Shutdown**
```go
// Handle SIGTERM for rolling deployments
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

log.Println("Shutting down gracefully...")

// Stop accepting new WebSocket connections
wsHandler.SetAcceptingConnections(false)

// Wait for active sessions to close (max 30s)
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

wsHandler.WaitForSessions(ctx)
```

**Expected Results:**
- Zero-downtime deployments
- P95 latency <50ms (Redis sessions)
- 99.9% uptime (multi-server HA)

---

### Phase 3: Kubernetes Migration (2000-10000 users)
**Timeline:** 4-6 weeks
**Cost:** $300-600/month
**Goal:** Auto-scaling, container orchestration

#### Tasks

1. **Deploy K3s Cluster**
```bash
# Master node
curl -sfL https://get.k3s.io | sh -

# Get token
sudo cat /var/lib/rancher/k3s/server/node-token

# Worker nodes
curl -sfL https://get.k3s.io | K3S_URL=https://master:6443 \
  K3S_TOKEN=<token> sh -
```

2. **Deploy Backend with Helm**
```yaml
# values.yaml
replicaCount: 5

image:
  repository: businessos-backend
  tag: latest

resources:
  requests:
    memory: "1Gi"
    cpu: "500m"
  limits:
    memory: "2Gi"
    cpu: "1000m"

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 20
  targetCPUUtilizationPercentage: 70
```

3. **User Container Orchestration**
```go
// internal/container/k8s.go
import "k8s.io/client-go/kubernetes"

type K8sContainerManager struct {
    clientset *kubernetes.Clientset
}

func (k *K8sContainerManager) CreateWorkspace(userID string) error {
    pod := &corev1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name: fmt.Sprintf("workspace-%s", userID),
            Namespace: "user-workspaces",
            Labels: map[string]string{
                "app": "workspace",
                "user": userID,
            },
        },
        Spec: corev1.PodSpec{
            Containers: []corev1.Container{{
                Name:  "terminal",
                Image: "businessos-workspace:latest",
                Resources: corev1.ResourceRequirements{
                    Requests: corev1.ResourceList{
                        corev1.ResourceMemory: resource.MustParse("128Mi"),
                        corev1.ResourceCPU:    resource.MustParse("100m"),
                    },
                    Limits: corev1.ResourceList{
                        corev1.ResourceMemory: resource.MustParse("512Mi"),
                        corev1.ResourceCPU:    resource.MustParse("500m"),
                    },
                },
            }},
        },
    }

    _, err := k.clientset.CoreV1().Pods("user-workspaces").Create(
        context.Background(), pod, metav1.CreateOptions{})
    return err
}
```

4. **Observability Stack**
```yaml
# Install Prometheus + Grafana via Helm
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack

# Dashboards to create:
# - API request latency (P50, P95, P99)
# - Database connection pool usage
# - Active WebSocket sessions
# - Container resource usage (CPU, memory)
# - Error rate by endpoint
```

**Expected Results:**
- Auto-scale 3-20 backend pods based on CPU
- Support 10,000+ concurrent terminal sessions
- P95 latency <30ms
- 99.95% uptime

---

## Quick Decision Matrix

Use this table to pick the right stack based on current user count:

| Users | App Infra | Database | Session Store | Load Balancer | Containers | Monthly Cost |
|-------|-----------|----------|---------------|---------------|------------|--------------|
| **0-500** | Single VM | PostgreSQL | PostgreSQL | Cloudflare Tunnel | Docker | $10-20 |
| **500-2k** | 3 VMs | PostgreSQL + PgBouncer | Redis | Nginx | Docker Swarm | $60-100 |
| **2k-10k** | K3s (5 nodes) | PostgreSQL + Read Replica | Redis Cluster | K8s Ingress | Kubernetes | $300-600 |
| **10k+** | Managed K8s | Cloud SQL + Replicas | Managed Redis | Cloud LB | Kubernetes | $1500+ |

---

## Monitoring Triggers for Scaling

Set up alerts for these metrics:

| Metric | Threshold | Action |
|--------|-----------|--------|
| **P95 API Latency** | >100ms | Add Redis for sessions |
| **DB CPU** | >70% | Add read replica OR PgBouncer |
| **Active Connections** | >80% of pool | Increase pool size OR add PgBouncer |
| **Container Host CPU** | >80% | Add worker node (Swarm/K8s) |
| **WebSocket Connection Fails** | >1% | Add backend server |
| **Memory per Container** | >400MB avg | Optimize workspace image |

---

## Summary Recommendations

### Start Here (MVP - Next 2 Weeks)
1. Keep PostgreSQL for sessions (add indexes)
2. Optimize pgx pool settings (50 max conns, 3s acquire timeout)
3. Deploy on single VM with Docker ($10-20/month)
4. Use Cloudflare Tunnel for free SSL + WebSocket support
5. Add Prometheus metrics + Grafana dashboards

### Scale to 1000 Users (Month 2-3)
1. Add Redis for session caching
2. Deploy 3 app servers with Nginx load balancer
3. Add PgBouncer for connection pooling
4. Migrate to Docker Swarm for container orchestration
5. Set up read replica for database

### Scale to 10k Users (Month 6+)
1. Migrate to managed Kubernetes (GKE/EKS/AKS)
2. Implement horizontal pod autoscaling
3. Use managed PostgreSQL (Cloud SQL) with read replicas
4. Implement distributed tracing (Jaeger/Tempo)
5. Add CDN for static assets (Cloudflare/CloudFront)

**The beauty of this architecture:** Each phase is a non-breaking evolution. You can run the MVP for months and scale when metrics demand it, not prematurely.

---

## Next Steps

1. **Implement Phase 1 optimizations** (pgx pool, indexes, health checks)
2. **Set up monitoring** (Prometheus + Grafana Cloud free tier)
3. **Load test current setup** (k6 or Locust to find breaking points)
4. **Document infrastructure** (architecture diagrams, runbooks)
5. **Plan Phase 2 migration** (when monitoring shows need for scaling)

---

**Document Version:** 1.0
**Last Updated:** 2025-12-22
**Next Review:** When user count hits 500 or P95 latency >100ms
