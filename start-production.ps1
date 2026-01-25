# =============================================================================
# OMNI Production Startup Script
# =============================================================================
# Usage:
#   .\start-production.ps1              # Start all services
#   .\start-production.ps1 -WithAdmin   # Include pgAdmin
#   .\start-production.ps1 -WithTunnel  # Include Cloudflare tunnel
#   .\start-production.ps1 -Stop        # Stop all services
#   .\start-production.ps1 -Restart     # Restart all services
#   .\start-production.ps1 -Status      # Show status
#   .\start-production.ps1 -Logs        # Show logs
# =============================================================================

param(
    [switch]$WithAdmin,
    [switch]$WithTunnel,
    [switch]$Stop,
    [switch]$Restart,
    [switch]$Status,
    [switch]$Logs,
    [switch]$Build
)

$composeFile = "docker-compose.production.yml"

function Show-Status {
    Write-Host "`n📊 Container Status:" -ForegroundColor Cyan
    docker ps -a --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | Select-String "omni|NAMES"
    
    Write-Host "`n🔗 Service URLs:" -ForegroundColor Cyan
    Write-Host "   Frontend:   http://localhost" -ForegroundColor White
    Write-Host "   API:        http://localhost:8080/api/health" -ForegroundColor White
    Write-Host "   PostgreSQL: localhost:5432" -ForegroundColor White
    if ($WithAdmin) {
        Write-Host "   pgAdmin:    http://localhost:5050" -ForegroundColor White
    }
}

function Start-Services {
    Write-Host "`n🚀 Starting OMNI Production Stack..." -ForegroundColor Green
    
    # Build profiles
    $profiles = @()
    if ($WithAdmin) { $profiles += "--profile"; $profiles += "admin" }
    if ($WithTunnel) { $profiles += "--profile"; $profiles += "tunnel" }
    
    # Build command
    $buildFlag = if ($Build) { "--build" } else { "" }
    
    if ($profiles.Count -gt 0) {
        docker compose -f $composeFile $profiles up -d $buildFlag
    } else {
        docker compose -f $composeFile up -d $buildFlag
    }
    
    Write-Host "`n⏳ Waiting for services to be healthy..." -ForegroundColor Yellow
    Start-Sleep -Seconds 10
    
    Show-Status
}

function Stop-Services {
    Write-Host "`n🛑 Stopping OMNI Production Stack..." -ForegroundColor Red
    docker compose -f $composeFile --profile admin --profile tunnel down
    Write-Host "✅ All services stopped" -ForegroundColor Green
}

function Show-Logs {
    Write-Host "`n📜 Showing logs (Ctrl+C to exit)..." -ForegroundColor Cyan
    docker compose -f $composeFile logs -f --tail 50
}

# Main logic
if ($Stop) {
    Stop-Services
}
elseif ($Restart) {
    Stop-Services
    Start-Sleep -Seconds 3
    Start-Services
}
elseif ($Status) {
    Show-Status
}
elseif ($Logs) {
    Show-Logs
}
else {
    # Ensure network exists
    docker network create omni-network 2>$null
    Start-Services
}
