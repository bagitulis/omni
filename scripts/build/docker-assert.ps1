# ============================================
# Docker Assert Module
# SRP: Docker readiness verification
# ============================================

# Note: config.ps1, logger.ps1, docker-core.ps1, and docker-repair.ps1 should be loaded by the caller

function Assert-DockerReady {
    <#
    .SYNOPSIS
    Ensure Docker is ready for build (running + Linux mode)
    #>
    Write-Step "PRE-CHECK" "Ensuring Docker is ready..."
    
    # First, check if Docker Desktop process is running
    $dockerProcess = Get-Process -Name "Docker Desktop" -ErrorAction SilentlyContinue
    if (-not $dockerProcess) {
        Write-Info "Docker Desktop not running, starting it..."
        Start-DockerDesktop | Out-Null
        # Cold start needs more time
        Write-Info "Waiting for Docker Desktop cold start (this may take 60-90 seconds)..."
        if (-not (Wait-ForDocker -TimeoutSeconds 120 -Activity "Docker Desktop starting...")) {
            Write-Warning "Docker Desktop took too long to start, will retry..."
        }
    } else {
        # Docker Desktop is running - just wait for engine to be ready
        # DON'T immediately repair, just wait first
        Write-Info "Docker Desktop process found, waiting for engine..."
        $waitResult = Wait-ForDocker -TimeoutSeconds 60 -Activity "Waiting for Docker engine..."
        if ($waitResult) {
            Write-Success "Docker is ready (running in Linux mode)"
            return $true
        }
        Write-Warning "Docker engine not responding after wait, will attempt repair..."
    }
    
    $maxAttempts = 3
    $lastErrorType = ""
    
    for ($attempt = 1; $attempt -le $maxAttempts; $attempt++) {
        if ($attempt -gt 1) {
            # Longer wait between attempts - Docker needs time to stabilize
            $waitTime = 15 + ($attempt * 5)  # 20s, 25s
            Write-Info "Waiting ${waitTime}s for Docker to stabilize..."
            Start-Sleep -Seconds $waitTime
        }
        
        $status = Test-DockerReady
        
        # Escalate if same error persists
        $escalate = ($status.ErrorType -eq $lastErrorType -and $lastErrorType -ne "")
        $lastErrorType = $status.ErrorType
        
        switch ($status.ErrorType) {
            "EngineError" {
                Write-Warning "Docker engine error detected (attempt $attempt/$maxAttempts)..."
                if ($attempt -lt $maxAttempts) {
                    Repair-DockerEngineError | Out-Null
                    continue
                }
                return $false
            }
            "PipeError" {
                Write-Warning "Docker pipe error detected (attempt $attempt/$maxAttempts)..."
                if ($attempt -lt $maxAttempts) {
                    if ($escalate) {
                        Repair-DockerEngineError | Out-Null
                    } else {
                        Repair-DockerPipeError | Out-Null
                    }
                    continue
                }
                Write-Warning "Docker pipe may be unstable - continuing..."
            }
            "DNSError" {
                Write-Warning "Docker DNS error detected (attempt $attempt/$maxAttempts)..."
                if ($attempt -lt $maxAttempts) {
                    Repair-DockerDNSError | Out-Null
                    continue
                }
            }
            "NotRunning" {
                Write-Warning "Docker not running (attempt $attempt/$maxAttempts)..."
                if ($attempt -lt $maxAttempts) {
                    Restart-DockerDesktop | Out-Null
                    continue
                }
                return $false
            }
            "" {
                # Docker is ready, verify mode
                $mode = Get-DockerMode
                if ($mode -ne "linux") {
                    Write-Warning "Docker in $mode mode, switching to Linux..."
                    if (-not (Switch-DockerToLinux)) {
                        if ($attempt -lt $maxAttempts) { continue }
                        return $false
                    }
                }
                
                # Warm-up: Verify Docker is truly ready with a simple command
                Write-Info "Verifying Docker responsiveness..."
                $warmupResult = docker ps 2>&1 | Out-String
                if ($LASTEXITCODE -ne 0 -or $warmupResult -match "error|pipe|500") {
                    Write-Warning "Docker reported ready but still unstable, waiting..."
                    Start-Sleep -Seconds 10
                    $null = docker ps 2>&1  # Retry check
                    if ($LASTEXITCODE -ne 0) {
                        if ($attempt -lt $maxAttempts) { continue }
                    }
                }
                
                Write-Success "Docker is ready (running in Linux mode)"
                return $true
            }
        }
    }
    
    # Final check
    $finalStatus = Test-DockerReady
    if ($finalStatus.Ready) {
        Write-Warning "Docker may not be fully stable, but attempting to continue..."
        return $true
    }
    
    Write-Error "Docker is not ready after $maxAttempts attempts"
    return $false
}
