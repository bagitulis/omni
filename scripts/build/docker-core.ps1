# ============================================
# Docker Core Module
# Core Docker operations and utilities
# SRP: Only basic Docker operations
# ============================================

# Note: config.ps1 and logger.ps1 should be loaded by the caller

# ============================================
# Docker Process Management
# ============================================

function Stop-AllDockerProcesses {
    <#
    .SYNOPSIS
    Stop all Docker-related processes
    #>
    param([int]$WaitSeconds = 3)
    
    $dockerProcesses = @(
        "Docker Desktop",
        "com.docker.backend",
        "com.docker.proxy",
        "com.docker.service",
        "dockerd",
        "docker-compose"
    )
    
    foreach ($proc in $dockerProcesses) {
        Get-Process -Name $proc -ErrorAction SilentlyContinue | 
            Stop-Process -Force -ErrorAction SilentlyContinue
    }
    
    Start-Sleep -Seconds $WaitSeconds
}

function Start-DockerDesktop {
    <#
    .SYNOPSIS
    Start Docker Desktop application with process verification
    #>
    $dockerPath = "C:\Program Files\Docker\Docker\Docker Desktop.exe"
    if (-not (Test-Path $dockerPath)) {
        $dockerPath = "${env:ProgramFiles}\Docker\Docker\Docker Desktop.exe"
    }
    
    if (Test-Path $dockerPath) {
        Start-Process $dockerPath
        
        # Wait for Docker Desktop process to actually start
        $maxWait = 30
        $waited = 0
        while ($waited -lt $maxWait) {
            Start-Sleep -Seconds 2
            $waited += 2
            $proc = Get-Process -Name "Docker Desktop" -ErrorAction SilentlyContinue
            if ($proc) {
                # Process started, wait a bit more for initialization
                Start-Sleep -Seconds 5
                return $true
            }
        }
        
        Write-Warning "Docker Desktop process not detected after ${maxWait}s"
        return $true  # Continue anyway, might still work
    }
    
    Write-Error "Docker Desktop executable not found"
    return $false
}

function Test-DockerReady {
    <#
    .SYNOPSIS
    Test if Docker is running and responding
    Returns: hashtable with Status, ErrorType, Output
    #>
    $result = @{
        Ready = $false
        ErrorType = ""
        Output = ""
    }
    
    $dockerInfo = docker info 2>&1 | Out-String
    $result.Output = $dockerInfo
    
    # Check for specific error patterns (be more precise to avoid false positives)
    if ($dockerInfo -match "500 Internal Server Error|request returned 500") {
        $result.ErrorType = "EngineError"
    }
    elseif ($dockerInfo -match "failed to connect.*pipe|error during connect.*pipe|The system cannot find the file specified|cannot connect to the Docker daemon") {
        $result.ErrorType = "PipeError"
    }
    elseif ($dockerInfo -match "DNS lookup error|DNS.*name does not exist|no such host|dial tcp.*lookup") {
        $result.ErrorType = "DNSError"
    }
    elseif ($LASTEXITCODE -ne 0) {
        $result.ErrorType = "NotRunning"
    }
    else {
        $result.Ready = $true
    }
    
    return $result
}

function Get-DockerMode {
    <#
    .SYNOPSIS
    Get current Docker container mode (linux/windows)
    #>
    $mode = docker version --format '{{.Server.Os}}' 2>&1
    if ($LASTEXITCODE -eq 0 -and $mode -notmatch "error|pipe|500") {
        return $mode
    }
    return $null
}

function Wait-ForDocker {
    <#
    .SYNOPSIS
    Wait for Docker to be ready with timeout
    Uses progressive backoff for better cold-start handling
    #>
    param(
        [int]$TimeoutSeconds = 90,
        [string]$Activity = "Waiting for Docker...",
        [switch]$VerifyMode
    )
    
    $elapsed = 0
    $interval = 3  # Start with shorter interval
    $consecutiveReady = 0  # Need 2 consecutive ready checks
    
    while ($elapsed -lt $TimeoutSeconds) {
        Start-Sleep -Seconds $interval
        $elapsed += $interval
        
        # Progressive backoff - increase interval as we wait longer
        if ($elapsed -gt 30 -and $interval -lt 8) {
            $interval = 8
        }
        
        $status = Test-DockerReady
        if ($status.Ready) {
            $consecutiveReady++
            
            # Require 2 consecutive ready checks for stability
            if ($consecutiveReady -ge 2) {
                if ($VerifyMode) {
                    $mode = Get-DockerMode
                    if ($mode -eq "linux") {
                        return $true
                    }
                }
                else {
                    return $true
                }
            }
        } else {
            $consecutiveReady = 0  # Reset on failure
        }
        
        # Log progress (also ensures $consecutiveReady is used for PSScriptAnalyzer)
        Write-Verbose "Docker check: elapsed=${elapsed}s, consecutiveReady=$consecutiveReady"
        Write-Progress -Current $elapsed -Total $TimeoutSeconds -Activity $Activity -PercentComplete ([math]::Min(100, [math]::Round(($elapsed / $TimeoutSeconds) * 100)))
    }
    
    return $false
}

# ============================================
# WSL Operations
# ============================================

function Stop-WSL {
    <#
    .SYNOPSIS
    Shutdown WSL completely
    #>
    param([int]$WaitSeconds = 5)
    
    wsl --shutdown 2>&1 | Out-Null
    Start-Sleep -Seconds $WaitSeconds
    
    # Kill lingering WSL processes
    Get-Process -Name "wsl", "wslhost" -ErrorAction SilentlyContinue | 
        Stop-Process -Force -ErrorAction SilentlyContinue
}

function Stop-DockerWSLDistros {
    <#
    .SYNOPSIS
    Terminate Docker-specific WSL distributions
    #>
    wsl --terminate docker-desktop 2>&1 | Out-Null
    wsl --terminate docker-desktop-data 2>&1 | Out-Null
    Start-Sleep -Seconds 3
}

# ============================================
# Docker Mode Switching
# ============================================

function Switch-DockerToLinux {
    <#
    .SYNOPSIS
    Switch Docker Desktop to Linux container mode
    #>
    Write-Fix "Switching Docker to Linux container mode..."
    
    if (-not (Test-Path $script:DockerCliPath)) {
        Write-Error "Docker Desktop CLI not found at $script:DockerCliPath"
        return $false
    }
    
    try {
        & $script:DockerCliPath -SwitchLinuxEngine 2>&1 | Out-Null
        
        $timeout = 30
        $elapsed = 0
        while ($elapsed -lt $timeout) {
            Start-Sleep -Seconds 2
            $elapsed += 2
            
            $mode = Get-DockerMode
            if ($mode -eq "linux") {
                Write-Success "Successfully switched to Linux container mode"
                return $true
            }
            
            Write-Progress -Current $elapsed -Total $timeout -Activity "Switching to Linux mode..."
        }
        
        Write-Error "Timeout switching to Linux mode"
        return $false
    }
    catch {
        Write-Error "Failed to switch Docker mode: $_"
        return $false
    }
}

# ============================================
# Memory & Resource Checks
# ============================================

function Test-AvailableMemory {
    <#
    .SYNOPSIS
    Check if there's enough available memory
    #>
    param([int]$MinMemoryMB = 1024)
    
    try {
        $os = Get-CimInstance -ClassName Win32_OperatingSystem
        $freeMemoryMB = [math]::Round($os.FreePhysicalMemory / 1024)
        
        Write-Debug "Available memory: ${freeMemoryMB}MB"
        
        if ($freeMemoryMB -lt $MinMemoryMB) {
            Write-Warning "Low memory: ${freeMemoryMB}MB available (need ${MinMemoryMB}MB)"
            return $false
        }
        return $true
    }
    catch {
        Write-Debug "Could not check memory: $_"
        return $true
    }
}

function Invoke-MemoryCleanup {
    <#
    .SYNOPSIS
    Free up memory before Docker operations
    #>
    Write-Fix "Freeing up system memory..."
    
    try {
        # Stop unused containers
        $stoppedContainers = docker ps -aq --filter "status=exited" 2>$null
        if ($stoppedContainers) {
            docker rm $stoppedContainers 2>&1 | Out-Null
        }
        
        # Remove unused images
        docker image prune -f 2>&1 | Out-Null
        
        # Clear build cache (partial)
        docker builder prune -f --keep-storage 1GB 2>&1 | Out-Null
        
        # Request garbage collection
        [System.GC]::Collect()
        
        Write-Success "Memory cleanup completed"
        return $true
    }
    catch {
        Write-Warning "Memory cleanup had issues: $_"
        return $false
    }
}

# ============================================
# Port Management
# ============================================

function Repair-PortConflict {
    <#
    .SYNOPSIS
    Kill processes using required ports
    #>
    Write-Fix "Checking for port conflicts..."
    
    $portsToCheck = $script:RequiredPorts
    $fixed = $false
    $systemProcesses = @("System", "docker", "dockerd", "com.docker.backend", "svchost")
    
    foreach ($port in $portsToCheck) {
        try {
            $conn = Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue | 
                    Select-Object -First 1
            
            if ($conn) {
                $procInfo = Get-Process -Id $conn.OwningProcess -ErrorAction SilentlyContinue
                $procName = if ($procInfo) { $procInfo.ProcessName } else { "Unknown" }
                
                if ($procName -notin $systemProcesses) {
                    Write-Warning "Port $port in use by $procName - killing..."
                    Stop-Process -Id $conn.OwningProcess -Force -ErrorAction SilentlyContinue
                    $fixed = $true
                    Start-Sleep -Seconds 1
                }
            }
        }
        catch {
            # Process may have already terminated - safe to ignore
            $null = $_
        }
    }
    
    if ($fixed) {
        Write-Success "Port conflicts resolved"
        Start-Sleep -Seconds 2
    }
    else {
        Write-Info "No conflicting processes found"
    }
    
    return $true
}

function Repair-ContainerNameConflict {
    <#
    .SYNOPSIS
    Fix container name conflict by removing orphaned containers
    #>
    Write-Fix "Fixing container name conflict..."
    
    $projectContainers = @(
        $script:ContainerBackend,
        $script:ContainerFrontend,
        $script:ContainerNginx,
        $script:ContainerRedis,
        "omni-postgres"
    )
    
    foreach ($container in $projectContainers) {
        try {
            # Check if container exists (running or stopped)
            $exists = docker ps -aq --filter "name=^${container}$" 2>$null
            if ($exists) {
                Write-Info "Removing conflicting container: $container"
                docker stop $container 2>&1 | Out-Null
                docker rm -f $container 2>&1 | Out-Null
            }
        }
        catch {
            Write-Debug "Could not remove container $container : $_"
        }
    }
    
    # Also clean up any orphaned containers from compose
    try {
        docker-compose -f $script:DockerComposeTunnel down --remove-orphans 2>&1 | Out-Null
    }
    catch {
        Write-Debug "Compose down had issues: $_"
    }
    
    Write-Success "Container name conflicts resolved"
    Start-Sleep -Seconds 2
    return $true
}
