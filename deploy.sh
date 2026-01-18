#!/bin/bash
set -e

PROJECT_ID="gen-lang-client-0338297768"
REGION="asia-northeast1"
SERVICE_NAME="suiten-server"
IMAGE_NAME="gcr.io/${PROJECT_ID}/${SERVICE_NAME}"

echo "=== Building Docker image ==="
docker build -t ${IMAGE_NAME}:latest .

echo "=== Pushing to Container Registry ==="
docker push ${IMAGE_NAME}:latest

echo "=== Deploying to Cloud Run ==="
gcloud run deploy ${SERVICE_NAME} \
  --image ${IMAGE_NAME}:latest \
  --platform managed \
  --region ${REGION} \
  --project ${PROJECT_ID} \
  --allow-unauthenticated \
  --set-env-vars "DATABASE_URL=${DATABASE_URL}" \
  --set-env-vars "JWT_SECRET=${JWT_SECRET}"

echo "=== Deployment complete ==="
gcloud run services describe ${SERVICE_NAME} --region ${REGION} --project ${PROJECT_ID} --format="value(status.url)"
