@echo off
REM Run database migrations for Windows
REM Usage: scripts\run-migration.bat [up|down|version|force]

setlocal

set COMMAND=%1

if "%COMMAND%"=="" (
    set COMMAND=up
)

echo ===================================
echo Running Database Migration: %COMMAND%
echo ===================================
echo.

if "%COMMAND%"=="up" (
    echo Applying all pending migrations...
    go run cmd\migrator\main.go up
) else if "%COMMAND%"=="down" (
    echo Reverting last migration...
    go run cmd\migrator\main.go down
) else if "%COMMAND%"=="version" (
    echo Checking migration version...
    go run cmd\migrator\main.go version
) else if "%COMMAND%"=="force" (
    if "%2"=="" (
        echo ERROR: Please specify version number
        echo Usage: scripts\run-migration.bat force [version]
        exit /b 1
    )
    echo Forcing migration to version %2...
    go run cmd\migrator\main.go force %2
) else (
    echo ERROR: Unknown command: %COMMAND%
    echo Usage: scripts\run-migration.bat [up^|down^|version^|force]
    exit /b 1
)

if errorlevel 1 (
    echo.
    echo ERROR: Migration failed!
    exit /b 1
)

echo.
echo ✓ Migration completed successfully!
echo.

endlocal
