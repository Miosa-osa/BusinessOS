#!/bin/bash
# BusinessOS Development Environment
# Usage: ./dev.sh [start|stop|status|logs]
set -Eeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
BACKEND_DIR="$SCRIPT_DIR/desktop/backend-go"
FRONTEND_DIR="$SCRIPT_DIR/frontend"
LOG_DIR="$SCRIPT_DIR/.startup-logs"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() { echo -e "${BLUE}[dev]${NC} $*"; }
success() { echo -e "${GREEN}[ok]${NC} $*"; }
warn() { echo -e "${YELLOW}[warn]${NC} $*"; }
error() { echo -e "${RED}[error]${NC} $*"; }

mkdir -p "$LOG_DIR"

check_deps() {
    local missing=()
    command -v go >/dev/null 2>&1 || missing+=("go")
    command -v node >/dev/null 2>&1 || missing+=("node")
    command -v docker >/dev/null 2>&1 || missing+=("docker")

    if [ ${#missing[@]} -gt 0 ]; then
        error "Missing dependencies: ${missing[*]}"
        exit 1
    fi
    success "Dependencies: go, node, docker"
}

check_ports() {
    local ports_in_use=()
    for port in 8001 5173 6379; do
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            ports_in_use+=("$port")
        fi
    done

    if [ ${#ports_in_use[@]} -gt 0 ]; then
        warn "Ports in use: ${ports_in_use[*]}"
        return 1
    fi
    return 0
}

start_postgres() {
    log "Starting PostgreSQL..."
    if command -v brew >/dev/null 2>&1; then
        brew services start postgresql@14 2>/dev/null || brew services start postgresql 2>/dev/null || true
    fi

    # Wait for postgres
    for i in {1..10}; do
        if pg_isready -q 2>/dev/null; then
            success "PostgreSQL ready"
            return 0
        fi
        sleep 1
    done
    warn "PostgreSQL may not be running - check manually"
}

start_redis() {
    log "Starting Redis via Docker..."
    cd "$BACKEND_DIR"

    if docker ps --format '{{.Names}}' | grep -q 'businessos-redis'; then
        success "Redis already running"
        return 0
    fi

    docker-compose up -d redis 2>/dev/null || {
        # Fallback: start Redis without compose
        docker run -d --name businessos-redis \
            -p 6379:6379 \
            redis:7-alpine redis-server --requirepass "dev-password" 2>/dev/null || true
    }

    # Wait for Redis
    for i in {1..10}; do
        if docker exec businessos-redis redis-cli -a "dev-password" ping 2>/dev/null | grep -q PONG; then
            success "Redis ready"
            return 0
        fi
        sleep 1
    done
    warn "Redis may not be healthy"
}

start_backend() {
    log "Starting Go backend on :8001..."
    cd "$BACKEND_DIR"

    # Ensure .env exists
    if [ ! -f .env ]; then
        if [ -f .env.example ]; then
            cp .env.example .env
            log "Created .env from .env.example"
        fi
    fi

    # Build and run
    go build -o server cmd/server/main.go 2>"$LOG_DIR/backend-build.log" || {
        error "Backend build failed - check $LOG_DIR/backend-build.log"
        return 1
    }

    DATABASE_URL="${DATABASE_URL:-postgres://rhl:password@localhost:5432/business_os?sslmode=disable}" \
    SERVER_PORT=8001 \
    ENVIRONMENT=development \
    ./server >"$LOG_DIR/backend.log" 2>&1 &
    echo $! > "$LOG_DIR/backend.pid"

    # Wait for health
    for i in {1..15}; do
        if curl -s http://localhost:8001/health >/dev/null 2>&1; then
            success "Backend ready at http://localhost:8001"
            return 0
        fi
        sleep 1
    done
    error "Backend failed to start - check $LOG_DIR/backend.log"
}

start_frontend() {
    log "Starting frontend on :5173..."
    cd "$FRONTEND_DIR"

    # Install deps if needed
    if [ ! -d node_modules ]; then
        log "Installing frontend dependencies..."
        npm install >"$LOG_DIR/npm-install.log" 2>&1
    fi

    npm run dev >"$LOG_DIR/frontend.log" 2>&1 &
    echo $! > "$LOG_DIR/frontend.pid"

    # Wait for frontend
    for i in {1..20}; do
        if curl -s http://localhost:5173 >/dev/null 2>&1; then
            success "Frontend ready at http://localhost:5173"
            return 0
        fi
        sleep 1
    done
    warn "Frontend may still be starting - check $LOG_DIR/frontend.log"
}

stop_all() {
    log "Stopping all services..."

    # Stop backend
    if [ -f "$LOG_DIR/backend.pid" ]; then
        kill "$(cat "$LOG_DIR/backend.pid")" 2>/dev/null || true
        rm -f "$LOG_DIR/backend.pid"
    fi
    pkill -f "desktop/backend-go/server" 2>/dev/null || true

    # Stop frontend
    if [ -f "$LOG_DIR/frontend.pid" ]; then
        kill "$(cat "$LOG_DIR/frontend.pid")" 2>/dev/null || true
        rm -f "$LOG_DIR/frontend.pid"
    fi
    pkill -f "vite.*frontend" 2>/dev/null || true

    # Stop Redis
    docker stop businessos-redis 2>/dev/null || true
    docker rm businessos-redis 2>/dev/null || true

    success "All services stopped"
}

show_status() {
    echo ""
    echo "Service Status:"
    echo "==============="

    # Backend
    if curl -s http://localhost:8001/health >/dev/null 2>&1; then
        echo -e "Backend:  ${GREEN}running${NC} at http://localhost:8001"
    else
        echo -e "Backend:  ${RED}stopped${NC}"
    fi

    # Frontend
    if curl -s http://localhost:5173 >/dev/null 2>&1; then
        echo -e "Frontend: ${GREEN}running${NC} at http://localhost:5173"
    else
        echo -e "Frontend: ${RED}stopped${NC}"
    fi

    # Redis
    if docker ps --format '{{.Names}}' | grep -q 'businessos-redis'; then
        echo -e "Redis:    ${GREEN}running${NC} at localhost:6379"
    else
        echo -e "Redis:    ${RED}stopped${NC}"
    fi

    # PostgreSQL
    if pg_isready -q 2>/dev/null; then
        echo -e "Postgres: ${GREEN}running${NC}"
    else
        echo -e "Postgres: ${YELLOW}unknown${NC}"
    fi
    echo ""
}

show_logs() {
    local service="${1:-all}"
    case "$service" in
        backend)
            tail -f "$LOG_DIR/backend.log"
            ;;
        frontend)
            tail -f "$LOG_DIR/frontend.log"
            ;;
        all)
            tail -f "$LOG_DIR/backend.log" "$LOG_DIR/frontend.log"
            ;;
    esac
}

start_all() {
    echo ""
    echo "BusinessOS Development Environment"
    echo "==================================="
    echo ""

    check_deps

    if ! check_ports; then
        warn "Some ports in use - services may conflict"
    fi

    start_postgres
    start_redis
    start_backend
    start_frontend

    echo ""
    echo "All services started!"
    echo "====================="
    echo "Backend:  http://localhost:8001"
    echo "Frontend: http://localhost:5173"
    echo "Redis:    localhost:6379"
    echo ""
    echo "Commands:"
    echo "  ./dev.sh status  - Check service status"
    echo "  ./dev.sh logs    - Tail all logs"
    echo "  ./dev.sh stop    - Stop all services"
    echo ""
}

case "${1:-start}" in
    start)
        start_all
        ;;
    stop)
        stop_all
        ;;
    status)
        show_status
        ;;
    logs)
        show_logs "${2:-all}"
        ;;
    restart)
        stop_all
        sleep 2
        start_all
        ;;
    *)
        echo "Usage: $0 [start|stop|status|logs|restart]"
        exit 1
        ;;
esac
