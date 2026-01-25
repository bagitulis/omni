# ============================================
# Environment Checker Module
# Checks all prerequisites before build
# ============================================

. "$PSScriptRoot\config.ps1"
. "$PSScriptRoot\logger.ps1"

function Test-DockerRunning {
    <#
    .SYNOPSIS
    Check if Docker daemon is running
    #>
    try {
        $null = docker info 2>&1
        if ($LASTEXITCODE -eq 0) {
            return $true
        }
        return $false
    }
    catch {
        return $false
    }
}

function Get-DockerContainerMode {
    <#
    .SYNOPSIS
    Get current Docker container mode (linux/windows)
    #>
    try {
        $os = docker version --format '{{.Server.Os}}' 2>&1
        if ($LASTEXITCODE -eq 0) {
            return $os.Trim()
        }
        return $null
    }
    catch {
        return $null
    }
}

function Test-DockerLinuxMode {
    <#
    .SYNOPSIS
    Check if Docker is in Linux container mode
    #>
    $mode = Get-DockerContainerMode
    return $mode -eq "linux"
}

function Test-WslRunning {
    <#
    .SYNOPSIS
    Check if WSL is running
    #>
    try {
        $null = wsl --status 2>&1
        return $LASTEXITCODE -eq 0
    }
    catch {
        return $false
    }
}

function Test-WslMountHealth {
    <#
    .SYNOPSIS
    Check if WSL mount is working properly (prevents "mkdir /run/desktop/mnt/host/c: file exists" error)
    This is a proactive check to detect mount cache corruption before deployment fails
    #>
    try {
        # Quick test: try to mount a volume with Docker
        $testDir = $script:ProjectRoot -replace '\\', '/'
        $testResult = docker run --rm -v "${testDir}:/test" alpine:latest ls /test 2>&1
        
        if ($LASTEXITCODE -eq 0) {
            return @{
                Healthy = $true
                Error = $null
            }
        }
        
        # Check for the specific mount error
        $errorStr = $testResult | Out-String
        if ($errorStr -match "mkdir /run/desktop/mnt/host|creating mount source path|file exists") {
            return @{
                Healthy = $false
                Error = "WSL mount cache corrupted"
            }
        }
        
        return @{
            Healthy = $false
            Error = "Mount test failed: $errorStr"
        }
    }
    catch {
        return @{
            Healthy = $false
            Error = $_.Exception.Message
        }
    }
}

function Get-AvailableDiskSpace {
    <#
    .SYNOPSIS
    Get available disk space in GB for the project drive
    #>
    param([string]$Path = $script:ProjectRoot)
    
    $drive = (Get-Item $Path).PSDrive.Name
    $disk = Get-PSDrive $drive
    return [math]::Round($disk.Free / 1GB, 2)
}

function Test-PortAvailable {
    <#
    .SYNOPSIS
    Check if a port is available
    #>
    param([int]$Port)
    
    try {
        $listener = [System.Net.Sockets.TcpListener]::new([System.Net.IPAddress]::Loopback, $Port)
        $listener.Start()
        $listener.Stop()
        return $true
    }
    catch {
        # Port might be in use by our containers which is fine
        $containerUsing = docker ps --format "{{.Ports}}" 2>&1 | Select-String ":$Port->"
        if ($containerUsing) {
            return $true  # Our container is using it
        }
        return $false
    }
}

function Test-RequiredPorts {
    <#
    .SYNOPSIS
    Check all required ports
    #>
    $results = @{}
    foreach ($port in $script:RequiredPorts) {
        $results[$port] = Test-PortAvailable -Port $port
    }
    return $results
}

function Test-RequiredDirectories {
    <#
    .SYNOPSIS
    Ensure required directories exist
    #>
    $created = @()
    foreach ($dir in $script:RequiredDirs) {
        $fullPath = Join-Path $script:ProjectRoot $dir
        if (-not (Test-Path $fullPath)) {
            New-Item -ItemType Directory -Path $fullPath -Force | Out-Null
            $created += $dir
        }
    }
    return $created
}

function Test-DockerDesktopInstalled {
    <#
    .SYNOPSIS
    Check if Docker Desktop is installed
    #>
    return Test-Path $script:DockerCliPath
}

function Get-EnvironmentStatus {
    <#
    .SYNOPSIS
    Get full environment status report
    #>
    $status = @{
        DockerInstalled = Test-DockerDesktopInstalled
        DockerRunning = Test-DockerRunning
        DockerLinuxMode = $false
        DockerMode = $null
        WslRunning = Test-WslRunning
        WslMountHealthy = $true
        WslMountError = $null
        DiskSpaceGB = Get-AvailableDiskSpace
        DiskSpaceOK = $false
        PortsAvailable = @{}
        AllPortsOK = $true
        DirectoriesCreated = @()
        AllChecksPass = $false
    }
    
    if ($status.DockerRunning) {
        $status.DockerMode = Get-DockerContainerMode
        $status.DockerLinuxMode = $status.DockerMode -eq "linux"
        
        # Check WSL mount health (critical for deployment)
        $mountCheck = Test-WslMountHealth
        $status.WslMountHealthy = $mountCheck.Healthy
        $status.WslMountError = $mountCheck.Error
    }
    
    $status.DiskSpaceOK = $status.DiskSpaceGB -ge $script:MinDiskSpaceGB
    $status.PortsAvailable = Test-RequiredPorts
    $status.AllPortsOK = ($status.PortsAvailable.Values | Where-Object { -not $_ }).Count -eq 0
    $status.DirectoriesCreated = Test-RequiredDirectories
    
    # Overall status
    $status.AllChecksPass = $status.DockerInstalled -and 
                            $status.DockerRunning -and 
                            $status.DockerLinuxMode -and 
                            $status.WslMountHealthy -and
                            $status.DiskSpaceOK -and 
                            $status.AllPortsOK
    
    return $status
}

function Show-EnvironmentStatus {
    <#
    .SYNOPSIS
    Display environment status to console
    #>
    param([hashtable]$Status)
    
    Write-Header "Environment Check"
    
    # Docker Desktop
    if ($Status.DockerInstalled) {
        Write-Success "Docker Desktop installed"
    } else {
        Write-Error "Docker Desktop not found at $script:DockerCliPath"
    }
    
    # Docker Running
    if ($Status.DockerRunning) {
        Write-Success "Docker daemon is running"
    } else {
        Write-Error "Docker daemon is not running"
    }
    
    # Docker Mode
    if ($Status.DockerLinuxMode) {
        Write-Success "Docker is in Linux container mode"
    } elseif ($Status.DockerMode) {
        Write-Warning "Docker is in $($Status.DockerMode) container mode (needs Linux)"
    } else {
        Write-Warning "Cannot determine Docker container mode"
    }
    
    # WSL
    if ($Status.WslRunning) {
        Write-Success "WSL is running"
    } else {
        Write-Warning "WSL status unknown"
    }
    
    # WSL Mount Health (critical check)
    if ($Status.WslMountHealthy) {
        Write-Success "WSL mount cache is healthy"
    } else {
        Write-Warning "WSL mount cache issue detected: $($Status.WslMountError)"
        Write-Info "This will be auto-fixed before deployment"
    }
    
    # Disk Space
    if ($Status.DiskSpaceOK) {
        Write-Success "Disk space: $($Status.DiskSpaceGB) GB available (min: $script:MinDiskSpaceGB GB)"
    } else {
        Write-Error "Low disk space: $($Status.DiskSpaceGB) GB (need: $script:MinDiskSpaceGB GB)"
    }
    
    # Ports
    foreach ($port in $Status.PortsAvailable.Keys) {
        if ($Status.PortsAvailable[$port]) {
            Write-Success "Port $port is available"
        } else {
            Write-Error "Port $port is in use by another application"
        }
    }
    
    # Directories created
    if ($Status.DirectoriesCreated.Count -gt 0) {
        Write-Info "Created directories: $($Status.DirectoriesCreated -join ', ')"
    }
    
    Write-Host ""
    
    return $Status.AllChecksPass
}
