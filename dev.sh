#!/bin/bash
# BusinessOS Development Environment Launcher
# One command to start everything

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Configuration
BACKEND_DIR="desktop/backend-go"
FRONTEND_DIR="frontend"
DATABASE_URL="${DATABASE_URL:?ERROR: DATABASE_URL must be set}"

echo -e "${GREEN}BusinessOS Development Environment${NC}"
echo "=================================="

# Function to cleanup on exit
cleanup() {
    echo -e "\n${YELLOW}Shutting down...${NC}"
    [ ! -z "$BACKEND_PID" ] && kill $BACKEND_PID 2>/dev/null || true
    [ ! -z "$FRONTEND_PID" ] && kill $FRONTEND_PID 2>/dev/null || true
    exit 0
}

trap cleanup EXIT INT TERM

# Kill existing processes
echo "Cleaning up existing processes..."
pkill -f "bin/backend" 2>/dev/null || true
pkill -f "pnpm.*dev" 2>/dev/null || true

# Backend
echo -e "\n${GREEN}Starting Backend...${NC}"
cd $BACKEND_DIR
export DATABASE_URL="$DATABASE_URL"
go build -o bin/backend ./cmd/server
./bin/backend > backend.log 2>&1 &
BACKEND_PID=$!
echo -e "${GREEN}Backend started (PID: $BACKEND_PID)${NC}"

sleep 2

# Frontend
echo -e "\n${GREEN}Starting Frontend...${NC}"
cd ../../$FRONTEND_DIR
[ ! -d "node_modules" ] && pnpm install
pnpm dev > ../frontend.log 2>&1 &
FRONTEND_PID=$!
echo -e "${GREEN}Frontend started (PID: $FRONTEND_PID)${NC}"

echo -e "\n${GREEN}Ready! Backend: :8001 | Frontend: :5173${NC}"
echo -e "${YELLOW}Press Ctrl+C to stop${NC}"

wait
