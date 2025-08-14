#!/bin/bash

# Build and push Docker images to ECR
# Usage: ./build-and-push.sh [aws-region] [image-tag]

set -e

AWS_REGION=${1:-us-west-1}
IMAGE_TAG=${2:-latest}
PROJECT_NAME="realtime-docs"

echo "Building and pushing Docker images..."

# Check if AWS CLI is configured
if ! aws sts get-caller-identity &> /dev/null; then
    echo "ERROR: AWS CLI is not configured. Please configure it first."
    exit 1
fi

# Get AWS account ID
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
ECR_REGISTRY="$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com"

echo "Using ECR registry: $ECR_REGISTRY"

# Create ECR repositories if they don't exist
echo "Creating ECR repositories..."
aws ecr create-repository --repository-name "$PROJECT_NAME-backend" --region $AWS_REGION || true
aws ecr create-repository --repository-name "$PROJECT_NAME-frontend" --region $AWS_REGION || true

# Login to ECR
echo "Logging into ECR..."
aws ecr get-login-password --region $AWS_REGION | docker login --username AWS --password-stdin $ECR_REGISTRY

# Build and push backend
echo "Building backend image..."
cd ../backend
docker build -t "$ECR_REGISTRY/$PROJECT_NAME-backend:$IMAGE_TAG" .
docker tag "$ECR_REGISTRY/$PROJECT_NAME-backend:$IMAGE_TAG" "$ECR_REGISTRY/$PROJECT_NAME-backend:latest"

echo "Pushing backend image..."
docker push "$ECR_REGISTRY/$PROJECT_NAME-backend:$IMAGE_TAG"
docker push "$ECR_REGISTRY/$PROJECT_NAME-backend:latest"

# Build and push frontend
echo "Building frontend image..."
cd ../frontend
docker build -t "$ECR_REGISTRY/$PROJECT_NAME-frontend:$IMAGE_TAG" .
docker tag "$ECR_REGISTRY/$PROJECT_NAME-frontend:$IMAGE_TAG" "$ECR_REGISTRY/$PROJECT_NAME-frontend:latest"

echo "Pushing frontend image..."
docker push "$ECR_REGISTRY/$PROJECT_NAME-frontend:$IMAGE_TAG"
docker push "$ECR_REGISTRY/$PROJECT_NAME-frontend:latest"

echo "SUCCESS: Images built and pushed successfully"
echo "Backend image: $ECR_REGISTRY/$PROJECT_NAME-backend:$IMAGE_TAG"
echo "Frontend image: $ECR_REGISTRY/$PROJECT_NAME-frontend:$IMAGE_TAG"

# Update Helm values with new image repositories
cd ../scripts
echo "Updating Helm values with ECR repositories..."
sed -i.bak "s|your-registry/realtime-docs-backend|$ECR_REGISTRY/$PROJECT_NAME-backend|g" ../helm/$PROJECT_NAME/values.yaml
sed -i.bak "s|your-registry/realtime-docs-frontend|$ECR_REGISTRY/$PROJECT_NAME-frontend|g" ../helm/$PROJECT_NAME/values.yaml

echo "Next steps:"
echo "  1. Run ./deploy-app.sh to deploy the application"
echo "  2. Or use the CI/CD pipeline for automated deployment"
