# Fix dirty database migration script for Windows (PowerShell)
# Usage: .\scripts\fix-dirty-migration.ps1 [version]
# If no version specified, it will force to version 5 (before the dirty migration)

param(
    [int]$Version = 5
)

Write-Host "===================================" -ForegroundColor Cyan
Write-Host "Dirty Migration Fix Script (PowerShell)" -ForegroundColor Cyan
Write-Host "===================================" -ForegroundColor Cyan
Write-Host ""

if ($PSBoundParameters.ContainsKey('Version')) {
    Write-Host "Will force migration to version $Version" -ForegroundColor Yellow
} else {
    Write-Host "No version specified. Will force to version 5 (rollback before dirty migration 6)." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Current migration status:" -ForegroundColor Green
try {
    & go run cmd\migrator\main.go version
} catch {
    Write-Host "Failed to get version - database might not be accessible" -ForegroundColor Red
}

Write-Host ""
Write-Host "-----------------------------------" -ForegroundColor Yellow
Write-Host "IMPORTANT: Before proceeding, you should:" -ForegroundColor Yellow
Write-Host "1. Check if migration 6 (create_users_table) was partially applied"
Write-Host "2. Manually verify the database state"
Write-Host "3. Decide whether to:"
Write-Host "   - Force to version 5 (rollback) if migration failed early"
Write-Host "   - Force to version 6 (complete) if migration mostly succeeded"
Write-Host "-----------------------------------" -ForegroundColor Yellow
Write-Host ""

$confirmation = Read-Host "Do you want to force migration to version $Version? (yes/no)"

if ($confirmation -ne "yes") {
    Write-Host "Operation cancelled." -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Forcing migration to version $Version..." -ForegroundColor Green
try {
    & go run cmd\migrator\main.go force $Version
    if ($LASTEXITCODE -ne 0) {
        throw "Migration force command failed"
    }
} catch {
    Write-Host ""
    Write-Host "ERROR: Failed to force migration!" -ForegroundColor Red
    Write-Host "Please check the error message above." -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "✓ Successfully forced to version $Version" -ForegroundColor Green
Write-Host ""
Write-Host "Current migration status:" -ForegroundColor Green
& go run cmd\migrator\main.go version

Write-Host ""
Write-Host "-----------------------------------" -ForegroundColor Cyan
Write-Host "Next steps:" -ForegroundColor Cyan
if ($Version -eq 5) {
    Write-Host "1. Run: go run cmd\migrator\main.go up"
    Write-Host "   This will re-apply migration 6 and subsequent migrations"
} elseif ($Version -eq 6) {
    Write-Host "1. Verify migration 6 was completed correctly"
    Write-Host "2. Run: go run cmd\migrator\main.go up"
    Write-Host "   This will apply any remaining migrations"
}
Write-Host "-----------------------------------" -ForegroundColor Cyan
Write-Host ""
