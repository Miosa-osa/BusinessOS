#!/bin/bash
# ═══════════════════════════════════════════════════════════════════════════════
# BusinessOS Development Startup Script
# ═══════════════════════════════════════════════════════════════════════════════

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo -e "${BLUE}"
    echo "╔═══════════════════════════════════════════════════════════════════════════════╗"
    echo "║                        BusinessOS Development Server                          ║"
    echo "╚═══════════════════════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

print_status() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

# Check if running in the right directory
if [ ! -f "CLAUDE.md" ]; then
    print_error "Please run this script from the BusinessOS-1 root directory"
    exit 1
fi

print_header

MODE=${1:-"local"}

case $MODE in
    "local")
        echo -e "${YELLOW}Starting in LOCAL mode (no containers)${NC}\n"

        # Check dependencies
        command -v go >/dev/null 2>&1 || { print_error "Go is not installed"; exit 1; }
        command -v node >/dev/null 2>&1 || { print_error "Node.js is not installed"; exit 1; }

        # Check for .env file
        if [ ! -f "desktop/backend-go/.env" ]; then
            print_warning "No .env file found. Copying from .env.example..."
            cp desktop/backend-go/.env.example desktop/backend-go/.env
            print_warning "Please edit desktop/backend-go/.env with your credentials"
        fi

        # Start backend in background
        print_status "Starting Go Backend on :8001..."
        cd desktop/backend-go
        go run ./cmd/server &
        BACKEND_PID=$!
        cd ../..

        # Wait for backend to start
        sleep 3

        # Start frontend
        print_status "Starting SvelteKit Frontend on :5173..."
        cd frontend
        npm run dev &
        FRONTEND_PID=$!
        cd ..

        echo ""
        echo -e "${GREEN}═══════════════════════════════════════════════════════════════════════════════${NC}"
        echo -e "${GREEN}  BusinessOS is running!${NC}"
        echo -e "${GREEN}═══════════════════════════════════════════════════════════════════════════════${NC}"
        echo ""
        echo "  Frontend:  http://localhost:5173"
        echo "  Backend:   http://localhost:8001"
        echo "  API Docs:  http://localhost:8001/api/docs"
        echo ""
        echo "  Press Ctrl+C to stop all services"
        echo ""

        # Wait for Ctrl+C
        trap "kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; exit" SIGINT SIGTERM
        wait
        ;;

    "docker")
        echo -e "${YELLOW}Starting in DOCKER mode${NC}\n"

        command -v docker >/dev/null 2>&1 || { print_error "Docker is not installed"; exit 1; }
        command -v docker-compose >/dev/null 2>&1 || { print_error "Docker Compose is not installed"; exit 1; }

        # Check for .env
        if [ ! -f ".env" ]; then
            print_warning "Creating .env from template..."
            cat > .env << 'EOF'
# Database (Supabase)
DATABASE_URL=your-supabase-connection-string

# AI Providers
ANTHROPIC_API_KEY=your-key
GROQ_API_KEY=your-key

# Security
SECRET_KEY=change-me-in-production-32-chars
TOKEN_ENCRYPTION_KEY=change-me-32-characters-long!!!

# Redis
REDIS_PASSWORD=changeme

# OSA Integration
OSA_WEBHOOK_SECRET=your-osa-webhook-secret
EOF
            print_warning "Please edit .env with your credentials"
            exit 1
        fi

        print_status "Building and starting containers..."
        docker-compose -f docker-compose.full.yml up --build -d

        echo ""
        echo -e "${GREEN}═══════════════════════════════════════════════════════════════════════════════${NC}"
        echo -e "${GREEN}  BusinessOS Docker Stack Running!${NC}"
        echo -e "${GREEN}═══════════════════════════════════════════════════════════════════════════════${NC}"
        echo ""
        echo "  Frontend:  http://localhost:5173"
        echo "  Backend:   http://localhost:8001"
        echo "  Redis:     localhost:6379"
        echo ""
        echo "  Commands:"
        echo "    docker-compose -f docker-compose.full.yml logs -f   # View logs"
        echo "    docker-compose -f docker-compose.full.yml down      # Stop all"
        echo ""
        ;;

    "stop")
        echo -e "${YELLOW}Stopping Docker containers...${NC}"
        docker-compose -f docker-compose.full.yml down
        print_status "All containers stopped"
        ;;

    *)
        echo "Usage: ./start-dev.sh [local|docker|stop]"
        echo ""
        echo "  local  - Run without containers (default)"
        echo "  docker - Run with Docker Compose"
        echo "  stop   - Stop Docker containers"
        ;;
esac
