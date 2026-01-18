# Deployment Guide - OSA Voice Agent to Google Cloud Run

## Prerequisites

1. Google Cloud SDK installed (`gcloud`)
2. Project ID: `businessos-431800` (or your GCP project)
3. Region: `us-central1` (recommended)
4. Docker installed locally (for testing)

## Environment Variables Needed

These must be set in Cloud Run:

```bash
LIVEKIT_URL=wss://macstudiosystems-yn61tekm.livekit.cloud
LIVEKIT_API_KEY=APIcFNUEtCEkZpa
LIVEKIT_API_SECRET=iBtjeSlz2ioQ8Ptd9SiOOW5B2ihO1Ff6gSjWtKanflxA
GROQ_API_KEY=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
ELEVENLABS_API_KEY=sk_4fd29ef975197a42a9d5d9b0b4ac809720e6a7c2ee8ef657
ELEVENLABS_VOICE_ID=KoVIHoyLDrQyd4pGalbs
GO_BACKEND_URL=http://localhost:8001
```

## Deployment Steps

### 1. Test Docker Build Locally

```bash
cd python-voice-agent

# Build image
docker build -t osa-voice-agent:latest .

# Test run (requires env vars)
docker run -p 8080:8080 \
  -e LIVEKIT_URL=wss://... \
  -e LIVEKIT_API_KEY=... \
  -e LIVEKIT_API_SECRET=... \
  -e GROQ_API_KEY=... \
  -e ELEVENLABS_API_KEY=... \
  -e ELEVENLABS_VOICE_ID=... \
  osa-voice-agent:latest
```

### 2. Configure GCloud

```bash
# Login
gcloud auth login

# Set project
gcloud config set project businessos-431800

# Enable required APIs
gcloud services enable \
  run.googleapis.com \
  containerregistry.googleapis.com \
  cloudbuild.googleapis.com
```

### 3. Deploy to Cloud Run

```bash
cd python-voice-agent

gcloud run deploy osa-voice-agent \
  --source . \
  --region us-central1 \
  --platform managed \
  --allow-unauthenticated \
  --set-env-vars LIVEKIT_URL=wss://macstudiosystems-yn61tekm.livekit.cloud \
  --set-env-vars LIVEKIT_API_KEY=APIcFNUEtCEkZpa \
  --set-env-vars LIVEKIT_API_SECRET=iBtjeSlz2ioQ8Ptd9SiOOW5B2ihO1Ff6gSjWtKanflxA \
  --set-env-vars GROQ_API_KEY=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 \
  --set-env-vars ELEVENLABS_API_KEY=sk_4fd29ef975197a42a9d5d9b0b4ac809720e6a7c2ee8ef657 \
  --set-env-vars ELEVENLABS_VOICE_ID=KoVIHoyLDrQyd4pGalbs \
  --set-env-vars GO_BACKEND_URL=https://your-backend.run.app \
  --memory 1Gi \
  --cpu 2 \
  --timeout 300 \
  --concurrency 10 \
  --min-instances 0 \
  --max-instances 10
```

### 4. Using Secret Manager (Recommended for Production)

Instead of passing secrets as env vars in the command, use Secret Manager:

```bash
# Create secrets
echo -n "wss://macstudiosystems-yn61tekm.livekit.cloud" | \
  gcloud secrets create livekit-url --data-file=-

echo -n "APIcFNUEtCEkZpa" | \
  gcloud secrets create livekit-api-key --data-file=-

echo -n "iBtjeSlz2ioQ8Ptd9SiOOW5B2ihO1Ff6gSjWtKanflxA" | \
  gcloud secrets create livekit-api-secret --data-file=-

echo -n "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" | \
  gcloud secrets create groq-api-key --data-file=-

echo -n "sk_4fd29ef975197a42a9d5d9b0b4ac809720e6a7c2ee8ef657" | \
  gcloud secrets create elevenlabs-api-key --data-file=-

echo -n "KoVIHoyLDrQyd4pGalbs" | \
  gcloud secrets create elevenlabs-voice-id --data-file=-

# Deploy with secrets
gcloud run deploy osa-voice-agent \
  --source . \
  --region us-central1 \
  --set-secrets LIVEKIT_URL=livekit-url:latest \
  --set-secrets LIVEKIT_API_KEY=livekit-api-key:latest \
  --set-secrets LIVEKIT_API_SECRET=livekit-api-secret:latest \
  --set-secrets GROQ_API_KEY=groq-api-key:latest \
  --set-secrets ELEVENLABS_API_KEY=elevenlabs-api-key:latest \
  --set-secrets ELEVENLABS_VOICE_ID=elevenlabs-voice-id:latest \
  --set-env-vars GO_BACKEND_URL=https://your-backend.run.app
```

### 5. Verify Deployment

```bash
# Get service URL
gcloud run services describe osa-voice-agent \
  --region us-central1 \
  --format='value(status.url)'

# Check logs
gcloud run services logs read osa-voice-agent \
  --region us-central1 \
  --limit 50
```

## Architecture After Deployment

```
Frontend (https://businessos.app)
    ↓ WebRTC
LiveKit Cloud (wss://macstudiosystems-yn61tekm.livekit.cloud)
    ↓ Assigns room
Python Voice Agent (Cloud Run: https://osa-voice-agent-....run.app)
    ↓ Fetches context
Go Backend (Cloud Run: https://your-backend.run.app)
    ↓ Returns user data
Python Voice Agent responds via LiveKit
    ↓ WebRTC audio
Frontend plays audio
```

## Monitoring

### View Logs
```bash
gcloud run services logs read osa-voice-agent --region us-central1 --follow
```

### View Metrics
```bash
# Open in GCP Console
gcloud run services describe osa-voice-agent --region us-central1
```

## Scaling Configuration

Current settings:
- **Min instances**: 0 (scales to zero when idle)
- **Max instances**: 10
- **Concurrency**: 10 (handles 10 concurrent LiveKit rooms)
- **Memory**: 1Gi
- **CPU**: 2
- **Timeout**: 300s (5 minutes max per room)

Adjust based on load:
- High usage: Increase min-instances to 1-2 (reduces cold starts)
- More concurrent users: Increase max-instances
- Memory issues: Increase memory to 2Gi

## Troubleshooting

**Issue**: Cold starts are slow
- **Solution**: Set `--min-instances 1` to keep one instance warm

**Issue**: Agent not connecting to LiveKit
- **Check**: LIVEKIT_URL, LIVEKIT_API_KEY, LIVEKIT_API_SECRET are correct
- **Check**: Network connectivity from Cloud Run to LiveKit cloud

**Issue**: No user context in responses
- **Check**: GO_BACKEND_URL points to correct backend
- **Check**: Backend endpoint `/api/voice/user-context/:user_id` is accessible
- **Check**: Backend is deployed and running

**Issue**: "Insufficient permissions" error
- **Solution**: Grant Cloud Run service account access to Secret Manager:
```bash
gcloud projects add-iam-policy-binding businessos-431800 \
  --member=serviceAccount:SERVICE_ACCOUNT_EMAIL \
  --role=roles/secretmanager.secretAccessor
```

## Cost Optimization

Cloud Run billing:
- Charged per request + CPU/memory time
- Free tier: 2 million requests/month
- Voice sessions are long-running (billed for duration)

Tips:
- Use `min-instances=0` for development (scales to zero)
- Use `min-instances=1` for production (better UX, slight cost increase)
- Monitor usage in GCP Console → Cloud Run → Metrics

## Updating the Agent

```bash
# Make code changes to agent.py
# Then redeploy:
gcloud run deploy osa-voice-agent \
  --source . \
  --region us-central1
# Secrets/env vars are preserved across deployments
```

## Rollback

```bash
# List revisions
gcloud run revisions list --service osa-voice-agent --region us-central1

# Rollback to previous revision
gcloud run services update-traffic osa-voice-agent \
  --region us-central1 \
  --to-revisions REVISION_NAME=100
```
