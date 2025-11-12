# Start PostgreSQL and Redis using Docker Compose for Windows (PowerShell)
# Usage: .\scripts\start-database.ps1

Write-Host "===================================" -ForegroundColor Cyan
Write-Host "Starting Database Services (PowerShell)" -ForegroundColor Cyan
Write-Host "===================================" -ForegroundColor Cyan
Write-Host ""

# Check if Docker is installed
try {
    $null = Get-Command docker -ErrorAction Stop
} catch {
    Write-Host "ERROR: Docker is not installed or not in PATH" -ForegroundColor Red
    Write-Host "Please install Docker Desktop for Windows from:" -ForegroundColor Yellow
    Write-Host "https://www.docker.com/products/docker-desktop" -ForegroundColor Yellow
    exit 1
}

# Check if Docker is running
try {
    $null = docker info 2>$null
    if ($LASTEXITCODE -ne 0) {
        throw "Docker not running"
    }
} catch {
    Write-Host "ERROR: Docker is not running" -ForegroundColor Red
    Write-Host "Please start Docker Desktop and try again" -ForegroundColor Yellow
    exit 1
}

Write-Host "✓ Docker is available and running" -ForegroundColor Green
Write-Host ""

# Start services
Write-Host "Starting PostgreSQL and Redis..." -ForegroundColor Green
try {
    docker-compose -f deployments\docker-compose.dev.yml up -d postgres redis
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to start services"
    }
} catch {
    Write-Host ""
    Write-Host "ERROR: Failed to start database services" -ForegroundColor Red
    Write-Host "Please check Docker logs for details" -ForegroundColor Yellow
    exit 1
}

Write-Host ""
Write-Host "===================================" -ForegroundColor Green
Write-Host "✓ Database services started successfully!" -ForegroundColor Green
Write-Host "===================================" -ForegroundColor Green
Write-Host ""
Write-Host "Services running:" -ForegroundColor Cyan
Write-Host "- PostgreSQL on localhost:5432"
Write-Host "  Database: fish_db"
Write-Host "  User: user"
Write-Host "  Password: password"
Write-Host ""
Write-Host "- Redis on localhost:6379"
Write-Host ""
Write-Host "Commands:" -ForegroundColor Yellow
Write-Host "  Check status: docker-compose -f deployments\docker-compose.dev.yml ps"
Write-Host "  View logs: docker-compose -f deployments\docker-compose.dev.yml logs -f"
Write-Host "  Stop: docker-compose -f deployments\docker-compose.dev.yml down"
Write-Host ""

# Wait for PostgreSQL to be ready
Write-Host "Waiting for PostgreSQL to be ready..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

$maxAttempts = 15
$attempt = 0
$ready = $false

while ($attempt -lt $maxAttempts) {
    try {
        $null = docker exec fish_server-postgres-1 pg_isready -U user -d fish_db 2>$null
        if ($LASTEXITCODE -eq 0) {
            $ready = $true
            break
        }
    } catch {
        # Continue waiting
    }

    Write-Host "PostgreSQL is still starting... (attempt $($attempt + 1)/$maxAttempts)" -ForegroundColor Yellow
    Start-Sleep -Seconds 2
    $attempt++
}

Write-Host ""
if ($ready) {
    Write-Host "✓ PostgreSQL is ready to accept connections!" -ForegroundColor Green
    Write-Host ""
    Write-Host "You can now run migrations:" -ForegroundColor Cyan
    Write-Host "  go run cmd\migrator\main.go up" -ForegroundColor White
} else {
    Write-Host "⚠ PostgreSQL may still be starting. Please wait a moment." -ForegroundColor Yellow
    Write-Host "Check status with: docker-compose -f deployments\docker-compose.dev.yml ps" -ForegroundColor Yellow
}
Write-Host ""
