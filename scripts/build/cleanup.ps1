# ============================================
# Cleanup Utilities Module
# Smart cleanup for Docker resources
# ============================================

. "$PSScriptRoot\config.ps1"
. "$PSScriptRoot\logger.ps1"

function Stop-ProjectContainers {
    <#
    .SYNOPSIS
    Stop all project containers gracefully
    #>
    Write-Step "CLEANUP" "Stopping project containers..."
    
    $containers = @($script:ContainerBackend, $script:ContainerFrontend, $script:ContainerNginx, $script:ContainerRedis)
    
    foreach ($container in $containers) {
        $exists = docker ps -a --format "{{.Names}}" | Where-Object { $_ -eq $container }
        if ($exists) {
            docker stop $container 2>&1 | Out-Null
            Write-Debug "Stopped: $container"
        }
    }
    
    Write-Success "Project containers stopped"
}

function Remove-ProjectContainers {
    <#
    .SYNOPSIS
    Remove all project containers
    #>
    Write-Step "CLEANUP" "Removing project containers..."
    
    # Stop and remove via compose files
    docker-compose -f $script:DockerComposeTunnel down --remove-orphans 2>&1 | Out-Null
    
    Write-Success "Project containers removed"
}

function Remove-DanglingImages {
    <#
    .SYNOPSIS
    Remove dangling (untagged) images
    #>
    Write-Step "CLEANUP" "Removing dangling images..."
    
    $danglingCount = (docker images -f "dangling=true" -q 2>$null | Measure-Object).Count
    
    if ($danglingCount -gt 0) {
        docker image prune -f 2>&1 | Out-Null
        Write-Success "Removed $danglingCount dangling images"
    } else {
        Write-Info "No dangling images to remove"
    }
}

function Remove-UnusedVolumes {
    <#
    .SYNOPSIS
    Remove unused Docker volumes
    #>
    Write-Step "CLEANUP" "Removing unused volumes..."
    
    docker volume prune -f 2>&1 | Out-Null
    Write-Success "Unused volumes removed"
}

function Remove-BuildCache {
    <#
    .SYNOPSIS
    Remove Docker build cache
    #>
    Write-Step "CLEANUP" "Clearing build cache..."
    
    docker builder prune -af 2>&1 | Out-Null
    Write-Success "Build cache cleared"
}

function Remove-UnusedNetworks {
    <#
    .SYNOPSIS
    Remove unused Docker networks
    #>
    Write-Step "CLEANUP" "Removing unused networks..."
    
    docker network prune -f 2>&1 | Out-Null
    Write-Success "Unused networks removed"
}

function Invoke-LightCleanup {
    <#
    .SYNOPSIS
    Light cleanup - just remove orphaned/dangling resources
    #>
    Write-Header "Light Cleanup"
    
    Remove-DanglingImages
    Remove-UnusedVolumes
    Remove-UnusedNetworks
    
    Write-Success "Light cleanup complete"
}

function Invoke-MediumCleanup {
    <#
    .SYNOPSIS
    Medium cleanup - stop containers, remove dangling resources
    #>
    Write-Header "Medium Cleanup"
    
    Remove-ProjectContainers
    Remove-DanglingImages
    Remove-UnusedVolumes
    Remove-UnusedNetworks
    Remove-BuildCache
    
    Write-Success "Medium cleanup complete"
}

function Invoke-AggressiveCleanup {
    <#
    .SYNOPSIS
    Aggressive cleanup - remove everything (USE WITH CAUTION)
    #>
    Write-Header "Aggressive Cleanup"
    Write-Warning "This will remove ALL Docker images, volumes, and cache!"
    
    # Stop all containers first
    $allContainers = docker ps -aq 2>$null
    if ($allContainers) {
        docker stop $allContainers 2>&1 | Out-Null
        Write-Debug "All containers stopped"
    }
    
    # Remove all stopped containers
    docker container prune -f 2>&1 | Out-Null
    Write-Success "Removed stopped containers"
    
    # Remove all images, volumes, and networks
    docker system prune -a --volumes -f 2>&1 | Out-Null
    Write-Success "System pruned (images, volumes, networks)"
    
    # Remove build cache
    docker builder prune -af 2>&1 | Out-Null
    Write-Success "Build cache cleared"
    
    # Show disk usage after cleanup
    Write-Info "Docker disk usage after cleanup:"
    docker system df
    
    Write-Success "Aggressive cleanup complete"
}

function Clear-OldBuildArtifacts {
    <#
    .SYNOPSIS
    Clear old build artifacts (dist folders, node_modules)
    #>
    param([switch]$IncludeNodeModules)
    
    Write-Step "CLEANUP" "Clearing build artifacts..."
    
    # Clear frontend dist
    $frontendDist = Join-Path $script:FrontendPath "dist"
    if (Test-Path $frontendDist) {
        Remove-Item -Path $frontendDist -Recurse -Force
        Write-Debug "Removed frontend/dist"
    }
    
    # Clear backend dist
    $backendDist = Join-Path $script:BackendPath "dist"
    if (Test-Path $backendDist) {
        Remove-Item -Path $backendDist -Recurse -Force
        Write-Debug "Removed backend/dist"
    }
    
    if ($IncludeNodeModules) {
        # Clear frontend node_modules
        $frontendModules = Join-Path $script:FrontendPath "node_modules"
        if (Test-Path $frontendModules) {
            Remove-Item -Path $frontendModules -Recurse -Force
            Write-Debug "Removed frontend/node_modules"
        }
        
        # Clear backend node_modules
        $backendModules = Join-Path $script:BackendPath "node_modules"
        if (Test-Path $backendModules) {
            Remove-Item -Path $backendModules -Recurse -Force
            Write-Debug "Removed backend/node_modules"
        }
    }
    
    Write-Success "Build artifacts cleared"
}

function Get-DockerDiskUsage {
    <#
    .SYNOPSIS
    Get Docker disk usage
    #>
    return docker system df --format "{{json .}}" 2>$null | ConvertFrom-Json
}
