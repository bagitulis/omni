#!/usr/bin/env pwsh
# Traffic Split Control Script
# Manages gradual migration from Node.js to Go backend

param(
    [Parameter(Position=0)]
    [ValidateSet("status", "10", "50", "100", "rollback")]
    [string]$Action = "status"
)

$ErrorActionPreference = "Stop"
$ConfigFile = "$PSScriptRoot/nginx-traffic-split.conf"

function Show-Banner {
    Write-Host ""
    Write-Host "╔══════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
    Write-Host "║         TRAFFIC SPLIT CONTROLLER: Node.js → Go           ║" -ForegroundColor Cyan
    Write-Host "╚══════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
    Write-Host ""
}

function Get-CurrentSplit {
    $content = Get-Content $ConfigFile -Raw
    
    if ($content -match '(?m)^split_clients.*\{[\s\S]*?(\d+)%\s+go_backend') {
        return [int]$Matches[1]
    } elseif ($content -match '(?m)^split_clients.*\{\s*\*\s+go_backend') {
        return 100
    }
    return 0
}

function Set-TrafficSplit {
    param([int]$Percentage)
    
    $content = Get-Content $ConfigFile -Raw
    
    # Comment out all split_clients blocks
    $content = $content -replace '(?m)^(split_clients)', '# $1'
    
    # Uncomment the appropriate block
    switch ($Percentage) {
        10 {
            $content = $content -replace '(?m)^# (split_clients.*\{[\s\S]*?10%\s+go_backend[\s\S]*?\})', '$1'
        }
        50 {
            $content = $content -replace '(?m)^# (split_clients.*\{[\s\S]*?50%\s+go_backend[\s\S]*?\})', '$1'
        }
        100 {
            $content = $content -replace '(?m)^# (split_clients.*\{\s*\*\s+go_backend[\s\S]*?\})', '$1'
        }
    }
    
    Set-Content $ConfigFile -Value $content
}

function Reload-Nginx {
    Write-Host "Reloading nginx configuration..." -ForegroundColor Yellow
    docker exec omni-nginx nginx -t 2>&1 | Out-Null
    if ($LASTEXITCODE -eq 0) {
        docker exec omni-nginx nginx -s reload
        Write-Host "  ✓ Nginx reloaded successfully" -ForegroundColor Green
    } else {
        Write-Host "  ✗ Nginx config test failed!" -ForegroundColor Red
        docker exec omni-nginx nginx -t
        exit 1
    }
}

function Show-Status {
    $currentSplit = Get-CurrentSplit
    
    Write-Host "Current Traffic Distribution:" -ForegroundColor Yellow
    Write-Host ""
    
    $goBar = "█" * ($currentSplit / 5)
    $nodeBar = "█" * ((100 - $currentSplit) / 5)
    
    Write-Host "  Go Backend:     " -NoNewline
    Write-Host "$goBar" -NoNewline -ForegroundColor Green
    Write-Host " $currentSplit%" -ForegroundColor Green
    
    Write-Host "  Node.js Backend:" -NoNewline
    Write-Host "$nodeBar" -NoNewline -ForegroundColor Blue
    Write-Host " $((100 - $currentSplit))%" -ForegroundColor Blue
    
    Write-Host ""
    Write-Host "Commands:" -ForegroundColor Gray
    Write-Host "  .\traffic-split.ps1 10       # Set 10% to Go (testing)"
    Write-Host "  .\traffic-split.ps1 50       # Set 50% to Go (parallel)"
    Write-Host "  .\traffic-split.ps1 100      # Set 100% to Go (full migration)"
    Write-Host "  .\traffic-split.ps1 rollback # Rollback to 0% Go"
    Write-Host ""
}

# Main
Show-Banner

switch ($Action) {
    "status" {
        Show-Status
    }
    "10" {
        Write-Host "Setting traffic split to 10% Go / 90% Node.js..." -ForegroundColor Yellow
        Set-TrafficSplit 10
        Reload-Nginx
        Show-Status
    }
    "50" {
        Write-Host "Setting traffic split to 50% Go / 50% Node.js..." -ForegroundColor Yellow
        Set-TrafficSplit 50
        Reload-Nginx
        Show-Status
    }
    "100" {
        Write-Host "Setting traffic split to 100% Go / 0% Node.js..." -ForegroundColor Yellow
        Write-Host ""
        Write-Host "⚠️  WARNING: This will route ALL traffic to Go backend!" -ForegroundColor Red
        $confirm = Read-Host "Type 'yes' to confirm"
        if ($confirm -eq "yes") {
            Set-TrafficSplit 100
            Reload-Nginx
            Show-Status
        } else {
            Write-Host "Cancelled." -ForegroundColor Gray
        }
    }
    "rollback" {
        Write-Host "Rolling back to 0% Go (all traffic to Node.js)..." -ForegroundColor Yellow
        Set-TrafficSplit 0
        Reload-Nginx
        Show-Status
    }
}
