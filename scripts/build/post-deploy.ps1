# ============================================
# Post-Deploy Module
# Post-deployment tasks (copy files, permissions, etc.)
# ============================================

. "$PSScriptRoot\config.ps1"
. "$PSScriptRoot\logger.ps1"
. "$PSScriptRoot\retry-logic.ps1"

function Test-PostgresConnection {
    <#
    .SYNOPSIS
    Verify PostgreSQL connection is working
    NOTE: SQLite migration completed - now using PostgreSQL only
    #>
    Write-Step "POST-DEPLOY" "Verifying PostgreSQL connection..."
    
    try {
        # Check if postgres container is running
        $pgStatus = docker inspect --format='{{.State.Status}}' "omni-postgres" 2>$null
        
        if ($pgStatus -ne "running") {
            Write-Warning "PostgreSQL container not running (status: $pgStatus)"
            return $false
        }
        
        # Test database connection via pgbouncer
        $result = docker exec omni-postgres pg_isready -U omni -d omni_main 2>&1
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "PostgreSQL connection verified"
            return $true
        }
        else {
            Write-Warning "PostgreSQL not ready: $result"
            return $false
        }
    }
    catch {
        Write-Warning "Error checking PostgreSQL: $_"
        return $false
    }
}

function Test-DatabaseEmpty {
    <#
    .SYNOPSIS
    Check if database is empty (fresh install / migration from another PC)
    Returns $true if database has no tenant data
    Returns $false if postgres not running or has data
    #>
    try {
        # First check if postgres is running
        $pgStatus = docker inspect --format='{{.State.Status}}' "omni-postgres" 2>$null
        if ($pgStatus -ne "running") {
            return $false  # Can't check - assume not empty
        }
        
        # Count tables in tenant schemas
        $query = "SELECT COUNT(*) FROM pg_stat_user_tables WHERE schemaname LIKE 'tenant_%';"
        $result = docker exec omni-postgres psql -U omni -d omni_main -t -A -c $query 2>&1
        
        if ($LASTEXITCODE -ne 0) {
            return $false
        }
        
        $tableCount = 0
        if ($result -is [string]) {
            $tableCount = [int]($result.Trim())
        } elseif ($result -is [array]) {
            $firstLine = ($result | Where-Object { $_ -match '^\d+$' } | Select-Object -First 1)
            if ($firstLine) { $tableCount = [int]$firstLine }
        }
        
        if ($tableCount -eq 0) {
            return $true
        }
        
        # Also check if tables exist but are empty (row count = 0)
        $rowQuery = "SELECT COALESCE(SUM(n_live_tup), 0) FROM pg_stat_user_tables WHERE schemaname LIKE 'tenant_%';"
        $rowResult = docker exec omni-postgres psql -U omni -d omni_main -t -A -c $rowQuery 2>&1
        
        $totalRows = 0
        if ($rowResult -is [string]) {
            $totalRows = [int]($rowResult.Trim())
        } elseif ($rowResult -is [array]) {
            $firstLine = ($rowResult | Where-Object { $_ -match '^\d+$' } | Select-Object -First 1)
            if ($firstLine) { $totalRows = [int]$firstLine }
        }
        
        return $totalRows -eq 0
    }
    catch {
        Write-Warning "Error checking database: $_"
        return $false
    }
}

function Test-BackupExists {
    <#
    .SYNOPSIS
    Check if smart backup exists
    #>
    $manifestPath = Join-Path $script:ProjectRoot "backups\smart\manifest.json"
    return Test-Path $manifestPath
}

function Invoke-DatabaseRestorePrompt {
    <#
    .SYNOPSIS
    Prompt user to restore database from backup if database is empty
    #>
    Write-Step "POST-DEPLOY" "Checking database state..."
    
    # Skip if postgres not ready
    $pgStatus = docker inspect --format='{{.State.Status}}' "omni-postgres" 2>$null
    if ($pgStatus -ne "running") {
        Write-Warning "PostgreSQL not running - skipping database check"
        return
    }
    
    # Wait for postgres to be fully ready
    Start-Sleep -Seconds 2
    
    if (Test-DatabaseEmpty) {
        Write-Host ""
        Write-Host "  +-------------------------------------------+" -ForegroundColor Yellow
        Write-Host "  | DATABASE IS EMPTY                         |" -ForegroundColor Yellow
        Write-Host "  +-------------------------------------------+" -ForegroundColor Yellow
        Write-Host "  | Detected fresh database installation.     |" -ForegroundColor White
        Write-Host "  | This typically happens when:              |" -ForegroundColor White
        Write-Host "  |   - First time setup on new PC            |" -ForegroundColor Gray
        Write-Host "  |   - Docker volumes were reset             |" -ForegroundColor Gray
        Write-Host "  +-------------------------------------------+" -ForegroundColor Yellow
        Write-Host ""
        
        if (Test-BackupExists) {
            $manifestPath = Join-Path $script:ProjectRoot "backups\smart\manifest.json"
            $manifest = Get-Content $manifestPath -Raw | ConvertFrom-Json
            
            Write-Host "  Backup found: $($manifest.exported_at)" -ForegroundColor Green
            Write-Host "  Tables: $($manifest.total_tables) | Rows: $($manifest.total_rows)" -ForegroundColor Green
            Write-Host ""
            
            $choice = Read-Host "  Restore database from backup? (Y/n)"
            
            if ($choice -eq "" -or $choice -eq "Y" -or $choice -eq "y") {
                Write-Host ""
                Write-Info "Starting database restore..."
                
                $restoreScript = Join-Path $script:ProjectRoot "backups\db-tools\restore-smart.ps1"
                
                if (Test-Path $restoreScript) {
                    & $restoreScript -Force
                    Write-Success "Database restore completed!"
                }
                else {
                    Write-Warning "Restore script not found: $restoreScript"
                    Write-Info "Run manually: backups\db-tools\menu.bat -> [3] Smart Restore"
                }
            }
            else {
                Write-Info "Skipping restore - database will remain empty"
            }
        }
        else {
            Write-Warning "No backup found in backups\smart\"
            Write-Info "If migrating from another PC, copy the backups folder first"
            Write-Info "Then run: backups\db-tools\menu.bat -> [3] Smart Restore"
        }
        
        Write-Host ""
    }
    else {
        Write-Success "Database has data - no restore needed"
    }
}

function Test-BackendDbConfig {
    <#
    .SYNOPSIS
    Verify backend is configured to use PostgreSQL
    NOTE: SQLite migration completed - permissions check no longer needed
    #>
    Write-Step "POST-DEPLOY" "Verifying backend database config..."
    
    try {
        # Check backend environment for PostgreSQL config
        $dbDriver = docker exec $script:ContainerBackend sh -c 'echo $DB_DRIVER' 2>$null
        
        if ($dbDriver -eq "postgres") {
            Write-Success "Backend configured for PostgreSQL"
            return $true
        }
        elseif ([string]::IsNullOrEmpty($dbDriver)) {
            Write-Info "DB_DRIVER not set - defaulting to postgres"
            return $true
        }
        else {
            Write-Warning "Unexpected DB_DRIVER: $dbDriver"
            return $false
        }
    }
    catch {
        Write-Warning "Error checking backend config: $_"
        return $false
    }
}

function Wait-ForBackendReady {
    <#
    .SYNOPSIS
    Wait for backend to be ready (replaces old restart logic)
    NOTE: No need to restart for SQLite files - PostgreSQL is always ready
    #>
    Write-Step "POST-DEPLOY" "Waiting for backend to be ready..."
    
    try {
        $maxWait = 30  # seconds
        $waited = 0
        
        while ($waited -lt $maxWait) {
            $status = docker inspect --format='{{.State.Status}}' $script:ContainerBackend 2>$null
            
            if ($status -eq "running") {
                Write-Success "Backend container is ready"
                return $true
            }
            
            Start-Sleep -Seconds 2
            $waited += 2
        }
        
        Write-Warning "Backend not ready after ${maxWait}s"
        return $false
    }
    catch {
        Write-Warning "Error checking backend: $_"
        return $false
    }
}

function Test-NginxRunning {
    <#
    .SYNOPSIS
    Check if nginx container is running
    #>
    try {
        $status = docker ps --filter "name=$script:ContainerNginx" --format "{{.Status}}" 2>$null
        return $status -match "Up"
    }
    catch {
        return $false
    }
}

function Test-BackendHealth {
    <#
    .SYNOPSIS
    Check backend health endpoint
    #>
    try {
        $response = Invoke-RestMethod -Uri "http://localhost:3000/api/health" -TimeoutSec 10 -ErrorAction SilentlyContinue
        # Accept multiple valid health statuses
        return $null -ne $response -and ($response.status -eq "healthy" -or $response.status -eq "ok" -or $response.success -eq $true)
    }
    catch {
        return $false
    }
}

function Show-ContainerLogs {
    <#
    .SYNOPSIS
    Show recent logs from a container
    #>
    param(
        [string]$Container,
        [int]$TailLines = 50
    )
    
    docker logs $Container --tail $TailLines 2>&1
}

function Invoke-PostDeployTasks {
    <#
    .SYNOPSIS
    Run all post-deployment tasks
    NOTE: Updated for PostgreSQL - no more SQLite file copying
    #>
    Write-Header "Post-Deploy Tasks"
    
    # Wait for containers to stabilize
    Write-Info "Waiting for containers to stabilize..."
    Start-Sleep -Seconds 5
    
    # Verify PostgreSQL connection
    Test-PostgresConnection
    
    # Check if database is empty and offer restore
    Invoke-DatabaseRestorePrompt
    
    # Verify backend database config
    Test-BackendDbConfig
    
    # Wait for backend to be ready
    Wait-ForBackendReady
    
    # Verify nginx is running
    Write-Step "VERIFY" "Checking nginx status..."
    if (Test-NginxRunning) {
        Write-Success "Nginx is running"
    }
    else {
        Write-Warning "Nginx may not be running properly"
        Write-Info "Checking nginx logs..."
        Show-ContainerLogs -Container $script:ContainerNginx -TailLines 20
    }
    
    # Check backend health with auto-recovery
    Write-Step "VERIFY" "Checking backend health..."
    Start-Sleep -Seconds 3
    
    $healthRetries = 3
    $healthPassed = $false
    
    for ($i = 1; $i -le $healthRetries; $i++) {
        if (Test-BackendHealth) {
            Write-Success "Backend health check passed"
            $healthPassed = $true
            break
        }
        else {
            if ($i -lt $healthRetries) {
                Write-Warning "Backend health check failed (attempt $i/$healthRetries) - auto-fixing..."
                
                # Try to fix common issues
                $status = docker inspect --format='{{.State.Status}}' $script:ContainerBackend 2>$null
                
                if ($status -eq "exited" -or $status -eq "dead") {
                    Write-Fix "Backend container is $status - restarting..."
                    docker start $script:ContainerBackend 2>&1 | Out-Null
                }
                elseif ($status -eq "running") {
                    # Container running but unhealthy - check logs and restart
                    Write-Fix "Backend running but unhealthy - restarting..."
                    docker restart $script:ContainerBackend 2>&1 | Out-Null
                }
                else {
                    Write-Warning "Unknown container status: $status"
                }
                
                Start-Sleep -Seconds 10
            }
            else {
                Write-Warning "Backend health check failed after $healthRetries attempts"
                Write-Info "Showing backend logs for debugging:"
                Show-ContainerLogs -Container $script:ContainerBackend -TailLines 30
            }
        }
    }
    
    # Return health status - deployment proceeds but caller knows if health check passed
    if (-not $healthPassed) {
        Write-Warning "Deployment completed but backend health check did not pass"
    }
    return $healthPassed
}

function Show-DeploymentSummary {
    <#
    .SYNOPSIS
    Show deployment summary with endpoints
    #>
    param(
        [string]$Spec = "standard"
    )
    
    Write-Header "Deployment Complete"
    
    Write-Host ""
    Write-Host "Endpoints:"
    Write-Host "  Frontend:     http://localhost"
    Write-Host "  Backend API:  http://localhost:3000"
    Write-Host "  Health Check: http://localhost:3000/api/health"
    
    Write-Host ""
    Write-Host "Configuration:"
    Write-Host "  Environment: Tunnel (Cloudflare)"
    Write-Host "  RAM Spec:    $($script:RamSpecs[$Spec].Name) ($($script:RamSpecs[$Spec].RAM))"
    Write-Host ""
}
