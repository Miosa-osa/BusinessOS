#!/usr/bin/env bash
# =============================================================================
# BusinessOS — First-time setup script
# Usage: bash scripts/setup.sh
# =============================================================================
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# ── Colours ──────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
RESET='\033[0m'

ok()   { printf "  ${GREEN}✓${RESET} %s\n" "$*"; }
warn() { printf "  ${YELLOW}!${RESET} %s\n" "$*"; }
err()  { printf "  ${RED}✗${RESET} %s\n" "$*" >&2; }
step() { printf "\n${BOLD}%s${RESET}\n" "$*"; }

# ── Prerequisite checks ───────────────────────────────────────────────────────
step "Checking prerequisites..."

if ! command -v docker &>/dev/null; then
  err "Docker is not installed. Install it from https://docs.docker.com/get-docker/"
  exit 1
fi
ok "Docker $(docker --version | awk '{print $3}' | tr -d ',')"

if ! docker info &>/dev/null 2>&1; then
  err "Docker daemon is not running. Start Docker and try again."
  exit 1
fi
ok "Docker daemon is running"

# Compose can be `docker compose` (plugin) or `docker-compose` (standalone)
if docker compose version &>/dev/null 2>&1; then
  COMPOSE="docker compose"
elif command -v docker-compose &>/dev/null; then
  COMPOSE="docker-compose"
else
  err "Docker Compose not found. Install the Compose plugin: https://docs.docker.com/compose/install/"
  exit 1
fi
ok "Compose: $($COMPOSE version --short 2>/dev/null || $COMPOSE version)"

# ── Environment files ─────────────────────────────────────────────────────────
step "Setting up environment files..."

cd "$ROOT"

if [[ ! -f .env ]]; then
  cp .env.example .env
  ok "Created .env from .env.example"
  warn "Edit .env and set the REQUIRED values (POSTGRES_PASSWORD, REDIS_PASSWORD, SECRET_KEY, TOKEN_ENCRYPTION_KEY, REDIS_KEY_HMAC_SECRET) before continuing."
  printf "\n  Press ENTER when done, or Ctrl-C to exit.\n"
  read -r
else
  ok ".env already exists"
fi

if [[ ! -f desktop/backend-go/.env ]]; then
  cp desktop/backend-go/.env.example desktop/backend-go/.env
  ok "Created desktop/backend-go/.env"
else
  ok "desktop/backend-go/.env already exists"
fi

# ── Validate critical keys ────────────────────────────────────────────────────
step "Validating required keys in .env..."

# shellcheck disable=SC1091
source .env 2>/dev/null || true

MISSING=()
[[ -z "${POSTGRES_PASSWORD:-}" ]] && MISSING+=("POSTGRES_PASSWORD")
[[ -z "${REDIS_PASSWORD:-}" ]]    && MISSING+=("REDIS_PASSWORD")
[[ -z "${SECRET_KEY:-}" ]]        && MISSING+=("SECRET_KEY")
[[ -z "${TOKEN_ENCRYPTION_KEY:-}" ]] && MISSING+=("TOKEN_ENCRYPTION_KEY")
[[ -z "${REDIS_KEY_HMAC_SECRET:-}" ]] && MISSING+=("REDIS_KEY_HMAC_SECRET")

if [[ ${#MISSING[@]} -gt 0 ]]; then
  err "The following required keys are empty in .env:"
  for key in "${MISSING[@]}"; do
    err "  - $key"
  done
  printf "\n  Generate them with:\n"
  printf "    openssl rand -base64 64   # for SECRET_KEY\n"
  printf "    openssl rand -base64 32   # for all others\n\n"
  err "Fix .env and run this script again."
  exit 1
fi
ok "All required keys are set"

# ── Pull base images ──────────────────────────────────────────────────────────
step "Pulling base images..."
$COMPOSE pull postgres redis

# ── Start infrastructure ──────────────────────────────────────────────────────
step "Starting postgres and redis..."
$COMPOSE up -d postgres redis

bash "$ROOT/scripts/wait-healthy.sh" postgres redis

# ── Build and start application ───────────────────────────────────────────────
step "Building and starting backend..."
$COMPOSE up -d --build backend

bash "$ROOT/scripts/wait-healthy.sh" backend

step "Building and starting frontend..."
$COMPOSE up -d --build frontend

# ── Print URLs ────────────────────────────────────────────────────────────────
bash "$ROOT/scripts/print-urls.sh"
