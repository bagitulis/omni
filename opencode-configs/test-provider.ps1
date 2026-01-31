<#
.SYNOPSIS
    Test Oh My OpenCode provider configuration and sub-agents

.DESCRIPTION
    Runs verification tests on the current provider configuration.
    Tests config validity, all agents models, all categories, and triggers switch if errors found.

.PARAMETER Fix
    Auto-fix by switching to copilot if errors detected

.PARAMETER TestModels
    Test actual model calls via opencode run

.EXAMPLE
    .\test-provider.ps1
    Runs all tests and reports PASS/FAIL for each

.EXAMPLE
    .\test-provider.ps1 -Fix
    Run tests and auto-switch to copilot if errors found
#>

param(
    [switch]$Fix,
    [switch]$TestModels,
    [switch]$Verbose
)

$ErrorActionPreference = "Continue"
$TestResults = @()
$TargetConfig = "$env:USERPROFILE\.config\opencode\oh-my-opencode.json"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path

function Write-TestResult {
    param(
        [string]$TestName,
        [bool]$Passed,
        [string]$Message = ""
    )
    
    $status = if ($Passed) { "PASS" } else { "FAIL" }
    $color = if ($Passed) { "Green" } else { "Red" }
    
    Write-Host "[$status] " -ForegroundColor $color -NoNewline
    Write-Host $TestName -NoNewline
    if ($Message) {
        Write-Host " - $Message" -ForegroundColor Gray
    } else {
        Write-Host ""
    }
    
    $script:TestResults += @{
        Name = $TestName
        Passed = $Passed
        Message = $Message
    }
}

function Test-AgentModel {
    param(
        [string]$AgentName,
        [PSObject]$AgentConfig
    )
    
    if (-not $AgentConfig) {
        return @{ Passed = $false; Message = "Agent not defined" }
    }
    
    if (-not $AgentConfig.model) {
        return @{ Passed = $false; Message = "No model specified" }
    }
    
    $model = $AgentConfig.model
    $hasThinking = $AgentConfig.thinking -eq $true
    $hasVariant = $null -ne $AgentConfig.variant
    $hasReasoning = $null -ne $AgentConfig.reasoningEffort
    
    $thinkingStatus = ""
    if ($hasThinking) { $thinkingStatus = " [thinking]" }
    elseif ($hasVariant) { $thinkingStatus = " [variant:$($AgentConfig.variant)]" }
    elseif ($hasReasoning) { $thinkingStatus = " [reasoning:$($AgentConfig.reasoningEffort)]" }
    
    return @{ Passed = $true; Message = "$model$thinkingStatus" }
}

Write-Host ""
Write-Host "Oh My OpenCode Provider Test Suite" -ForegroundColor Cyan
Write-Host "===================================" -ForegroundColor Cyan
Write-Host ""

# Test 1: Config file exists
$configExists = Test-Path $TargetConfig
Write-TestResult -TestName "Config file exists" -Passed $configExists -Message $TargetConfig

if (-not $configExists) {
    Write-Host "`nCRITICAL: Config file not found!" -ForegroundColor Red
    if ($Fix) {
        Write-Host "Attempting to fix by switching to copilot..." -ForegroundColor Yellow
        & "$ScriptDir\switch-provider.ps1" copilot
        exit 0
    }
    exit 1
}

# Test 2: Config is valid JSON
$config = $null
try {
    $config = Get-Content $TargetConfig -Raw | ConvertFrom-Json
    Write-TestResult -TestName "Config is valid JSON" -Passed $true
} catch {
    Write-TestResult -TestName "Config is valid JSON" -Passed $false -Message $_.Exception.Message
    if ($Fix) {
        Write-Host "Attempting to fix by switching to copilot..." -ForegroundColor Yellow
        & "$ScriptDir\switch-provider.ps1" copilot
        exit 0
    }
    exit 1
}

# Test 3: Has default_model
if ($config.default_model) {
    Write-TestResult -TestName "Has default_model" -Passed $true -Message $config.default_model
} else {
    Write-TestResult -TestName "Has default_model" -Passed $false -Message "Missing default_model"
}

# Test 4: Provider detection
$defaultModel = $config.default_model
$providerName = "Unknown"
$providerValid = $false

if ($defaultModel -match "^google/antigravity-") {
    $providerName = "Antigravity"
    $providerValid = $true
} elseif ($defaultModel -match "^github-copilot/") {
    $providerName = "Copilot"
    $providerValid = $true
} elseif ($defaultModel -match "^openai/") {
    $providerName = "OpenAI"
    $providerValid = $true
} elseif ($defaultModel -match "^google/gemini") {
    $providerName = "Gemini"
    $providerValid = $true
}

Write-TestResult -TestName "Provider detected" -Passed $providerValid -Message $providerName

# Test 5: All agents configured
Write-Host ""
Write-Host "Agent Configuration Tests" -ForegroundColor Cyan
Write-Host "-------------------------" -ForegroundColor Cyan

$expectedAgents = @(
    "sisyphus", "oracle", "librarian", "explore", "multimodal-looker",
    "prometheus", "metis", "momus", "atlas", "executor", "reviewer",
    "tester", "security-auditor", "refactorer", "doc-writer", 
    "sisyphus-junior", "default"
)

foreach ($agentName in $expectedAgents) {
    $agentConfig = $config.agents.$agentName
    $result = Test-AgentModel -AgentName $agentName -AgentConfig $agentConfig
    Write-TestResult -TestName "Agent: $agentName" -Passed $result.Passed -Message $result.Message
}

# Test 6: All categories configured
Write-Host ""
Write-Host "Category Configuration Tests" -ForegroundColor Cyan
Write-Host "----------------------------" -ForegroundColor Cyan

$expectedCategories = @(
    "visual-engineering", "artistry", "writing", "quick", "ultrabrain",
    "implementation", "review", "testing", "security", 
    "unspecified-low", "unspecified-high"
)

foreach ($categoryName in $expectedCategories) {
    $categoryConfig = $config.categories.$categoryName
    $result = Test-AgentModel -AgentName $categoryName -AgentConfig $categoryConfig
    Write-TestResult -TestName "Category: $categoryName" -Passed $result.Passed -Message $result.Message
}

# Test 7: Model prefix consistency
Write-Host ""
Write-Host "Model Consistency Tests" -ForegroundColor Cyan
Write-Host "-----------------------" -ForegroundColor Cyan

$allModels = @()
foreach ($agent in ($config.agents | Get-Member -MemberType NoteProperty)) {
    $allModels += $config.agents.$($agent.Name).model
}
foreach ($category in ($config.categories | Get-Member -MemberType NoteProperty)) {
    $allModels += $config.categories.$($category.Name).model
}

$uniquePrefixes = $allModels | ForEach-Object { ($_ -split '/')[0] } | Sort-Object -Unique
$prefixList = $uniquePrefixes -join ", "

if ($uniquePrefixes.Count -eq 1) {
    Write-TestResult -TestName "All models use same provider" -Passed $true -Message $prefixList
} else {
    Write-TestResult -TestName "All models use same provider" -Passed $false -Message "Mixed providers: $prefixList"
}

# Test 8: Thinking mode check for heavy agents
$heavyAgents = @("sisyphus", "prometheus", "atlas", "security-auditor", "ultrabrain")
$thinkingEnabled = 0

foreach ($agentName in $heavyAgents) {
    $agentConfig = $null
    if ($config.agents.$agentName) {
        $agentConfig = $config.agents.$agentName
    } elseif ($config.categories.$agentName) {
        $agentConfig = $config.categories.$agentName
    }
    
    if ($agentConfig) {
        if ($agentConfig.thinking -eq $true -or $agentConfig.variant -or $agentConfig.reasoningEffort) {
            $thinkingEnabled++
        }
    }
}

$thinkingPassed = $thinkingEnabled -ge 3
Write-TestResult -TestName "Heavy agents have thinking mode" -Passed $thinkingPassed -Message "$thinkingEnabled/$($heavyAgents.Count) enabled"

# Optional: Test actual model calls
if ($TestModels) {
    Write-Host ""
    Write-Host "Model Call Tests (Live)" -ForegroundColor Cyan
    Write-Host "-----------------------" -ForegroundColor Cyan
    Write-Host "Testing actual model availability..." -ForegroundColor Yellow
    
    try {
        $testModel = $config.default_model
        $response = & opencode run "Say OK" --model=$testModel 2>&1
        if ($LASTEXITCODE -eq 0 -and $response -notmatch "error|fail|403|401") {
            Write-TestResult -TestName "Model call: $testModel" -Passed $true -Message "Response received"
        } else {
            Write-TestResult -TestName "Model call: $testModel" -Passed $false -Message $response
        }
    } catch {
        Write-TestResult -TestName "Model call: $testModel" -Passed $false -Message $_.Exception.Message
    }
}

# Summary
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Test Summary" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

$passedCount = ($TestResults | Where-Object { $_.Passed }).Count
$failedCount = ($TestResults | Where-Object { -not $_.Passed }).Count
$totalCount = $TestResults.Count
$allPassed = $failedCount -eq 0

Write-Host ""
Write-Host "Total Tests: $totalCount" -ForegroundColor White
Write-Host "Passed: " -NoNewline
Write-Host $passedCount -ForegroundColor Green
Write-Host "Failed: " -NoNewline
Write-Host $failedCount -ForegroundColor $(if ($failedCount -gt 0) { "Red" } else { "Green" })

if ($allPassed) {
    Write-Host ""
    Write-Host "ALL TESTS PASSED!" -ForegroundColor Green
    Write-Host "Provider '$providerName' is fully configured and ready." -ForegroundColor Green
    Write-Host ""
    exit 0
} else {
    Write-Host ""
    Write-Host "SOME TESTS FAILED" -ForegroundColor Red
    
    if ($Fix) {
        Write-Host ""
        Write-Host "Auto-fix enabled. Switching to copilot provider..." -ForegroundColor Yellow
        & "$ScriptDir\switch-provider.ps1" copilot
        Write-Host ""
        Write-Host "Please restart OpenCode and run tests again." -ForegroundColor Yellow
        exit 0
    } else {
        Write-Host "Run with -Fix to auto-switch to copilot provider" -ForegroundColor Yellow
        Write-Host ""
    }
    exit 1
}
