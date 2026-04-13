#!/bin/bash
# BusinessOS — Start all services
# Usage: ./start.sh [frontend|backend|desktop|all]

set -e

DIR="$(cd "$(dirname "$0")" && pwd)"
export NVM_DIR="$HOME/.nvm"
. "$NVM_DIR/nvm.sh"
nvm use 22 > /dev/null 2>&1

MODE="${1:-all}"

start_backend() {
  echo "Starting Go backend on port 8001..."
  cd "$DIR/desktop/backend-go"
  nohup go run ./cmd/server/ > /tmp/bos-backend.log 2>&1 &
  echo "  PID: $! → log: /tmp/bos-backend.log"
}

start_frontend() {
  echo "Starting SvelteKit frontend on port 5173..."
  cd "$DIR/frontend"
  nohup npx vite dev --port 5173 > /tmp/bos-frontend.log 2>&1 &
  echo "  PID: $! → log: /tmp/bos-frontend.log"
}

start_desktop() {
  echo "Starting Electron desktop..."
  cd "$DIR/desktop"
  npm start &
}

wait_for_port() {
  local port=$1 name=$2 max=30
  for i in $(seq 1 $max); do
    if lsof -i :$port -P -n 2>/dev/null | grep -q LISTEN; then
      echo "  ✓ $name ready on port $port"
      return 0
    fi
    sleep 1
  done
  echo "  ✗ $name failed to start on port $port"
  return 1
}

stop_all() {
  echo "Stopping all BusinessOS services..."
  lsof -i :5173 -t 2>/dev/null | xargs kill 2>/dev/null
  lsof -i :8001 -t 2>/dev/null | xargs kill 2>/dev/null
  pkill -f "Electron.*businessos" 2>/dev/null
  echo "  ✓ All stopped"
}

case "$MODE" in
  backend)
    start_backend
    wait_for_port 8001 "Backend"
    ;;
  frontend)
    start_frontend
    wait_for_port 5173 "Frontend"
    ;;
  desktop)
    start_desktop
    ;;
  stop)
    stop_all
    ;;
  all)
    echo "═══════════════════════════════════"
    echo "  BusinessOS — Starting all services"
    echo "═══════════════════════════════════"
    echo ""
    start_backend
    start_frontend
    echo ""
    echo "Waiting for services..."
    wait_for_port 8001 "Backend"
    wait_for_port 5173 "Frontend"
    echo ""
    echo "Starting Electron desktop..."
    start_desktop
    echo ""
    echo "═══════════════════════════════════"
    echo "  All services running:"
    echo "  Backend:  http://localhost:8001"
    echo "  Frontend: http://localhost:5173"
    echo "  Desktop:  Electron app"
    echo ""
    echo "  Stop all: ./start.sh stop"
    echo "═══════════════════════════════════"
    ;;
  *)
    echo "Usage: ./start.sh [frontend|backend|desktop|stop|all]"
    ;;
esac
