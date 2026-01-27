# ============================================
# Logging Utilities
# Uses native PowerShell colors for compatibility
# ============================================

# Log file path
$script:LogFile = $null

function Initialize-Logger {
    param([string]$LogDir)
    
    if (-not (Test-Path $LogDir)) {
        New-Item -ItemType Directory -Path $LogDir -Force | Out-Null
    }
    
    $timestamp = Get-Date -Format "yyyy-MM-dd_HH-mm-ss"
    $script:LogFile = Join-Path $LogDir "build_$timestamp.log"
    
    Write-LogFile "============================================"
    Write-LogFile "Build started at $(Get-Date)"
    Write-LogFile "============================================"
}

function Write-LogFile {
    param([string]$Message)
    
    if ($script:LogFile) {
        $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
        "$timestamp | $Message" | Out-File -FilePath $script:LogFile -Append -Encoding UTF8
    }
}

function Write-Header {
    param([string]$Title)
    
    $line = "=" * 50
    Write-Host ""
    Write-Host $line -ForegroundColor Cyan
    Write-Host "  $Title" -ForegroundColor Cyan
    Write-Host $line -ForegroundColor Cyan
    Write-Host ""
    
    Write-LogFile $line
    Write-LogFile "  $Title"
    Write-LogFile $line
}

function Write-Step {
    param(
        [string]$Step,
        [string]$Message
    )
    
    Write-Host "[$Step] " -ForegroundColor Blue -NoNewline
    Write-Host $Message
    Write-LogFile "[$Step] $Message"
}

function Write-Success {
    param([string]$Message)
    
    Write-Host "[OK] " -ForegroundColor Green -NoNewline
    Write-Host $Message
    Write-LogFile "[OK] $Message"
}

function Write-Warning {
    param([string]$Message)
    
    Write-Host "[WARN] " -ForegroundColor Yellow -NoNewline
    Write-Host $Message
    Write-LogFile "[WARN] $Message"
}

function Write-Error {
    param([string]$Message)
    
    Write-Host "[ERROR] " -ForegroundColor Red -NoNewline
    Write-Host $Message
    Write-LogFile "[ERROR] $Message"
}

function Write-Info {
    param([string]$Message)
    
    Write-Host "[INFO] " -ForegroundColor Gray -NoNewline
    Write-Host $Message
    Write-LogFile "[INFO] $Message"
}

function Write-Debug {
    param([string]$Message)
    
    Write-Host "[DEBUG] " -ForegroundColor DarkGray -NoNewline
    Write-Host $Message
    Write-LogFile "[DEBUG] $Message"
}

function Write-Fix {
    param([string]$Message)
    
    Write-Host "[AUTO-FIX] " -ForegroundColor Magenta -NoNewline
    Write-Host $Message
    Write-LogFile "[AUTO-FIX] $Message"
}

function Write-Retry {
    param(
        [int]$Attempt,
        [int]$MaxAttempts,
        [string]$Message
    )
    
    Write-Host "[RETRY $Attempt/$MaxAttempts] " -ForegroundColor Yellow -NoNewline
    Write-Host $Message
    Write-LogFile "[RETRY $Attempt/$MaxAttempts] $Message"
}

function Write-Progress {
    param(
        [int]$Current,
        [int]$Total,
        [string]$Activity
    )
    
    $percent = [math]::Round(($Current / $Total) * 100)
    $filled = [math]::Floor($percent / 5)
    $empty = 20 - $filled
    $progressBar = "[" + ("=" * $filled) + ("-" * $empty) + "]"
    
    Write-Host "`r$progressBar $percent% $Activity    " -ForegroundColor Cyan -NoNewline
    
    if ($Current -eq $Total) {
        Write-Host ""
    }
}

function Get-LogFilePath {
    return $script:LogFile
}
