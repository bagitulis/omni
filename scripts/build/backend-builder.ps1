# ============================================
# Backend Build Module (Go)
# Build Go backend with Docker
# ============================================

. "$PSScriptRoot\config.ps1"
. "$PSScriptRoot\logger.ps1"
. "$PSScriptRoot\retry-logic.ps1"

function Test-GoBackend {
    <#
    .SYNOPSIS
    Verify Go backend structure exists
    #>
    $goMod = Join-Path $script:BackendPath "go.mod"
    $dockerfile = Join-Path $script:BackendPath "Dockerfile"
    
    if (-not (Test-Path $goMod)) {
        Write-Warning "go.mod not found at $goMod"
        return $false
    }
    
    if (-not (Test-Path $dockerfile)) {
        Write-Warning "Dockerfile not found at $dockerfile"
        return $false
    }
    
    return $true
}

function Invoke-BackendBuildStep {
    <#
    .SYNOPSIS
    Build Go backend using Docker
    #>
    param([switch]$CleanFirst)
    
    Write-Step "BACKEND" "Building Go backend..."
    
    if (-not (Test-GoBackend)) {
        Write-Error "Go backend structure invalid"
        return $false
    }
    
    Push-Location $script:ProjectRoot
    
    try {
        # Clean build cache if requested
        if ($CleanFirst) {
            Write-Debug "Cleaning Docker build cache for backend..."
            docker builder prune -f --filter "label=stage=backend-builder" 2>&1 | Out-Null
        }
        
        $result = Invoke-WithRetry -ScriptBlock {
            # Build using docker compose
            $output = docker compose -f docker-compose.go.yml build backend-go 2>&1
            $success = $LASTEXITCODE -eq 0
            
            if (-not $success) {
                Write-Host $output -ForegroundColor Red
            }
            
            return $success
        } -RetryMessage "Retrying Go backend build" -MaxRetries 2
        
        if ($result.Success) {
            Write-Success "Go backend built successfully"
            return $true
        }
        else {
            Write-Error "Go backend build failed: $($result.Error)"
            return $false
        }
    }
    finally {
        Pop-Location
    }
}

function Invoke-BackendBuild {
    <#
    .SYNOPSIS
    Full backend build process
    #>
    param(
        [ValidateSet("quick", "incremental", "full")]
        [string]$Mode = "incremental"
    )
    
    Write-Header "Go Backend Build ($Mode)"
    
    switch ($Mode) {
        "quick" {
            Write-Info "Skipping backend build (quick mode)"
            return $true
        }
        "incremental" {
            return Invoke-BackendBuildStep
        }
        "full" {
            return Invoke-BackendBuildStep -CleanFirst
        }
    }
}

function Get-BackendHealth {
    <#
    .SYNOPSIS
    Check Go backend health endpoint
    #>
    param([int]$TimeoutSeconds = 30)
    
    $endpoint = "http://localhost:3000/api/health"
    $startTime = Get-Date
    
    while (((Get-Date) - $startTime).TotalSeconds -lt $TimeoutSeconds) {
        try {
            $response = Invoke-RestMethod -Uri $endpoint -TimeoutSec 5 -ErrorAction SilentlyContinue
            if ($response.status -eq "ok" -or $response.success -eq $true) {
                return $true
            }
        }
        catch {
            # Endpoint not ready yet
        }
        Start-Sleep -Seconds 2
    }
    
    return $false
}

# Functions are loaded via dot-sourcing in build.ps1, no export needed
