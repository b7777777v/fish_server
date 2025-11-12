# Complete database reset script for Windows (PowerShell)
# WARNING: This will DELETE all data and recreate the database from scratch

Write-Host "==========================================" -ForegroundColor Red
Write-Host "⚠️  DATABASE COMPLETE RESET" -ForegroundColor Red
Write-Host "==========================================" -ForegroundColor Red
Write-Host ""
Write-Host "This script will:" -ForegroundColor Yellow
Write-Host "1. Stop all connections to the database"
Write-Host "2. DROP the entire database"
Write-Host "3. CREATE a fresh database"
Write-Host "4. Run all migrations from scratch"
Write-Host ""
Write-Host "⚠️  WARNING: ALL DATA WILL BE LOST!" -ForegroundColor Red
Write-Host ""

$confirmation = Read-Host "Are you ABSOLUTELY SURE you want to continue? (type 'yes' to proceed)"

if ($confirmation -ne "yes") {
    Write-Host "Operation cancelled." -ForegroundColor Yellow
    exit 1
}

Write-Host ""
Write-Host "Starting database reset..." -ForegroundColor Green
Write-Host ""

# Database connection details
$DB_USER = "user"
$DB_NAME = "fish_db"

Write-Host "Step 1/4: Terminating existing connections..." -ForegroundColor Cyan
$terminateSQL = @"
SELECT pg_terminate_backend(pg_stat_activity.pid)
FROM pg_stat_activity
WHERE pg_stat_activity.datname = '$DB_NAME'
  AND pid <> pg_backend_pid();
"@

try {
    $terminateSQL | docker exec -i fish_server-postgres-1 psql -U $DB_USER -d postgres
    Write-Host "✓ Connections terminated" -ForegroundColor Green
} catch {
    Write-Host "⚠ Warning: Failed to terminate connections (may be none active)" -ForegroundColor Yellow
}
Write-Host ""

Write-Host "Step 2/4: Dropping database..." -ForegroundColor Cyan
try {
    docker exec fish_server-postgres-1 psql -U $DB_USER -d postgres -c "DROP DATABASE IF EXISTS $DB_NAME;"
    Write-Host "✓ Database dropped" -ForegroundColor Green
} catch {
    Write-Host "ERROR: Failed to drop database" -ForegroundColor Red
    exit 1
}
Write-Host ""

Write-Host "Step 3/4: Creating fresh database..." -ForegroundColor Cyan
try {
    docker exec fish_server-postgres-1 psql -U $DB_USER -d postgres -c "CREATE DATABASE $DB_NAME;"
    Write-Host "✓ Database created" -ForegroundColor Green
} catch {
    Write-Host "ERROR: Failed to create database" -ForegroundColor Red
    exit 1
}
Write-Host ""

Write-Host "Step 4/4: Running all migrations..." -ForegroundColor Cyan
try {
    & go run cmd\migrator\main.go up
    if ($LASTEXITCODE -ne 0) {
        throw "Migration failed"
    }
} catch {
    Write-Host ""
    Write-Host "ERROR: Migration failed!" -ForegroundColor Red
    Write-Host "Please check the error messages above." -ForegroundColor Yellow
    exit 1
}

Write-Host ""
Write-Host "==========================================" -ForegroundColor Green
Write-Host "✓ Database reset completed successfully!" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Checking migration status:" -ForegroundColor Cyan
& go run cmd\migrator\main.go version
Write-Host ""
