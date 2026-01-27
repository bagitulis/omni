# ============================================
# Docker Diagnostics Tool
# Run this script to diagnose Docker issues
# Usage: .\docker-diagnose.ps1 [-Fix]
# ============================================

param(
    [switch]$Fix  # Attempt to fix issues automatically
)

. "$PSScriptRoot\config.ps1"
. "$PSScriptRoot\logger.ps1"
. "$PSScriptRoot\docker-fixer.ps1"

function Write-Status {
    param(
        [string]$Label,
        [string]$Status,
        [string]$Details = ""
    )
    
    $color = switch ($Status) {
        "OK" { "Green" }
        "WARNING" { "Yellow" }
        "ERROR" { "Red" }
        default { "White" }
    }
    
    $statusText = "[$Status]".PadRight(10)
    Write-Host "  $Label".PadRight(35) -NoNewline
    Write-Host $statusText -ForegroundColor $color -NoNewline
    if ($Details) {
        Write-Host " $Details" -ForegroundColor DarkGray
    } else {
        Write-Host ""
    }
}

function Test-DockerDesktopRunning {
    $process = Get-Process "Docker Desktop" -ErrorAction SilentlyContinue
    return $null -ne $process
}

function Test-DockerDaemon {
    $output = docker info 2>&1 | Out-String
    return @{
        Success = ($LASTEXITCODE -eq 0)
        Output = $output
        Has500Error = ($output -match "500 Internal Server Error")
        HasPipeError = ($output -match "pipe|npipe")
        HasConnectionError = ($output -match "error during connect|connection refused")
    }
}

function Test-DockerVersion {
    $output = docker version 2>&1 | Out-String
    $os = docker version --format '{{.Server.Os}}' 2>&1
    return @{
        Success = ($LASTEXITCODE -eq 0)
        Output = $output
        ServerOS = $os
        IsLinux = ($os -eq "linux")
    }
}

function Test-WslStatus {
    $wslList = wsl --list --verbose 2>&1 | Out-String
    $dockerDesktop = $wslList -match "docker-desktop"
    $dockerDesktopData = $wslList -match "docker-desktop-data"
    
    return @{
        Output = $wslList
        HasDockerDesktop = $dockerDesktop
        HasDockerDesktopData = $dockerDesktopData
    }
}

function Test-HyperVStatus {
    try {
        $vmcompute = Get-Service "vmcompute" -ErrorAction SilentlyContinue
        $hns = Get-Service "hns" -ErrorAction SilentlyContinue
        
        return @{
            VmComputeRunning = ($vmcompute.Status -eq "Running")
            HnsRunning = ($hns.Status -eq "Running")
        }
    }
    catch {
        return @{
            VmComputeRunning = $false
            HnsRunning = $false
        }
    }
}

function Test-DockerPorts {
    $portsToCheck = @(80, 443, 3000, 5678)
    $conflicts = @()
    
    foreach ($port in $portsToCheck) {
        try {
            $conn = Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue | Select-Object -First 1
            if ($conn) {
                $proc = Get-Process -Id $conn.OwningProcess -ErrorAction SilentlyContinue
                $procName = if ($proc) { $proc.ProcessName } else { "Unknown" }
                $conflicts += "$port ($procName)"
            }
        }
        catch {
            # Port check failed - skip this port
            $null = $_
        }
    }
    
    return $conflicts
}

# Main diagnostic routine
Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Docker Desktop Diagnostics" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

$issues = @()

# 1. Check Docker Desktop process
Write-Host "[1/6] Docker Desktop Process" -ForegroundColor Yellow
$dockerRunning = Test-DockerDesktopRunning
if ($dockerRunning) {
    Write-Status "Docker Desktop" "OK" "Process is running"
} else {
    Write-Status "Docker Desktop" "ERROR" "Process not found"
    $issues += @{
        Type = "DockerDesktopNotRunning"
        Fix = { 
            $dockerPath = "C:\Program Files\Docker\Docker\Docker Desktop.exe"
            if (Test-Path $dockerPath) { Start-Process $dockerPath }
            Start-Sleep -Seconds 30
        }
    }
}
Write-Host ""

# 2. Check Docker daemon
Write-Host "[2/6] Docker Daemon" -ForegroundColor Yellow
$daemonStatus = Test-DockerDaemon
if ($daemonStatus.Success) {
    Write-Status "Docker Daemon" "OK" "Responding normally"
} else {
    if ($daemonStatus.Has500Error) {
        Write-Status "Docker Daemon" "ERROR" "500 Internal Server Error - Engine not responding"
        $issues += @{
            Type = "DockerEngine500"
            Fix = { Repair-DockerEngineError }
        }
    }
    elseif ($daemonStatus.HasPipeError) {
        Write-Status "Docker Daemon" "ERROR" "Pipe connection error"
        $issues += @{
            Type = "DockerPipeError"
            Fix = { Repair-DockerPipeError }
        }
    }
    elseif ($daemonStatus.HasConnectionError) {
        Write-Status "Docker Daemon" "ERROR" "Connection error - daemon not running"
        $issues += @{
            Type = "DockerDaemonNotRunning"
            Fix = { Restart-DockerDesktop }
        }
    }
    else {
        Write-Status "Docker Daemon" "ERROR" "Unknown error"
        $issues += @{
            Type = "DockerUnknownError"
            Fix = { Restart-DockerDesktop }
        }
    }
}
Write-Host ""

# 3. Check Docker version/mode
Write-Host "[3/6] Docker Container Mode" -ForegroundColor Yellow
$versionStatus = Test-DockerVersion
if ($versionStatus.Success -and $versionStatus.IsLinux) {
    Write-Status "Container Mode" "OK" "Linux containers"
} elseif ($versionStatus.Success) {
    Write-Status "Container Mode" "WARNING" "Windows containers (should be Linux)"
    $issues += @{
        Type = "WindowsContainerMode"
        Fix = { Switch-DockerToLinux }
    }
} else {
    Write-Status "Container Mode" "ERROR" "Could not determine"
}
Write-Host ""

# 4. Check WSL
Write-Host "[4/6] WSL2 Status" -ForegroundColor Yellow
$wslStatus = Test-WslStatus
if ($wslStatus.HasDockerDesktop -and $wslStatus.HasDockerDesktopData) {
    Write-Status "WSL2 Docker Distros" "OK" "Both distros present"
} else {
    Write-Status "WSL2 Docker Distros" "WARNING" "Some distros missing"
    Write-Host "    docker-desktop: $(if ($wslStatus.HasDockerDesktop) { 'Present' } else { 'Missing' })" -ForegroundColor DarkGray
    Write-Host "    docker-desktop-data: $(if ($wslStatus.HasDockerDesktopData) { 'Present' } else { 'Missing' })" -ForegroundColor DarkGray
}
Write-Host ""

# 5. Check Hyper-V services
Write-Host "[5/6] Hyper-V Services" -ForegroundColor Yellow
$hyperv = Test-HyperVStatus
if ($hyperv.VmComputeRunning -and $hyperv.HnsRunning) {
    Write-Status "Hyper-V Services" "OK" "vmcompute & hns running"
} else {
    Write-Status "Hyper-V Services" "WARNING" "Some services not running"
    Write-Host "    vmcompute: $(if ($hyperv.VmComputeRunning) { 'Running' } else { 'Stopped' })" -ForegroundColor DarkGray
    Write-Host "    hns: $(if ($hyperv.HnsRunning) { 'Running' } else { 'Stopped' })" -ForegroundColor DarkGray
    $issues += @{
        Type = "HyperVServicesStopped"
        Fix = { Repair-HyperVError }
    }
}
Write-Host ""

# 6. Check port conflicts
Write-Host "[6/6] Port Conflicts" -ForegroundColor Yellow
$portConflicts = Test-DockerPorts
if ($portConflicts.Count -eq 0) {
    Write-Status "Required Ports" "OK" "No conflicts"
} else {
    Write-Status "Required Ports" "WARNING" "Conflicts: $($portConflicts -join ', ')"
    $issues += @{
        Type = "PortConflict"
        Fix = { Repair-PortConflict }
    }
}
Write-Host ""

# Summary
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Summary" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

if ($issues.Count -eq 0) {
    Write-Host "  [OK] Docker Desktop is healthy!" -ForegroundColor Green
    Write-Host "  You should be able to run builds successfully." -ForegroundColor Green
} else {
    Write-Host "  [!] Found $($issues.Count) issue(s):" -ForegroundColor Yellow
    Write-Host ""
    foreach ($issue in $issues) {
        Write-Host "    - $($issue.Type)" -ForegroundColor Yellow
    }
    Write-Host ""
    
    if ($Fix) {
        Write-Host "  Attempting to fix issues..." -ForegroundColor Cyan
        Write-Host ""
        foreach ($issue in $issues) {
            Write-Host "  Fixing: $($issue.Type)..." -ForegroundColor Yellow
            try {
                & $issue.Fix
                Write-Host "    [OK] Fixed" -ForegroundColor Green
            }
            catch {
                Write-Host "    [FAIL] Could not fix: $_" -ForegroundColor Red
            }
        }
        Write-Host ""
        Write-Host "  Re-run this script to verify fixes." -ForegroundColor Cyan
    } else {
        Write-Host "  Run with -Fix parameter to attempt automatic fixes:" -ForegroundColor Cyan
        Write-Host "    .\docker-diagnose.ps1 -Fix" -ForegroundColor White
    }
}

Write-Host ""
