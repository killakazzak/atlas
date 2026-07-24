$ErrorActionPreference = "Stop"

$repoRoot = Split-Path $PSScriptRoot -Parent

Push-Location $repoRoot
try {
    Write-Host "Starting PostgreSQL..."
    docker compose up -d postgres

    Write-Host "Waiting for PostgreSQL to be ready..."
    $attempts = 0
    $maxAttempts = 30
    do {
        Start-Sleep -Seconds 1
        $attempts++
        $result = docker exec atlas-postgres pg_isready -U atlas 2>&1
    } while ($LASTEXITCODE -ne 0 -and $attempts -lt $maxAttempts)

    if ($LASTEXITCODE -ne 0) {
        Write-Error "PostgreSQL did not become ready after $maxAttempts seconds."
        exit 1
    }

    Write-Host "PostgreSQL is ready. (attempts: $attempts)"
} finally {
    Pop-Location
}
