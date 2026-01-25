# ============================================
# Docker Fixer Module (Unified Interface)
# Combines docker-core and docker-repair
# SRP: Single entry point for Docker fixes
# ============================================

# Load sub-modules
. "$PSScriptRoot\docker-core.ps1"
. "$PSScriptRoot\docker-repair.ps1"

# Re-export all functions from sub-modules for backward compatibility
# Functions are already available through dot-sourcing

<#
.SYNOPSIS
This module provides a unified interface for Docker fixes.

EXPORTED FROM docker-core.ps1:
- Stop-AllDockerProcesses
- Start-DockerDesktop
- Test-DockerReady
- Get-DockerMode
- Wait-ForDocker
- Stop-WSL
- Stop-DockerWSLDistros
- Switch-DockerToLinux
- Test-AvailableMemory
- Invoke-MemoryCleanup
- Repair-PortConflict
- Repair-ContainerNameConflict

EXPORTED FROM docker-repair.ps1:
- Repair-DockerEngineError
- Repair-DockerPipeError
- Repair-DockerDNSError
- Repair-WslMountCache
- Repair-WslKernelError
- Repair-NetworkError
- Repair-HyperVError
- Repair-DockerBuildKit
- Repair-ContainerHealth
- Restart-DockerDesktop
- Restart-WslAndDocker
- Assert-DockerReady
#>

function Invoke-DockerAutoFix {
    <#
    .SYNOPSIS
    Run automatic Docker fixes based on error message
    #>
    param([string]$ErrorMessage = "")
    
    $fixed = $false
    
    # Detect error type and apply appropriate fix
    switch -Regex ($ErrorMessage) {
        "500 Internal Server Error|dockerDesktopLinuxEngine|request returned 500" {
            $fixed = Repair-DockerEngineError
        }
        "pipe.*docker|npipe.*error|named pipe" {
            $fixed = Repair-DockerPipeError
        }
        "DNS lookup error|DNS.*name does not exist|fetch.*error" {
            $fixed = Repair-DockerDNSError
        }
        "no matching manifest" {
            $mode = Get-DockerMode
            if ($mode -ne "linux") {
                $fixed = Switch-DockerToLinux
            }
        }
        "mkdir /run/desktop/mnt/host|creating mount source path|file exists.*mnt" {
            $fixed = Repair-WslMountCache
        }
        "mount|volume|bind" {
            $fixed = Restart-WslAndDocker
        }
        "network|connect" {
            $fixed = Repair-NetworkError
        }
        "buildkit|builder" {
            $fixed = Repair-DockerBuildKit
        }
        "container name.*already in use|Conflict.*container name" {
            $fixed = Repair-ContainerNameConflict
        }
        "hyperv|hyper-v|virtualization|vmcompute" {
            $fixed = Repair-HyperVError
        }
        "wsl.*kernel|wsl2.*error|docker-desktop.*distro" {
            $fixed = Repair-WslKernelError
        }
        default {
            # Generic fix
            Write-Fix "Applying general Docker fixes..."
            $mode = Get-DockerMode
            if ($mode -ne "linux") {
                $fixed = Switch-DockerToLinux
            }
            if (-not $fixed) {
                $fixed = Restart-WslAndDocker
            }
        }
    }
    
    return $fixed
}
