@echo off
REM Windows PowerShell script to deploy infrastructure
REM Usage: deploy-infrastructure.bat [environment] [aws-region]

setlocal EnableDelayedExpansion

set ENVIRONMENT=%1
if "%ENVIRONMENT%"=="" set ENVIRONMENT=dev

set AWS_REGION=%2
if "%AWS_REGION%"=="" set AWS_REGION=us-west-1

set PROJECT_NAME=realtime-docs

echo Deploying infrastructure for %PROJECT_NAME% in %ENVIRONMENT% environment...

REM Check if Terraform is installed
terraform version >nul 2>&1
if errorlevel 1 (
    echo ERROR: Terraform is not installed. Please install it first.
    exit /b 1
)

REM Check if AWS CLI is installed
aws --version >nul 2>&1
if errorlevel 1 (
    echo ERROR: AWS CLI is not installed. Please install it first.
    exit /b 1
)

REM Navigate to terraform directory
cd terraform

REM Initialize Terraform
echo Initializing Terraform...
terraform init

REM Create terraform.tfvars if it doesn't exist
if not exist "terraform.tfvars" (
    echo Creating terraform.tfvars file...
    (
        echo aws_region = "%AWS_REGION%"
        echo project_name = "%PROJECT_NAME%"
        echo environment = "%ENVIRONMENT%"
        echo.
        echo # Update these values as needed
        echo db_password = "change-me-in-production"
        echo domain_name = "your-domain.com"
        echo certificate_arn = ""
        echo.
        echo # Node configuration
        echo node_instance_type = "t3.medium"
        echo node_desired_capacity = 2
        echo node_max_capacity = 4
        echo node_min_capacity = 1
        echo.
        echo # Database configuration
        echo db_instance_class = "db.t3.micro"
        echo db_allocated_storage = 20
        echo.
        echo # Redis configuration
        echo redis_node_type = "cache.t3.micro"
    ) > terraform.tfvars
    echo WARNING: Please update terraform.tfvars with your actual values before proceeding.
    echo Especially: domain_name, certificate_arn, and other production settings.
    pause
)

REM Plan the deployment
echo Planning Terraform deployment...
terraform plan -var-file="terraform.tfvars"

REM Ask for confirmation
set /p REPLY="Do you want to apply these changes? (y/N): "
if /i "%REPLY%"=="y" (
    echo Applying Terraform configuration...
    terraform apply -var-file="terraform.tfvars" -auto-approve
    
    echo SUCCESS: Infrastructure deployment completed
    echo Getting cluster information...
    
    REM Update kubeconfig
    for /f "tokens=*" %%i in ('terraform output -raw cluster_id') do set CLUSTER_NAME=%%i
    aws eks update-kubeconfig --region %AWS_REGION% --name !CLUSTER_NAME!
    
    echo EKS cluster ready: !CLUSTER_NAME!
    echo Next steps:
    echo   1. Run deploy-app.bat to deploy the application
    echo   2. Configure your domain DNS to point to the load balancer
    echo   3. Update secrets and configurations as needed
) else (
    echo Deployment cancelled.
)

pause
