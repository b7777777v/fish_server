# Cleanup partial migration 6 (create_users_table) - PowerShell version
# This script removes all objects that might have been partially created by migration 6

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Cleaning up partial migration 6" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "This script will remove the following objects if they exist:" -ForegroundColor Yellow
Write-Host "- users table"
Write-Host "- All indexes on users table"
Write-Host "- All constraints on users table"
Write-Host "- update_users_updated_at() function"
Write-Host "- trigger_update_users_updated_at trigger"
Write-Host ""

$confirmation = Read-Host "Do you want to proceed? (yes/no)"

if ($confirmation -ne "yes") {
    Write-Host "Operation cancelled." -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Connecting to database and cleaning up..." -ForegroundColor Green
Write-Host ""

# Create SQL cleanup script
$sqlScript = @"
-- Drop trigger
DROP TRIGGER IF EXISTS trigger_update_users_updated_at ON users;

-- Drop function
DROP FUNCTION IF EXISTS update_users_updated_at();

-- Drop all constraints
ALTER TABLE IF EXISTS users DROP CONSTRAINT IF EXISTS check_third_party;
ALTER TABLE IF EXISTS users DROP CONSTRAINT IF EXISTS check_regular_user;

-- Drop all indexes (explicitly)
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_third_party;
DROP INDEX IF EXISTS idx_users_is_guest;
DROP INDEX IF EXISTS idx_users_created_at;

-- Drop the table
DROP TABLE IF EXISTS users;
"@

# Execute cleanup using docker exec
try {
    $sqlScript | docker exec -i fish_server-postgres-1 psql -U user -d fish_db
    if ($LASTEXITCODE -ne 0) {
        throw "Cleanup command failed"
    }
} catch {
    Write-Host ""
    Write-Host "ERROR: Cleanup failed!" -ForegroundColor Red
    Write-Host "Please check the error messages above." -ForegroundColor Yellow
    exit 1
}

Write-Host ""
Write-Host "✓ Successfully cleaned up partial migration 6!" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "1. Force migration to version 5:"
Write-Host "   go run cmd\migrator\main.go force 5" -ForegroundColor White
Write-Host ""
Write-Host "2. Re-apply migrations:"
Write-Host "   go run cmd\migrator\main.go up" -ForegroundColor White
Write-Host ""
