# Real-time Collaborative Document Editor

A **production-ready, enterprise-grade** collaborative document editing platform that enables multiple users to edit documents simultaneously with real-time synchronization. Built with modern cloud-native technologies and deployed on AWS infrastructure for high availability and scalability.

## Live Demo

![Real-time Collaborative Editing](./assets/demo.gif)

## Key Features

**Real-time Collaboration:** Multi-user editing with operational transformation, visual cursors, and document isolation  
**Enterprise Architecture:** Auto-scaling Kubernetes on AWS with Multi-AZ deployment and load balancing  
**Security & Performance:** JWT auth, VPC isolation, Redis caching, and connection pooling  
**Developer Experience:** Complete CI/CD pipeline with Infrastructure as Code and Docker containerization

---

## Technology Stack

**Frontend:** Angular 17 + TypeScript + WebSocket client for real-time collaboration  
**Backend:** Go REST API + Gorilla WebSocket + JWT middleware + structured logging  
**Database:** PostgreSQL 15.13 (RDS Multi-AZ) + Redis (ElastiCache) for caching  
**Infrastructure:** EKS + ALB + ECR + VPC with NAT Gateway  
**DevOps:** Terraform + Helm + GitHub Actions + Docker multi-stage builds

---

## Skills

**Backend:** Go concurrency, WebSocket protocols, RESTful APIs, database optimization, caching strategies  
**Frontend:** Angular/TypeScript, real-time UI/UX, state management, performance optimization  
**Cloud:** AWS solutions architecture, Kubernetes operations, Infrastructure as Code, containerization  
**DevOps:** CI/CD pipelines, security implementation, monitoring, auto-scaling configuration

---

## Project Structure

```
realtime-docs/
├── Infrastructure & Deployment
│   ├── terraform/              # Infrastructure as Code
│   │   ├── main.tf            # AWS provider and core resources
│   │   ├── eks.tf             # Kubernetes cluster configuration
│   │   ├── database.tf        # RDS PostgreSQL and ElastiCache Redis
│   │   └── variables.tf       # Configurable input parameters
│   ├── helm/realtime-docs/    # Kubernetes application manifests
│   │   ├── templates/         # Deployment, service, and ingress resources
│   │   └── values.yaml        # Application configuration values
│   └── scripts/               # Deployment automation scripts
│       ├── deploy-infrastructure.sh  # Terraform deployment
│       └── deploy-app.sh             # Application deployment
│
├── Application Code
│   ├── backend/               # Go REST API and WebSocket server
│   │   ├── cmd/server/        # Application entry point and configuration
│   │   ├── internal/          # Business logic and domain models
│   │   │   ├── http/         # HTTP handlers, middleware, and routing
│   │   │   ├── ws/           # WebSocket hub and client connection management
│   │   │   └── store/        # Database layer, models, and migrations
│   │   └── pkg/              # Shared utilities and packages
│   │       ├── auth/         # JWT token management
│   │       ├── metrics/      # Application metrics and monitoring
│   │       └── ratelimit/    # API rate limiting
│   │
│   └── frontend/             # Angular single-page application
│       ├── src/app/          # Angular components, services, and modules
│       │   ├── auth.service.ts       # Authentication service
│       │   ├── docs.service.ts       # Document management API client
│       │   └── ws-sync.service.ts    # WebSocket synchronization service
│       ├── Dockerfile        # Multi-stage build for production
│       └── nginx.conf        # Production web server configuration
│
└── CI/CD & Automation
    ├── .github/workflows/     # GitHub Actions pipeline
    │   └── deploy.yml        # Automated build, test, and deployment
    └── k8s/                  # Additional Kubernetes resources
```

## Quick Start Deployment Guide

### Prerequisites

Required Technologies:

- **AWS CLI** configured with appropriate permissions
- **Terraform** (>= 1.0) installed
- **kubectl** configured for Kubernetes management
- **Helm** (>= 3.0) for Kubernetes package management
- **Docker** for building container images
- **Git** for cloning the repository

### Step 1: Clone and Configure

```bash
# Clone the repository
git clone https://github.com/your-username/realtime-docs.git
cd realtime-docs

# Copy example configuration files
cp scripts/terraform.tfvars.example scripts/terraform.tfvars
cp helm/realtime-docs/values.yaml.example helm/realtime-docs/values.yaml
```

### Step 2: Configure AWS Credentials

```bash
# Configure AWS CLI (if not already done)
aws configure

# Verify your AWS identity
aws sts get-caller-identity
```

### Step 3: Set Up Configuration

Edit `scripts/terraform.tfvars` with your specific values:

```hcl
# Required: Replace with your values
region = "us-west-1"  # Your preferred AWS region
project_name = "realtime-docs"
db_password = "your-password"

# Optional: Customize as needed
node_instance_type = "t3.medium"
min_node_count = 2
max_node_count = 10
```

### Step 4: Deploy Infrastructure

```bash
# Navigate to terraform directory
cd terraform

# Initialize Terraform
terraform init

# Plan the deployment (review changes)
terraform plan -var-file="../scripts/terraform.tfvars"

# Deploy infrastructure (takes ~15-20 minutes)
terraform apply -var-file="../scripts/terraform.tfvars" -auto-approve

# Save important outputs
terraform output
```

### Step 5: Configure Kubernetes

```bash
# Return to project root
cd ..

# Configure kubectl for your new EKS cluster
aws eks update-kubeconfig --region us-west-1 --name realtime-docs-cluster

# Verify cluster access
kubectl get nodes

# Install AWS Load Balancer Controller
kubectl apply -f k8s/aws-load-balancer-controller-service-account.yaml
helm repo add eks https://aws.github.io/eks-charts
helm install aws-load-balancer-controller eks/aws-load-balancer-controller \
  -n kube-system \
  --set clusterName=realtime-docs-cluster \
  --set serviceAccount.create=false \
  --set serviceAccount.name=aws-load-balancer-controller
```

### Step 6: Create Application Secrets

```bash
# Create Kubernetes secret with your configuration
kubectl create secret generic realtime-docs-config \
  --from-literal=jwt-secret="$(openssl rand -base64 32)" \
  --from-literal=database-url="postgres://postgres:your-secure-database-password@YOUR_RDS_ENDPOINT:5432/docs?sslmode=require" \
  --from-literal=redis-addr="YOUR_REDIS_ENDPOINT:6379"

# Replace YOUR_RDS_ENDPOINT and YOUR_REDIS_ENDPOINT with actual values from terraform output
```

### Step 7: Deploy Application

```bash
# Update Helm values with your specific configuration
# Edit helm/realtime-docs/values.yaml and update:
# - AWS account ID
# - ECR repository URLs
# - Subnet IDs (from terraform output)

# Deploy the application
helm install realtime-docs ./helm/realtime-docs

# Wait for deployment to complete
kubectl rollout status deployment/realtime-docs-frontend
kubectl rollout status deployment/realtime-docs-backend
```

### Step 8: Access Your Application

```bash
# Get the load balancer URL
kubectl get ingress realtime-docs-ingress

# Wait for DNS propagation (2-3 minutes), then visit:
# http://YOUR-LOAD-BALANCER-DNS-NAME
```

## Advanced Configuration

### Custom Domain Setup

To use your own domain:

1. **Update Helm values** in `helm/realtime-docs/values.yaml`:
   ```yaml
   ingress:
     host: "docs.yourdomain.com"
     annotations:
       alb.ingress.kubernetes.io/certificate-arn: "arn:aws:acm:..."
   ```

2. **Deploy the changes**:
   ```bash
   helm upgrade realtime-docs ./helm/realtime-docs
   ```

3. **Update DNS** to point your domain to the load balancer

### Scaling Configuration

Adjust auto-scaling parameters in `helm/realtime-docs/values.yaml`:

```yaml
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80
```

### Monitoring Setup

Enable additional monitoring:

```bash
# Install Prometheus and Grafana (optional)
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack
```

## CI/CD Pipeline Setup

### GitHub Actions Configuration

1. **Fork this repository** to your GitHub account

2. **Set up GitHub Secrets** in your repository settings:
   ```
   AWS_ACCESS_KEY_ID=your-access-key
   AWS_SECRET_ACCESS_KEY=your-secret-key
   AWS_REGION=us-west-1
   JWT_SECRET=your-jwt-secret
   DATABASE_URL=your-database-url
   ```

3. **Push changes** to trigger the automated deployment:
   ```bash
   git add .
   git commit -m "Initial deployment configuration"
   git push origin main
   ```

The pipeline will automatically:
- Build and test the application
- Create Docker images and push to ECR
- Deploy to your Kubernetes cluster
- Run health checks

## Cleanup

To remove all resources and avoid AWS charges:

```bash
# Delete Kubernetes resources
helm uninstall realtime-docs
kubectl delete secret realtime-docs-config

# Delete AWS infrastructure
cd terraform
terraform destroy -var-file="../scripts/terraform.tfvars" -auto-approve
```