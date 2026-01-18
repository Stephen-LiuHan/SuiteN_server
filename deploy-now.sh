#!/bin/bash
set -e

cd "$(dirname "$0")"

echo "=== SuiteN Server Deployment ==="

# Use gcloud from Docker
GCLOUD="docker run --rm -i -v /var/run/docker.sock:/var/run/docker.sock -v $(pwd):/app -v /home/liuhan/.config/gcloud:/root/.config/gcloud -w /app gcr.io/google.com/cloudsdktool/cloud-sdk:slim"

# Step 1: Authenticate
echo ""
echo "Step 1: Authenticating with Google Cloud..."
$GCLOUD gcloud auth login --no-launch-browser

# Step 2: Set project
echo ""
echo "Step 2: Setting project..."
$GCLOUD gcloud config set project gen-lang-client-0338297768

# Step 3: Enable APIs
echo ""
echo "Step 3: Enabling required APIs..."
$GCLOUD gcloud services enable containerregistry.googleapis.com run.googleapis.com

# Step 4: Configure Docker
echo ""
echo "Step 4: Configuring Docker for GCR..."
$GCLOUD gcloud auth configure-docker --quiet

# Step 5: Build image
echo ""
echo "Step 5: Building Docker image..."
docker build -t gcr.io/gen-lang-client-0338297768/suiten-server:latest .

# Step 6: Push image
echo ""
echo "Step 6: Pushing to Container Registry..."
docker push gcr.io/gen-lang-client-0338297768/suiten-server:latest

# Step 7: Deploy
echo ""
echo "Step 7: Deploying to Cloud Run..."
$GCLOUD gcloud run deploy suiten-server \
  --image gcr.io/gen-lang-client-0338297768/suiten-server:latest \
  --platform managed \
  --region asia-northeast1 \
  --allow-unauthenticated \
  --set-env-vars "DATABASE_URL=postgresql://neondb_owner:npg_RsQepyAW86Nw@ep-tiny-glitter-a12bte2m-pooler.ap-southeast-1.aws.neon.tech/neondb?sslmode=require" \
  --set-env-vars "JWT_SECRET=suiten-jwt-secret-2024-production"

echo ""
echo "=== Deployment Complete ==="
$GCLOUD gcloud run services describe suiten-server --region asia-northeast1 --format="value(status.url)"
