# ============================================
# Error Recovery Module (Refactored)
# Smart error detection and recovery
# SRP: Error pattern matching and recovery orchestration
# ============================================

# Note: Dependencies should be loaded by the caller

# ============================================
# Error Patterns Definition
# ============================================

$script:ErrorPatterns = @{
    # Docker infrastructure errors
    "DockerEngineError" = @{
        Pattern = "500 Internal Server Error|dockerDesktopLinuxEngine.*_ping|request returned 500|error during connect"
        Description = "Docker Desktop engine error (500)"
        Fix = { Repair-DockerEngineError }
    }
    "DockerPipeError" = @{
        Pattern = "pipe.*docker|npipe.*error|named pipe|\\\\\.\\/pipe"
        Description = "Docker pipe connection error"
        Fix = { Repair-DockerPipeError }
    }
    "DNSError" = @{
        Pattern = "DNS lookup error|DNS.*name does not exist|SERVFAIL|NXDOMAIN|fetch.*error|lookup.*no such host|dial tcp.*lookup"
        Description = "DNS resolution failure"
        Fix = { Repair-DockerDNSError }
    }
    "AlpineRepoError" = @{
        Pattern = "fetching https://dl-cdn\.alpinelinux\.org.*temporary error|unable to select packages.*alpine|apk.*error"
        Description = "Alpine repository network error"
        Fix = { Repair-AlpineRepositoryError }
    }
    "DockerRegistryError" = @{
        Pattern = "registry-1\.docker\.io.*no such host|failed to do request.*registry|registry.*connection|pull.*manifest.*error"
        Description = "Docker registry connection error"
        Fix = { Repair-DockerRegistryError }
    }
    "WslMountCacheError" = @{
        Pattern = "/run/desktop/mnt/host|mount source path|mkdir.*mnt|creating mount.*path|file exists.*mnt|mnt/host.*file exists"
        Description = "WSL2 mount cache corruption"
        Fix = { Repair-WslMountCache }
    }
    "WslKernelError" = @{
        Pattern = "wsl.*kernel|wsl2.*error|vmlinux|WSL.*failed|docker-desktop.*distro"
        Description = "WSL2 kernel error"
        Fix = { Repair-WslKernelError }
    }
    "HyperVError" = @{
        Pattern = "hyperv|hyper-v|virtualization|vmcompute|hv_sock"
        Description = "Hyper-V issue"
        Fix = { Repair-HyperVError }
    }
    "NetworkError" = @{
        Pattern = "network.*error|connect.*refused|ECONNREFUSED"
        Description = "Network connectivity issue"
        Fix = { Repair-NetworkError }
    }
    "BuildKitError" = @{
        Pattern = "buildkit|builder.*error|failed to solve"
        Description = "BuildKit issue"
        Fix = { Repair-DockerBuildKit }
    }
    "ManifestError" = @{
        Pattern = "no matching manifest|manifest.*not found"
        Description = "Docker manifest/platform error"
        Fix = { Switch-DockerToLinux }
    }

    # Resource errors
    "DiskSpaceError" = @{
        Pattern = "no space left|disk.*full|ENOSPC"
        Description = "Disk space issue"
        Fix = { Invoke-AggressiveCleanup }
    }
    "OOMError" = @{
        Pattern = "exit.*code.*137|killed|out of memory|OOM|ENOMEM"
        Description = "Out of Memory"
        Fix = {
            docker stop $(docker ps -q) 2>&1 | Out-Null
            docker system prune -f 2>&1 | Out-Null
            Stop-WSL -WaitSeconds 5
            Restart-DockerDesktop
        }
    }
    "PortConflictError" = @{
        Pattern = "port.*already.*use|address already in use|EADDRINUSE"
        Description = "Port conflict"
        Fix = { Repair-PortConflict }
    }
    "ContainerNameConflict" = @{
        Pattern = "container name.*already in use|Conflict.*container name"
        Description = "Container name already in use"
        Fix = { Repair-ContainerNameConflict }
    }
    "PostgresDataCorruption" = @{
        Pattern = "could not open directory.*pg_|pg_notify.*No such file|pg_wal.*No such file|pg_xact.*No such file|database.*shut down|FATAL.*postgres"
        Description = "PostgreSQL data directory corrupted"
        Fix = { Repair-PostgresDataDirectory }
    }
    "DependencyFailedToStart" = @{
        Pattern = "dependency.*failed to start|depends_on.*failed|service.*unhealthy"
        Description = "Dependency service failed to start"
        Fix = { Repair-DependencyFailure }
    }

    # NPM/build errors
    "NpmIntegrityError" = @{
        Pattern = "integrity checksum|EINTEGRITY|sha512|sha1.*mismatch"
        Description = "NPM integrity mismatch"
        Fix = {
            Write-Fix "Clearing npm cache..."
            npm cache clean --force 2>&1 | Out-Null
        }
    }
    "PrismaMigrateError" = @{
        Pattern = "prisma.*migrate|migration.*failed|SQLITE_BUSY"
        Description = "Prisma migration issue"
        Fix = {
            $backendPath = Join-Path $script:ProjectRoot "backend"
            Push-Location $backendPath
            try { npx prisma generate 2>&1 | Out-Null }
            finally { Pop-Location }
        }
    }

    # Code errors (require pause)
    "TypeScriptError" = @{
        Pattern = "error TS\d+:|Cannot find module|has no exported member|is not assignable to|Object is of type 'unknown'"
        Description = "TypeScript compilation error"
        Fix = $null
        IsCodeError = $true
    }
    "ImportResolutionError" = @{
        Pattern = "Could not resolve|Failed to resolve import|Module not found.*from"
        Description = "Import/Module resolution error - file may be missing or path incorrect"
        Fix = $null
        IsCodeError = $true
    }
    "SyntaxError" = @{
        Pattern = "SyntaxError|Unexpected token|Parse error"
        Description = "JavaScript/TypeScript syntax error"
        Fix = $null
        IsCodeError = $true
    }
    "ViteError" = @{
        Pattern = "vite.*error|rollup.*error"
        Description = "Vite/Rollup build error"
        Fix = $null
        IsCodeError = $true
    }
    "PrismaSchemaError" = @{
        Pattern = "error:.*prisma|Prisma schema.*invalid|Unknown.*model"
        Description = "Prisma schema error"
        Fix = $null
        IsCodeError = $true
    }
}

# ============================================
# Error Detection
# ============================================

function Get-ErrorType {
    <#
    .SYNOPSIS
    Detect error type from message
    #>
    param([string]$ErrorMessage)

    foreach ($errorName in $script:ErrorPatterns.Keys) {
        $pattern = $script:ErrorPatterns[$errorName].Pattern
        if ($ErrorMessage -match $pattern) {
            $isCodeError = $script:ErrorPatterns[$errorName].ContainsKey("IsCodeError") -and $script:ErrorPatterns[$errorName].IsCodeError
            return @{
                Type = $errorName
                Description = $script:ErrorPatterns[$errorName].Description
                Fix = $script:ErrorPatterns[$errorName].Fix
                IsCodeError = $isCodeError
            }
        }
    }

    return @{ Type = "Unknown"; Description = "Unknown error"; Fix = $null; IsCodeError = $false }
}

function Test-IsCodeError {
    <#
    .SYNOPSIS
    Check if error is a code error requiring manual fix
    #>
    param([string]$ErrorMessage)
    return (Get-ErrorType -ErrorMessage $ErrorMessage).IsCodeError
}

# ============================================
# Code Error Pause UI
# ============================================

function Show-CodeErrorPause {
    <#
    .SYNOPSIS
    Display code errors and pause for user action
    #>
    param([string]$ErrorMessage, [string]$ErrorType = "Code")

    Write-Host ""
    Write-Host "[$ErrorType ERROR DETECTED]" -ForegroundColor Red
    Write-Host "This error requires manual fix and cannot be auto-fixed." -ForegroundColor Yellow
    Write-Host ""

    $errorPattern = "error TS\d+:|Cannot find|has no exported|Property.*does not|is not assignable|npm ERR!|Error:|src/|backend/|frontend/|\.ts:"
    $errorLines = $ErrorMessage -split "`n" | Where-Object { $_ -match $errorPattern } | Select-Object -First 15

    if ($errorLines) {
        Write-Host "--- Error Details (top 15) ---" -ForegroundColor Cyan
        foreach ($line in $errorLines) {
            $color = if ($line -match "error|Error") { "Red" } elseif ($line -match "src/|backend/|frontend/") { "Yellow" } else { "Gray" }
            Write-Host $line -ForegroundColor $color
        }
        Write-Host "------------------------------" -ForegroundColor Cyan
    }

    Write-Host ""
    Write-Host "[R] Retry  [S] Skip  [Q] Quit" -ForegroundColor Cyan
    Write-Host ""

    while ($true) {
        $choice = Read-Host "Choice (R/S/Q)"
        switch ($choice.ToUpper()) {
            "R" { Write-Info "Retrying..."; return "retry" }
            "S" { Write-Warning "Skipping..."; return "skip" }
            "Q" { Write-Info "Quitting..."; return "quit" }
            default { Write-Host "Invalid. Enter R, S, or Q." -ForegroundColor Yellow }
        }
    }
}

# ============================================
# Error Recovery
# ============================================

function Invoke-ErrorRecovery {
    <#
    .SYNOPSIS
    Attempt automatic error recovery
    #>
    param([string]$ErrorMessage, [int]$AttemptNumber = 1)

    Write-Fix "Analyzing error (attempt $AttemptNumber)..."

    $errorInfo = Get-ErrorType -ErrorMessage $ErrorMessage

    if ($errorInfo.Type -ne "Unknown" -and $errorInfo.Fix) {
        Write-Info "Detected: $($errorInfo.Description)"
        try {
            & $errorInfo.Fix
            Write-Success "Applied fix for: $($errorInfo.Type)"
            return $true
        }
        catch {
            Write-Warning "Fix failed: $_"
        }
    }

    return Invoke-ProgressiveRecovery -AttemptNumber $AttemptNumber
}

function Invoke-ProgressiveRecovery {
    <#
    .SYNOPSIS
    Apply progressively aggressive fixes (6 levels)
    #>
    param([int]$AttemptNumber)

    switch ($AttemptNumber) {
        1 {
            Write-Fix "Level 1: Basic Docker health check + dependency repair..."
            $status = Test-DockerReady
            if ($status.ErrorType) { Invoke-DockerAutoFix -ErrorMessage $status.Output }
            Repair-PortConflict
            Repair-ContainerNameConflict
            # Check PostgreSQL and dependency issues early
            Repair-DependencyFailure
            Invoke-LightCleanup
            return $true
        }
        2 {
            Write-Fix "Level 2: Memory + WSL restart..."
            Invoke-MemoryCleanup
            Restart-WslAndDocker
            Invoke-MediumCleanup
            return $true
        }
        3 {
            Write-Fix "Level 3: Container health + network..."
            Repair-ContainerHealth
            Repair-NetworkError
            return $true
        }
        4 {
            Write-Fix "Level 4: Docker Desktop restart..."
            Repair-DockerEngineError
            Start-Sleep -Seconds 10
            return $true
        }
        5 {
            Write-Fix "Level 5: Aggressive cleanup..."
            Invoke-AggressiveCleanup
            Repair-DockerEngineError
            Start-Sleep -Seconds 15
            return $true
        }
        6 {
            Write-Fix "Level 6: Full system reset..."
            Stop-AllDockerProcesses -WaitSeconds 5
            Stop-WSL -WaitSeconds 5
            Stop-DockerWSLDistros
            wsl --update 2>&1 | Out-Null

            try {
                Restart-Service "vmcompute" -Force -ErrorAction SilentlyContinue
                Restart-Service "hns" -Force -ErrorAction SilentlyContinue
            } catch {
                # Service restart may require admin privileges - continue anyway
                $null = $_
            }

            Invoke-AggressiveCleanup
            Start-DockerDesktop
            Start-Sleep -Seconds 30
            return Assert-DockerReady
        }
        default {
            Write-Error "All recovery levels exhausted."
            Write-Info "  1. Restart your computer"
            Write-Info "  2. Open Docker Desktop manually"
            Write-Info "  3. Check Windows Event Viewer"
            return $false
        }
    }
}
