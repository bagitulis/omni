# ============================================
# Nginx Validator Module
# Validate nginx configuration before deploy
# ============================================

. "$PSScriptRoot\config.ps1"
. "$PSScriptRoot\logger.ps1"

function Test-NginxConfigExists {
    <#
    .SYNOPSIS
    Check if required nginx config files exist
    #>
    $mainConfig = Join-Path $script:NginxPath "nginx.conf"
    $tunnelConfig = Join-Path $script:NginxPath "conf.d\tunnel.conf"
    
    $results = @{
        MainConfig = Test-Path $mainConfig
        TunnelConfig = Test-Path $tunnelConfig
        AllExist = $false
    }
    
    $results.AllExist = $results.MainConfig -and $results.TunnelConfig
    
    return $results
}

function Test-NginxRateLimitZones {
    <#
    .SYNOPSIS
    Check for rate limit zone definitions
    #>
    $mainConfig = Join-Path $script:NginxPath "nginx.conf"
    
    if (-not (Test-Path $mainConfig)) {
        return @{ HasZones = $false; Error = "Config file not found" }
    }
    
    $content = Get-Content $mainConfig -Raw
    $hasZones = $content -match "limit_req_zone"
    
    return @{
        HasZones = $hasZones
        Error = if (-not $hasZones) { "No rate limit zones defined" } else { $null }
    }
}

function Test-NginxDuplicateZones {
    <#
    .SYNOPSIS
    Check for duplicate zone definitions across different config files
    #>
    $mainConfig = Join-Path $script:NginxPath "nginx.conf"
    $confDPath = Join-Path $script:NginxPath "conf.d"
    
    $allConfigs = @($mainConfig)
    if (Test-Path $confDPath) {
        $allConfigs += Get-ChildItem -Path $confDPath -Filter "*.conf" | Select-Object -ExpandProperty FullName
    }
    
    $zoneDefinitions = @()
    foreach ($config in $allConfigs) {
        if (Test-Path $config) {
            $content = Get-Content $config -Raw
            $foundZones = [regex]::Matches($content, "limit_req_zone[^;]+zone=([a-zA-Z_][a-zA-Z0-9_]*):")
            foreach ($zone in $foundZones) {
                $zoneName = $zone.Groups[1].Value
                if ($zoneName) {
                    $zoneDefinitions += [PSCustomObject]@{
                        Zone = $zoneName
                        File = $config
                    }
                }
            }
        }
    }
    
    # Check for duplicate zone names defined in DIFFERENT files
    $grouped = $zoneDefinitions | Group-Object -Property Zone | Where-Object { 
        $_.Count -gt 1 -and ($_.Group | Select-Object -ExpandProperty File -Unique).Count -gt 1
    }
    
    return @{
        HasDuplicates = $grouped.Count -gt 0
        Duplicates = $grouped
    }
}

function Test-NginxZeroSizeZones {
    <#
    .SYNOPSIS
    Check for zero-size zone definitions
    #>
    $mainConfig = Join-Path $script:NginxPath "nginx.conf"
    
    if (-not (Test-Path $mainConfig)) {
        return @{ HasZeroSize = $false }
    }
    
    $content = Get-Content $mainConfig -Raw
    $hasZeroSize = $content -match "zone=[^:]+:0[^m]"
    
    return @{
        HasZeroSize = $hasZeroSize
        Error = if ($hasZeroSize) { "Found zero-size zone definition" } else { $null }
    }
}

function Test-NginxSyntaxWithDocker {
    <#
    .SYNOPSIS
    Test nginx config syntax using Docker
    #>
    $mainConfig = Join-Path $script:NginxPath "nginx.conf"
    
    Write-Debug "Testing nginx syntax with Docker..."
    
    try {
        $result = docker run --rm -v "${mainConfig}:/etc/nginx/nginx.conf:ro" nginx:alpine nginx -t 2>&1
        $success = $result -match "successful"
        
        return @{
            Success = $success
            Output = $result
        }
    }
    catch {
        return @{
            Success = $false
            Output = $_.Exception.Message
        }
    }
}

function Invoke-NginxValidation {
    <#
    .SYNOPSIS
    Run full nginx configuration validation
    #>
    Write-Step "VALIDATE" "Validating Nginx configuration..."
    
    $issues = @()
    
    # Check files exist
    $fileCheck = Test-NginxConfigExists
    if (-not $fileCheck.MainConfig) {
        $issues += "nginx.conf not found"
    }
    if (-not $fileCheck.TunnelConfig) {
        $issues += "conf.d/tunnel.conf not found"
    }
    
    if ($issues.Count -gt 0) {
        foreach ($issue in $issues) {
            Write-Error $issue
        }
        return @{ Valid = $false; Issues = $issues }
    }
    
    # Check rate limit zones
    $zoneCheck = Test-NginxRateLimitZones
    if (-not $zoneCheck.HasZones) {
        Write-Warning "No rate limit zones defined (optional)"
    }
    
    # Check for duplicates
    $dupCheck = Test-NginxDuplicateZones
    if ($dupCheck.HasDuplicates) {
        $issues += "Duplicate zone definitions found"
        foreach ($dup in $dupCheck.Duplicates) {
            Write-Error "Duplicate zone '$($dup.Name)' in: $($dup.Group.File -join ', ')"
        }
    }
    
    # Check for zero-size zones
    $zeroCheck = Test-NginxZeroSizeZones
    if ($zeroCheck.HasZeroSize) {
        $issues += "Zero-size zone definition found"
        Write-Error $zeroCheck.Error
    }
    
    if ($issues.Count -eq 0) {
        Write-Success "Nginx configuration is valid"
    }
    
    return @{
        Valid = $issues.Count -eq 0
        Issues = $issues
    }
}

function Repair-NginxCommonIssues {
    <#
    .SYNOPSIS
    Attempt to auto-fix common nginx issues
    #>
    Write-Fix "Attempting to fix nginx issues..."
    
    # Clean nginx volumes
    $nginxVolumes = docker volume ls -q 2>$null | Where-Object { $_ -match "nginx" }
    foreach ($vol in $nginxVolumes) {
        docker volume rm $vol -f 2>&1 | Out-Null
        Write-Debug "Removed volume: $vol"
    }
    
    # Remove nginx container if exists
    docker rm -f omni-nginx 2>&1 | Out-Null
    
    Write-Success "Nginx cleanup complete"
    return $true
}
