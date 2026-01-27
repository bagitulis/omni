# ============================================
# Build Modes Module
# SRP: Individual build mode implementations
# ============================================

. "$PSScriptRoot\config.ps1"
. "$PSScriptRoot\logger.ps1"

# ============================================
# Pre-Build Checks
# ============================================

function Invoke-PreBuildChecks {
    <#
    .SYNOPSIS
    Run all pre-build checks and auto-fix issues
    #>
    Write-Header "Pre-Build Checks (Auto-Fix Enabled)"
    
    # 1. Docker
    Write-Step "1/7" "Checking Docker status..."
    if (-not (Assert-DockerReady)) {
        Write-Error "Failed to ensure Docker is ready"
        return $false
    }
    
    # 2. Environment
    Write-Step "2/7" "Checking environment..."
    $envStatus = Get-EnvironmentStatus
    
    if (-not $envStatus.DiskSpaceOK) {
        Write-Warning "Low disk space. Running cleanup..."
        Invoke-LightCleanup
        $newSpace = Get-AvailableDiskSpace
        if ($newSpace -lt $script:MinDiskSpaceGB) {
            Write-Info "Running aggressive cleanup..."
            Invoke-AggressiveCleanup
        }
    }
    
    # 3. WSL Mount
    Write-Step "3/7" "Checking WSL mount health..."
    if (-not $envStatus.WslMountHealthy) {
        Write-Warning "WSL mount cache corruption: $($envStatus.WslMountError)"
        Write-Fix "Repairing WSL mount cache..."
        if (-not (Repair-WslMountCache)) {
            Write-Error "Failed to repair WSL mount cache"
            return $false
        }
    } else {
        Write-Success "WSL mount cache healthy"
    }
    
    # 4. Memory
    Write-Step "4/7" "Checking available memory..."
    if (-not (Test-AvailableMemory -MinMemoryMB 1024)) {
        Write-Warning "Low memory. Freeing resources..."
        Invoke-MemoryCleanup
        Start-Sleep -Seconds 3
    }
    
    # 5. Ports
    Write-Step "5/7" "Checking port availability..."
    Repair-PortConflict
    
    # 6. Nginx
    Write-Step "6/7" "Validating Nginx configuration..."
    $nginxResult = Invoke-NginxValidation
    if (-not $nginxResult.Valid) {
        Write-Warning "Nginx issues detected"
        Repair-NginxCommonIssues
    }
    
    # 7. Directories
    Write-Step "7/7" "Ensuring required directories..."
    $created = Test-RequiredDirectories
    if ($created.Count -gt 0) {
        Write-Info "Created: $($created -join ', ')"
    }
    
    Write-Success "Pre-build checks completed"
    return $true
}

# ============================================
# Build Mode Router
# ============================================

function Invoke-BuildMode {
    <#
    .SYNOPSIS
    Route to appropriate build mode
    #>
    param(
        [string]$Mode,
        [string]$Spec,
        [bool]$ShowBuildOutput = $true
    )
    
    switch ($Mode) {
        "quick"       { return Invoke-QuickApply -Spec $Spec }
        "smart"       { return Invoke-SmartBuild -Spec $Spec -ShowBuildOutput $ShowBuildOutput }
        "full"        { return Invoke-FullRebuild -Spec $Spec -ShowBuildOutput $ShowBuildOutput }
        default       { Write-Error "Unknown mode: $Mode"; return $false }
    }
}

# ============================================
# Quick Apply Mode
# ============================================

function Invoke-QuickApply {
    <#
    .SYNOPSIS
    Quick apply - restart containers only, with auto-fallback to Smart Build if fails
    #>
    param([string]$Spec = "standard")
    
    Write-Header "Quick Apply Mode"
    
    if (-not (Assert-DockerReady)) { return $false }
    
    $result = Invoke-QuickRestart -Spec $Spec
    
    if ($result) {
        $null = Wait-ForContainerHealth
        Show-ContainerStatus -Spec $Spec | Out-Host
        Show-DeploymentSummary -Spec $Spec
        return $true
    }
    
    # Quick Apply failed - offer fallback options
    Write-Host ""
    Write-Host "============================================" -ForegroundColor Yellow
    Write-Host "  Quick Apply Failed - Choose Next Action" -ForegroundColor Yellow
    Write-Host "============================================" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Quick restart couldn't fix the issue. Options:" -ForegroundColor White
    Write-Host ""
    Write-Host "  [1] Smart Build      - Rebuild code, keep cached deps (RECOMMENDED)" -ForegroundColor Green
    Write-Host "  [2] Full Rebuild     - Clean rebuild everything"
    Write-Host "  [3] Clean Docker     - Aggressive cleanup, then retry"
    Write-Host "  [0] Cancel           - Return to menu"
    Write-Host ""
    
    $fallbackChoice = (Read-Host "Enter choice (0-3)").Trim()
    
    switch ($fallbackChoice) {
        "1" {
            Write-Host ""
            Write-Info "Escalating to Smart Build..."
            return Invoke-SmartBuild -Spec $Spec -ShowBuildOutput $true
        }
        "2" {
            Write-Host ""
            Write-Info "Escalating to Full Rebuild..."
            return Invoke-FullRebuild -Spec $Spec -ShowBuildOutput $true
        }
        "3" {
            Write-Host ""
            Write-Info "Running aggressive cleanup..."
            Invoke-AggressiveCleanup
            Write-Host ""
            Write-Info "Retrying Quick Apply after cleanup..."
            # Retry quick apply once after cleanup
            $retryResult = Invoke-QuickRestart -Spec $Spec
            if ($retryResult) {
                $null = Wait-ForContainerHealth
                Show-ContainerStatus -Spec $Spec | Out-Host
                Show-DeploymentSummary -Spec $Spec
                return $true
            }
            Write-Warning "Still failing after cleanup. Escalating to Smart Build..."
            return Invoke-SmartBuild -Spec $Spec -ShowBuildOutput $true
        }
        "0" {
            Write-Info "Cancelled by user"
            return $false
        }
        default {
            Write-Warning "Invalid choice. Defaulting to Smart Build..."
            return Invoke-SmartBuild -Spec $Spec -ShowBuildOutput $true
        }
    }
}

# ============================================
# Smart Build Mode (RECOMMENDED)
# ============================================

function Invoke-SmartBuild {
    <#
    .SYNOPSIS
    Smart build - cache dependencies (npm install, apk add), rebuild source code
    This is the fastest mode for code changes while ensuring fresh builds.
    
    How it works:
    - Docker caches layers: apk add, npm install (slow steps)
    - COPY src/. happens AFTER npm install, so code changes trigger rebuild
    - Source code is ALWAYS rebuilt fresh (no cache)
    - Dependencies use cache (fast when package.json unchanged)
    #>
    param(
        [string]$Spec = "standard",
        [bool]$ShowBuildOutput = $true
    )
    
    Write-Header "Smart Build Mode (Cache deps, Fresh code)"
    Write-Info "Dependencies: CACHED (npm install, apk add)"
    Write-Info "Source code:  FRESH (always rebuilt)"
    Write-Host ""
    
    # Pre-checks (minimal - skip WSL restart for speed)
    Write-Step "1/5" "Checking Docker..."
    if (-not (Assert-DockerReady)) {
        Write-Error "Docker not ready"
        return $false
    }
    
    # Frontend build (includes its own TypeScript check)
    Write-Step "2/5" "Building frontend..."
    if (-not (Invoke-FrontendBuild -Mode "incremental")) { return $false }
    
    # Stop containers
    Write-Step "3/5" "Stopping containers..."
    Stop-ExistingContainers -Spec $Spec
    
    # Build with strict error handling
    Write-Step "4/5" "Building Docker images..."
    if (-not (Invoke-DockerBuild -Spec $Spec -ShowProgress $ShowBuildOutput)) {
        Write-Error "Docker build failed"
        return $false
    }
    
    # Deploy with strict health verification
    Write-Step "5/5" "Deploying and verifying..."
    if (-not (Invoke-DockerDeploy -Spec $Spec)) {
        Write-Error "Docker deploy failed"
        return $false
    }
    
    # Strict health check - MUST pass
    $healthResult = Invoke-StrictHealthCheck -MaxRetries 10 -RetryDelaySeconds 3
    if (-not $healthResult.Success) {
        Write-Error "Health check failed after deployment!"
        Write-Host ""
        Write-Host "  Errors:" -ForegroundColor Red
        $healthResult.Errors | ForEach-Object { Write-Host "    - $_" -ForegroundColor Red }
        Write-Host ""
        Write-Host "  Showing backend logs:" -ForegroundColor Yellow
        docker logs --tail 30 omni-backend 2>&1 | ForEach-Object { Write-Host "    $_" -ForegroundColor Gray }
        return $false
    }
    
    # Success
    $null = Invoke-PostDeployTasks
    Show-ContainerStatus -Spec $Spec | Out-Host
    Show-DeploymentSummary -Spec $Spec
    
    return $true
}

function Invoke-StrictHealthCheck {
    <#
    .SYNOPSIS
    Strict health verification - ensures deployment success
    Uses Docker's built-in health check status instead of manual HTTP check
    #>
    param(
        [int]$MaxRetries = 15,
        [int]$RetryDelaySeconds = 4
    )
    
    $errors = @()
    
    Write-Info "Verifying deployment health..."
    
    # Give containers time to start their health checks
    Start-Sleep -Seconds 5
    
    for ($i = 1; $i -le $MaxRetries; $i++) {
        Write-Host "  Health check $i/$MaxRetries..." -ForegroundColor Gray -NoNewline
        
        # Check Docker's built-in health status
        $backendHealth = docker inspect --format='{{.State.Health.Status}}' omni-backend 2>$null
        $frontendHealth = docker inspect --format='{{.State.Health.Status}}' omni-frontend 2>$null
        $backendStatus = docker inspect --format='{{.State.Status}}' omni-backend 2>$null
        $frontendStatus = docker inspect --format='{{.State.Status}}' omni-frontend 2>$null
        
        # Check if containers are running
        if ($backendStatus -ne "running") {
            Write-Host " backend not running ($backendStatus)" -ForegroundColor Yellow
            Start-Sleep -Seconds $RetryDelaySeconds
            continue
        }
        
        if ($frontendStatus -ne "running") {
            Write-Host " frontend not running ($frontendStatus)" -ForegroundColor Yellow
            Start-Sleep -Seconds $RetryDelaySeconds
            continue
        }
        
        # Check Docker health status (healthy/starting/unhealthy)
        if ($backendHealth -eq "healthy" -and $frontendHealth -eq "healthy") {
            Write-Host " ✓" -ForegroundColor Green
            Write-Success "All containers are healthy!"
            return @{ Success = $true; Errors = @() }
        }
        
        # If health check is still starting, wait
        if ($backendHealth -eq "starting" -or $frontendHealth -eq "starting") {
            Write-Host " starting (backend:$backendHealth, frontend:$frontendHealth)" -ForegroundColor Yellow
            Start-Sleep -Seconds $RetryDelaySeconds
            continue
        }
        
        # Try API health endpoint as fallback
        try {
            $response = Invoke-RestMethod -Uri "http://localhost:3000/api/health" -TimeoutSec 5 -ErrorAction Stop
            if ($response.status -eq "ok" -or $response.status -eq "healthy" -or $response.success -eq $true) {
                Write-Host " ✓ (API OK)" -ForegroundColor Green
                Write-Success "API health check passed!"
                return @{ Success = $true; Errors = @() }
            }
        }
        catch {
            # API not ready yet, continue waiting
        }
        
        Write-Host " waiting (backend:$backendHealth, frontend:$frontendHealth)..." -ForegroundColor Yellow
        Start-Sleep -Seconds $RetryDelaySeconds
    }
    
    # Collect error info only if truly unhealthy
    $backendHealth = docker inspect --format='{{.State.Health.Status}}' omni-backend 2>$null
    if ($backendHealth -eq "unhealthy") {
        $errors += "Backend container is unhealthy"
        # Show last health check log
        $healthLog = docker inspect --format='{{range .State.Health.Log}}{{.Output}}{{end}}' omni-backend 2>$null | Select-Object -Last 1
        if ($healthLog) { $errors += "Health log: $healthLog" }
    }
    
    $frontendHealth = docker inspect --format='{{.State.Health.Status}}' omni-frontend 2>$null
    if ($frontendHealth -eq "unhealthy") {
        $errors += "Frontend container is unhealthy"
    }
    
    if ($errors.Count -eq 0) {
        $errors += "Health endpoint not responding after $MaxRetries attempts"
    }
    
    return @{ Success = $false; Errors = $errors }
}

# ============================================
# Full Rebuild Mode
# ============================================

function Invoke-FullRebuild {
    <#
    .SYNOPSIS
    Full rebuild - clean rebuild from scratch with robust error handling
    Includes database reset with automatic restore from backup
    #>
    param(
        [string]$Spec = "standard",
        [bool]$ShowBuildOutput = $true
    )
    
    Write-Header "Full Rebuild Mode"
    
    # Check if database has data and warn user
    $hasLocalData = $false
    $hasBackup = Test-Path (Join-Path $script:ProjectRoot "backups\smart\manifest.json")
    
    # Check if postgres container exists and has data
    $pgExists = docker inspect "omni-postgres" 2>$null
    if ($LASTEXITCODE -eq 0) {
        $pgStatus = docker inspect --format='{{.State.Status}}' "omni-postgres" 2>$null
        if ($pgStatus -eq "running") {
            $rowQuery = "SELECT COALESCE(SUM(n_live_tup), 0) FROM pg_stat_user_tables WHERE schemaname LIKE 'tenant_%';"
            $rowResult = docker exec omni-postgres psql -U omni -d omni_main -t -A -c $rowQuery 2>&1
            if ($LASTEXITCODE -eq 0 -and [int]($rowResult.Trim()) -gt 0) {
                $hasLocalData = $true
            }
        }
    }
    
    # Show warning if local database has data
    if ($hasLocalData) {
        Write-Host ""
        Write-Host "  +-------------------------------------------+" -ForegroundColor Red
        Write-Host "  | WARNING: DATABASE WILL BE RESET!          |" -ForegroundColor Red
        Write-Host "  +-------------------------------------------+" -ForegroundColor Red
        Write-Host "  | Full Rebuild will:                        |" -ForegroundColor White
        Write-Host "  |   - Remove all Docker volumes             |" -ForegroundColor Yellow
        Write-Host "  |   - Delete current database data          |" -ForegroundColor Yellow
        if ($hasBackup) {
            Write-Host "  |   - Auto-restore from backup             |" -ForegroundColor Green
            Write-Host "  +-------------------------------------------+" -ForegroundColor Red
            Write-Host ""
            Write-Host "  Backup available - data will be restored automatically" -ForegroundColor Cyan
        } else {
            Write-Host "  |   - Start with EMPTY database            |" -ForegroundColor Red
            Write-Host "  +-------------------------------------------+" -ForegroundColor Red
            Write-Host ""
            Write-Host "  NO BACKUP FOUND! Run backup first:" -ForegroundColor Red
            Write-Host "    backups\db-tools\menu.bat -> [2] Smart Backup" -ForegroundColor Yellow
        }
        Write-Host ""
        
        $choice = Read-Host "  Continue with Full Rebuild? (y/N)"
        if ($choice -ne "y" -and $choice -ne "Y") {
            Write-Info "Full Rebuild cancelled"
            Write-Info "TIP: Use Smart Build [2] to preserve database"
            return $false
        }
        Write-Host ""
    }
    
    # Pre-checks
    Write-Step "1/7" "Running pre-build checks..."
    if (-not (Invoke-PreBuildChecks)) { return $false }
    
    # Clean including Docker volumes
    Write-Step "2/7" "Removing old build artifacts and volumes..."
    Clear-OldBuildArtifacts
    Invoke-MediumCleanup
    
    # Remove postgres volume to ensure fresh database
    Write-Info "Removing database volume for fresh start..."
    docker volume rm omni_postgres_data 2>$null | Out-Null
    
    # Frontend
    Write-Step "3/7" "Building frontend..."
    if (-not (Invoke-FrontendBuild -Mode "full")) { return $false }
    
    # Reset WSL mount cache with proper Docker recovery
    Write-Step "4/7" "Resetting WSL mount cache..."
    wsl --shutdown 2>&1 | Out-Null
    Start-Sleep -Seconds 3
    
    # CRITICAL: Wait for Docker to be ready after WSL shutdown
    Write-Info "Waiting for Docker engine to recover..."
    if (-not (Assert-DockerReady)) {
        Write-Warning "Docker not ready after WSL reset, attempting recovery..."
        # Try to restart Docker Desktop
        Repair-DockerEngineError
        Start-Sleep -Seconds 10
        if (-not (Assert-DockerReady)) {
            Write-Error "Failed to recover Docker after WSL shutdown"
            return $false
        }
    }
    Write-Success "Docker engine ready"
    
    # Build & deploy (no cache) with strict error handling
    Write-Step "5/7" "Building Docker images (no cache)..."
    if (-not (Invoke-DockerBuild -Spec $Spec -NoCache -ShowProgress $ShowBuildOutput)) {
        Write-Error "Docker build failed"
        return $false
    }
    
    Write-Step "6/7" "Deploying and verifying..."
    if (-not (Invoke-DockerDeploy -Spec $Spec)) {
        Write-Error "Docker deploy failed"
        return $false
    }
    
    # Strict health check like Smart Build
    $healthResult = Invoke-StrictHealthCheck -MaxRetries 10 -RetryDelaySeconds 3
    if (-not $healthResult.Success) {
        Write-Error "Health check failed after deployment!"
        Write-Host ""
        Write-Host "  Errors:" -ForegroundColor Red
        $healthResult.Errors | ForEach-Object { Write-Host "    - $_" -ForegroundColor Red }
        Write-Host ""
        Write-Host "  Showing backend logs:" -ForegroundColor Yellow
        docker logs --tail 30 omni-backend 2>&1 | ForEach-Object { Write-Host "    $_" -ForegroundColor Gray }
        return $false
    }
    
    # Post-deploy with AUTO-RESTORE (Full Build always auto-restores)
    Write-Step "7/7" "Running post-deploy tasks with auto-restore..."
    $null = Invoke-PostDeployTasks -AutoRestore
    Show-ContainerStatus -Spec $Spec | Out-Host
    Show-DeploymentSummary -Spec $Spec
    
    return $true
}