## BusinessOS Scalability Architecture Diagrams

Visual representations of the system architecture at different scales.

---

## Phase 1: MVP (0-500 Users) - Single VM

```bash
┌─────────────────────────────────────────────────────────────────┐
│                     Cloudflare Tunnel (Free)                    │
│                    SSL/TLS + WebSocket Support                  │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ HTTPS/WSS
                             │
┌────────────────────────────▼────────────────────────────────────┐
│                   Single VM (4 vCPU, 8GB RAM)                   │
│                         $10-20/month                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌───────────────────────────────────────────────────────┐    │
│  │        Go Backend (Gin Framework)                     │    │
│  │                                                        │    │
│  │  ├─ HTTP Handlers                                     │    │
│  │  ├─ WebSocket Terminal Handler                        │    │
│  │  ├─ Auth Middleware (BetterAuth)                      │    │
│  │  ├─ Prometheus Metrics (/metrics)                     │    │
│  │  └─ Health Checks (/health)                           │    │
│  │                                                        │    │
│  │  Connection Pool:                                     │    │
│  │  ├─ MaxConns: 50                                      │    │
│  │  ├─ MinConns: 10                                      │    │
│  │  └─ Acquire Timeout: 3s                               │    │
│  └────────────────────┬──────────────────────────────────┘    │
│                       │                                        │
│                       │ pgx/v5                                 │
│                       │                                        │
│  ┌────────────────────▼──────────────────────────────────┐    │
│  │         PostgreSQL 14                                 │    │
│  │                                                        │    │
│  │  ├─ Sessions (with indexes)                           │    │
│  │  ├─ Users                                             │    │
│  │  ├─ Terminal Sessions                                 │    │
│  │  └─ Application Data                                  │    │
│  │                                                        │    │
│  │  Optimizations:                                       │    │
│  │  ├─ idx_sessions_token_active (WHERE expires_at)     │    │
│  │  ├─ idx_terminal_sessions_user_active                │    │
│  │  └─ Cleanup function (daily cron)                     │    │
│  └───────────────────────────────────────────────────────┘    │
│                                                                 │
│  ┌───────────────────────────────────────────────────────┐    │
│  │        Docker Engine (User Containers)                │    │
│  │                                                        │    │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐           │    │
│  │  │ User 1   │  │ User 2   │  │ User 3   │  ...      │    │
│  │  │ Terminal │  │ Terminal │  │ Terminal │           │    │
│  │  │ 512MB    │  │ 512MB    │  │ 512MB    │           │    │
│  │  └──────────┘  └──────────┘  └──────────┘           │    │
│  │                                                        │    │
│  │  Max: 50 containers (with resource limits)           │    │
│  └───────────────────────────────────────────────────────┘    │
│                                                                 │
│  ┌───────────────────────────────────────────────────────┐    │
│  │        Redis (Optional - Future)                      │    │
│  │        Currently using PostgreSQL for sessions        │    │
│  └───────────────────────────────────────────────────────┘    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
                             │
                             │ Remote Write
                             │
                    ┌────────▼────────┐
                    │  Grafana Cloud  │
                    │   (Free Tier)   │
                    │                 │
                    │ ├─ Prometheus   │
                    │ ├─ Dashboards   │
                    │ └─ Alerts       │
                    └─────────────────┘

Key Metrics:
├─ Concurrent Users: 500
├─ WebSocket Sessions: 200
├─ Database Connections: 50 max
├─ Monthly Cost: $10-20
└─ P95 Latency: <100ms
```
---

## Phase 2: Multi-Server HA (500-2000 Users)

```text
┌─────────────────────────────────────────────────────────────────┐
│                    Cloudflare Pro ($20/month)                   │
│              SSL/TLS + DDoS Protection + CDN                    │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ HTTPS/WSS
                             │
┌────────────────────────────▼────────────────────────────────────┐
│                  Nginx Load Balancer                            │
│                  (4 vCPU, 8GB RAM) - $12/month                  │
│                                                                  │
│  ├─ IP Hash (Sticky Sessions)                                   │
│  ├─ WebSocket Upgrade Headers                                   │
│  ├─ Health Checks (30s interval)                                │
│  └─ SSL Termination                                             │
└──────┬───────────────────┬────────────────────┬─────────────────┘
       │                   │                    │
       │                   │                    │
   ┌───▼────┐         ┌────▼────┐         ┌────▼────┐
   │ App 1  │         │ App 2   │         │ App 3   │
   │ 4 vCPU │         │ 4 vCPU  │         │ 4 vCPU  │
   │ 8GB    │         │ 8GB     │         │ 8GB     │
   │$12/mo  │         │$12/mo   │         │$12/mo   │
   └────┬───┘         └────┬────┘         └────┬────┘
        │                  │                   │
        │                  │                   │
        │                  │                   │
        └──────────┬───────┴──────┬────────────┘
                   │              │
                   │              │
        ┌──────────▼───┐   ┌──────▼───────────┐
        │  PgBouncer   │   │  Redis Cluster   │
        │              │   │  (Session Store) │
        │ Transaction  │   │                  │
        │ Mode         │   │  ├─ Master       │
        │              │   │  └─ Replica      │
        │ 25 DB conns  │   │                  │
        │ 1000 clients │   │  5GB data        │
        └──────┬───────┘   │  $10/month       │
               │           └──────────────────┘
               │
        ┌──────▼────────────────────────┐
        │   PostgreSQL 14               │
        │   (4 vCPU, 16GB RAM)          │
        │   $20/month                   │
        │                               │
        │   ├─ max_connections: 100     │
        │   ├─ shared_buffers: 4GB      │
        │   └─ effective_cache_size: 12GB│
        └───────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│              Docker Swarm (Container Orchestration)             │
│                                                                  │
│   Manager Node (App 1)         Worker Nodes (App 2, App 3)      │
│                                                                  │
│   ┌──────────────────────────────────────────────────────────┐ │
│   │  User Container Pool (500 containers)                    │ │
│   │                                                           │ │
│   │  ┌────────┐  ┌────────┐  ┌────────┐        ┌────────┐  │ │
│   │  │ User 1 │  │ User 2 │  │ User 3 │  ...   │User 500│  │ │
│   │  │ 512MB  │  │ 512MB  │  │ 512MB  │        │ 512MB  │  │ │
│   │  └────────┘  └────────┘  └────────┘        └────────┘  │ │
│   │                                                           │ │
│   │  Resource Limits per Node:                               │ │
│   │  ├─ CPU: 0.5 core/container                              │ │
│   │  ├─ Memory: 512MB/container                              │ │
│   │  └─ Max: ~150 containers/node                            │ │
│   └──────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                  Monitoring Stack                               │
│                                                                  │
│   Grafana Cloud Pro ($29/month)                                 │
│   ├─ 50k metrics/month                                          │
│   ├─ Custom Dashboards                                          │
│   ├─ Slack Alerts                                               │
│   └─ Uptime Monitoring                                          │
└─────────────────────────────────────────────────────────────────┘

Total Cost: $110/month
Concurrent Users: 2,000
P95 Latency: <50ms
Uptime: 99.9%
```
---

## Phase 3: Kubernetes (2000-10000 Users)

```text
┌─────────────────────────────────────────────────────────────────────┐
│                     Cloudflare Enterprise                           │
│                  ($200/month - 10TB bandwidth)                      │
│                                                                      │
│   ├─ Global CDN (250+ locations)                                    │
│   ├─ DDoS Protection (L3/L4/L7)                                     │
│   ├─ Rate Limiting                                                  │
│   └─ Web Application Firewall                                       │
└───────────────────────────┬─────────────────────────────────────────┘
                            │
                            │ HTTPS/WSS
                            │
┌───────────────────────────▼──────────────────────────────────────────┐
│                  GCP Cloud Load Balancer                             │
│                      ($40/month - 1TB data)                          │
│                                                                       │
│   ├─ Session Affinity (Client IP + Cookie)                          │
│   ├─ Health Checks                                                   │
│   ├─ SSL Offloading                                                  │
│   └─ WebSocket Support                                               │
└────────────────────────┬─────────────────────────────────────────────┘
                         │
                         │
┌────────────────────────▼─────────────────────────────────────────────┐
│              Google Kubernetes Engine (GKE Autopilot)                │
│                      Control Plane: Managed (Free)                   │
└──────────────────────────────────────────────────────────────────────┘
│                                                                       │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │           Backend Service Node Pool                            │ │
│  │           (10 pods across 3 nodes)                             │ │
│  │                                                                 │ │
│  │   Node Type: n2-standard-4 (4 vCPU, 16GB RAM)                  │ │
│  │   Autoscaling: 3-10 nodes                                      │ │
│  │                                                                 │ │
│  │   ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐         │ │
│  │   │Backend 1│  │Backend 2│  │Backend 3│  │Backend..│         │ │
│  │   │ 2 vCPU  │  │ 2 vCPU  │  │ 2 vCPU  │  │         │  ...    │ │
│  │   │ 4GB RAM │  │ 4GB RAM │  │ 4GB RAM │  │         │         │ │
│  │   └─────────┘  └─────────┘  └─────────┘  └─────────┘         │ │
│  │                                                                 │ │
│  │   HorizontalPodAutoscaler:                                     │ │
│  │   ├─ Min Replicas: 3                                           │ │
│  │   ├─ Max Replicas: 20                                          │ │
│  │   └─ Target CPU: 70%                                           │ │
│  └────────────────────────────────────────────────────────────────┘ │
│                                                                       │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │           User Workspace Node Pool                             │ │
│  │           (5000 pods across 20 spot nodes)                     │ │
│  │                                                                 │ │
│  │   Node Type: n2-standard-8 (8 vCPU, 32GB RAM) - Spot instances│ │
│  │   Autoscaling: 5-50 nodes                                      │ │
│  │   70% cost savings with spot pricing                           │ │
│  │                                                                 │ │
│  │   ┌────────┐  ┌────────┐  ┌────────┐                          │ │
│  │   │ User 1 │  │ User 2 │  │ User 3 │  ...  (5000 pods)        │ │
│  │   │ 0.1 CPU│  │ 0.1 CPU│  │ 0.1 CPU│                          │ │
│  │   │ 128MB  │  │ 128MB  │  │ 128MB  │                          │ │
│  │   └────────┘  └────────┘  └────────┘                          │ │
│  │                                                                 │ │
│  │   Resource Quota (namespace: user-workspaces):                 │ │
│  │   ├─ requests.cpu: 200 cores                                   │ │
│  │   ├─ requests.memory: 200Gi                                    │ │
│  │   ├─ limits.cpu: 400 cores                                     │ │
│  │   └─ limits.memory: 400Gi                                      │ │
│  └────────────────────────────────────────────────────────────────┘ │
│                                                                       │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │              Supporting Services                               │ │
│  │                                                                 │ │
│  │   ┌─────────────┐   ┌──────────────┐   ┌─────────────┐       │ │
│  │   │  PgBouncer  │   │    Redis     │   │  Prometheus │       │ │
│  │   │  (3 pods)   │   │   (cluster)  │   │  Operator   │       │ │
│  │   └─────────────┘   └──────────────┘   └─────────────┘       │ │
│  └────────────────────────────────────────────────────────────────┘ │
└───────────────────────────────────────────────────────────────────────┘
           │                          │                        │
           │                          │                        │
   ┌───────▼────────┐      ┌──────────▼────────┐   ┌─────────▼─────────┐
   │  Cloud SQL     │      │  Memorystore      │   │  Cloud Monitoring │
   │  (PostgreSQL)  │      │  (Redis)          │   │                   │
   │                │      │                   │   │  ├─ Metrics       │
   │  ├─ Master     │      │  ├─ 5GB Standard  │   │  ├─ Logs          │
   │  │  4vCPU,16GB │      │  └─ Auto-failover │   │  ├─ Traces        │
   │  │  $250/mo    │      │                   │   │  └─ Dashboards    │
   │  │             │      │  $50/month        │   │                   │
   │  ├─ Read Rep 1 │      └───────────────────┘   │  $50/month        │
   │  │  4vCPU,16GB │                               └───────────────────┘
   │  │  $200/mo    │
   │  │             │
   │  └─ Backups    │
   │     Daily full │
   │     $30/mo     │
   └────────────────┘

┌───────────────────────────────────────────────────────────────────────┐
│                         Network Architecture                          │
│                                                                        │
│   ├─ VPC (10.0.0.0/16)                                                │
│   │                                                                    │
│   │  ├─ Backend Subnet (10.0.1.0/24)                                  │
│   │  │  └─ Backend pods, PgBouncer, Redis                             │
│   │  │                                                                 │
│   │  ├─ Workspace Subnet (10.0.2.0/23)                                │
│   │  │  └─ User container pods (isolated)                             │
│   │  │                                                                 │
│   │  └─ Database Subnet (10.0.10.0/24)                                │
│   │     └─ Cloud SQL (private IP only)                                │
│   │                                                                    │
│   └─ Network Policies:                                                │
│      ├─ User pods CAN'T communicate with each other                   │
│      ├─ User pods CAN'T access database directly                      │
│      ├─ Backend pods CAN access database via PgBouncer                │
│      └─ All egress traffic through Cloud NAT                          │
└───────────────────────────────────────────────────────────────────────┘

Total Cost: $1,620/month
Concurrent Users: 10,000
Active Containers: 5,000
P95 Latency: <30ms
Uptime: 99.95%
Monthly Revenue: $50,000 (@ $5/user)
Profit Margin: 96.8%
```
---

## Data Flow Diagrams

### Terminal WebSocket Connection Flow

```bash
┌─────────┐                                              ┌──────────────┐
│ Client  │                                              │   Backend    │
│ Browser │                                              │   Server     │
└────┬────┘                                              └──────┬───────┘
     │                                                          │
     │  1. GET /api/terminal/ws?cols=80&rows=24               │
     │     Authorization: Bearer <token>                       │
     ├────────────────────────────────────────────────────────▶│
     │                                                          │
     │                                             2. Validate token
     │                                                (Query PostgreSQL)
     │                                                          │
     │  3. 101 Switching Protocols                             │
     │     Upgrade: websocket                                  │
     │◀────────────────────────────────────────────────────────┤
     │                                                          │
     │           WebSocket Connection Established              │
     │═════════════════════════════════════════════════════════│
     │                                                          │
     │                                             4. Create Docker container
     │                                                OR Kubernetes pod
     │                                                          │
     │  5. {"type":"ready","container_id":"abc123"}            │
     │◀────────────────────────────────────────────────────────┤
     │                                                          │
     │  6. {"type":"input","data":"ls -la\n"}                  │
     ├────────────────────────────────────────────────────────▶│
     │                                                          │
     │                                             7. Forward to PTY
     │                                                in container
     │                                                          │
     │  8. {"type":"output","data":"total 24\ndrwx..."}        │
     │◀────────────────────────────────────────────────────────┤
     │                                                          │
     │  9. {"type":"resize","cols":120,"rows":30}              │
     ├────────────────────────────────────────────────────────▶│
     │                                                          │
     │                                             10. Resize PTY
     │                                                          │
     │                                                          │
     │  11. Close connection (idle timeout or user logout)     │
     ├────────────────────────────────────────────────────────▶│
     │                                                          │
     │                                             12. Cleanup container
     │                                                 Update DB
     │                                                          │
     │  13. Connection closed                                  │
     │◀────────────────────────────────────────────────────────┤
     │                                                          │
```
### Database Query Flow (with PgBouncer)

```text
┌──────────┐       ┌───────────┐       ┌────────────┐
│ Backend  │       │ PgBouncer │       │ PostgreSQL │
│   Pod    │       │           │       │            │
└────┬─────┘       └─────┬─────┘       └──────┬─────┘
     │                   │                     │
     │ 1. Query Session  │                     │
     │  (user token)     │                     │
     ├──────────────────▶│                     │
     │                   │                     │
     │                   │ 2. Acquire DB conn  │
     │                   │   from pool         │
     │                   ├────────────────────▶│
     │                   │                     │
     │                   │ 3. Execute query    │
     │                   │                     │
     │                   │◀────────────────────┤
     │                   │                     │
     │ 4. Return result  │                     │
     │   (50ms latency)  │                     │
     │◀──────────────────┤                     │
     │                   │                     │
     │                   │ 5. Return conn      │
     │                   │   to pool           │
     │                   │   (transaction mode)│
     │                   │                     │
     │                   │                     │
     │ 6. Another query  │                     │
     │   (same request)  │                     │
     ├──────────────────▶│                     │
     │                   │                     │
     │                   │ 7. Reuse conn       │
     │                   │   (same transaction)│
     │                   ├────────────────────▶│
     │                   │◀────────────────────┤
     │◀──────────────────┤                     │
     │                   │                     │

PgBouncer Stats:
├─ 1000 client connections
├─ 25 PostgreSQL connections
├─ 40:1 multiplexing ratio
└─ <1ms proxy overhead
```
### Session Storage: PostgreSQL vs Redis

```bash
┌──────────────────────────────────────────────────────────────┐
│                    Session Lookup Flow                       │
└──────────────────────────────────────────────────────────────┘

CURRENT (PostgreSQL Only):
┌────────┐          ┌──────────────┐          ┌─────────────┐
│Request │─────────▶│ Auth Check   │─────────▶│ PostgreSQL  │
│        │          │ Middleware   │          │             │
│        │          │              │◀─────────│ 10-20ms     │
│        │◀─────────│              │          │ latency     │
└────────┘          └──────────────┘          └─────────────┘

FUTURE (Redis Cache):
┌────────┐          ┌──────────────┐          ┌─────────────┐
│Request │─────────▶│ Auth Check   │─────────▶│   Redis     │
│        │          │ Middleware   │          │             │
│        │          │              │◀─────────│ 1-2ms       │
│        │          │              │          │ latency     │
│        │          │      │       │          └─────────────┘
│        │          │      │       │                  │
│        │          │      │       │                  │ Cache miss
│        │          │      │       │                  │ (rare)
│        │          │      │       │                  │
│        │          │      ▼       │          ┌───────▼─────┐
│        │          │  Fallback    │─────────▶│ PostgreSQL  │
│        │          │  to DB       │◀─────────│ 10-20ms     │
│        │◀─────────│              │          │             │
└────────┘          └──────────────┘          └─────────────┘

Performance Comparison:
├─ PostgreSQL: 10-20ms (disk I/O)
├─ Redis: 1-2ms (in-memory)
└─ Improvement: 10x faster session lookups
```
---

## Scaling Triggers

Visual representation of when to scale:

```bash
User Count                 Architecture                      Monthly Cost
────────────────────────────────────────────────────────────────────────

0-100                      Single VM (MVP)                   $5-10
│                          ├─ Go Backend
│                          ├─ PostgreSQL (same VM)
│                          └─ Docker containers
│
├─ Trigger: P95 latency >100ms
│
500-1000                   Single VM (Optimized)             $20-40
│                          ├─ Tuned connection pool
│                          ├─ Database indexes
│                          ├─ Prometheus monitoring
│                          └─ More containers
│
├─ Trigger: CPU >70% OR Need HA
│
1000-2000                  Multi-Server HA                   $60-110
│                          ├─ 3× Backend servers
│                          ├─ Nginx load balancer
│                          ├─ Redis session cache
│                          ├─ PgBouncer
│                          └─ Docker Swarm
│
├─ Trigger: Need autoscaling OR >2000 users
│
2000-5000                  K3s (Self-managed K8s)            $200-400
│                          ├─ 5-10 nodes
│                          ├─ HPA (autoscaling)
│                          ├─ PostgreSQL read replica
│                          └─ Container orchestration
│
├─ Trigger: Operational overhead OR >5000 users
│
5000-10000                 Managed Kubernetes                $600-1200
│                          ├─ GKE/EKS/AKS
│                          ├─ Cluster autoscaler
│                          ├─ Managed database
│                          └─ Managed Redis
│
├─ Trigger: Global expansion OR >10000 users
│
10000+                     Multi-Region K8s                  $2000+
                           ├─ Regional clusters
                           ├─ Global load balancing
                           ├─ Database replication
                           └─ CDN for static assets

Key Metrics to Monitor:
├─ P95 API latency
├─ Database connection pool usage
├─ CPU/Memory utilization
├─ Error rate
├─ WebSocket connection failures
└─ Container resource usage
```
---

## Security Layers

```bash
┌─────────────────────────────────────────────────────────────────┐
│                  Layer 7: Application Security                  │
│                                                                  │
│  ├─ BetterAuth (session-based authentication)                   │
│  ├─ JWT validation middleware                                   │
│  ├─ Rate limiting (per user, per endpoint)                      │
│  ├─ Input sanitization (terminal commands)                      │
│  └─ SQL injection protection (prepared statements)              │
└──────────────────────────┬──────────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────────┐
│               Layer 6: Container Isolation                       │
│                                                                  │
│  User Containers (Namespaces):                                  │
│  ├─ PID namespace (isolated process tree)                       │
│  ├─ Network namespace (private network stack)                   │
│  ├─ Mount namespace (isolated filesystem)                       │
│  ├─ User namespace (root inside ≠ root outside)                 │
│  │                                                               │
│  Security Policies:                                             │
│  ├─ AppArmor profile (restrict system calls)                    │
│  ├─ Seccomp filter (block dangerous syscalls)                   │
│  ├─ Read-only root filesystem                                   │
│  ├─ No privileged containers                                    │
│  └─ Resource limits (CPU, memory, disk I/O)                     │
└──────────────────────────┬──────────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────────┐
│                Layer 5: Network Security                         │
│                                                                  │
│  Kubernetes Network Policies:                                   │
│  ├─ Deny all ingress by default                                 │
│  ├─ Allow backend → user pods (WebSocket only)                  │
│  ├─ Deny user pod → user pod communication                      │
│  ├─ Allow user pods → internet (egress via NAT)                 │
│  └─ Private subnet for database (no public IP)                  │
│                                                                  │
│  Firewall Rules:                                                │
│  ├─ Allow 443 (HTTPS/WSS) from internet                         │
│  ├─ Allow 22 (SSH) from bastion host only                       │
│  ├─ Deny all other inbound traffic                              │
│  └─ Egress: Allow to specific endpoints (OAuth, APIs)           │
└──────────────────────────┬──────────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────────┐
│               Layer 4: Data Encryption                           │
│                                                                  │
│  In Transit:                                                     │
│  ├─ TLS 1.3 (HTTPS, WebSocket Secure)                           │
│  ├─ Certificate auto-renewal (Let's Encrypt)                    │
│  └─ Strong cipher suites only (no SSLv3, TLS 1.0)               │
│                                                                  │
│  At Rest:                                                        │
│  ├─ Database encryption (transparent data encryption)           │
│  ├─ Disk encryption (LUKS/dm-crypt)                             │
│  └─ Secrets encrypted (Kubernetes Secrets, sealed)              │
└──────────────────────────┬──────────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────────┐
│               Layer 3: Access Control                            │
│                                                                  │
│  IAM Policies (Cloud Provider):                                 │
│  ├─ Least privilege principle                                   │
│  ├─ Service accounts (no user credentials)                      │
│  ├─ MFA for admin access                                        │
│  └─ Audit logging (all API calls)                               │
│                                                                  │
│  Kubernetes RBAC:                                               │
│  ├─ Namespace-level isolation                                   │
│  ├─ Pod security policies                                       │
│  └─ Service account restrictions                                │
└──────────────────────────┬──────────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────────┐
│          Layer 2: Monitoring & Incident Response                 │
│                                                                  │
│  ├─ Intrusion detection (anomaly detection)                     │
│  ├─ Audit logs (all database queries, API calls)                │
│  ├─ Real-time alerts (suspicious activity)                      │
│  ├─ Automated response (block IPs, kill sessions)               │
│  └─ Compliance reporting (SOC2, GDPR)                           │
└──────────────────────────┬──────────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────────┐
│               Layer 1: DDoS Protection                           │
│                                                                  │
│  Cloudflare:                                                     │
│  ├─ Rate limiting (per IP, per endpoint)                        │
│  ├─ Bot detection (challenge pages)                             │
│  ├─ L3/L4 DDoS mitigation (SYN flood, UDP flood)                │
│  ├─ L7 DDoS mitigation (HTTP flood)                             │
│  └─ Geo-blocking (if needed)                                    │
└──────────────────────────────────────────────────────────────────┘
```
---

## Summary

These diagrams show:

1. **Phase 1 (MVP):** Simple single-VM setup for $10-20/month
2. **Phase 2 (HA):** Multi-server architecture for high availability at $60-110/month
3. **Phase 3 (Scale):** Kubernetes orchestration for 10k+ users at $1,620/month
4. **Data Flows:** WebSocket connections, database queries, session lookups
5. **Scaling Triggers:** Clear metrics to know when to upgrade
6. **Security:** Defense-in-depth across all layers

The architecture scales smoothly from startup to enterprise without breaking changes.
