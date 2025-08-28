#!/bin/bash

# Deploy Square Service to GCP Cloud Run
# Usage: ./deploy.sh

set -e

echo "🚀 Deploying Rise Square Service to GCP..."

# Set project
PROJECT_ID="sacred-armor-452904-c0"
SERVICE_NAME="rise-square-service"
REGION="us-central1"

# Check if gcloud is authenticated
if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | head -1 > /dev/null; then
    echo "❌ Not authenticated with gcloud. Run: gcloud auth login"
    exit 1
fi

# Set the project
gcloud config set project $PROJECT_ID

# Enable required APIs
echo "📋 Enabling required APIs..."
gcloud services enable cloudbuild.googleapis.com
gcloud services enable run.googleapis.com
gcloud services enable containerregistry.googleapis.com

# Build and deploy using Cloud Build
echo "🔨 Building and deploying with Cloud Build..."
cd "$(dirname "$0")"
gcloud builds submit --config cloudbuild.yaml .

# Get the service URL
SERVICE_URL=$(gcloud run services describe $SERVICE_NAME --region=$REGION --format="value(status.url)")

echo "✅ Deployment complete!"
echo "🌐 Service URL: $SERVICE_URL"
echo "📊 Health check: $SERVICE_URL/health"
echo "📈 Metrics: $SERVICE_URL/metrics"

echo ""
echo "🔧 Next steps:"
echo "1. Set up your environment variables in Cloud Run console"
echo "2. Configure your database connection"
echo "3. Update Square webhook URLs to point to: $SERVICE_URL/webhook"
echo "4. Test the deployment with: curl $SERVICE_URL/health"