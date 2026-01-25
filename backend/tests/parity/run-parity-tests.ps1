#!/usr/bin/env pwsh
# Parity Test Runner
# Compares Node.js vs Go API responses

param(
    [string]$NodeJSURL = "http://localhost:3000",
    [string]$GoURL = "http://localhost:8080",
    [string]$TenantID = "yumna_bertigamart",
    [string]$AuthToken = "",
    [switch]$Verbose
)

$ErrorActionPreference = "Stop"

Write-Host "╔══════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║           PARITY TEST: Node.js vs Go Backend             ║" -ForegroundColor Cyan
Write-Host "╚══════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

# Set environment variables
$env:NODEJS_URL = $NodeJSURL
$env:GO_URL = $GoURL
$env:TENANT_ID = $TenantID
$env:AUTH_TOKEN = $AuthToken

Write-Host "Configuration:" -ForegroundColor Yellow
Write-Host "  Node.js URL: $NodeJSURL"
Write-Host "  Go URL:      $GoURL"
Write-Host "  Tenant ID:   $TenantID"
Write-Host "  Auth Token:  $(if ($AuthToken) { '***' } else { '(not set)' })"
Write-Host ""

# Check if servers are running
Write-Host "Checking server availability..." -ForegroundColor Yellow

try {
    $nodeHealth = Invoke-RestMethod -Uri "$NodeJSURL/api/health" -TimeoutSec 5
    Write-Host "  ✓ Node.js server is running" -ForegroundColor Green
} catch {
    Write-Host "  ✗ Node.js server not available at $NodeJSURL" -ForegroundColor Red
    Write-Host "    Start with: docker compose up -d backend" -ForegroundColor Gray
}

try {
    $goHealth = Invoke-RestMethod -Uri "$GoURL/api/health" -TimeoutSec 5
    Write-Host "  ✓ Go server is running" -ForegroundColor Green
} catch {
    Write-Host "  ✗ Go server not available at $GoURL" -ForegroundColor Red
    Write-Host "    Start with: docker compose up -d backend-go" -ForegroundColor Gray
}

Write-Host ""
Write-Host "Running parity tests..." -ForegroundColor Yellow
Write-Host ""

# Run Go tests
$testArgs = @("-v", "./tests/parity/...")
if ($Verbose) {
    $testArgs += "-count=1"
}

Push-Location $PSScriptRoot/..
try {
    go test @testArgs
    $exitCode = $LASTEXITCODE
} finally {
    Pop-Location
}

Write-Host ""
if ($exitCode -eq 0) {
    Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Green
    Write-Host "  All parity tests passed!" -ForegroundColor Green
    Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Green
} else {
    Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Red
    Write-Host "  Some parity tests failed. Check differences above." -ForegroundColor Red
    Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Red
}

exit $exitCode
