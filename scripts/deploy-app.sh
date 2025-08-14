#!/bin/bash

# Deploy application to Kubernetes
# Usage: ./deploy-app.sh [environment] [image-tag]

set -e

ENVIRONMENT=${1:-dev}
IMAGE_TAG=${2:-latest}
PROJECT_NAME="realtime-docs"

echo "Deploying $PROJECT_NAME application..."

# Check if kubectl is installed and configured
if ! command -v kubectl &> /dev/null; then
    echo "ERROR: kubectl is not installed. Please install it first."
    exit 1
fi

# Check if helm is installed
if ! command -v helm &> /dev/null; then
    echo "ERROR: Helm is not installed. Please install it first."
    exit 1
fi

# Check cluster connectivity
if ! kubectl cluster-info &> /dev/null; then
    echo "ERROR: Cannot connect to Kubernetes cluster. Please check your kubeconfig."
    exit 1
fi

echo "Current cluster context:"
kubectl config current-context

# Install AWS Load Balancer Controller if not exists
echo "Checking AWS Load Balancer Controller..."
if ! kubectl get deployment -n kube-system aws-load-balancer-controller &> /dev/null; then
    echo "Installing AWS Load Balancer Controller..."
    
    # Create IAM role for AWS Load Balancer Controller
    CLUSTER_NAME=$(kubectl config current-context | cut -d'/' -f2)
    AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
    
    # Download and apply the controller
    curl -o iam_policy.json https://raw.githubusercontent.com/kubernetes-sigs/aws-load-balancer-controller/v2.5.4/docs/install/iam_policy.json
    
    aws iam create-policy \
        --policy-name AWSLoadBalancerControllerIAMPolicy \
        --policy-document file://iam_policy.json || true
    
    eksctl create iamserviceaccount \
        --cluster=$CLUSTER_NAME \
        --namespace=kube-system \
        --name=aws-load-balancer-controller \
        --role-name AmazonEKSLoadBalancerControllerRole \
        --attach-policy-arn=arn:aws:iam::$AWS_ACCOUNT_ID:policy/AWSLoadBalancerControllerIAMPolicy \
        --approve || true
    
    helm repo add eks https://aws.github.io/eks-charts
    helm repo update
    
    helm install aws-load-balancer-controller eks/aws-load-balancer-controller \
        -n kube-system \
        --set clusterName=$CLUSTER_NAME \
        --set serviceAccount.create=false \
        --set serviceAccount.name=aws-load-balancer-controller || true
fi

# Create namespace if it doesn't exist
kubectl create namespace $PROJECT_NAME --dry-run=client -o yaml | kubectl apply -f -

# Deploy using Helm
echo "Deploying application with Helm..."

# Check if we have custom values file
VALUES_FILE="../helm/$PROJECT_NAME/values.yaml"
if [ -f "values-$ENVIRONMENT.yaml" ]; then
    VALUES_FILE="values-$ENVIRONMENT.yaml"
    echo "Using environment-specific values: $VALUES_FILE"
fi

helm upgrade --install $PROJECT_NAME ../helm/$PROJECT_NAME \
    --namespace $PROJECT_NAME \
    --values $VALUES_FILE \
    --set image.backend.tag=$IMAGE_TAG \
    --set image.frontend.tag=$IMAGE_TAG \
    --wait \
    --timeout=600s

echo "SUCCESS: Application deployment completed"

# Show deployment status
echo "Deployment status:"
kubectl get pods -n $PROJECT_NAME
kubectl get services -n $PROJECT_NAME
kubectl get ingress -n $PROJECT_NAME

# Get load balancer URL
echo "Getting load balancer information..."
sleep 10
LOAD_BALANCER=$(kubectl get ingress -n $PROJECT_NAME -o jsonpath='{.items[0].status.loadBalancer.ingress[0].hostname}')

if [ ! -z "$LOAD_BALANCER" ]; then
    echo "Application URL: http://$LOAD_BALANCER"
    echo "NOTE: It may take a few minutes for the load balancer to be fully ready."
else
    echo "Load balancer is still being created. Check again in a few minutes."
fi

echo "Useful commands:"
echo "  kubectl get pods -n $PROJECT_NAME"
echo "  kubectl logs -f deployment/backend -n $PROJECT_NAME"
echo "  kubectl logs -f deployment/frontend -n $PROJECT_NAME"
