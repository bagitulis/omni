param(
    [Parameter(Position=0)]
    [string]$Provider,
    [switch]$List,
    [switch]$Current
)

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$TargetConfig = "$env:USERPROFILE\.config\opencode\oh-my-opencode.json"

$Profiles = @{
    "antigravity" = "oh-my-opencode.json"
    "copilot"     = "oh-my-opencode-copilot.json"
    "gemini"      = "oh-my-opencode-full-gemini.json"
    "openai"      = "oh-my-opencode-openai.json"
}

# Show available providers
if ($List) {
    Write-Host "`nAvailable providers:" -ForegroundColor Cyan
    foreach ($name in $Profiles.Keys | Sort-Object) {
        $profilePath = Join-Path $ScriptDir $Profiles[$name]
        if (Test-Path $profilePath) {
            Write-Host "  $name" -ForegroundColor Green -NoNewline
            Write-Host " -> $($Profiles[$name])"
        } else {
            Write-Host "  $name" -ForegroundColor Red -NoNewline
            Write-Host " -> $($Profiles[$name]) (missing)"
        }
    }
    Write-Host ""
    exit 0
}

# Show current provider
if ($Current) {
    if (-not (Test-Path $TargetConfig)) {
        Write-Host "No active provider config found at: $TargetConfig" -ForegroundColor Red
        exit 1
    }

    try {
        $config = Get-Content $TargetConfig -Raw | ConvertFrom-Json
        $defaultModel = $config.default_model

        $providerName = "Unknown"
        if ($defaultModel -match "^google/antigravity-") {
            $providerName = "Antigravity"
        } elseif ($defaultModel -match "^github-copilot/") {
            $providerName = "Copilot"
        } elseif ($defaultModel -match "^openai/") {
            $providerName = "OpenAI"
        } elseif ($defaultModel -match "^gemini-") {
            $providerName = "Gemini"
        }

        Write-Host "`nCurrent provider: " -NoNewline -ForegroundColor Cyan
        Write-Host $providerName -ForegroundColor Green
        Write-Host "  Default model: $defaultModel" -ForegroundColor Gray
        
        if ($config.thinking -eq $true) {
            Write-Host "  Thinking mode: Enabled (via thinking: true)" -ForegroundColor Yellow
        } elseif ($config.agents) {
            $thinkingAgents = @($config.agents.PSObject.Properties | Where-Object { $_.Value.thinking -eq $true })
            if ($thinkingAgents.Count -gt 0) {
                Write-Host "  Thinking mode: Enabled for $($thinkingAgents.Count) agents" -ForegroundColor Yellow
            } else {
                Write-Host "  Thinking mode: Disabled" -ForegroundColor Gray
            }
        }

        if ($config.agents) {
            Write-Host "  Agents configured: $($config.agents.PSObject.Properties.Count)" -ForegroundColor Gray
        }
        
        if ($config.categories) {
            Write-Host "  Categories configured: $($config.categories.PSObject.Properties.Count)" -ForegroundColor Gray
        }

        Write-Host ""
        exit 0
    } catch {
        Write-Host "Error reading config: $_" -ForegroundColor Red
        exit 1
    }
}

# Validate provider argument
if (-not $Provider) {
    Write-Host "Usage: .\switch-provider.ps1 <provider>" -ForegroundColor Yellow
    Write-Host "       .\switch-provider.ps1 --list" -ForegroundColor Yellow
    Write-Host "       .\switch-provider.ps1 --current" -ForegroundColor Yellow
    Write-Host "`nAvailable providers: $($Profiles.Keys -join ', ')" -ForegroundColor Cyan
    exit 1
}

$Provider = $Provider.ToLower()

if (-not $Profiles.ContainsKey($Provider)) {
    Write-Host "Invalid provider: $Provider" -ForegroundColor Red
    Write-Host "Available providers: $($Profiles.Keys -join ', ')" -ForegroundColor Cyan
    exit 1
}

# Get source profile path
$sourceProfile = Join-Path $ScriptDir $Profiles[$Provider]

if (-not (Test-Path $sourceProfile)) {
    Write-Host "Source profile not found: $sourceProfile" -ForegroundColor Red
    exit 1
}

# Ensure target directory exists
$targetDir = Split-Path -Parent $TargetConfig
if (-not (Test-Path $targetDir)) {
    New-Item -ItemType Directory -Path $targetDir -Force | Out-Null
}

# Copy profile
try {
    Copy-Item -Path $sourceProfile -Destination $TargetConfig -Force
    
    # Read the config to show details
    $config = Get-Content $TargetConfig -Raw | ConvertFrom-Json
    
    Write-Host "`nSwitched to " -NoNewline -ForegroundColor Green
    Write-Host ($Provider.Substring(0,1).ToUpper() + $Provider.Substring(1)) -ForegroundColor Cyan -NoNewline
    Write-Host " provider" -ForegroundColor Green
    
    Write-Host "  Default model: $($config.default_model)" -ForegroundColor Gray
    
    if ($config.thinking -eq $true) {
        Write-Host "  Thinking mode: Enabled (via thinking: true)" -ForegroundColor Yellow
    } elseif ($config.agents) {
        $thinkingAgents = @($config.agents.PSObject.Properties | Where-Object { $_.Value.thinking -eq $true })
        if ($thinkingAgents.Count -gt 0) {
            Write-Host "  Thinking mode: Enabled for $($thinkingAgents.Count) agents" -ForegroundColor Yellow
        } else {
            Write-Host "  Thinking mode: Disabled" -ForegroundColor Gray
        }
    }
    
    if ($config.agents) {
        Write-Host "  Agents configured: $($config.agents.PSObject.Properties.Count)" -ForegroundColor Gray
    }
    
    if ($config.categories) {
        Write-Host "  Categories configured: $($config.categories.PSObject.Properties.Count)" -ForegroundColor Gray
    }
    
    Write-Host ""
    
} catch {
    Write-Host "Error switching provider: $_" -ForegroundColor Red
    exit 1
}
