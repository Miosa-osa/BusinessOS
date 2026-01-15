#!/bin/bash
# BusinessOS Development Startup Script
# Ensures Node 22+, installs dependencies, and starts both backend and frontend

set -e  # Exit on error

echo "🚀 BusinessOS Development Startup"
echo "=================================="

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if nvm is available
if [ ! -d "$HOME/.nvm" ]; then
    echo "${YELLOW}⚠️  nvm not found. Please install nvm first:${NC}"
    echo "   curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash"
    exit 1
fi

# Load nvm
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"

# Check and switch to Node 22+
echo "${BLUE}📦 Checking Node version...${NC}"
CURRENT_NODE=$(node -v 2>/dev/null || echo "none")
if [[ "$CURRENT_NODE" != v22* ]]; then
    echo "${YELLOW}⚠️  Node 22+ required. Current: $CURRENT_NODE${NC}"
    echo "${BLUE}Installing Node 22...${NC}"
    nvm install 22
    nvm use 22
else
    echo "${GREEN}✅ Node 22 already active: $CURRENT_NODE${NC}"
fi

# Verify Node version
NODE_VERSION=$(node -v)
echo "${GREEN}✅ Using Node: $NODE_VERSION${NC}"

# Install/Update Frontend Dependencies
echo ""
echo "${BLUE}📦 Installing Frontend Dependencies...${NC}"
cd frontend
if [ ! -d "node_modules" ]; then
    echo "${YELLOW}⚠️  node_modules not found. Running npm install...${NC}"
    npm install
else
    echo "${GREEN}✅ node_modules exists. Checking for updates...${NC}"
    npm install
fi

# Install/Update Backend Dependencies
echo ""
echo "${BLUE}📦 Checking Backend Dependencies (Go)...${NC}"
cd ../desktop/backend-go
if [ ! -d "vendor" ]; then
    echo "${YELLOW}⚠️  Running go mod download...${NC}"
    go mod download
fi

# Check if sqlc is installed
if ! command -v sqlc &> /dev/null; then
    echo "${YELLOW}⚠️  sqlc not found. Install with: brew install sqlc${NC}"
fi

# Build backend
echo ""
echo "${BLUE}🔨 Building Backend...${NC}"
go build -o bin/server ./cmd/server

# Start services
echo ""
echo "${GREEN}🚀 Starting Services...${NC}"
echo ""

# Start backend in background
echo "${BLUE}Starting Backend on port 8001...${NC}"
./bin/server > /tmp/businessos-backend.log 2>&1 &
BACKEND_PID=$!
echo "${GREEN}✅ Backend started (PID: $BACKEND_PID)${NC}"
echo "   Logs: /tmp/businessos-backend.log"

# Wait for backend to be ready
sleep 3

# Start frontend in background
echo ""
echo "${BLUE}Starting Frontend on port 5173...${NC}"
cd ../../frontend
npm run dev > /tmp/businessos-frontend.log 2>&1 &
FRONTEND_PID=$!
echo "${GREEN}✅ Frontend started (PID: $FRONTEND_PID)${NC}"
echo "   Logs: /tmp/businessos-frontend.log"

# Wait for frontend to be ready
sleep 5

echo ""
echo "${GREEN}=================================="
echo "✨ BusinessOS is running!"
echo "=================================="
echo ""
echo "  Frontend: ${BLUE}http://localhost:5173${NC}"
echo "  Backend:  ${BLUE}http://localhost:8001${NC}"
echo ""
echo "  Backend PID:  $BACKEND_PID"
echo "  Frontend PID: $FRONTEND_PID"
echo ""
echo "To stop:"
echo "  kill $BACKEND_PID $FRONTEND_PID"
echo ""
echo "Logs:"
echo "  tail -f /tmp/businessos-backend.log"
echo "  tail -f /tmp/businessos-frontend.log"
echo ""
