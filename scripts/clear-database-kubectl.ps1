# Cle# Create a temporary pod with psql client
kubectl run postgres-client --rm -it --restart=Never --image=postgres:15 -- bash -c "
export PGPASSWORD='CHANGE_ME_IN_PRODUCTION'
psql -h YOUR_RDS_ENDPOINT -p 5432 -U postgres -d docs -c \"
TRUNCATE TABLE documents CASCADE;
SELECT 'Documents cleared successfully!' as status;
SELECT COUNT(*) as remaining_documents FROM documents;
SELECT COUNT(*) as remaining_users FROM users;
\""nt history using kubectl and a temporary PostgreSQL client pod
# This approach doesn't require installing PostgreSQL client locally

Write-Host "Creating temporary PostgreSQL client pod..." -ForegroundColor Yellow

# Create a temporary pod with psql client
kubectl run postgres-client --rm -it --restart=Never --image=postgres:15 -- bash -c "
export PGPASSWORD='CHANGE_ME_IN_PRODUCTION'
psql -h realtime-docs-postgres.cbakcuiowct4.us-west-1.rds.amazonaws.com -p 5432 -U postgres -d docs -c \"
TRUNCATE TABLE documents CASCADE;
SELECT 'Documents cleared successfully!' as status;
SELECT COUNT(*) as remaining_documents FROM documents;
SELECT COUNT(*) as remaining_users FROM users;
\"
"

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Document history cleared successfully!" -ForegroundColor Green
} else {
    Write-Host "❌ Failed to clear document history!" -ForegroundColor Red
}
