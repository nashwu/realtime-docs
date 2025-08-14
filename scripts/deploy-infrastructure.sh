#!/bin/bash

# Deploy infrastructure with Terraform
# Usage: ./deploy-infrastructure.sh [environment] [aws-region]

set -e

ENVIRONMENT=${1:-dev}
AWS_REGION=${2:-us-west-1}
PROJECT_NAME="realtime-docs"

echo "Deploying infrastructure for $PROJECT_NAME in $ENVIRONMENT environment..."

# Check if Terraform is installed
if ! command -v terraform &> /dev/null; then
    echo "ERROR: Terraform is not installed. Please install it first."
    exit 1
fi

# Check if AWS CLI is installed and configured
if ! command -v aws &> /dev/null; then
    echo "ERROR: AWS CLI is not installed. Please install it first."
    exit 1
fi

# Navigate to terraform directory
cd terraform

# Initialize Terraform
echo "Initializing Terraform..."
terraform init

# Create terraform.tfvars if it doesn't exist
if [ ! -f "terraform.tfvars" ]; then
    echo "Creating terraform.tfvars file..."
    cat > terraform.tfvars << EOF
aws_region = "$AWS_REGION"
project_name = "$PROJECT_NAME"
environment = "$ENVIRONMENT"

# Update these values as needed
db_password = "$(openssl rand -base64 32)"
domain_name = "your-domain.com"
certificate_arn = ""

# Node configuration
node_instance_type = "t3.medium"
node_desired_capacity = 2
node_max_capacity = 4
node_min_capacity = 1

# Database configuration
db_instance_class = "db.t3.micro"
db_allocated_storage = 20

# Redis configuration
redis_node_type = "cache.t3.micro"
EOF
    echo "WARNING: Please update terraform.tfvars with your actual values before proceeding."
    echo "Especially: domain_name, certificate_arn, and other production settings."
    read -p "Press Enter to continue or Ctrl+C to exit..."
fi

# Plan the deployment
echo "Planning Terraform deployment..."
terraform plan -var-file="terraform.tfvars"

# Ask for confirmation
read -p "Do you want to apply these changes? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Applying Terraform configuration..."
    terraform apply -var-file="terraform.tfvars" -auto-approve
    
    echo "SUCCESS: Infrastructure deployment completed"
    echo "Getting cluster information..."
    
    # Update kubeconfig
    CLUSTER_NAME=$(terraform output -raw cluster_id)
    aws eks update-kubeconfig --region $AWS_REGION --name $CLUSTER_NAME
    
    echo "EKS cluster ready: $CLUSTER_NAME"
    echo "Next steps:"
    echo "  1. Run ./deploy-app.sh to deploy the application"
    echo "  2. Configure your domain DNS to point to the load balancer"
    echo "  3. Update secrets and configurations as needed"
else
    echo "Deployment cancelled."
fi
