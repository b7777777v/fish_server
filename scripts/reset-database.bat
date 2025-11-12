@echo off
REM Complete database reset script for Windows
REM WARNING: This will DELETE all data and recreate the database from scratch

setlocal enabledelayedexpansion

echo ==========================================
echo WARNING: DATABASE COMPLETE RESET
echo ==========================================
echo.
echo This script will:
echo 1. Stop all connections to the database
echo 2. DROP the entire database
echo 3. CREATE a fresh database
echo 4. Run all migrations from scratch
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

echo Step 1/4: Terminating existing connections...

REM Create temporary SQL file for terminating connections
set TEMP_SQL=%TEMP%\terminate_connections_%RANDOM%.sql
echo SELECT pg_terminate_backend(pg_stat_activity.pid) > "%TEMP_SQL%"
echo FROM pg_stat_activity >> "%TEMP_SQL%"
echo WHERE pg_stat_activity.datname = '%DB_NAME%' >> "%TEMP_SQL%"
echo   AND pid ^<^> pg_backend_pid(); >> "%TEMP_SQL%"

REM Execute the SQL
docker exec -i fish_server-postgres-1 psql -U %DB_USER% -d postgres -f - < "%TEMP_SQL%" 2>nul

REM Clean up temp file
del "%TEMP_SQL%" 2>nul

echo Successfully terminated connections
echo.

echo Step 2/4: Dropping database...
docker exec -i fish_server-postgres-1 psql -U %DB_USER% -d postgres -c "DROP DATABASE IF EXISTS %DB_NAME%;"

echo Successfully dropped database
echo.

echo Step 3/4: Creating fresh database...
docker exec -i fish_server-postgres-1 psql -U %DB_USER% -d postgres -c "CREATE DATABASE %DB_NAME%;"

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
