# ============================================
# Docker Operations Module (Refactored)
# SRP: Docker compose build and deploy operations
# ============================================

# Note: Dependencies should be loaded by the caller

# ============================================
# Compose Command Helpers
# ============================================

function Get-DockerComposeCommand {
    <#
    .SYNOPSIS
    Get docker-compose arguments based on settings
    #>
    param(
        [string]$Spec = "standard"
    )
    
    $files = @("-f", $script:DockerComposeTunnel)
    switch ($Spec) {
        "lowspec"  { $files += @("-f", $script:DockerComposeLowSpec) }
        "standard" { $files += @("-f", $script:DockerComposeStandard) }
        "highspec" { $files += @("-f", $script:DockerComposeHighSpec) }
    }
    return $files
}

function Stop-ExistingContainers {
    <#
    .SYNOPSIS
    Stop and remove existing containers
    .PARAMETER Spec
    Spec level (used for logging purposes)
    #>
    param(
        [string]$Spec = "standard"
    )
    
    Write-Step "DOCKER" "Stopping existing containers (spec: $Spec)..."
    # Stop all compose configurations to ensure clean state
    docker-compose -f $script:DockerComposeTunnel down --remove-orphans 2>&1 | Out-Null
    Write-Success "Existing containers stopped"
}

# ============================================
# Condensed Output Parser
# ============================================

function Show-CondensedBuildProgress {
    <#
    .SYNOPSIS
    Parse and display minimal build output - only current step + errors
    Shows: Current docker step being executed, errors/warnings
    Output updates IN-PLACE using carriage return
    #>
    param([string]$Line, [ref]$LastStep, [ref]$LastImage, [ref]$ErrorList)
    
    # Trim leading whitespace for matching
    $trimmedLine = $Line.TrimStart()
    
    # Capture errors and warnings - always show these on new line
    if ($trimmedLine -match "error|ERROR|failed|FAILED|fatal|FATAL" -and $trimmedLine -notmatch "0 error|no error|without error|transferring") {
        Write-Host ""
        Write-Host "  [ERROR] $trimmedLine" -ForegroundColor Red
        if ($null -ne $ErrorList -and $null -ne $ErrorList.Value) {
            $ErrorList.Value += $trimmedLine
        }
        return
    }
    
    # Show npm errors only (skip noisy warnings)
    if ($trimmedLine -match "npm ERR!") {
        Write-Host ""
        Write-Host "  [NPM ERROR] $trimmedLine" -ForegroundColor Red
        return
    }
    
    # Track current Docker build step: "#63 [n8n] exporting to image" or "#63 [n8n 2/9] RUN..."
    if ($trimmedLine -match "^#(\d+)\s+\[([^\]]+)\]\s+(.+)$") {
        $stepNum = $matches[1]
        $context = $matches[2]
        $action = $matches[3]
        
        # Skip intermediate progress lines like "transferring context"
        if ($action -match "^transferring|^reading|^resolve") {
            return
        }
        
        $currentStep = "#$stepNum [$context] $action"
        
        # Only update if different step
        if ($currentStep -ne $LastStep.Value) {
            # Truncate long lines
            if ($currentStep.Length -gt 70) {
                $currentStep = $currentStep.Substring(0, 67) + "..."
            }
            # Update in place with carriage return
            Write-Host "`r  $currentStep".PadRight(80) -ForegroundColor Cyan -NoNewline
            $LastStep.Value = $currentStep
        }
        return
    }
    
    # Track image completion - show on new line
    if ($trimmedLine -match "naming to docker.io/library/(omni-\w+):latest done") {
        $LastImage.Value = $matches[1]
        Write-Host ""
        Write-Host "  ✓ $($matches[1]) built" -ForegroundColor Green
        return
    }
    
    # Silent - ignore other lines (DONE, auth, internal steps, etc.)
}

# ============================================
# Build Function
# ============================================

function Invoke-DockerBuild {
    <#
    .SYNOPSIS
    Build Docker images with auto-recovery and robust error handling
    #>
    param(
        [string]$Spec = "standard",
        [switch]$NoCache,
        [string[]]$Services = @(),
        [bool]$ShowProgress = $false,
        [bool]$CondensedOutput = $true
    )
    
    Write-Step "DOCKER" "Building Docker images..."
    
    # Pre-build Docker readiness check
    Write-Info "Verifying Docker is ready for build..."
    $dockerCheck = docker info 2>&1 | Out-String
    if ($LASTEXITCODE -ne 0 -or $dockerCheck -match "error|Cannot connect|pipe") {
        Write-Warning "Docker not ready, attempting recovery..."
        if (-not (Assert-DockerReady)) {
            Write-Error "Docker is not available for build"
            return $false
        }
    }
    
    $composeArgs = Get-DockerComposeCommand -Spec $Spec
    $buildArgs = $composeArgs + @("build", "--progress=plain")
    if ($NoCache) { $buildArgs += "--no-cache" }
    if ($Services.Count -gt 0) { $buildArgs += $Services }
    
    $maxAttempts = 7
    $attempt = 0
    
    while ($attempt -lt $maxAttempts) {
        $attempt++
        Write-Info "Build attempt $attempt of $maxAttempts..."
        
        # Apply progressive fix on retry
        if ($attempt -gt 1) {
            Invoke-ProgressiveBuildFix -Level ($attempt - 1)
            Start-Sleep -Seconds 3
            
            # Re-verify Docker readiness after fix
            $dockerRecheck = docker info 2>&1 | Out-String
            if ($LASTEXITCODE -ne 0 -or $dockerRecheck -match "error|Cannot connect|pipe") {
                Write-Warning "Docker still not ready after fix, waiting..."
                if (-not (Assert-DockerReady)) {
                    Write-Warning "Docker recovery failed, trying next fix level..."
                    continue
                }
            }
        }
        
        $fullOutput = [System.Text.StringBuilder]::new()
        $lastStep = ""; $lastImage = ""
        $errorList = [System.Collections.ArrayList]@()
        
        Write-Host "  ─────────────── Build Progress ───────────────" -ForegroundColor Cyan
        
        # Note: Docker BuildKit writes directly to console TTY and cannot be fully redirected.
        # We capture output for error detection but display is controlled by Docker itself.
        # Use --quiet for minimal output or --progress=plain for text-only mode.
        
        if ($ShowProgress) {
            # Verbose mode - show all output
            & docker-compose @buildArgs 2>&1 | ForEach-Object {
                Write-Host "  $_" -ForegroundColor Gray
                [void]$fullOutput.AppendLine($_)
            }
        }
        else {
            # Normal mode - let Docker show its progress (no way to suppress BuildKit TTY output)
            # But we still capture output for error detection
            & docker-compose @buildArgs 2>&1 | ForEach-Object {
                [void]$fullOutput.AppendLine($_)
                # Silent - Docker BuildKit handles its own display
            }
            Write-Host ""
        }
        
        $fullOutput = $fullOutput.ToString()
        
        Write-Host "  ────────────────────────────────────────────" -ForegroundColor Cyan
        
        # Check exit code from LASTEXITCODE
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Docker build completed (attempt $attempt)"
            return $true
        }
        
        # Handle errors
        $errorHandled = Resolve-BuildError -Output $fullOutput -Attempt $attempt
        if ($errorHandled -eq "quit") { return $false }
        if ($errorHandled -eq "skip") { continue }
    }
    
    Write-Error "Docker build failed after $maxAttempts attempts"
    return $false
}

function Resolve-BuildError {
    <#
    .SYNOPSIS
    Handle build errors with smart detection
    #>
    param([string]$Output, [int]$Attempt)
    
    Write-Warning "Build failed on attempt $Attempt"
    
    # Code errors - cannot auto-fix
    if (Test-IsCodeError -ErrorMessage $Output) {
        $action = Show-CodeErrorPause -ErrorMessage $Output -ErrorType "Build"
        return $action
    }
    
    # DNS/Network errors
    if ($Output -match "DNS lookup error|DNS.*name does not exist|fetch.*error") {
        Write-Fix "DNS error detected - repairing..."
        Repair-DockerDNSError
        Start-Sleep -Seconds 10
        return "continue"
    }
    
    # Docker engine errors
    if ($Output -match "500 Internal Server Error|request returned 500") {
        Write-Fix "Docker engine error - repairing..."
        Repair-DockerEngineError
        Start-Sleep -Seconds 15
        return "continue"
    }
    
    # Pipe errors
    if ($Output -match "pipe.*docker|npipe.*error") {
        Write-Fix "Docker pipe error - repairing..."
        Repair-DockerPipeError
        Start-Sleep -Seconds 10
        return "continue"
    }
    
    # WSL mount errors
    if ($Output -match "mkdir /run/desktop/mnt/host|file exists.*mnt") {
        Write-Fix "WSL mount error - repairing..."
        Repair-WslMountCache
        Start-Sleep -Seconds 10
        return "continue"
    }
    
    return "continue"
}

function Invoke-ProgressiveBuildFix {
    <#
    .SYNOPSIS
    Progressive fix levels for build failures with Docker readiness verification
    #>
    param([int]$Level)
    
    Write-Info "Applying fix level $Level..."
    
    switch ($Level) {
        1 { 
            Write-Fix "L1: Docker health check and port repair..."
            Repair-PortConflict
            Start-Sleep -Seconds 3
        }
        2 { 
            Write-Fix "L2: WSL restart..."
            wsl --shutdown 2>&1 | Out-Null
            Start-Sleep -Seconds 10
            # Wait for Docker to recover
            $null = Assert-DockerReady
        }
        3 { 
            Write-Fix "L3: Memory cleanup..."
            Invoke-MemoryCleanup
            Invoke-MediumCleanup
            Start-Sleep -Seconds 5
        }
        4 { 
            Write-Fix "L4: Docker engine restart..."
            Repair-DockerEngineError | Out-Null
            Start-Sleep -Seconds 15
            # Verify Docker is back
            $null = Assert-DockerReady
        }
        5 { 
            Write-Fix "L5: Aggressive cleanup..."
            Invoke-AggressiveCleanup
            Start-Sleep -Seconds 10
        }
        6 { 
            Write-Fix "L6: Full WSL and Docker reset..."
            Restart-WslAndDocker
            Start-Sleep -Seconds 30
            # Must verify Docker is ready after full reset
            if (-not (Assert-DockerReady)) {
                Write-Warning "Docker may not be fully ready after reset"
            }
        }
    }
    
    Write-Info "Fix level $Level applied"
}

# ============================================
# Deploy Function
# ============================================

function Invoke-DockerDeploy {
    <#
    .SYNOPSIS
    Deploy containers with auto-recovery
    #>
    param(
        [string]$Spec = "standard",
        [string[]]$Services = @()
    )
    
    Write-Step "DOCKER" "Deploying containers..."
    
    $composeArgs = Get-DockerComposeCommand -Spec $Spec
    $upArgs = $composeArgs + @("up", "-d", "--build")
    if ($Services.Count -gt 0) { $upArgs += $Services }
    
    $maxAttempts = $script:MaxRetries
    $attempt = 0
    
    while ($attempt -lt $maxAttempts) {
        $attempt++
        Write-Host "  [DEPLOY] Attempt $attempt of $maxAttempts..." -ForegroundColor Gray
        
        # Capture output for error detection (don't suppress with Out-Null)
        $deployOutput = ""
        & docker-compose @upArgs 2>&1 | ForEach-Object {
            $deployOutput += "$_`n"
            # Show progress for key events
            if ($_ -match "Created|Started|Pulled|Error|error") {
                Write-Host "  $_" -ForegroundColor $(if ($_ -match "Error|error") { "Red" } else { "DarkGray" })
            }
        }
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Deployment completed (attempt $attempt)"
            return $true
        }
        
        # Show error details
        Write-Warning "Deployment failed on attempt $attempt"
        
        # Check for code errors that cannot be auto-fixed
        if (Test-IsCodeError -ErrorMessage $deployOutput) {
            $action = Show-CodeErrorPause -ErrorMessage $deployOutput -ErrorType "Deploy"
            if ($action -eq "quit") { return $false }
            if ($action -eq "skip") { continue }
        }
        
        # Enhanced: Check container logs for dependency failures
        if ($deployOutput -match "dependency.*failed to start|postgres.*failed") {
            Write-Info "Checking PostgreSQL container logs..."
            $pgLogs = docker logs omni-postgres 2>&1 | Out-String
            if ($pgLogs) {
                # Append PostgreSQL logs to deploy output for better error detection
                $deployOutput += "`n[PostgreSQL Logs]:`n$pgLogs"
            }
        }
        
        if ($attempt -lt $maxAttempts) {
            Write-Retry -Attempt $attempt -MaxAttempts $maxAttempts -Message "Retrying..."
            # Pass error message so specific error types can be detected and fixed
            Invoke-ErrorRecovery -ErrorMessage $deployOutput -AttemptNumber $attempt
            docker-compose @($composeArgs + @("down", "--remove-orphans")) 2>&1 | Out-Null
            Start-Sleep -Seconds 5
        }
    }
    
    Write-Error "Deployment failed after $maxAttempts attempts"
    return $false
}

function Invoke-QuickRestart {
    <#
    .SYNOPSIS
    Quick restart without rebuilding - with comprehensive error handling and auto-fix
    Returns: $true if success, $false if failed (caller should fallback to Smart Build)
    #>
    param([string]$Spec = "standard")
    
    Write-Step "DOCKER" "Quick restart (with auto-fix)..."
    $composeArgs = Get-DockerComposeCommand -Spec $Spec
    $maxAttempts = 3
    
    for ($attempt = 1; $attempt -le $maxAttempts; $attempt++) {
        Write-Info "Attempt $attempt of $maxAttempts..."
        
        # Step 1: Try simple restart first
        if ($attempt -eq 1) {
            $restartOutput = & docker-compose @($composeArgs + @("restart")) 2>&1
            if ($LASTEXITCODE -eq 0) {
                # Verify containers are actually running
                Start-Sleep -Seconds 3
                if (Test-ContainersRunning) {
                    Write-Success "Restarted successfully"
                    return $true
                }
                Write-Warning "Restart command succeeded but containers not running"
            }
        }
        
        # Step 2: Try docker-compose up -d
        Write-Info "Trying docker-compose up -d..."
        $upOutput = & docker-compose @($composeArgs + @("up", "-d")) 2>&1
        if ($LASTEXITCODE -eq 0) {
            Start-Sleep -Seconds 5
            if (Test-ContainersRunning) {
                Write-Success "Started successfully"
                return $true
            }
            Write-Warning "Up command succeeded but containers not running"
        }
        
        # Step 3: Diagnose and auto-fix common issues (NO rebuild, NO clean)
        if ($attempt -lt $maxAttempts) {
            Write-Warning "Quick restart failed, running diagnostics..."
            $fixApplied = Invoke-QuickFixDiagnostics -Spec $Spec -ErrorOutput "$restartOutput $upOutput"
            
            if (-not $fixApplied) {
                Write-Info "No quick fix available, waiting before retry..."
                Start-Sleep -Seconds 3
            }
        }
    }
    
    # All attempts failed - show diagnostic info
    Write-Error "Quick restart failed after $maxAttempts attempts"
    Show-QuickRestartDiagnostics -Spec $Spec
    return $false
}

function Test-ContainersRunning {
    <#
    .SYNOPSIS
    Check if essential containers are running
    #>
    param()
    
    $essentialContainers = @($script:ContainerBackend, $script:ContainerFrontend, "omni-postgres", "omni-redis")
    
    foreach ($container in $essentialContainers) {
        $status = docker inspect --format='{{.State.Status}}' $container 2>$null
        if ($status -ne "running") {
            return $false
        }
    }
    return $true
}

function Invoke-QuickFixDiagnostics {
    <#
    .SYNOPSIS
    Diagnose and fix common issues WITHOUT rebuilding or cleaning images
    Only fixes: network, orphan containers, port conflicts, daemon issues, WSL mount
    #>
    param(
        [string]$Spec = "standard",
        [string]$ErrorOutput = ""
    )
    
    $fixApplied = $false
    $composeArgs = Get-DockerComposeCommand -Spec $Spec
    
    # Try to start a container to check for WSL mount error (common issue)
    $testStartOutput = docker start omni-postgres 2>&1
    $hasWslMountError = $testStartOutput -match "mkdir /run/desktop/mnt/host|file exists.*mnt|mount source path"
    
    # Fix 0: WSL Mount Cache Error (PRIORITY - most common issue)
    if ($hasWslMountError -or $ErrorOutput -match "mkdir /run/desktop/mnt/host|file exists.*mnt|mount source path") {
        Write-Warning "WSL mount cache error detected - this is a common Docker Desktop issue"
        Write-Info "Fixing WSL mount cache (this will take ~30 seconds)..."
        
        # Use the existing Repair-WslMountCache function
        if (Repair-WslMountCache) {
            Write-Success "WSL mount cache fixed"
            $fixApplied = $true
            # Return immediately to retry with fixed WSL
            return $fixApplied
        } else {
            Write-Warning "WSL mount fix may need manual intervention"
        }
    }
    
    # Fix 1: Remove orphan containers (lightweight, no rebuild)
    Write-Info "Removing orphan containers..."
    & docker-compose @($composeArgs + @("down", "--remove-orphans")) 2>&1 | Out-Null
    $fixApplied = $true
    
    # Fix 2: Check and fix network issues
    if ($ErrorOutput -match "network|subnet|already exists|pool overlaps") {
        Write-Info "Fixing network issues..."
        # Remove only orphan networks, not all networks
        docker network prune -f 2>&1 | Out-Null
        $fixApplied = $true
    }
    
    # Fix 3: Check port conflicts
    $portConflicts = @()
    $portsToCheck = @(3000, 80, 5432, 6379)
    
    foreach ($port in $portsToCheck) {
        $inUse = Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue | Where-Object { $_.State -eq "Listen" }
        if ($inUse) {
            $process = Get-Process -Id $inUse.OwningProcess -ErrorAction SilentlyContinue
            if ($process -and $process.ProcessName -notmatch "docker|com.docker") {
                $portConflicts += "Port $port in use by $($process.ProcessName) (PID: $($process.Id))"
            }
        }
    }
    
    if ($portConflicts.Count -gt 0) {
        Write-Warning "Port conflicts detected:"
        $portConflicts | ForEach-Object { Write-Host "  - $_" -ForegroundColor Yellow }
        Write-Info "Please close conflicting applications or change ports"
    }
    
    # Fix 4: Check if Docker daemon is responsive
    $dockerPing = docker info 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Warning "Docker daemon not responding, attempting restart..."
        try {
            # Try to restart Docker Desktop (Windows)
            $dockerProcess = Get-Process "Docker Desktop" -ErrorAction SilentlyContinue
            if ($dockerProcess) {
                Write-Info "Restarting Docker Desktop..."
                Stop-Process -Name "Docker Desktop" -Force -ErrorAction SilentlyContinue
                Start-Sleep -Seconds 5
                Start-Process "$env:ProgramFiles\Docker\Docker\Docker Desktop.exe" -ErrorAction SilentlyContinue
                
                # Wait for Docker to be ready
                $timeout = 60
                $waited = 0
                while ($waited -lt $timeout) {
                    Start-Sleep -Seconds 2
                    $waited += 2
                    $check = docker info 2>&1
                    if ($LASTEXITCODE -eq 0) {
                        Write-Success "Docker Desktop restarted"
                        $fixApplied = $true
                        break
                    }
                    Write-Progress -Activity "Waiting for Docker" -Status "$waited seconds..." -PercentComplete (($waited / $timeout) * 100)
                }
                Write-Progress -Activity "Waiting for Docker" -Completed
            }
        } catch {
            Write-Warning "Could not restart Docker: $_"
        }
    }
    
    # Fix 5: Check container-specific errors (containers stuck in "created" status)
    $createdContainers = docker ps -a --filter "status=created" --format "{{.Names}}" 2>$null
    if ($createdContainers) {
        Write-Info "Found containers stuck in 'created' status, attempting to start..."
        foreach ($container in ($createdContainers -split "`n" | Where-Object { $_ -match "omni-" })) {
            Write-Info "  Starting $container..."
            $startOutput = docker start $container 2>&1
            if ($LASTEXITCODE -ne 0) {
                # Check specific error
                if ($startOutput -match "mkdir /run/desktop/mnt/host|file exists.*mnt") {
                    Write-Warning "  $container has WSL mount error - will be fixed in next attempt"
                } else {
                    Write-Warning "  $container failed: $($startOutput | Select-Object -Last 1)"
                }
            } else {
                Write-Success "  $container started"
            }
        }
    }
    
    # Fix 6: Check exited containers for errors
    $stoppedContainers = docker ps -a --filter "status=exited" --format "{{.Names}}" 2>$null
    if ($stoppedContainers) {
        Write-Info "Found stopped containers, checking logs..."
        foreach ($container in ($stoppedContainers -split "`n" | Where-Object { $_ -match "omni-" })) {
            $lastLog = docker logs --tail 5 $container 2>&1
            if ($lastLog -match "error|fatal|failed") {
                Write-Warning "Container $container has errors:"
                Write-Host "  $($lastLog | Select-Object -Last 2)" -ForegroundColor Gray
            }
        }
    }
    
    return $fixApplied
}

function Show-QuickRestartDiagnostics {
    <#
    .SYNOPSIS
    Show diagnostic information when quick restart fails
    #>
    param([string]$Spec = "standard")
    
    Write-Host ""
    Write-Host "=== Quick Restart Diagnostics ===" -ForegroundColor Cyan
    
    # Show container status
    Write-Host "`nContainer Status:" -ForegroundColor Yellow
    docker ps -a --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" 2>$null | Select-Object -First 10
    
    # Show recent errors from key containers
    $keyContainers = @("omni-postgres", "omni-backend", "omni-frontend", "omni-redis")
    
    Write-Host "`nRecent Errors:" -ForegroundColor Yellow
    foreach ($container in $keyContainers) {
        $exists = docker ps -a --format "{{.Names}}" | Where-Object { $_ -eq $container }
        if ($exists) {
            $status = docker inspect --format='{{.State.Status}}' $container 2>$null
            if ($status -ne "running") {
                Write-Host "  [$container] Status: $status" -ForegroundColor Red
                $logs = docker logs --tail 3 $container 2>&1
                if ($logs) {
                    Write-Host "    Last log: $($logs | Select-Object -Last 1)" -ForegroundColor Gray
                }
            }
        } else {
            Write-Host "  [$container] Not found" -ForegroundColor DarkGray
        }
    }
    
    Write-Host ""
}

function Wait-ForContainerHealth {
    <#
    .SYNOPSIS
    Wait for containers to be healthy
    #>
    param([int]$TimeoutSeconds = 60)
    
    Write-Step "HEALTH" "Waiting for containers..."
    $containers = @($script:ContainerBackend, $script:ContainerFrontend)
    
    $healthy = Wait-ForCondition -Condition {
        foreach ($c in $containers) {
            $status = docker inspect --format='{{.State.Status}}' $c 2>$null
            if ($status -ne "running") { return $false }
        }
        return $true
    } -TimeoutSeconds $TimeoutSeconds -ActivityMessage "Waiting..."
    
    if ($healthy) { Write-Success "All containers running" }
    else { Write-Warning "Some containers unhealthy" }
    return $healthy
}

function Show-ContainerStatus {
    <#
    .SYNOPSIS
    Display container status
    #>
    param([string]$Spec = "standard")
    
    Write-Header "Container Status"
    $composeArgs = Get-DockerComposeCommand -Spec $Spec
    & docker-compose @($composeArgs + @("ps"))
}
