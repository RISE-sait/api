# 🚀 GCP Cloud Run Deployment Guide

## Quick Deploy

```bash
# Make deploy script executable
chmod +x deploy.sh

# Deploy to GCP
./deploy.sh
```

## Manual Setup (if needed)

### 1. Prerequisites
```bash
# Install gcloud CLI
# https://cloud.google.com/sdk/docs/install

# Authenticate
gcloud auth login
gcloud config set project sacred-armor-452904-c0

# Enable APIs
gcloud services enable cloudbuild.googleapis.com
gcloud services enable run.googleapis.com
gcloud services enable containerregistry.googleapis.com
```

### 2. Deploy
```bash
cd square_service
gcloud builds submit --config cloudbuild.yaml .
```

### 3. Configure Environment Variables

**Option A: Via Console**
1. Go to [Cloud Run Console](https://console.cloud.google.com/run)
2. Select `rise-square-service`
3. Click "EDIT & DEPLOY NEW REVISION"
4. Add environment variables from `env-template.yaml`

**Option B: Via CLI**
```bash
# Set each environment variable
gcloud run services update rise-square-service \
  --region=us-central1 \
  --set-env-vars="DATABASE_URL=postgresql://...,SQUARE_ACCESS_TOKEN=..." 
```

**Option C: Use Secret Manager (Recommended)**
```bash
# Create secrets
gcloud secrets create square-access-token --data-file=token.txt
gcloud secrets create jwt-secret --data-file=jwt.txt

# Grant Cloud Run access
gcloud secrets add-iam-policy-binding square-access-token \
  --member="serviceAccount:your-compute@developer.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"
```

## 🔒 Security Checklist

- [x] **Authentication required** (`--allow-unauthenticated=false`)
- [x] **Resource limits** (1GB memory, 1 CPU, max 10 instances)
- [x] **Production environment** (`SQUARE_ENV=production`)
- [x] **Non-root container** user
- [x] **HTTPS only** (Cloud Run default)
- [ ] **Environment variables configured**
- [ ] **Database connection secured**
- [ ] **Square webhook URLs updated**

## 📊 Monitoring

After deployment:
- **Health**: `https://your-service-url/health`
- **Metrics**: `https://your-service-url/metrics`
- **Logs**: [Cloud Run Logs](https://console.cloud.google.com/run)

## 🔧 Troubleshooting

**Build fails?**
```bash
# Check build logs
gcloud builds log [BUILD_ID]
```

**Service not starting?**
```bash
# Check service logs
gcloud run services logs tail rise-square-service --region=us-central1
```

**Environment issues?**
- Verify all required env vars are set
- Check database connectivity
- Ensure Square credentials are valid

## 📝 Post-Deployment

1. **Test health endpoint**: `curl https://your-service-url/health`
2. **Update Square webhook URLs** in Square Dashboard
3. **Configure monitoring alerts** in GCP Console
4. **Set up log-based metrics** for errors
5. **Create backup/rollback procedures**