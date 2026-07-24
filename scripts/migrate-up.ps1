$ErrorActionPreference = "Stop"

$repoRoot = Split-Path $PSScriptRoot -Parent

if (-not (Get-Command migrate -ErrorAction SilentlyContinue)) {
    Write-Error "migrate.exe not found. Install it: https://github.com/golang-migrate/migrate"
    exit 1
}

if (-not $env:DATABASE_URL) {
    $env:DATABASE_URL = "postgres://atlas:atlas@localhost:5432/atlas?sslmode=disable"
}

$migrationsPath = Join-Path $repoRoot "backend\migrations"

Write-Host "Running migrations from $migrationsPath..."
migrate -path $migrationsPath -database $env:DATABASE_URL up
Write-Host "Migrations applied."
