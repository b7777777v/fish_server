# Complete database reset script for Windows (PowerShell)
# WARNING: This will DELETE all data and recreate the database from scratch

Write-Host "==========================================" -ForegroundColor Red
Write-Host "⚠️  DATABASE COMPLETE RESET" -ForegroundColor Red
Write-Host "==========================================" -ForegroundColor Red
Write-Host ""
Write-Host "This script will:" -ForegroundColor Yellow
Write-Host "1. Ensure database is running"
Write-Host "2. Stop all connections to the database"
Write-Host "3. DROP the entire database"
Write-Host "4. CREATE a fresh database"
Write-Host "5. Run all migrations from scratch"
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
$CONTAINER_NAME = "fish_server-postgres-1"

# Check if Docker is running
try {
    $null = docker info 2>$null
    if ($LASTEXITCODE -ne 0) {
        throw "Docker not running"
    }
} catch {
    Write-Host "ERROR: Docker is not running!" -ForegroundColor Red
    Write-Host "Please start Docker Desktop and try again." -ForegroundColor Yellow
    exit 1
}

Write-Host "Checking database container status..." -ForegroundColor Cyan

# Check if container exists
$containerExists = docker ps -a --format "{{.Names}}" | Select-String -Pattern "^$CONTAINER_NAME$" -Quiet
if (-not $containerExists) {
    Write-Host "ERROR: Database container '$CONTAINER_NAME' does not exist!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please start the database first using:" -ForegroundColor Yellow
    Write-Host "  scripts\start-database.bat" -ForegroundColor White
    Write-Host ""
    Write-Host "Or use Docker Compose:" -ForegroundColor Yellow
    Write-Host "  docker-compose -f deployments\docker-compose.dev.yml up -d postgres" -ForegroundColor White
    exit 1
}

# Check if container is running
$containerRunning = docker ps --format "{{.Names}}" | Select-String -Pattern "^$CONTAINER_NAME$" -Quiet
if (-not $containerRunning) {
    Write-Host "Database container exists but is not running. Starting it..." -ForegroundColor Yellow
    docker start $CONTAINER_NAME | Out-Null
    Write-Host "Waiting for database to be ready..." -ForegroundColor Yellow
    Start-Sleep -Seconds 10
} else {
    Write-Host "Database container is running." -ForegroundColor Green
}

# Wait for PostgreSQL to be ready
Write-Host "Verifying database is ready..." -ForegroundColor Cyan
$maxAttempts = 30
$attempt = 0
$ready = $false

while ($attempt -lt $maxAttempts) {
    try {
        $null = docker exec $CONTAINER_NAME pg_isready -U $DB_USER 2>$null
        if ($LASTEXITCODE -eq 0) {
            $ready = $true
            break
        }
    } catch {
        # Continue waiting
    }

    Write-Host "Waiting for PostgreSQL to be ready... (attempt $($attempt + 1)/$maxAttempts)" -ForegroundColor Yellow
    Start-Sleep -Seconds 2
    $attempt++
}

if (-not $ready) {
    Write-Host "ERROR: PostgreSQL did not become ready in time!" -ForegroundColor Red
    exit 1
}

Write-Host "Database is ready!" -ForegroundColor Green
Write-Host ""

Write-Host "Step 1/4: Terminating existing connections..." -ForegroundColor Cyan
$terminateSQL = @"
SELECT pg_terminate_backend(pg_stat_activity.pid)
FROM pg_stat_activity
WHERE pg_stat_activity.datname = '$DB_NAME'
  AND pid <> pg_backend_pid();
"@

try {
    $terminateSQL | docker exec -i $CONTAINER_NAME psql -U $DB_USER -d postgres 2>$null | Out-Null
    Write-Host "✓ Successfully terminated connections" -ForegroundColor Green
} catch {
    Write-Host "⚠ Warning: Failed to terminate connections (may be none active)" -ForegroundColor Yellow
}
Write-Host ""

Write-Host "Step 2/4: Dropping database..." -ForegroundColor Cyan
try {
    docker exec $CONTAINER_NAME psql -U $DB_USER -d postgres -c "DROP DATABASE IF EXISTS $DB_NAME;" | Out-Null
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to drop database"
    }
    Write-Host "✓ Successfully dropped database" -ForegroundColor Green
} catch {
    Write-Host "ERROR: Failed to drop database!" -ForegroundColor Red
    exit 1
}
Write-Host ""

Write-Host "Step 3/4: Creating fresh database..." -ForegroundColor Cyan
try {
    docker exec $CONTAINER_NAME psql -U $DB_USER -d postgres -c "CREATE DATABASE $DB_NAME;" | Out-Null
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to create database"
    }
    Write-Host "✓ Successfully created database" -ForegroundColor Green
} catch {
    Write-Host "ERROR: Failed to create database!" -ForegroundColor Red
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
