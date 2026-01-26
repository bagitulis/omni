# ============================================
# Docker Repair Module
# Error-specific repair functions
# SRP: Only repair/fix operations
# ============================================

# Note: config.ps1, logger.ps1, and docker-core.ps1 should be loaded by the caller

# ============================================
# Engine & Pipe Repairs
# ============================================

function Repair-DockerEngineError {
    <#
    .SYNOPSIS
    Fix Docker Desktop engine 500 Internal Server Error
    #>
    Write-Fix "Repairing Docker Desktop engine (500 error)..."
    
    try {
        Write-Info "Stopping all Docker processes..."
        Stop-AllDockerProcesses -WaitSeconds 3
        
        Write-Info "Stopping Docker services..."
        Stop-Service "com.docker.service" -Force -ErrorAction SilentlyContinue
        Stop-Service "docker" -Force -ErrorAction SilentlyContinue
        Start-Sleep -Seconds 2
        
        Write-Info "Shutting down WSL..."
        Stop-WSL -WaitSeconds 5
        
        Write-Info "Starting Docker Desktop..."
        if (-not (Start-DockerDesktop)) {
            return $false
        }
        
        # Give Docker Desktop time to initialize before checking
        Write-Info "Waiting for Docker Desktop initialization..."
        Start-Sleep -Seconds 15
        
        if (Wait-ForDocker -TimeoutSeconds 150 -Activity "Waiting for Docker engine..." -VerifyMode) {
            # Additional stabilization after engine reports ready
            Write-Info "Engine ready, allowing stabilization..."
            Start-Sleep -Seconds 10
            Write-Success "Docker Desktop engine recovered successfully"
            return $true
        }
        
        Write-Warning "Docker engine recovery timeout - may need Windows restart"
        return $false
    }
    catch {
        Write-Error "Failed to repair Docker engine: $_"
        return $false
    }
}

function Repair-DockerPipeError {
    <#
    .SYNOPSIS
    Fix Docker named pipe connection errors
    #>
    Write-Fix "Repairing Docker pipe connection..."
    
    try {
        Write-Info "Stopping all Docker processes..."
        Stop-AllDockerProcesses -WaitSeconds 5
        
        Write-Info "Shutting down WSL completely..."
        Stop-WSL -WaitSeconds 8
        
        Write-Info "Terminating Docker WSL distros..."
        Stop-DockerWSLDistros
        
        Write-Info "Starting Docker Desktop..."
        if (-not (Start-DockerDesktop)) {
            return $false
        }
        
        # Wait for Docker Desktop process to be fully up
        Write-Info "Waiting for Docker Desktop process..."
        Start-Sleep -Seconds 20
        
        Write-Info "Waiting for Docker pipe to be ready..."
        if (Wait-ForDocker -TimeoutSeconds 120 -Activity "Waiting for Docker pipe...") {
            # Verify pipe is stable with a test command
            Write-Info "Verifying pipe stability..."
            Start-Sleep -Seconds 5
            $testResult = docker ps 2>&1 | Out-String
            if ($LASTEXITCODE -eq 0 -and $testResult -notmatch "error|pipe") {
                Write-Success "Docker pipe connection restored"
                return $true
            }
            Write-Warning "Pipe connected but unstable, allowing more time..."
            Start-Sleep -Seconds 15
        }
        
        Write-Warning "Docker pipe may still be unstable - continuing anyway"
        return $true  # Allow retry
    }
    catch {
        Write-Error "Failed to repair Docker pipe: $_"
        return $false
    }
}

function Repair-DockerDNSError {
    <#
    .SYNOPSIS
    Fix DNS resolution errors inside Docker containers
    Handles: "no such host", "dial tcp: lookup", registry connection failures
    #>
    Write-Fix "Repairing Docker DNS/network configuration..."
    
    try {
        Write-Info "Flushing system DNS cache..."
        ipconfig /flushdns 2>&1 | Out-Null
        
        Write-Info "Pruning Docker networks..."
        docker network prune -f 2>&1 | Out-Null
        
        Write-Info "Resetting Docker services..."
        Stop-AllDockerProcesses -WaitSeconds 3
        
        Write-Info "Resetting network stack..."
        netsh winsock reset 2>&1 | Out-Null
        netsh int ip reset 2>&1 | Out-Null
        
        Write-Info "Resetting WSL network..."
        Stop-WSL -WaitSeconds 5
        
        # Additional wait for network interfaces to stabilize
        Write-Info "Waiting for network interfaces to stabilize..."
        Start-Sleep -Seconds 10
        
        Write-Info "Starting Docker Desktop..."
        if (-not (Start-DockerDesktop)) {
            return $false
        }
        
        if (Wait-ForDocker -TimeoutSeconds 90 -Activity "Waiting for Docker DNS...") {
            # Test actual DNS resolution
            Write-Info "Testing network connectivity..."
            Start-Sleep -Seconds 5
            $pingTest = Test-NetConnection -ComputerName "registry-1.docker.io" -Port 443 -WarningAction SilentlyContinue -ErrorAction SilentlyContinue
            if ($pingTest.TcpTestSucceeded) {
                Write-Success "Docker DNS and network restored"
                return $true
            }
            else {
                Write-Warning "DNS restored but registry still unreachable - may need VPN check or retry"
                # Still return true to allow retry
                return $true
            }
        }
        
        Write-Warning "Docker DNS recovery timeout - continuing anyway"
        return $true
    }
    catch {
        Write-Error "Failed to repair Docker DNS: $_"
        return $false
    }
}

# ============================================
# WSL & Mount Repairs
# ============================================

function Repair-WslMountCache {
    <#
    .SYNOPSIS
    Fix WSL2 mount cache corruption issue
    #>
    Write-Fix "Repairing WSL2 mount cache corruption..."
    
    try {
        Write-Info "Stopping all running containers..."
        docker stop $(docker ps -aq) 2>&1 | Out-Null
        Start-Sleep -Seconds 2
        
        Write-Info "Stopping Docker Desktop..."
        Stop-AllDockerProcesses -WaitSeconds 3
        
        Write-Info "Terminating WSL mount cache..."
        Stop-WSL -WaitSeconds 5
        Stop-DockerWSLDistros
        
        Write-Info "Restarting Docker Desktop..."
        if (-not (Start-DockerDesktop)) {
            return $false
        }
        
        if (Wait-ForDocker -TimeoutSeconds 90 -Activity "Waiting for mount cache rebuild...") {
            Write-Success "WSL mount cache fixed"
            return $true
        }
        
        Write-Warning "Mount cache may still have issues - consider restarting Windows"
        return $false
    }
    catch {
        Write-Error "Failed to repair WSL mount cache: $_"
        return $false
    }
}

function Repair-WslKernelError {
    <#
    .SYNOPSIS
    Fix WSL2 kernel or distribution errors
    #>
    Write-Fix "Repairing WSL2 kernel/distribution..."
    
    try {
        Write-Info "Shutting down all WSL instances..."
        Stop-WSL -WaitSeconds 5
        
        Write-Info "Attempting WSL update..."
        wsl --update 2>&1 | Out-Null
        Start-Sleep -Seconds 3
        
        return Restart-DockerDesktop
    }
    catch {
        Write-Error "Failed to repair WSL: $_"
        return $false
    }
}

# ============================================
# Network & Service Repairs
# ============================================

function Repair-NetworkError {
    <#
    .SYNOPSIS
    Fix Docker network issues
    #>
    Write-Fix "Fixing Docker network issues..."
    
    try {
        docker network prune -f 2>&1 | Out-Null
        docker network rm omni_default 2>&1 | Out-Null
        Write-Success "Docker networks cleaned"
        return $true
    }
    catch {
        Write-Warning "Network cleanup had issues, continuing..."
        return $true
    }
}

function Repair-HyperVError {
    <#
    .SYNOPSIS
    Fix Hyper-V virtualization issues
    #>
    Write-Fix "Repairing Hyper-V issues..."
    
    try {
        Write-Info "Restarting Hyper-V services..."
        Restart-Service "vmcompute" -Force -ErrorAction SilentlyContinue
        Start-Sleep -Seconds 3
        
        Restart-Service "hns" -Force -ErrorAction SilentlyContinue
        Start-Sleep -Seconds 2
        
        return Restart-DockerDesktop
    }
    catch {
        Write-Warning "Could not restart Hyper-V services: $_"
        return Restart-DockerDesktop
    }
}

function Repair-DockerBuildKit {
    <#
    .SYNOPSIS
    Fix BuildKit issues
    #>
    Write-Fix "Resetting Docker BuildKit..."
    
    try {
        docker builder prune -af 2>&1 | Out-Null
        Write-Success "BuildKit cache cleared"
        return $true
    }
    catch {
        Write-Error "Failed to clear BuildKit: $_"
        return $false
    }
}

# ============================================
# Container Health
# ============================================

function Repair-ContainerHealth {
    <#
    .SYNOPSIS
    Fix unhealthy containers by restarting them
    #>
    param(
        [string[]]$Containers = @("omni-backend", "omni-frontend", "omni-nginx")
    )
    
    Write-Fix "Checking container health..."
    
    foreach ($container in $Containers) {
        try {
            $health = docker inspect --format='{{.State.Health.Status}}' $container 2>$null
            $status = docker inspect --format='{{.State.Status}}' $container 2>$null
            
            if ($status -eq "exited" -or $health -eq "unhealthy") {
                Write-Warning "$container is $status/$health - restarting..."
                docker restart $container 2>&1 | Out-Null
                Start-Sleep -Seconds 3
            }
        }
        catch {
            # Container inspection failed - container may not exist
            $null = $_
        }
    }
    
    return $true
}

# ============================================
# Restart Functions
# ============================================

function Restart-DockerDesktop {
    <#
    .SYNOPSIS
    Restart Docker Desktop with proper stabilization
    #>
    Write-Fix "Restarting Docker Desktop..."
    
    Stop-AllDockerProcesses -WaitSeconds 5
    Stop-WSL -WaitSeconds 3
    
    if (-not (Start-DockerDesktop)) {
        return $false
    }
    
    if (Wait-ForDocker -TimeoutSeconds 90 -Activity "Waiting for Docker restart...") {
        Write-Success "Docker Desktop restarted successfully"
        return $true
    }
    
    Write-Error "Timeout waiting for Docker restart"
    return $false
}

function Restart-WslAndDocker {
    <#
    .SYNOPSIS
    Restart WSL and wait for Docker
    #>
    Write-Fix "Restarting WSL..."
    
    Stop-WSL -WaitSeconds 3
    
    if (Wait-ForDocker -TimeoutSeconds $script:WslRestartTimeout -Activity "Waiting for Docker after WSL restart...") {
        Write-Success "WSL restarted, Docker recovered"
        return $true
    }
    
    Write-Warning "Docker not ready after WSL restart, attempting full restart..."
    return Restart-DockerDesktop
}

# Assert-DockerReady function moved to docker-assert.ps1 for SRP

# ============================================
# PostgreSQL Repairs
# ============================================

function Repair-PostgresDataDirectory {
    <#
    .SYNOPSIS
    Fix corrupted PostgreSQL data directory by removing it
    PostgreSQL will reinitialize on next container start
    #>
    Write-Fix "Repairing corrupted PostgreSQL data directory..."
    
    try {
        # Get project root from config
        $pgDataPath = Join-Path $script:ProjectRoot "data\postgres"
        
        if (Test-Path $pgDataPath) {
            Write-Info "Found PostgreSQL data directory: $pgDataPath"
            
            # First, stop the postgres container if running
            Write-Info "Stopping PostgreSQL container..."
            docker stop omni-postgres 2>&1 | Out-Null
            docker rm omni-postgres 2>&1 | Out-Null
            Start-Sleep -Seconds 2
            
            # Backup check - if there's a postmaster.pid, postgres might still be running
            $postmasterPid = Join-Path $pgDataPath "postmaster.pid"
            if (Test-Path $postmasterPid) {
                Write-Warning "PostgreSQL lock file found, removing..."
                Remove-Item $postmasterPid -Force -ErrorAction SilentlyContinue
            }
            
            # Check if data directory is corrupted (missing required dirs)
            $requiredPgDirs = @("pg_notify", "pg_tblspc", "pg_stat", "pg_stat_tmp", "pg_replslot", "pg_twophase", "pg_snapshots", "pg_commit_ts", "pg_dynshmem", "pg_serial")
            $missingDirs = @()
            foreach ($dir in $requiredPgDirs) {
                $dirPath = Join-Path $pgDataPath $dir
                if (-not (Test-Path $dirPath)) {
                    $missingDirs += $dir
                }
            }
            
            if ($missingDirs.Count -gt 0) {
                Write-Warning "Corrupted PostgreSQL data detected! Missing: $($missingDirs -join ', ')"
                Write-Info "Removing corrupted data directory..."
                
                # Remove the entire postgres data directory
                Remove-Item -Path $pgDataPath -Recurse -Force -ErrorAction SilentlyContinue
                
                # Recreate empty directory for Docker volume mount
                New-Item -ItemType Directory -Path $pgDataPath -Force | Out-Null
                
                Write-Success "PostgreSQL data directory cleared - will reinitialize on next start"
                Write-Warning "NOTE: All existing database data has been removed!"
            } else {
                Write-Info "PostgreSQL data directory appears intact"
                # Try just removing lock files
                Get-ChildItem -Path $pgDataPath -Filter "*.pid" | Remove-Item -Force -ErrorAction SilentlyContinue
                Get-ChildItem -Path $pgDataPath -Filter "*.lock" | Remove-Item -Force -ErrorAction SilentlyContinue
            }
        } else {
            Write-Info "PostgreSQL data directory not found - will be created on start"
            # Create the directory
            New-Item -ItemType Directory -Path $pgDataPath -Force | Out-Null
        }
        
        return $true
    }
    catch {
        Write-Error "Failed to repair PostgreSQL data directory: $_"
        return $false
    }
}

function Repair-DependencyFailure {
    <#
    .SYNOPSIS
    Fix dependency service failures (postgres, redis, etc.)
    Checks logs and applies appropriate fixes
    #>
    Write-Fix "Repairing dependency service failure..."
    
    try {
        # Check PostgreSQL logs for specific errors
        $pgLogs = docker logs omni-postgres 2>&1 | Out-String
        
        if ($pgLogs -match "could not open directory|No such file or directory|pg_notify|pg_wal|pg_xact") {
            Write-Info "PostgreSQL data corruption detected in logs"
            return Repair-PostgresDataDirectory
        }
        
        if ($pgLogs -match "FATAL:|database system is shut down") {
            Write-Info "PostgreSQL fatal error detected - clearing data"
            return Repair-PostgresDataDirectory
        }
        
        # Check Redis
        $redisLogs = docker logs omni-redis 2>&1 | Out-String
        if ($redisLogs -match "Can't handle RDB format|FATAL") {
            Write-Warning "Redis data corruption detected - clearing..."
            $redisDataPath = Join-Path $script:ProjectRoot "data\redis"
            if (Test-Path $redisDataPath) {
                docker stop omni-redis 2>&1 | Out-Null
                docker rm omni-redis 2>&1 | Out-Null
                Remove-Item -Path $redisDataPath -Recurse -Force -ErrorAction SilentlyContinue
                New-Item -ItemType Directory -Path $redisDataPath -Force | Out-Null
            }
        }
        
        # Remove all dependency containers and let them restart fresh
        Write-Info "Removing dependency containers..."
        docker rm -f omni-postgres omni-redis omni-pgbouncer omni-pgweb 2>&1 | Out-Null
        
        # Clean up any orphaned networks
        docker network prune -f 2>&1 | Out-Null
        
        Write-Success "Dependency cleanup complete"
        return $true
    }
    catch {
        Write-Error "Failed to repair dependency failure: $_"
        return $false
    }
}
