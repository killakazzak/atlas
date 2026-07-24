$ErrorActionPreference = "Stop"

$repoRoot = Split-Path $PSScriptRoot -Parent

if (-not $env:DATABASE_URL) {
    $env:DATABASE_URL = "postgres://atlas:atlas@localhost:5432/atlas?sslmode=disable"
}

Push-Location (Join-Path $repoRoot "backend")
try {
    Write-Host "Starting Atlas server..."
    go run ./cmd/server
} finally {
    Pop-Location
}
