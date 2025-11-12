@echo off
REM Start PostgreSQL and Redis using Docker Compose for Windows
REM Usage: scripts\start-database.bat

echo ===================================
echo Starting Database Services (Windows)
echo ===================================
echo.

REM Check if Docker is installed
where docker >nul 2>nul
if errorlevel 1 (
    echo ERROR: Docker is not installed or not in PATH
    echo Please install Docker Desktop for Windows from:
    echo https://www.docker.com/products/docker-desktop
    exit /b 1
)

REM Check if Docker is running
docker info >nul 2>nul
if errorlevel 1 (
    echo ERROR: Docker is not running
    echo Please start Docker Desktop and try again
    exit /b 1
)

echo Docker is available and running
echo.

REM Navigate to deployments directory and start services
echo Starting PostgreSQL and Redis...
docker-compose -f deployments\docker-compose.dev.yml up -d postgres redis

if errorlevel 1 (
    echo.
    echo ERROR: Failed to start database services
    echo Please check Docker logs for details
    exit /b 1
)

echo.
echo ===================================
echo ✓ Database services started successfully!
echo ===================================
echo.
echo Services running:
echo - PostgreSQL on localhost:5432
echo   Database: fish_db
echo   User: user
echo   Password: password
echo.
echo - Redis on localhost:6379
echo.
echo To check status: docker-compose -f deployments\docker-compose.dev.yml ps
echo To view logs: docker-compose -f deployments\docker-compose.dev.yml logs -f
echo To stop: docker-compose -f deployments\docker-compose.dev.yml down
echo.

REM Wait for PostgreSQL to be ready
echo Waiting for PostgreSQL to be ready...
timeout /t 5 /nobreak >nul

:check_loop
docker exec fish_server-postgres-1 pg_isready -U user -d fish_db >nul 2>nul
if errorlevel 1 (
    echo PostgreSQL is still starting...
    timeout /t 2 /nobreak >nul
    goto check_loop
)

echo.
echo ✓ PostgreSQL is ready to accept connections!
echo.
echo You can now run migrations:
echo   go run cmd\migrator\main.go up
echo.
