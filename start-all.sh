#!/bin/bash
# ==============================================================================
# BusinessOS + OSA-5 Complete Startup Script
# ==============================================================================
# One command to start everything: BusinessOS + OSA-5
#
# Usage:
#   ./start-all.sh            # Start everything
#   ./start-all.sh docker     # Use Docker Compose
#   ./start-all.sh stop       # Stop all services
#   ./start-all.sh status     # Check status

set -e

# ==============================================================================
# Configuration
# ==============================================================================
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$SCRIPT_DIR/desktop/backend-go"
FRONTEND_DIR="$SCRIPT_DIR/frontend"
OSA_DIR="${OSA_PATH:-$HOME/Desktop/OSA-5}"
LOG_DIR="$SCRIPT_DIR/.startup-logs"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Create log directory
mkdir -p "$LOG_DIR"

# ==============================================================================
# Logging Functions
# ==============================================================================
log_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
}

log_step() {
    echo -e "\n${CYAN}═══ $1 ═══${NC}\n"
}

# ==============================================================================
# Check Prerequisites
# ==============================================================================
check_prerequisites() {
    log_step "Checking Prerequisites"

    local missing=0

    # Check PostgreSQL
    if ! command -v psql &> /dev/null; then
        log_error "PostgreSQL not found. Install: brew install postgresql@16"
        missing=1
    else
        log_success "PostgreSQL: $(psql --version | head -1)"
    fi

    # Check Go
    if ! command -v go &> /dev/null; then
        log_error "Go not found. Install: brew install go"
        missing=1
    else
        log_success "Go: $(go version)"
    fi

    # Check Node.js
    if ! command -v node &> /dev/null; then
        log_error "Node.js not found. Install: brew install node"
        missing=1
    else
        log_success "Node.js: $(node --version)"
    fi

    # Check Redis
    if ! command -v redis-server &> /dev/null; then
        log_warn "Redis not found. Install: brew install redis (optional)"
    else
        log_success "Redis: $(redis-server --version | head -1)"
    fi

    # Check jq
    if ! command -v jq &> /dev/null; then
        log_warn "jq not found. Install: brew install jq (optional, for pretty JSON)"
    else
        log_success "jq: $(jq --version)"
    fi

    # Check OSA-5 directory
    if [ ! -d "$OSA_DIR" ]; then
        log_error "OSA-5 not found at: $OSA_DIR"
        log_info "Set OSA_PATH environment variable or clone to ~/Desktop/OSA-5"
        missing=1
    else
        log_success "OSA-5: $OSA_DIR"
    fi

    if [ $missing -eq 1 ]; then
        log_error "Missing prerequisites. Install and try again."
        exit 1
    fi
}

# ==============================================================================
# Start Services (Local Mode)
# ==============================================================================
start_local() {
    log_step "Starting Services (Local Mode)"

    # 1. Start PostgreSQL
    log_info "Starting PostgreSQL..."
    if brew services list | grep -q "postgresql.*started"; then
        log_warn "PostgreSQL already running"
    else
        brew services start postgresql@16
        sleep 2
    fi

    # Wait for PostgreSQL
    local retry=0
    while ! psql -U postgres -h localhost -c "SELECT 1" &>/dev/null; do
        if [ $retry -ge 30 ]; then
            log_error "PostgreSQL failed to start"
            exit 1
        fi
        sleep 1
        ((retry++))
    done
    log_success "PostgreSQL ready"

    # 2. Start Redis
    log_info "Starting Redis..."
    if pgrep redis-server &>/dev/null; then
        log_warn "Redis already running"
    else
        redis-server --daemonize yes --port 6379
        sleep 1
    fi
    log_success "Redis ready"

    # 3. Setup Database
    log_info "Setting up database..."
    if ! psql -U postgres -h localhost -lqt | cut -d \| -f 1 | grep -qw businessos; then
        psql -U postgres -h localhost <<EOF
CREATE DATABASE businessos;
CREATE USER businessos_user WITH PASSWORD 'password';
GRANT ALL PRIVILEGES ON DATABASE businessos TO businessos_user;
EOF
        log_success "Database created"
    else
        log_warn "Database already exists"
    fi

    # Run migrations
    cd "$BACKEND_DIR"
    if [ -f "cmd/migrate/main.go" ]; then
        log_info "Running database migrations..."
        go run ./cmd/migrate
        log_success "Migrations complete"
    fi

    # 4. Start Backend
    log_info "Starting BusinessOS backend..."
    cd "$BACKEND_DIR"

    # Check for .env
    if [ ! -f ".env" ]; then
        log_warn "No .env file found. Creating from .env.example..."
        if [ -f "../../.env.example" ]; then
            cp ../../.env.example .env
            log_info "Edit .env and set OSA_ENABLED=true, OSA_WEBHOOK_SECRET, etc."
        fi
    fi

    # Start backend
    go run cmd/server/main.go > "$LOG_DIR/backend.log" 2>&1 &
    local backend_pid=$!
    echo $backend_pid > "$LOG_DIR/backend.pid"

    sleep 3
    if kill -0 $backend_pid 2>/dev/null; then
        log_success "Backend started (PID: $backend_pid, Port: 8001)"
    else
        log_error "Backend failed to start. Check: $LOG_DIR/backend.log"
        exit 1
    fi

    # 5. Start Frontend
    log_info "Starting BusinessOS frontend..."
    cd "$FRONTEND_DIR"

    if [ ! -d "node_modules" ]; then
        log_info "Installing frontend dependencies..."
        npm install
    fi

    npm run dev > "$LOG_DIR/frontend.log" 2>&1 &
    local frontend_pid=$!
    echo $frontend_pid > "$LOG_DIR/frontend.pid"

    sleep 3
    if kill -0 $frontend_pid 2>/dev/null; then
        log_success "Frontend started (PID: $frontend_pid, Port: 5173)"
    else
        log_error "Frontend failed to start. Check: $LOG_DIR/frontend.log"
        exit 1
    fi

    # 6. Start OSA-5
    log_info "Starting OSA-5 orchestrator..."
    cd "$OSA_DIR"

    # Check for .env
    if [ ! -f ".env" ]; then
        log_warn "No .env file in OSA-5. Create one with:"
        log_info "  PORT=3003"
        log_info "  BUSINESSOS_URL=http://localhost:8001"
        log_info "  WEBHOOK_SECRET=<same as BusinessOS OSA_WEBHOOK_SECRET>"
    fi

    if [ ! -d "node_modules" ]; then
        log_info "Installing OSA-5 dependencies..."
        npm install
    fi

    npm start > "$LOG_DIR/osa.log" 2>&1 &
    local osa_pid=$!
    echo $osa_pid > "$LOG_DIR/osa.pid"

    sleep 3
    if kill -0 $osa_pid 2>/dev/null; then
        log_success "OSA-5 started (PID: $osa_pid, Port: 3003)"
    else
        log_error "OSA-5 failed to start. Check: $LOG_DIR/osa.log"
        exit 1
    fi
}

# ==============================================================================
# Start Services (Docker Mode)
# ==============================================================================
start_docker() {
    log_step "Starting Services (Docker Mode)"

    if ! command -v docker &> /dev/null; then
        log_error "Docker not found. Install Docker Desktop"
        exit 1
    fi

    cd "$SCRIPT_DIR"

    log_info "Starting Docker containers..."
    docker-compose -f docker-compose.complete.yml up -d

    log_info "Waiting for services to be healthy..."
    local retry=0
    while [ $retry -lt 60 ]; do
        if docker-compose -f docker-compose.complete.yml ps | grep -q "healthy"; then
            log_success "All services healthy"
            break
        fi
        sleep 1
        ((retry++))
    done

    if [ $retry -ge 60 ]; then
        log_error "Services failed to become healthy"
        docker-compose -f docker-compose.complete.yml logs
        exit 1
    fi
}

# ==============================================================================
# Stop All Services
# ==============================================================================
stop_services() {
    log_step "Stopping All Services"

    # Stop backend
    if [ -f "$LOG_DIR/backend.pid" ]; then
        local pid=$(cat "$LOG_DIR/backend.pid")
        if kill -0 $pid 2>/dev/null; then
            kill $pid
            log_success "Backend stopped"
        fi
        rm "$LOG_DIR/backend.pid"
    fi

    # Stop frontend
    if [ -f "$LOG_DIR/frontend.pid" ]; then
        local pid=$(cat "$LOG_DIR/frontend.pid")
        if kill -0 $pid 2>/dev/null; then
            kill $pid
            log_success "Frontend stopped"
        fi
        rm "$LOG_DIR/frontend.pid"
    fi

    # Stop OSA-5
    if [ -f "$LOG_DIR/osa.pid" ]; then
        local pid=$(cat "$LOG_DIR/osa.pid")
        if kill -0 $pid 2>/dev/null; then
            kill $pid
            log_success "OSA-5 stopped"
        fi
        rm "$LOG_DIR/osa.pid"
    fi

    # Stop Docker
    if [ -f "docker-compose.complete.yml" ]; then
        cd "$SCRIPT_DIR"
        docker-compose -f docker-compose.complete.yml down 2>/dev/null || true
    fi

    log_success "All services stopped"
}

# ==============================================================================
# Show Status
# ==============================================================================
show_status() {
    log_step "Service Status"

    echo ""
    echo "╔════════════════════════════════════════════════════════╗"
    echo "║           BusinessOS + OSA-5 Status                     ║"
    echo "╚════════════════════════════════════════════════════════╝"
    echo ""

    # PostgreSQL
    if pgrep -x postgres &>/dev/null; then
        log_success "PostgreSQL: Running"
        if psql -U postgres -h localhost -d businessos -c "SELECT 1" &>/dev/null; then
            echo "           └─ Database connection: OK"
        else
            echo "           └─ Database connection: FAILED"
        fi
    else
        log_warn "PostgreSQL: Stopped"
    fi

    # Redis
    if pgrep redis-server &>/dev/null; then
        log_success "Redis: Running"
    else
        log_warn "Redis: Stopped"
    fi

    # Backend
    if [ -f "$LOG_DIR/backend.pid" ] && kill -0 $(cat "$LOG_DIR/backend.pid") 2>/dev/null; then
        log_success "Backend: Running (Port: 8001)"
        if curl -s http://localhost:8001/health | grep -q "healthy" 2>/dev/null; then
            echo "           └─ Health check: OK"
        else
            echo "           └─ Health check: Not responding"
        fi
    else
        log_warn "Backend: Not running"
    fi

    # Frontend
    if [ -f "$LOG_DIR/frontend.pid" ] && kill -0 $(cat "$LOG_DIR/frontend.pid") 2>/dev/null; then
        log_success "Frontend: Running (Port: 5173)"
    else
        log_warn "Frontend: Not running"
    fi

    # OSA-5
    if [ -f "$LOG_DIR/osa.pid" ] && kill -0 $(cat "$LOG_DIR/osa.pid") 2>/dev/null; then
        log_success "OSA-5: Running (Port: 3003)"
        if curl -s http://localhost:3003/health 2>/dev/null | grep -q "healthy"; then
            echo "           └─ Health check: OK"
        else
            echo "           └─ Health check: Not responding"
        fi
    else
        log_warn "OSA-5: Not running"
    fi

    echo ""
    echo "Endpoints:"
    echo "  Frontend:   http://localhost:5173"
    echo "  Backend:    http://localhost:8001"
    echo "  OSA-5:      http://localhost:3003"
    echo "  Database:   postgresql://postgres@localhost:5432/businessos"
    echo ""
    echo "Logs:"
    echo "  Backend:    $LOG_DIR/backend.log"
    echo "  Frontend:   $LOG_DIR/frontend.log"
    echo "  OSA-5:      $LOG_DIR/osa.log"
    echo ""
}

# ==============================================================================
# Main
# ==============================================================================
main() {
    local mode="${1:-local}"

    case "$mode" in
        local|"")
            check_prerequisites
            start_local
            show_status
            echo ""
            log_success "All services started!"
            log_info "Access BusinessOS: http://localhost:5173"
            log_info "Test OSA: Type 'osa health' in BusinessOS terminal"
            ;;
        docker)
            check_prerequisites
            start_docker
            show_status
            log_success "All services started with Docker!"
            ;;
        stop)
            stop_services
            ;;
        status)
            show_status
            ;;
        *)
            echo "Usage: $0 [mode]"
            echo ""
            echo "Modes:"
            echo "  local   - Start all services locally (default)"
            echo "  docker  - Start all services with Docker Compose"
            echo "  stop    - Stop all services"
            echo "  status  - Show service status"
            echo ""
            echo "Examples:"
            echo "  ./start-all.sh          # Start everything locally"
            echo "  ./start-all.sh docker   # Start with Docker"
            echo "  ./start-all.sh status   # Check what's running"
            echo "  ./start-all.sh stop     # Stop everything"
            exit 1
            ;;
    esac
}

main "$@"
