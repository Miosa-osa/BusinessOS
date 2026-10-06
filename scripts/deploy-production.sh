#!/bin/bash
# =============================================================================
# BusinessOS - Production Deploy Script
# =============================================================================
# Deploys:
#   Backend  -> Railway (service businessos-api)
#   Frontend -> Cloudflare Pages (app.businessos.dev), which proxies /api to Railway
#
# Prerequisites:
#   - railway CLI authenticated: railway login
#   - wrangler authenticated: npx wrangler login
#   - Backend runtime variables already set on the Railway service (Variables tab)
#
# Usage:
#   ./scripts/deploy-production.sh              # deploy both
#   ./scripts/deploy-production.sh backend      # backend only
#   ./scripts/deploy-production.sh frontend     # frontend only

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="$REPO_ROOT/desktop/backend-go"
FRONTEND_DIR="$REPO_ROOT/frontend"

RAILWAY_SERVICE="${RAILWAY_SERVICE:-businessos-api}"
RAILWAY_ENVIRONMENT="${RAILWAY_ENVIRONMENT:-production}"
CF_PROJECT_NAME="${CF_PROJECT_NAME:-businessos-5}"

TARGET="${1:-all}"

log()  { echo "[$(date '+%H:%M:%S')] $*"; }
fail() { echo "[ERROR] $*" >&2; exit 1; }

# =============================================================================
# Backend: Railway (railway up from a copy staged outside the git tree)
# =============================================================================
deploy_backend() {
  log "=== Deploying Go backend to Railway ($RAILWAY_SERVICE) ==="

  if ! command -v railway &>/dev/null; then
    fail "railway CLI not found. Install: npm install --global @railway/cli"
  fi

  # Stage a copy outside the repo. Railway prunes uploads with the repository-root
  # .gitignore, which drops cmd/server from the backend; staging outside git (and
  # --no-gitignore) keeps the archive complete.
  STAGE="$(mktemp -d)/businessos-backend"
  mkdir -p "$STAGE"
  cp -R "$BACKEND_DIR/." "$STAGE"/

  log "Uploading $STAGE to Railway..."
  railway up "$STAGE" \
    --path-as-root \
    --no-gitignore \
    --service "$RAILWAY_SERVICE" \
    --environment "$RAILWAY_ENVIRONMENT" \
    --ci

  log "Backend deployed: https://businessos-api-production.up.railway.app"
}

# =============================================================================
# Frontend: Cloudflare Pages via wrangler
# =============================================================================
deploy_frontend() {
  log "=== Deploying SvelteKit frontend to Cloudflare Pages ==="

  if ! command -v node &>/dev/null; then
    fail "Node.js not found."
  fi

  cd "$FRONTEND_DIR"

  log "Installing dependencies..."
  corepack pnpm install --frozen-lockfile

  # VITE_BACKEND_URL / VITE_API_URL stay UNSET: the web app calls same-origin
  # /api/*, which the Pages Function proxies to Railway.
  log "Building for Cloudflare Pages (static SPA)..."
  CLOUDFLARE_BUILD=true corepack pnpm run build

  log "Deploying to Cloudflare Pages (project: $CF_PROJECT_NAME)..."
  npx wrangler pages deploy build \
    --project-name="$CF_PROJECT_NAME" \
    --branch=main \
    --commit-dirty=true

  log "Frontend deployed: https://app.businessos.dev"
}

# =============================================================================
# Dispatch
# =============================================================================
case "$TARGET" in
  backend)
    deploy_backend
    ;;
  frontend)
    deploy_frontend
    ;;
  all)
    deploy_backend
    deploy_frontend
    log "=== All deployments complete ==="
    log "  API (Railway): https://businessos-api-production.up.railway.app"
    log "  App:           https://app.businessos.dev"
    ;;
  *)
    echo "Usage: $0 [all|backend|frontend]"
    exit 1
    ;;
esac
