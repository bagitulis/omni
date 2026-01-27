# ============================================
# Retry Logic Module
# Exponential backoff and smart retry strategies
# ============================================

. "$PSScriptRoot\config.ps1"
. "$PSScriptRoot\logger.ps1"

function Invoke-WithRetry {
    <#
    .SYNOPSIS
    Execute a script block with retry logic and exponential backoff
    
    .PARAMETER ScriptBlock
    The code to execute
    
    .PARAMETER MaxRetries
    Maximum number of retry attempts
    
    .PARAMETER RetryDelayBase
    Base delay in seconds (will be multiplied for exponential backoff)
    
    .PARAMETER RetryMessage
    Message to display on retry
    
    .PARAMETER OnRetry
    Script block to execute before retrying (e.g., cleanup)
    #>
    param(
        [Parameter(Mandatory)]
        [scriptblock]$ScriptBlock,
        
        [int]$MaxRetries = $script:MaxRetries,
        [int]$RetryDelayBase = $script:RetryDelayBase,
        [string]$RetryMessage = "Retrying...",
        [scriptblock]$OnRetry = $null
    )
    
    $attempt = 0
    $lastError = $null
    
    while ($attempt -lt $MaxRetries) {
        $attempt++
        
        try {
            $result = & $ScriptBlock
            
            # Check if script block returned success
            if ($LASTEXITCODE -eq 0 -or $result -eq $true) {
                return @{
                    Success = $true
                    Result = $result
                    Attempts = $attempt
                }
            }
            else {
                throw "Command failed with exit code $LASTEXITCODE"
            }
        }
        catch {
            $lastError = $_
            
            if ($attempt -lt $MaxRetries) {
                $delay = $RetryDelayBase * [math]::Pow(2, $attempt - 1)
                Write-Retry -Attempt $attempt -MaxAttempts $MaxRetries -Message "$RetryMessage (waiting ${delay}s)"
                
                # Execute retry callback if provided
                if ($OnRetry) {
                    & $OnRetry
                }
                
                Start-Sleep -Seconds $delay
            }
        }
    }
    
    return @{
        Success = $false
        Error = $lastError
        Attempts = $attempt
    }
}

function Invoke-WithProgressiveFixRetry {
    <#
    .SYNOPSIS
    Execute with retry, applying progressively more aggressive fixes on each failure
    
    .PARAMETER ScriptBlock
    The code to execute
    
    .PARAMETER FixStrategies
    Array of script blocks, each representing a more aggressive fix strategy
    
    .PARAMETER OperationName
    Name of the operation for logging
    #>
    param(
        [Parameter(Mandatory)]
        [scriptblock]$ScriptBlock,
        
        [Parameter(Mandatory)]
        [scriptblock[]]$FixStrategies,
        
        [string]$OperationName = "Operation"
    )
    
    # First try without any fixes
    try {
        $result = & $ScriptBlock
        if ($LASTEXITCODE -eq 0 -or $result -eq $true) {
            return @{
                Success = $true
                Result = $result
                FixLevel = 0
            }
        }
    }
    catch {
        Write-Debug "Initial attempt failed: $_"
    }
    
    # Apply fixes progressively
    for ($i = 0; $i -lt $FixStrategies.Count; $i++) {
        $fixLevel = $i + 1
        Write-Fix "Applying fix strategy level $fixLevel for $OperationName..."
        
        try {
            # Apply the fix
            & $FixStrategies[$i]
            
            # Retry the operation
            $result = & $ScriptBlock
            if ($LASTEXITCODE -eq 0 -or $result -eq $true) {
                Write-Success "$OperationName succeeded after fix level $fixLevel"
                return @{
                    Success = $true
                    Result = $result
                    FixLevel = $fixLevel
                }
            }
        }
        catch {
            Write-Debug "Attempt with fix level $fixLevel failed: $_"
        }
    }
    
    return @{
        Success = $false
        Error = "All fix strategies exhausted for $OperationName"
        FixLevel = $FixStrategies.Count
    }
}

function Test-CommandSuccess {
    <#
    .SYNOPSIS
    Test if a command completed successfully
    #>
    param(
        [string]$Command,
        [string[]]$Arguments
    )
    
    try {
        $null = & $Command @Arguments 2>&1
        return $LASTEXITCODE -eq 0
    }
    catch {
        return $false
    }
}

function Wait-ForCondition {
    <#
    .SYNOPSIS
    Wait for a condition to become true
    
    .PARAMETER Condition
    Script block that returns true when condition is met
    
    .PARAMETER TimeoutSeconds
    Maximum seconds to wait
    
    .PARAMETER PollingIntervalSeconds
    Seconds between condition checks
    
    .PARAMETER ActivityMessage
    Message to display during wait
    #>
    param(
        [Parameter(Mandatory)]
        [scriptblock]$Condition,
        
        [int]$TimeoutSeconds = 60,
        [int]$PollingIntervalSeconds = 2,
        [string]$ActivityMessage = "Waiting for condition..."
    )
    
    $elapsed = 0
    
    while ($elapsed -lt $TimeoutSeconds) {
        try {
            $conditionResult = & $Condition
            if ($conditionResult) {
                return $true
            }
        }
        catch {
            # Condition threw an exception - continue polling
            Write-Verbose "Condition check failed: $_"
        }
        
        Write-Progress -Current $elapsed -Total $TimeoutSeconds -Activity $ActivityMessage
        Start-Sleep -Seconds $PollingIntervalSeconds
        $elapsed += $PollingIntervalSeconds
    }
    
    Write-Host ""  # Clear progress line
    return $false
}
