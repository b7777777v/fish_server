# Run database migrations for Windows (PowerShell)
# Usage: .\scripts\run-migration.ps1 [up|down|version|force] [version_number]

param(
    [string]$Command = "up",
    [int]$Version = 0
)

Write-Host "===================================" -ForegroundColor Cyan
Write-Host "Running Database Migration: $Command" -ForegroundColor Cyan
Write-Host "===================================" -ForegroundColor Cyan
Write-Host ""

try {
    switch ($Command.ToLower()) {
        "up" {
            Write-Host "Applying all pending migrations..." -ForegroundColor Green
            & go run cmd\migrator\main.go up
        }
        "down" {
            Write-Host "Reverting last migration..." -ForegroundColor Yellow
            & go run cmd\migrator\main.go down
        }
        "version" {
            Write-Host "Checking migration version..." -ForegroundColor Green
            & go run cmd\migrator\main.go version
        }
        "force" {
            if ($Version -eq 0) {
                Write-Host "ERROR: Please specify version number" -ForegroundColor Red
                Write-Host "Usage: .\scripts\run-migration.ps1 force -Version [number]" -ForegroundColor Yellow
                exit 1
            }
            Write-Host "Forcing migration to version $Version..." -ForegroundColor Yellow
            & go run cmd\migrator\main.go force $Version
        }
        default {
            Write-Host "ERROR: Unknown command: $Command" -ForegroundColor Red
            Write-Host "Usage: .\scripts\run-migration.ps1 [up|down|version|force]" -ForegroundColor Yellow
            exit 1
        }
    }

    if ($LASTEXITCODE -ne 0) {
        throw "Migration command failed"
    }
} catch {
    Write-Host ""
    Write-Host "ERROR: Migration failed!" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "✓ Migration completed successfully!" -ForegroundColor Green
Write-Host ""
