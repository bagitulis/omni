# ============================================
# Build Orchestrator - Main Entry Point
# SRP: Menu UI and build mode orchestration
# ============================================

param(
    [ValidateSet("menu", "quick", "smart", "full", "clean", "validate")]
    [string]$Mode = "menu",
    [ValidateSet("lowspec", "standard", "highspec")]
    [string]$Spec = "standard",
    [bool]$ShowBuildOutput = $true
)

# Note: Previously had $SkipChecks and $Silent parameters - removed as unused
# Add back if implementing silent/skip mode in the future

# Load modules
$ScriptDir = $PSScriptRoot
. "$ScriptDir\config.ps1"
. "$ScriptDir\logger.ps1"
. "$ScriptDir\environment-checker.ps1"
. "$ScriptDir\docker-core.ps1"
. "$ScriptDir\docker-repair.ps1"
. "$ScriptDir\docker-assert.ps1"
. "$ScriptDir\cleanup.ps1"
. "$ScriptDir\nginx-validator.ps1"
. "$ScriptDir\retry-logic.ps1"
. "$ScriptDir\error-recovery.ps1"
. "$ScriptDir\frontend-builder.ps1"
. "$ScriptDir\backend-builder.ps1"
. "$ScriptDir\docker-operations.ps1"
. "$ScriptDir\post-deploy.ps1"
. "$ScriptDir\build-modes.ps1"

Set-Location $script:ProjectRoot
# Note: Logging disabled - uncomment below line to enable log file creation
# Initialize-Logger -LogDir (Join-Path $script:ProjectRoot "logs")

# ============================================
# Menu UI
# ============================================

function Show-Menu {
    Clear-Host
    Write-Host ""
    Write-Host "============================================" -ForegroundColor Cyan
    Write-Host "  BUILD AND DEPLOY PIPELINE (Auto-Fix)" -ForegroundColor Cyan
    Write-Host "============================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "  All errors are automatically detected and fixed!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Choose deployment mode:"
    Write-Host ""
    Write-Host "  [1] Quick Apply      - Restart only (config changes)"
    Write-Host "  [2] Smart Build      - Cache deps, Fresh code [RECOMMENDED]" -ForegroundColor Yellow
    Write-Host "  [3] Full Rebuild     - Clean rebuild (dependency changes)"
    Write-Host "  ---"
    Write-Host "  [4] Validate Only    - Check environment and nginx"
    Write-Host "  [5] Clean Docker     - Aggressive cleanup"
    Write-Host "  [0] Exit"
    Write-Host ""
    Write-Host "  TIP: Use [2] Smart Build for fastest code changes!" -ForegroundColor Yellow
    Write-Host ""
    return (Read-Host "Enter choice (0-5)").Trim()
}

function Show-SpecMenu {
    Write-Host ""
    Write-Host "[STEP 2] Choose Server RAM Specification:"
    Write-Host ""
    Write-Host "  [1] Low Spec (2GB RAM) - VPS kecil"
    Write-Host "  [2] Standard (4GB RAM) - Production [RECOMMENDED]"
    Write-Host "  [3] High Spec (8GB+ RAM) - High traffic"
    Write-Host "  [0] Back"
    Write-Host ""
    return (Read-Host "Enter choice (0-3)").Trim()
}

# ============================================
# Interactive Mode Handler
# ============================================

function Invoke-InteractiveMode {
    :menuLoop while ($true) {
        $choice = Show-Menu
        
        switch ($choice) {
            "0" { Write-Host "Exiting..."; return }
            "1" { $buildMode = "quick" }
            "2" { $buildMode = "smart" }
            "3" { $buildMode = "full" }
            "4" { Invoke-ValidateOnly; continue menuLoop }
            "5" { Invoke-CleanDocker; continue menuLoop }
            default {
                Write-Host "Invalid choice" -ForegroundColor Red
                Start-Sleep -Seconds 1
                continue menuLoop
            }
        }
        
        # Spec choice (skip for quick)
        $specChoice = "standard"
        if ($buildMode -ne "quick") {
            $specInput = Show-SpecMenu
            switch ($specInput) {
                "0" { continue menuLoop }
                "1" { $specChoice = "lowspec" }
                "2" { $specChoice = "standard" }
                "3" { $specChoice = "highspec" }
                default { Write-Host "Invalid" -ForegroundColor Red; continue menuLoop }
            }
        }
        
        # Execute build
        $result = Invoke-BuildMode -Mode $buildMode -Spec $specChoice -ShowBuildOutput $ShowBuildOutput
        
        # Show result
        Show-BuildResult -Success $result
        
        Write-Host ""
        Write-Host "Press any key to return to menu..." -ForegroundColor Yellow
        $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
    }
}

# ============================================
# Non-Interactive Mode Handler
# ============================================

function Invoke-NonInteractiveMode {
    param([string]$Mode, [string]$Spec)
    
    if ($Mode -eq "clean") {
        Invoke-CleanDocker
        return $true
    }
    
    if ($Mode -eq "validate") {
        Invoke-ValidateOnly
        return $true
    }
    
    $result = Invoke-BuildMode -Mode $Mode -Spec $Spec -ShowBuildOutput $ShowBuildOutput
    Show-BuildResult -Success $result
    return $result
}

# ============================================
# Result Display
# ============================================

function Show-BuildResult {
    param([bool]$Success)
    
    Write-Host ""
    if ($Success) {
        Write-Host "============================================" -ForegroundColor Green
        Write-Host "  BUILD COMPLETED SUCCESSFULLY" -ForegroundColor Green
        Write-Host "============================================" -ForegroundColor Green
    } else {
        Write-Host "============================================" -ForegroundColor Red
        Write-Host "  BUILD FAILED" -ForegroundColor Red
        Write-Host "============================================" -ForegroundColor Red
        Write-Host ""
        Write-Host "Log file: $(Get-LogFilePath)" -ForegroundColor Yellow
    }
}

# ============================================
# Utility Modes
# ============================================

function Invoke-ValidateOnly {
    Write-Header "Validation Mode"
    $envStatus = Get-EnvironmentStatus
    Show-EnvironmentStatus -Status $envStatus
    Invoke-NginxValidation
    Write-Host ""
    Write-Host "Press any key to continue..."
    $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
}

function Invoke-CleanDocker {
    Write-Header "Aggressive Docker Cleanup"
    Write-Warning "This will remove ALL Docker images, volumes, and cache!"
    Write-Host ""
    $confirm = Read-Host "Are you sure? (y/N)"
    
    if ($confirm -eq "y" -or $confirm -eq "Y") {
        Invoke-AggressiveCleanup
        Write-Host ""
        Write-Host "Next steps:" -ForegroundColor Yellow
        Write-Host "  1. Restart Docker Desktop (recommended)"
        Write-Host "  2. Run deployment again"
        Write-Host ""
    } else {
        Write-Info "Cancelled"
    }
    
    Write-Host "Press any key to continue..."
    $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
}

# ============================================
# Main Entry Point
# ============================================

if ($Mode -eq "menu") {
    Invoke-InteractiveMode
} else {
    Invoke-NonInteractiveMode -Mode $Mode -Spec $Spec
}
