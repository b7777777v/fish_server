@echo off
REM Complete database reset script for Windows
REM WARNING: This will DELETE all data and recreate the database from scratch

setlocal enabledelayedexpansion

echo ==========================================
echo WARNING: DATABASE COMPLETE RESET
echo ==========================================
echo.
echo This script will:
echo 1. Ensure database is running
echo 2. Stop all connections to the database
echo 3. DROP the entire database
echo 4. CREATE a fresh database
echo 5. Run all migrations from scratch
echo.
echo WARNING: ALL DATA WILL BE LOST!
echo.

set /p confirmation="Are you ABSOLUTELY SURE you want to continue? (type 'yes' to proceed): "

if /i not "!confirmation!"=="yes" (
    echo Operation cancelled.
    exit /b 1
)

echo.
echo Starting database reset...
echo.

REM Database connection details
set DB_HOST=localhost
set DB_PORT=5432
set DB_USER=user
set DB_NAME=fish_db
set DB_PASSWORD=password
set PGPASSWORD=%DB_PASSWORD%
set CONTAINER_NAME=fish_server-postgres-1

REM Check if Docker is running
docker info >nul 2>nul
if errorlevel 1 (
    echo ERROR: Docker is not running!
    echo Please start Docker Desktop and try again.
    exit /b 1
)

echo Checking database container status...
docker ps -a | findstr /C:"%CONTAINER_NAME%" >nul
if errorlevel 1 (
    echo ERROR: Database container '%CONTAINER_NAME%' does not exist!
    echo.
    echo Please start the database first using:
    echo   scripts\start-database.bat
    echo.
    echo Or use Docker Compose:
    echo   docker-compose -f deployments\docker-compose.dev.yml up -d postgres
    exit /b 1
)

REM Check if container is running
docker ps | findstr /C:"%CONTAINER_NAME%" >nul
if errorlevel 1 (
    echo Database container exists but is not running. Starting it...
    docker start %CONTAINER_NAME%
    echo Waiting for database to be ready...
    timeout /t 10 /nobreak >nul
) else (
    echo Database container is running.
)

REM Wait for PostgreSQL to be ready
echo Verifying database is ready...
:wait_loop
docker exec %CONTAINER_NAME% pg_isready -U %DB_USER% >nul 2>nul
if errorlevel 1 (
    echo Waiting for PostgreSQL to be ready...
    timeout /t 2 /nobreak >nul
    goto wait_loop
)

echo Database is ready!
echo.

echo Step 1/4: Terminating existing connections...

REM Create temporary SQL file for terminating connections
set TEMP_SQL=%TEMP%\terminate_connections_%RANDOM%.sql
echo SELECT pg_terminate_backend(pg_stat_activity.pid) > "%TEMP_SQL%"
echo FROM pg_stat_activity >> "%TEMP_SQL%"
echo WHERE pg_stat_activity.datname = '%DB_NAME%' >> "%TEMP_SQL%"
echo   AND pid ^<^> pg_backend_pid(); >> "%TEMP_SQL%"

REM Execute the SQL
docker exec -i %CONTAINER_NAME% psql -U %DB_USER% -d postgres -f - < "%TEMP_SQL%" 2>nul

REM Clean up temp file
del "%TEMP_SQL%" 2>nul

echo Successfully terminated connections
echo.

echo Step 2/4: Dropping database...
docker exec %CONTAINER_NAME% psql -U %DB_USER% -d postgres -c "DROP DATABASE IF EXISTS %DB_NAME%;"

if errorlevel 1 (
    echo ERROR: Failed to drop database!
    exit /b 1
)

echo Successfully dropped database
echo.

echo Step 3/4: Creating fresh database...
docker exec %CONTAINER_NAME% psql -U %DB_USER% -d postgres -c "CREATE DATABASE %DB_NAME%;"

if errorlevel 1 (
    echo ERROR: Failed to create database!
    exit /b 1
)

echo Successfully created database
echo.

echo Step 4/4: Running all migrations...
go run cmd\migrator\main.go up

if errorlevel 1 (
    echo.
    echo ERROR: Migration failed!
    echo Please check the error messages above.
    exit /b 1
)

echo.
echo ==========================================
echo Successfully reset database!
echo ==========================================
echo.
echo Checking migration status:
go run cmd\migrator\main.go version
echo.

endlocal
