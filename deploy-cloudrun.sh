#!/bin/bash
set -e

# Configuration
PROJECT_ID="gen-lang-client-0338297768"
REGION="asia-northeast1"
SERVICE_NAME="suiten-server"
IMAGE_NAME="gcr.io/${PROJECT_ID}/${SERVICE_NAME}"

# Database configuration (Neon)
DATABASE_URL="postgresql://neondb_owner:npg_RsQepyAW86Nw@ep-tiny-glitter-a12bte2m-pooler.ap-southeast-1.aws.neon.tech/neondb?sslmode=require"

# Generate a secure JWT secret if not provided
JWT_SECRET="${JWT_SECRET:-$(openssl rand -base64 32)}"

echo "============================================"
echo "SuiteN Server - Cloud Run Deployment"
echo "============================================"
echo "Project: ${PROJECT_ID}"
echo "Region: ${REGION}"
echo "Service: ${SERVICE_NAME}"
echo "============================================"

# Check if gcloud is installed
if ! command -v gcloud &> /dev/null; then
    echo "Error: gcloud CLI is not installed."
    echo "Install it from: https://cloud.google.com/sdk/docs/install"
    exit 1
fi

# Set project
echo ""
echo "=== Setting GCP Project ==="
gcloud config set project ${PROJECT_ID}

# Enable required APIs
echo ""
echo "=== Enabling Required APIs ==="
gcloud services enable containerregistry.googleapis.com run.googleapis.com cloudbuild.googleapis.com

# Configure Docker for GCR
echo ""
echo "=== Configuring Docker for GCR ==="
gcloud auth configure-docker --quiet

# Build and push image
echo ""
echo "=== Building Docker Image ==="
cd "$(dirname "$0")"
docker build -t ${IMAGE_NAME}:latest .

echo ""
echo "=== Pushing Image to GCR ==="
docker push ${IMAGE_NAME}:latest

# Deploy to Cloud Run
echo ""
echo "=== Deploying to Cloud Run ==="
gcloud run deploy ${SERVICE_NAME} \
  --image ${IMAGE_NAME}:latest \
  --platform managed \
  --region ${REGION} \
  --allow-unauthenticated \
  --set-env-vars "DATABASE_URL=${DATABASE_URL}" \
  --set-env-vars "JWT_SECRET=${JWT_SECRET}" \
  --memory 256Mi \
  --cpu 1 \
  --min-instances 0 \
  --max-instances 10 \
  --timeout 60

# Get service URL
echo ""
echo "============================================"
echo "Deployment Complete!"
echo "============================================"
SERVICE_URL=$(gcloud run services describe ${SERVICE_NAME} --region ${REGION} --format="value(status.url)")
echo "Service URL: ${SERVICE_URL}"
echo ""
echo "Test with:"
echo "  curl ${SERVICE_URL}/health"
echo ""
echo "JWT_SECRET used: ${JWT_SECRET}"
echo "(Save this for your mobile app configuration)"
echo "============================================"
