# Business OS - Production Deployment Guide

> Last Updated: December 21, 2025

> **See also**: [CLOUD-INFRASTRUCTURE.md](./CLOUD-INFRASTRUCTURE.md) for detailed GCP setup and sync architecture.

This guide covers deploying Business OS in two modes:

**Web Deployment:**
- **Frontend**: Vercel
- **Backend**: Google Cloud Run
- **Database**: Google Cloud SQL (PostgreSQL)

**Desktop App:**
- **Platform**: Electron Forge
- **Bundled**: Frontend + Go Backend
- **Formats**: DMG (macOS), EXE (Windows), DEB (Linux)

---

## Architecture Overview

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│     Vercel      │────▶│   Cloud Run     │────▶│   Cloud SQL     │
│   (Frontend)    │     │   (Go Backend)  │     │  (PostgreSQL)   │
│   SvelteKit     │     │   Port 8080     │     │                 │
└─────────────────┘     └─────────────────┘     └─────────────────┘
        │                       │
        │                       ▼
        │               ┌─────────────────┐
        │               │  AI Providers   │
        │               │ - Anthropic     │
        │               │ - OpenAI        │
        │               │ - Groq          │
        └───────────────┴─────────────────┘
```

---

## Prerequisites

1. **Google Cloud Account** with billing enabled
2. **Vercel Account** (free tier works)
3. **Domain** (optional, for custom domain)
4. **API Keys** for AI providers (Anthropic, OpenAI, Groq)

---

## Step 1: Set Up Google Cloud SQL

### Create PostgreSQL Instance

```bash
# Set your project (BusinessOS uses miosa-460433)
gcloud config set project miosa-460433

# Create Cloud SQL instance (already created)
gcloud sql instances create businessos-db \
  --database-version=POSTGRES_15 \
  --tier=db-f1-micro \
  --region=us-central1 \
  --root-password=YOUR_STRONG_PASSWORD

# Create database (already created as 'businessos')
gcloud sql databases create businessos \
  --instance=businessos-db

# Create user (optional, using postgres user)
gcloud sql users create businessos_user \
  --instance=businessos-db \
  --password=YOUR_USER_PASSWORD
```

### Get Connection Details

```bash
# Get instance connection name
gcloud sql instances describe businessos-db --format='value(connectionName)'
# Output: miosa-460433:us-central1:businessos-db
```

**Current Production Values:**
- Instance: `businessos-db`
- Database: `businessos`
- Connection: `miosa-460433:us-central1:businessos-db`

---

## Step 2: Deploy Backend to Cloud Run

### Option A: Using Cloud Build (Recommended)

```bash
# From project root
cd /path/to/BusinessOS

# Submit build
gcloud builds submit --config backend-go/cloudbuild.yaml
```

### Option B: Manual Deploy

```bash
# Build and push image (using Artifact Registry)
cd desktop/backend-go
docker build -t us-central1-docker.pkg.dev/miosa-460433/businessos-repo/businessos-api:latest .
docker push us-central1-docker.pkg.dev/miosa-460433/businessos-repo/businessos-api:latest

# Deploy to Cloud Run
gcloud run deploy businessos-api \
  --image us-central1-docker.pkg.dev/miosa-460433/businessos-repo/businessos-api:latest \
  --platform managed \
  --region us-central1 \
  --add-cloudsql-instances miosa-460433:us-central1:businessos-db \
  --set-env-vars "ENVIRONMENT=production"
```

**Note**: The service uses IAM authentication (not `--allow-unauthenticated`) due to organization policy.

### Set Environment Variables

In Cloud Run console or via CLI:

```bash
gcloud run services update businessos-api \
  --region us-central1 \
  --set-env-vars "\
DATABASE_URL=postgres://postgres:PASSWORD@/business_os?host=/cloudsql/PROJECT:us-central1:businessos-db,\
ENVIRONMENT=production,\
AI_PROVIDER=anthropic,\
ANTHROPIC_API_KEY=sk-ant-xxx,\
GROQ_API_KEY=gsk_xxx,\
CORS_ORIGINS=https://your-app.vercel.app,\
ENABLE_LOCAL_MODELS=false"
```

### Get Backend URL

```bash
gcloud run services describe businessos-api \
  --region us-central1 \
  --format='value(status.url)'
# Output: https://businessos-api-460433387676.us-central1.run.app
```

**Current Production URL**: `https://businessos-api-460433387676.us-central1.run.app`

---

## Step 3: Deploy Frontend to Vercel

### Connect Repository

1. Go to [vercel.com](https://vercel.com)
2. Import your GitHub repository
3. Select the `frontend` directory as root

### Configure Build Settings

- **Framework Preset**: SvelteKit
- **Build Command**: `npm run build`
- **Output Directory**: `.svelte-kit`
- **Install Command**: `npm install`

### Set Environment Variables

In Vercel dashboard → Settings → Environment Variables:

| Variable | Value |
|----------|-------|
| `VITE_API_URL` | `https://businessos-api-xxxxx.run.app/api` |

### Deploy

```bash
# Or trigger from Vercel dashboard
vercel --prod
```

---

## Step 4: Run Database Migrations

Connect to Cloud SQL and run migrations:

```bash
# Connect via Cloud SQL Proxy
./cloud-sql-proxy YOUR_PROJECT:us-central1:businessos-db &

# Run migrations (from backend-go directory)
# Or connect directly and run SQL
psql "postgres://postgres:PASSWORD@localhost:5432/business_os" < migrations/schema.sql
```

---

## Git Workflow

### Branches

| Branch | Purpose | Deploys To |
|--------|---------|------------|
| `main-go` | Production-ready code | Cloud Run + Vercel (prod) |
| `develop` | Development/staging | Preview deployments |
| Feature branches | New features | PR previews |

### Workflow

```bash
# Start new feature
git checkout develop
git pull
git checkout -b feature/my-feature

# Work on feature...
git add .
git commit -m "feat: add new feature"
git push -u origin feature/my-feature

# Create PR to develop
# After review, merge to develop

# When ready for production
git checkout main-go
git merge develop
git push
```

---

## Environment Variables Reference

### Backend (Cloud Run)

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Yes | Cloud SQL connection string |
| `ENVIRONMENT` | Yes | `production` |
| `AI_PROVIDER` | Yes | `anthropic`, `groq`, or `ollama_cloud` |
| `ANTHROPIC_API_KEY` | If using | Anthropic API key |
| `GROQ_API_KEY` | If using | Groq API key |
| `CORS_ORIGINS` | Yes | Comma-separated allowed origins |
| `ENABLE_LOCAL_MODELS` | No | `false` for production |
| `GOOGLE_CLIENT_ID` | If using | Google OAuth client ID |
| `GOOGLE_CLIENT_SECRET` | If using | Google OAuth secret |

### Frontend (Vercel)

| Variable | Required | Description |
|----------|----------|-------------|
| `VITE_API_URL` | Yes | Backend API URL (Cloud Run) |

---

## Local vs Production AI Providers

| Provider | Local Dev | Production |
|----------|-----------|------------|
| `ollama_local` | ✅ Works | ❌ Not available |
| `ollama_cloud` | ✅ Works | ✅ Works |
| `anthropic` | ✅ Works | ✅ Works (recommended) |
| `groq` | ✅ Works | ✅ Works |

**Note**: In production, set `ENABLE_LOCAL_MODELS=false` and use cloud AI providers.

---

## Monitoring & Logs

### Cloud Run Logs

```bash
gcloud logging read "resource.type=cloud_run_revision \
  AND resource.labels.service_name=businessos-api" \
  --limit 50
```

### View in Console

- **Cloud Run**: [console.cloud.google.com/run](https://console.cloud.google.com/run)
- **Cloud SQL**: [console.cloud.google.com/sql](https://console.cloud.google.com/sql)
- **Vercel**: [vercel.com/dashboard](https://vercel.com/dashboard)

---

## Costs Estimate (Monthly)

| Service | Tier | Est. Cost |
|---------|------|-----------|
| Cloud SQL | db-f1-micro | ~$10-15 |
| Cloud Run | Pay-per-use | ~$5-20 |
| Vercel | Free/Pro | $0-20 |
| AI APIs | Pay-per-use | Variable |

**Total**: ~$15-55/month for small to medium usage

---

## Troubleshooting

### Backend won't connect to database

1. Verify Cloud SQL instance is running
2. Check connection string format
3. Ensure Cloud Run has Cloud SQL Admin role

### CORS errors

1. Verify `CORS_ORIGINS` includes your Vercel domain
2. Check for trailing slashes in URLs
3. Ensure both http and https are covered if needed

### AI provider errors

1. Verify API keys are correct
2. Check provider status pages
3. Ensure `AI_PROVIDER` matches your configured keys

---

## Security Checklist

- [ ] Database password is strong and unique
- [ ] API keys are stored in Secret Manager or env vars (not in code)
- [ ] CORS is restricted to your domains only
- [ ] Cloud Run service account has minimal permissions
- [ ] Enable Cloud SQL SSL/TLS
- [ ] Set up Cloud Armor if needed for DDoS protection

---

# Desktop App Deployment

## Overview

The desktop app bundles the frontend and Go backend into a single distributable application using Electron Forge.

```
desktop/
├── backend-go/          # Go backend (embedded)
├── src/
│   ├── main/            # Electron main process
│   ├── preload/         # Preload scripts
│   └── renderer/        # Built frontend
├── forge.config.ts      # Electron Forge config
└── package.json
```

---

## Step 1: Build Frontend

```bash
cd frontend
npm install
npm run build
```

This creates the production build in `frontend/build/`.

---

## Step 2: Copy Frontend to Desktop

```bash
# Copy built frontend to desktop renderer
cp -r frontend/build/* desktop/src/renderer/
```

---

## Step 3: Build Go Backend

Build for your target platform:

```bash
cd desktop/backend-go

# macOS (current)
go build -o server cmd/server/main.go

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o server-darwin-amd64 cmd/server/main.go

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o server-darwin-arm64 cmd/server/main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o server.exe cmd/server/main.go

# Linux
GOOS=linux GOARCH=amd64 go build -o server-linux cmd/server/main.go
```

---

## Step 4: Package Desktop App

```bash
cd desktop
npm install

# Development run
npm start

# Package for distribution
npm run make
```

Output will be in `desktop/out/make/`:
- macOS: `BusinessOS-darwin-arm64.dmg` or `BusinessOS-darwin-x64.dmg`
- Windows: `BusinessOS Setup.exe`
- Linux: `businessos_1.0.0_amd64.deb`

---

## Electron Forge Configuration

The `forge.config.ts` configures packaging:

```typescript
export default {
  packagerConfig: {
    asar: true,
    name: 'BusinessOS',
    executableName: 'businessos',
    icon: './resources/icon',
    extraResource: [
      './backend-go/server',  // Embedded Go binary
    ],
  },
  makers: [
    {
      name: '@electron-forge/maker-squirrel',
      config: { name: 'BusinessOS' },
    },
    {
      name: '@electron-forge/maker-dmg',
      config: { format: 'ULFO' },
    },
    {
      name: '@electron-forge/maker-deb',
      config: {},
    },
  ],
};
```

---

## Auto-starting Backend

The Electron main process spawns the Go backend:

```typescript
// desktop/src/main/index.ts
import { spawn } from 'child_process';
import path from 'path';

let backendProcess: ChildProcess | null = null;

function startBackend() {
  const backendPath = app.isPackaged
    ? path.join(process.resourcesPath, 'server')
    : path.join(__dirname, '../../backend-go/server');

  backendProcess = spawn(backendPath, [], {
    env: { ...process.env, SERVER_PORT: '8000' },
  });
}

app.on('ready', () => {
  startBackend();
  createWindow();
});

app.on('quit', () => {
  backendProcess?.kill();
});
```

---

## Desktop Environment Variables

Create `.env` in `desktop/backend-go/`:

```env
# Database (local SQLite or PostgreSQL)
DATABASE_URL=postgres://user:pass@localhost:5432/business_os

# Server
SERVER_PORT=8000

# AI Provider (ollama_local recommended for desktop)
AI_PROVIDER=ollama_local
OLLAMA_LOCAL_URL=http://localhost:11434
DEFAULT_MODEL=llama3.2:3b

# CORS (allow Electron renderer)
ALLOWED_ORIGINS=http://localhost:*,file://*
```

---

## Distribution Checklist

- [ ] Frontend is built and copied to `desktop/src/renderer/`
- [ ] Go backend is compiled for target platform
- [ ] Environment variables are configured
- [ ] App icon is set in `desktop/resources/`
- [ ] Code signing (macOS/Windows) is configured
- [ ] Auto-update server is set up (optional)

---

## Platform-Specific Notes

### macOS

- Requires code signing for distribution outside App Store
- Use `electron-notarize` for notarization
- DMG is the standard distribution format

### Windows

- Squirrel.Windows handles installation and updates
- Code signing recommended for trust
- `.exe` installer is created

### Linux

- `.deb` package for Debian/Ubuntu
- Can also create `.rpm` or AppImage

---

## Troubleshooting Desktop App

### Backend won't start

1. Check binary exists at expected path
2. Verify executable permissions: `chmod +x server`
3. Check logs in Electron DevTools

### Frontend won't load

1. Verify files exist in `desktop/src/renderer/`
2. Check Vite renderer config paths
3. Open DevTools to check for errors

### API calls failing

1. Verify backend is running on port 8000
2. Check CORS configuration
3. Inspect network requests in DevTools
