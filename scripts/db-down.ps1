$ErrorActionPreference = "Stop"

$repoRoot = Split-Path $PSScriptRoot -Parent

Push-Location $repoRoot
try {
    Write-Host "Stopping PostgreSQL..."
    docker compose down
    Write-Host "Done."
} finally {
    Pop-Location
}
