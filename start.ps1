# Load .env file and start the Boeing AI Gateway server
$envFile = Join-Path $PSScriptRoot ".env"
if (Test-Path $envFile) {
    Get-Content $envFile | ForEach-Object {
        if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
            [System.Environment]::SetEnvironmentVariable($matches[1].Trim(), $matches[2].Trim(), "Process")
        }
    }
    Write-Host "Loaded environment from .env" -ForegroundColor Green
}

# Remove stale DB if requested
if ($args -contains "--fresh") {
    $db = Join-Path $PSScriptRoot "boeing.db"
    if (Test-Path $db) {
        Remove-Item $db -Force
        Write-Host "Deleted boeing.db (fresh start)" -ForegroundColor Yellow
    }
}

Write-Host "Starting Boeing AI Gateway..." -ForegroundColor Cyan
go run main.go server --dev-mode
