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
        [bool]$WithN8n = $false,
        [string]$Spec = "standard"
    )
    
    if ($WithN8n) {
        return @("-f", $script:DockerComposeN8n)
    }
    
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
    .PARAMETER WithN8n
    Include n8n compose file in stop operation
    .PARAMETER Spec
    Spec level (used for logging purposes)
    #>
    param(
        [bool]$WithN8n = $false,
        [string]$Spec = "standard"
    )
    
    Write-Step "DOCKER" "Stopping existing containers (spec: $Spec)..."
    # Stop all compose configurations to ensure clean state
    docker-compose -f $script:DockerComposeTunnel down --remove-orphans 2>&1 | Out-Null
    if ($WithN8n) {
        docker-compose -f $script:DockerComposeN8n down --remove-orphans 2>&1 | Out-Null
    }
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
    #>
    param([string]$Line, [ref]$LastStep, [ref]$LastImage, [ref]$ErrorList)
    
    # Capture errors and warnings - always show these
    if ($Line -match "error|ERROR|failed|FAILED|fatal|FATAL" -and $Line -notmatch "0 error|no error|without error") {
        Write-Host ""
        Write-Host "  [ERROR] $Line" -ForegroundColor Red
        if ($null -ne $ErrorList -and $null -ne $ErrorList.Value) {
            $ErrorList.Value += $Line
        }
        return
    }
    
    # Show npm warnings for deprecated/vulnerability
    if ($Line -match "npm warn|npm ERR!|vulnerability|WARN.*deprecated") {
        # Skip noisy warnings, only show important ones
        if ($Line -match "vulnerability|ERR!|EBADENGINE") {
            Write-Host ""
            Write-Host "  [WARN] $Line" -ForegroundColor Yellow
        }
        return
    }
    
    # Track current Docker build step: "#63 [n8n] exporting to image"
    if ($Line -match "^#(\d+)\s+\[([^\]]+)\]\s+(.+)$") {
        $stepNum = $matches[1]
        $context = $matches[2]
        $action = $matches[3]
        $currentStep = "#$stepNum [$context] $action"
        
        # Only update if different step
        if ($currentStep -ne $LastStep.Value) {
            # Truncate long lines
            if ($currentStep.Length -gt 75) {
                $currentStep = $currentStep.Substring(0, 72) + "..."
            }
            Write-Host "`r  $currentStep".PadRight(80) -ForegroundColor Cyan -NoNewline
            $LastStep.Value = $currentStep
        }
        return
    }
    
    # Track image completion
    if ($Line -match "naming to docker.io/library/(omni-\w+):latest done") {
        $LastImage.Value = $matches[1]
        Write-Host ""
        Write-Host "  [DONE] $($matches[1]) built successfully" -ForegroundColor Green
        return
    }
    
    # Show DONE steps briefly
    if ($Line -match "^#(\d+)\s+DONE\s+([\d.]+)s$") {
        # Silent - don't spam DONE messages
        return
    }
}

# ============================================
# Build Function
# ============================================

function Invoke-DockerBuild {
    <#
    .SYNOPSIS
    Build Docker images with auto-recovery
    #>
    param(
        [bool]$WithN8n = $false,
        [string]$Spec = "standard",
        [switch]$NoCache,
        [string[]]$Services = @(),
        [bool]$ShowProgress = $false,
        [bool]$CondensedOutput = $true
    )
    
    Write-Step "DOCKER" "Building Docker images..."
    
    $composeArgs = Get-DockerComposeCommand -WithN8n $WithN8n -Spec $Spec
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
        }
        
        $fullOutput = ""
        $lastStep = ""; $lastImage = ""
        $errorList = [System.Collections.ArrayList]@()
        
        Write-Host "  ─────────────── Build Progress ───────────────" -ForegroundColor Cyan
        
        if ($ShowProgress) {
            & docker-compose @buildArgs 2>&1 | ForEach-Object {
                Write-Host "  $_" -ForegroundColor Gray
                $fullOutput += "$_`n"
            }
        }
        elseif ($CondensedOutput) {
            & docker-compose @buildArgs 2>&1 | ForEach-Object {
                $fullOutput += "$_`n"
                Show-CondensedBuildProgress -Line $_ -LastStep ([ref]$lastStep) -LastImage ([ref]$lastImage) -ErrorList ([ref]$errorList)
            }
            Write-Host ""
        }
        
        Write-Host "  ────────────────────────────────────────────" -ForegroundColor Cyan
        
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
    Progressive fix levels for build failures
    #>
    param([int]$Level)
    
    switch ($Level) {
        1 { Write-Fix "L1: Docker health check..."; Repair-PortConflict }
        2 { Write-Fix "L2: WSL restart..."; wsl --shutdown 2>&1 | Out-Null; Start-Sleep 10 }
        3 { Write-Fix "L3: Memory cleanup..."; Invoke-MemoryCleanup; Invoke-MediumCleanup }
        4 { Write-Fix "L4: Docker restart..."; Repair-DockerEngineError; Start-Sleep 10 }
        5 { Write-Fix "L5: Aggressive cleanup..."; Invoke-AggressiveCleanup; Start-Sleep 10 }
        6 { Write-Fix "L6: Full reset..."; Restart-WslAndDocker; Start-Sleep 30 }
    }
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
        [bool]$WithN8n = $false,
        [string]$Spec = "standard",
        [string[]]$Services = @()
    )
    
    Write-Step "DOCKER" "Deploying containers..."
    
    $composeArgs = Get-DockerComposeCommand -WithN8n $WithN8n -Spec $Spec
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
    Quick restart without rebuilding
    #>
    param([bool]$WithN8n = $false, [string]$Spec = "standard")
    
    Write-Step "DOCKER" "Quick restart..."
    $composeArgs = Get-DockerComposeCommand -WithN8n $WithN8n -Spec $Spec
    
    & docker-compose @($composeArgs + @("restart")) 2>&1 | Out-Null
    if ($LASTEXITCODE -eq 0) { Write-Success "Restarted"; return $true }
    
    & docker-compose @($composeArgs + @("up", "-d")) 2>&1 | Out-Null
    if ($LASTEXITCODE -eq 0) { Write-Success "Started"; return $true }
    
    Write-Error "Failed"; return $false
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
    param([bool]$WithN8n = $false, [string]$Spec = "standard")
    
    Write-Header "Container Status"
    $composeArgs = Get-DockerComposeCommand -WithN8n $WithN8n -Spec $Spec
    & docker-compose @($composeArgs + @("ps"))
}
