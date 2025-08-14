@echo off
REM Windows PowerShell script to deploy application
REM Usage: deploy-app.bat [environment] [image-tag]

setlocal EnableDelayedExpansion

set ENVIRONMENT=%1
if "%ENVIRONMENT%"=="" set ENVIRONMENT=dev

set IMAGE_TAG=%2
if "%IMAGE_TAG%"=="" set IMAGE_TAG=latest

set PROJECT_NAME=realtime-docs

echo Deploying %PROJECT_NAME% application...

REM Check if kubectl is installed
kubectl version --client >nul 2>&1
if errorlevel 1 (
    echo ERROR: kubectl is not installed. Please install it first.
    exit /b 1
)

REM Check if helm is installed
helm version >nul 2>&1
if errorlevel 1 (
    echo ERROR: Helm is not installed. Please install it first.
    exit /b 1
)

REM Check cluster connectivity
kubectl cluster-info >nul 2>&1
if errorlevel 1 (
    echo ERROR: Cannot connect to Kubernetes cluster. Please check your kubeconfig.
    exit /b 1
)

echo Current cluster context:
kubectl config current-context

REM Create namespace if it doesn't exist
kubectl create namespace %PROJECT_NAME% --dry-run=client -o yaml | kubectl apply -f -

REM Deploy using Helm
echo Deploying application with Helm...

REM Check if we have custom values file
set VALUES_FILE=..\helm\%PROJECT_NAME%\values.yaml
if exist "values-%ENVIRONMENT%.yaml" (
    set VALUES_FILE=values-%ENVIRONMENT%.yaml
    echo Using environment-specific values: !VALUES_FILE!
)

helm upgrade --install %PROJECT_NAME% ..\helm\%PROJECT_NAME% ^
    --namespace %PROJECT_NAME% ^
    --values !VALUES_FILE! ^
    --set image.backend.tag=%IMAGE_TAG% ^
    --set image.frontend.tag=%IMAGE_TAG% ^
    --wait ^
    --timeout=600s

echo SUCCESS: Application deployment completed

REM Show deployment status
echo Deployment status:
kubectl get pods -n %PROJECT_NAME%
kubectl get services -n %PROJECT_NAME%
kubectl get ingress -n %PROJECT_NAME%

REM Get load balancer URL
echo Getting load balancer information...
timeout /t 10 /nobreak >nul
for /f "tokens=*" %%i in ('kubectl get ingress -n %PROJECT_NAME% -o jsonpath^="{.items[0].status.loadBalancer.ingress[0].hostname}"') do set LOAD_BALANCER=%%i

if not "!LOAD_BALANCER!"=="" (
    echo Application URL: http://!LOAD_BALANCER!
    echo NOTE: It may take a few minutes for the load balancer to be fully ready.
) else (
    echo Load balancer is still being created. Check again in a few minutes.
)

echo Useful commands:
echo    kubectl get pods -n %PROJECT_NAME%
echo    kubectl logs -f deployment/backend -n %PROJECT_NAME%
echo    kubectl logs -f deployment/frontend -n %PROJECT_NAME%

pause
