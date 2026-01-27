# ============================================
# Frontend Build Module
# Build frontend with error handling
# ============================================
. "$PSScriptRoot\config.ps1"
. "$PSScriptRoot\logger.ps1"
. "$PSScriptRoot\retry-logic.ps1"

function Show-BuildError {
    <#
    .SYNOPSIS
    Display build error details clearly to user
    #>
    param(
        [string]$Output,
        [string]$ErrorType = "Build"
    )
    
    Write-Host ""
    Write-Host "============================================" -ForegroundColor Red
    Write-Host "  $ErrorType ERROR - REQUIRES MANUAL FIX" -ForegroundColor Red
    Write-Host "============================================" -ForegroundColor Red
    Write-Host ""
    
    # Extract relevant error lines
    $errorPatterns = @(
        "error TS\d+:",
        "Could not resolve",
        "Failed to resolve",
        "Cannot find module",
        "SyntaxError",
        "Unexpected token",
        "Parse error",
        "Error:",
        "error:",
        "ERROR:",
        "from "".*""",
        "file:.*"
    )
    $combinedPattern = ($errorPatterns -join "|")
    
    $errorLines = $Output -split "`n" | Where-Object { 
        $_ -match $combinedPattern -or $_ -match "^\s+at\s+" 
    } | Select-Object -First 20
    
    if ($errorLines) {
        Write-Host "--- Error Details ---" -ForegroundColor Cyan
        foreach ($line in $errorLines) {
            $trimmedLine = $line.Trim()
            if ($trimmedLine) {
                # Color code based on content
                $color = "White"
                if ($trimmedLine -match "error|Error|ERROR") { $color = "Red" }
                elseif ($trimmedLine -match "Could not resolve|Cannot find|Failed to resolve") { $color = "Yellow" }
                elseif ($trimmedLine -match "file:|from ""|\.vue|\.ts|\.js") { $color = "Cyan" }
                
                Write-Host "  $trimmedLine" -ForegroundColor $color
            }
        }
        Write-Host "---------------------" -ForegroundColor Cyan
    }
    else {
        # If no specific error found, show last 30 lines
        Write-Host "--- Build Output (last 30 lines) ---" -ForegroundColor Cyan
        $lastLines = $Output -split "`n" | Select-Object -Last 30
        foreach ($line in $lastLines) {
            if ($line.Trim()) {
                Write-Host "  $line" -ForegroundColor Gray
            }
        }
        Write-Host "------------------------------------" -ForegroundColor Cyan
    }
    
    Write-Host ""
    Write-Host "Please fix the error above and try again." -ForegroundColor Yellow
    Write-Host ""
}

function Get-FileHashCompat {
    <#
    .SYNOPSIS
    Get file hash with fallback for older PowerShell versions
    #>
    param([string]$Path, [string]$Algorithm = "SHA256")
    
    if (-not (Test-Path $Path)) { return $null }
    
    # Try native Get-FileHash first (PS 4.0+)
    if (Get-Command Get-FileHash -ErrorAction SilentlyContinue) {
        return (Get-FileHash -Path $Path -Algorithm $Algorithm).Hash
    }
    
    # Fallback for older PowerShell
    try {
        $hasher = [System.Security.Cryptography.HashAlgorithm]::Create($Algorithm)
        $stream = [System.IO.File]::OpenRead($Path)
        try {
            $hash = $hasher.ComputeHash($stream)
            return [BitConverter]::ToString($hash) -replace '-', ''
        }
        finally {
            $stream.Close()
        }
    }
    catch {
        return $null
    }
}

function Get-FrontendLockInfo {
    $lockPath = Join-Path $script:FrontendPath "package-lock.json"
    $hashPath = Join-Path $script:ScriptsPath ".frontend-lock.hash"
    return @{ LockPath = $lockPath; HashPath = $hashPath }
}
function Get-FrontendLockHash {
    param([string]$LockPath)
    if (-not (Test-Path $LockPath)) { return $null }
    return (Get-FileHashCompat -Path $LockPath -Algorithm SHA256)
}
function Test-FrontendDependenciesState {
    $info = Get-FrontendLockInfo
    $nodeModulesPath = Join-Path $script:FrontendPath "node_modules"
    $lockHash = Get-FrontendLockHash -LockPath $info.LockPath
    $storedHash = if (Test-Path $info.HashPath) { (Get-Content -Path $info.HashPath -ErrorAction SilentlyContinue | Select-Object -First 1).Trim() }
    $needsInstall = $false
    $reasons = @()
    if (-not (Test-Path $nodeModulesPath)) {
        $needsInstall = $true
        $reasons += "node_modules missing"
    }
    if (-not $lockHash) {
        $needsInstall = $true
        $reasons += "package-lock.json missing"
    }
    elseif ($storedHash -and $lockHash -ne $storedHash) {
        $needsInstall = $true
        $reasons += "package-lock.json changed"
    }
    return @{
        NeedsInstall = $needsInstall
        Reasons = $reasons
        LockHash = $lockHash
        LockHashPath = $info.HashPath
        NodeModulesPath = $nodeModulesPath
        LockPath = $info.LockPath
    }
}
function Save-FrontendLockHash {
    param($State)
    if (-not $State.LockHash) { return }
    $hashDir = Split-Path -Parent $State.LockHashPath
    if (-not (Test-Path $hashDir)) {
        New-Item -Path $hashDir -ItemType Directory -Force | Out-Null
    }
    Set-Content -Path $State.LockHashPath -Value $State.LockHash -Encoding ASCII
}
function Get-NpmVersionInfo {
    try {
        $npmVersion = (npm -v 2>&1 | Select-Object -First 1).Trim()
    }
    catch {
        $npmVersion = ""
    }
    try {
        $nodeVersion = (node -v 2>&1 | Select-Object -First 1).Trim()
    }
    catch {
        $nodeVersion = ""
    }
    return @{ npm = $npmVersion; node = $nodeVersion }
}
function Get-LockfileRequiredNpmMajor {
    param([string]$LockPath)
    if (-not (Test-Path $LockPath)) { return 0 }
    try {
        $lockData = Get-Content -Path $LockPath -Raw | ConvertFrom-Json -ErrorAction Stop
        $lockfileVersion = $lockData.lockfileVersion
        if ($lockfileVersion -ge 3) { return 7 }
        if ($lockfileVersion -eq 2) { return 7 }
        if ($lockfileVersion -eq 1) { return 6 }
    }
    catch {
        Write-Warning "Unable to read package-lock.json for compatibility check"
    }
    return 6
}
function Install-FrontendDependencies {
    <#
    .SYNOPSIS
    Install frontend npm dependencies with auto-fix for common issues
    #>
    param([switch]$ForceClean)
    
    Write-Step "FRONTEND" "Installing dependencies..."
    
    Push-Location $script:FrontendPath
    
    try {
        $state = Test-FrontendDependenciesState
        if (-not $ForceClean -and -not $state.NeedsInstall) {
            Write-Info "Dependencies already match package-lock.json"
            return $true
        }
        
        if ($ForceClean -and (Test-Path $state.NodeModulesPath)) {
            Write-Fix "Force clean enabled - removing existing node_modules and caches..."
            Remove-Item -Path $state.NodeModulesPath -Recurse -Force -ErrorAction SilentlyContinue
            Remove-Item -Path ".vite" -Recurse -Force -ErrorAction SilentlyContinue
            Remove-Item -Path "node_modules/.cache" -Recurse -Force -ErrorAction SilentlyContinue
        }
        
        $npmInfo = Get-NpmVersionInfo
        if (-not $npmInfo.npm) {
            Write-Error "npm is not available in PATH"
            return $false
        }
        if (-not $npmInfo.node) {
            Write-Warning "Node.js version not detected - continuing but build may fail"
        }
        $npmMajor = 0
        if ($npmInfo.npm -match "^(\d+)") { $npmMajor = [int]$matches[1] }
        $requiredNpmMajor = Get-LockfileRequiredNpmMajor -LockPath $state.LockPath
        $preferCi = (Test-Path $state.LockPath) -and ($npmMajor -ge $requiredNpmMajor)
        if (-not $preferCi) {
            Write-Warning "npm $($npmInfo.npm) may not fully match lockfile requirements (need npm $requiredNpmMajor+). Falling back to npm install with peer-deps fixes."
        }
        
        $attempt = 0
        $maxAttempts = 3
        
        while ($attempt -lt $maxAttempts) {
            $attempt++
            $cmdArgs = @()
            if ($preferCi) {
                $cmdArgs = @("ci", "--no-fund", "--no-audit", "--prefer-offline")
            }
            else {
                $cmdArgs = @("install", "--production=false", "--no-fund", "--no-audit")
            }
            if ($attempt -gt 1 -and -not $preferCi) {
                $cmdArgs += "--legacy-peer-deps"
            }
            Write-Info "npm $($cmdArgs -join ' ') attempt $attempt of $maxAttempts..."
            
            $output = npm @cmdArgs 2>&1 | Out-String
            
            if ($LASTEXITCODE -eq 0) {
                Write-Success "Frontend dependencies installed (attempt $attempt)"
                if (Test-Path $state.LockPath) {
                    $state.LockHash = Get-FrontendLockHash -LockPath $state.LockPath
                }
                Save-FrontendLockHash -State $state
                return $true
            }
            
            if ($output -match "EINTEGRITY|integrity checksum|sha512.*mismatch") {
                Write-Fix "Detected integrity error - clearing cache and lock file..."
                npm cache clean --force 2>&1 | Out-Null
                Remove-Item -Path "package-lock.json" -Force -ErrorAction SilentlyContinue
                $state = Test-FrontendDependenciesState
                $preferCi = $false
            }
            elseif ($output -match "ERESOLVE|peer dep|Could not resolve dependency") {
                Write-Fix "Detected dependency conflict - retrying with legacy peer deps..."
                $preferCi = $false
            }
            elseif ($output -match "ENOENT|EACCES|permission denied") {
                Write-Fix "Detected permission/path error - removing node_modules..."
                Remove-Item -Path $state.NodeModulesPath -Recurse -Force -ErrorAction SilentlyContinue
                npm cache clean --force 2>&1 | Out-Null
            }
            elseif ($output -match "ECONNRESET|ETIMEDOUT|network") {
                Write-Fix "Detected network error - waiting and retrying..."
                Start-Sleep -Seconds 5
            }
            elseif ($output -match "EBADENGINE") {
                Write-Error "Node/npm version mismatch detected (EBADENGINE). Please match Node.js version defined in package.json engines."
                return $false
            }
            else {
                Write-Fix "Unknown npm error - clearing cache..."
                npm cache clean --force 2>&1 | Out-Null
            }
            
            if ($attempt -lt $maxAttempts) {
                Start-Sleep -Seconds 3
            }
        }
        
        Write-Error "Failed to install frontend dependencies after $maxAttempts attempts"
        return $false
    }
    finally {
        Pop-Location
    }
}
function Invoke-FrontendBuildStep {
    <#
    .SYNOPSIS
    Build frontend with Vite and auto-fix for common build errors
    #>
    param([switch]$CleanFirst)
    
    Write-Step "FRONTEND" "Building frontend..."
    
    Push-Location $script:FrontendPath
    
    try {
        if (-not (Install-FrontendDependencies -ForceClean:$CleanFirst)) {
            return $false
        }
        
        # Clean dist if requested
        if ($CleanFirst) {
            $distPath = Join-Path $script:FrontendPath "dist"
            if (Test-Path $distPath) {
                Remove-Item -Path $distPath -Recurse -Force
                Write-Debug "Cleaned frontend/dist"
            }
        }
        
        $attempt = 0
        $maxAttempts = 3
        
        while ($attempt -lt $maxAttempts) {
            $attempt++
            Write-Info "Frontend build attempt $attempt of $maxAttempts..."
            
            $output = npm run build 2>&1 | Out-String
            
            if ($LASTEXITCODE -eq 0) {
                Write-Success "Frontend built successfully (attempt $attempt)"
                return $true
            }
            
            # Detect and fix specific build errors
            if ($output -match "ENOMEM|JavaScript heap out of memory") {
                Write-Fix "Detected memory issue - increasing Node.js heap size..."
                $env:NODE_OPTIONS = "--max-old-space-size=4096"
            }
            elseif ($output -match "Cannot find module|Module not found") {
                Write-Fix "Detected missing module - reinstalling dependencies..."
                    if (-not (Install-FrontendDependencies -ForceClean)) {
                        return $false
                    }
            }
            elseif ($output -match "ENOENT|no such file") {
                Write-Fix "Detected missing file - cleaning and rebuilding..."
                Remove-Item -Path ".vite" -Recurse -Force -ErrorAction SilentlyContinue
                Remove-Item -Path "node_modules/.vite" -Recurse -Force -ErrorAction SilentlyContinue
            }
            elseif ($output -match "TypeScript|type error|TS\d{4}") {
                Write-Warning "TypeScript error detected - this requires code fix"
                Write-Error "Frontend build failed due to TypeScript errors"
                Show-BuildError -Output $output -ErrorType "TypeScript"
                return $false
            }
            elseif ($output -match "Could not resolve|Failed to resolve import|Cannot find module") {
                Write-Warning "Import/Module resolution error detected - this requires code fix"
                Write-Error "Frontend build failed due to missing import/module"
                Show-BuildError -Output $output -ErrorType "Import Resolution"
                return $false
            }
            elseif ($output -match "SyntaxError|Unexpected token|Parse error") {
                Write-Warning "Syntax error detected - this requires code fix"
                Write-Error "Frontend build failed due to syntax errors"
                Show-BuildError -Output $output -ErrorType "Syntax"
                return $false
            }
            else {
                # Only retry for unknown errors on first 2 attempts
                if ($attempt -lt $maxAttempts) {
                    Write-Fix "Unknown build error - clearing cache and retrying..."
                    Remove-Item -Path ".vite" -Recurse -Force -ErrorAction SilentlyContinue
                    Remove-Item -Path "node_modules/.cache" -Recurse -Force -ErrorAction SilentlyContinue
                    Start-Sleep -Seconds 3
                }
            }
        }
        
        # If we get here, all attempts failed - show the actual error
        Write-Error "Frontend build failed after $maxAttempts attempts"
        Show-BuildError -Output $output -ErrorType "Build"
        return $false
    }
    finally {
        Pop-Location
    }
}
function Invoke-FrontendBuild {
    <#
    .SYNOPSIS
    Full frontend build process (install + build)
    #>
    param(
        [ValidateSet("quick", "incremental", "full")]
        [string]$Mode = "incremental"
    )
    
    Write-Header "Frontend Build ($Mode)"
    
    switch ($Mode) {
        "quick" {
            Write-Info "Skipping frontend build (quick mode)"
            return $true
        }
        "incremental" {
            return Invoke-FrontendBuildStep
        }
        "full" {
            # Clean install and build
            return Invoke-FrontendBuildStep -CleanFirst
        }
    }
}
