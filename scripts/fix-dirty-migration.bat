@echo off
REM Fix dirty database migration script for Windows
REM Usage: scripts\fix-dirty-migration.bat [version]
REM If no version specified, it will force to version 5 (before the dirty migration)

setlocal enabledelayedexpansion

echo ===================================
echo Dirty Migration Fix Script (Windows)
echo ===================================
echo.

REM Check if version argument is provided
if "%1"=="" (
    echo No version specified. Will force to version 5 (rollback before dirty migration 6).
    set VERSION=5
) else (
    set VERSION=%1
    echo Will force migration to version !VERSION!
)

echo.
echo Current migration status:
go run cmd\migrator\main.go version
if errorlevel 1 (
    echo Failed to get version - database might not be accessible
)

echo.
echo -----------------------------------
echo IMPORTANT: Before proceeding, you should:
echo 1. Check if migration 6 (create_users_table) was partially applied
echo 2. Manually verify the database state
echo 3. Decide whether to:
echo    - Force to version 5 (rollback) if migration failed early
echo    - Force to version 6 (complete) if migration mostly succeeded
echo -----------------------------------
echo.

set /p confirmation="Do you want to force migration to version !VERSION!? (yes/no): "

if /i not "!confirmation!"=="yes" (
    echo Operation cancelled.
    exit /b 1
)

echo.
echo Forcing migration to version !VERSION!...
go run cmd\migrator\main.go force !VERSION!
if errorlevel 1 (
    echo.
    echo ERROR: Failed to force migration!
    echo Please check the error message above.
    exit /b 1
)

echo.
echo Successfully forced to version !VERSION!
echo.
echo Current migration status:
go run cmd\migrator\main.go version

echo.
echo -----------------------------------
echo Next steps:
if "!VERSION!"=="5" (
    echo 1. Run: go run cmd\migrator\main.go up
    echo    This will re-apply migration 6 and subsequent migrations
) else if "!VERSION!"=="6" (
    echo 1. Verify migration 6 was completed correctly
    echo 2. Run: go run cmd\migrator\main.go up
    echo    This will apply any remaining migrations
)
echo -----------------------------------
echo.

endlocal
