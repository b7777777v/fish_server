@echo off
REM Stop PostgreSQL and Redis using Docker Compose for Windows
REM Usage: scripts\stop-database.bat

echo ===================================
echo Stopping Database Services (Windows)
echo ===================================
echo.

docker-compose -f deployments\docker-compose.dev.yml down

if errorlevel 1 (
    echo.
    echo ERROR: Failed to stop database services
    exit /b 1
)

echo.
echo ✓ Database services stopped successfully!
echo.
