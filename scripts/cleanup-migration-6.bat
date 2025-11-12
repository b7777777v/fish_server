@echo off
REM Cleanup partial migration 6 (create_users_table) - Windows version
REM This script removes all objects that might have been partially created by migration 6

setlocal enabledelayedexpansion

echo ==========================================
echo Cleaning up partial migration 6
echo ==========================================
echo.

echo This script will remove the following objects if they exist:
echo - users table
echo - All indexes on users table
echo - All constraints on users table
echo - update_users_updated_at() function
echo - trigger_update_users_updated_at trigger
echo.

set /p confirmation="Do you want to proceed? (yes/no): "

if /i not "!confirmation!"=="yes" (
    echo Operation cancelled.
    exit /b 1
)

echo.
echo Connecting to database and cleaning up...
echo.

REM Create temporary SQL file
set TEMP_SQL=%TEMP%\cleanup_migration_6.sql
echo -- Drop trigger > "%TEMP_SQL%"
echo DROP TRIGGER IF EXISTS trigger_update_users_updated_at ON users; >> "%TEMP_SQL%"
echo. >> "%TEMP_SQL%"
echo -- Drop function >> "%TEMP_SQL%"
echo DROP FUNCTION IF EXISTS update_users_updated_at(); >> "%TEMP_SQL%"
echo. >> "%TEMP_SQL%"
echo -- Drop all constraints >> "%TEMP_SQL%"
echo ALTER TABLE IF EXISTS users DROP CONSTRAINT IF EXISTS check_third_party; >> "%TEMP_SQL%"
echo ALTER TABLE IF EXISTS users DROP CONSTRAINT IF EXISTS check_regular_user; >> "%TEMP_SQL%"
echo. >> "%TEMP_SQL%"
echo -- Drop all indexes (explicitly) >> "%TEMP_SQL%"
echo DROP INDEX IF EXISTS idx_users_username; >> "%TEMP_SQL%"
echo DROP INDEX IF EXISTS idx_users_third_party; >> "%TEMP_SQL%"
echo DROP INDEX IF EXISTS idx_users_is_guest; >> "%TEMP_SQL%"
echo DROP INDEX IF EXISTS idx_users_created_at; >> "%TEMP_SQL%"
echo. >> "%TEMP_SQL%"
echo -- Drop the table >> "%TEMP_SQL%"
echo DROP TABLE IF EXISTS users; >> "%TEMP_SQL%"

REM Execute cleanup using docker exec
docker exec -i fish_server-postgres-1 psql -U user -d fish_db < "%TEMP_SQL%"

if errorlevel 1 (
    echo.
    echo ERROR: Cleanup failed!
    echo Please check the error messages above.
    del "%TEMP_SQL%"
    exit /b 1
)

REM Clean up temp file
del "%TEMP_SQL%"

echo.
echo Successfully cleaned up partial migration 6!
echo.
echo Next steps:
echo 1. Force migration to version 5:
echo    go run cmd\migrator\main.go force 5
echo.
echo 2. Re-apply migrations:
echo    go run cmd\migrator\main.go up
echo.

endlocal
