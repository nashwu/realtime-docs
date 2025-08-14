# Clear document history from PostgreSQL database
# This script connects to the RDS PostgreSQL instance and clears all documents

$DB_HOST = "YOUR_RDS_ENDPOINT"
$DB_PORT = "5432"
$DB_NAME = "docs"
$DB_USER = "postgres"
$DB_PASSWORD = "CHANGE_ME_IN_PRODUCTION"  # Update this with your actual password

Write-Host "Connecting to PostgreSQL database..." -ForegroundColor Yellow
Write-Host "Host: $DB_HOST" -ForegroundColor Gray
Write-Host "Database: $DB_NAME" -ForegroundColor Gray

# Check if psql is available
if (!(Get-Command psql -ErrorAction SilentlyContinue)) {
    Write-Host "ERROR: psql client not found!" -ForegroundColor Red
    Write-Host "Please install PostgreSQL client tools:" -ForegroundColor Yellow
    Write-Host "1. Download from: https://www.postgresql.org/download/windows/" -ForegroundColor Yellow
    Write-Host "2. Or use Docker: docker run --rm -it postgres:15 psql" -ForegroundColor Yellow
    exit 1
}

# Set PGPASSWORD environment variable
$env:PGPASSWORD = $DB_PASSWORD

Write-Host "Executing clear-documents.sql..." -ForegroundColor Yellow

# Execute the SQL script
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "clear-documents.sql"

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Document history cleared successfully!" -ForegroundColor Green
    Write-Host "All documents have been removed from the database." -ForegroundColor Green
} else {
    Write-Host "❌ Failed to clear document history!" -ForegroundColor Red
    Write-Host "Please check the connection details and try again." -ForegroundColor Red
}

# Clear the password from environment
$env:PGPASSWORD = $null
