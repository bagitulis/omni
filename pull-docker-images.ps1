# ============================================
# Pull Docker Images with Retry Logic
# ============================================

$images = @(
    "cloudflare/cloudflared:latest",
    "redis:7-alpine",
    "nginx:alpine",
    "n8nio/n8n:latest"
)

$maxRetries = 5
$retryDelay = 10

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Pulling Docker Images with Retry Logic" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

foreach ($image in $images) {
    Write-Host "Pulling: $image" -ForegroundColor Yellow
    
    # Check if image already exists
    $existingImage = docker images -q $image 2>$null
    if ($existingImage) {
        Write-Host "  [SKIP] Image already exists locally" -ForegroundColor Green
        Write-Host ""
        continue
    }
    
    $success = $false
    
    for ($i = 1; $i -le $maxRetries; $i++) {
        Write-Host "  Attempt $i/$maxRetries..." -ForegroundColor Gray
        
        try {
            # Increase Docker timeout for large images
            $env:DOCKER_CLIENT_TIMEOUT = "300"
            $env:COMPOSE_HTTP_TIMEOUT = "300"
            
            $result = docker pull $image 2>&1
            
            if ($LASTEXITCODE -eq 0) {
                Write-Host "  [OK] $image pulled successfully" -ForegroundColor Green
                $success = $true
                break
            } else {
                $errorMsg = $result | Out-String
                Write-Host "  [FAILED] Exit code: $LASTEXITCODE" -ForegroundColor Red
                
                # Show last 3 lines of error
                $errorLines = ($errorMsg -split "`n") | Select-Object -Last 3
                foreach ($line in $errorLines) {
                    if ($line.Trim()) {
                        Write-Host "    $line" -ForegroundColor DarkRed
                    }
                }
                
                if ($i -lt $maxRetries) {
                    Write-Host "  Retrying in $retryDelay seconds..." -ForegroundColor Yellow
                    Start-Sleep -Seconds $retryDelay
                }
            }
        } catch {
            Write-Host "  [ERROR] $_" -ForegroundColor Red
            
            if ($i -lt $maxRetries) {
                Write-Host "  Retrying in $retryDelay seconds..." -ForegroundColor Yellow
                Start-Sleep -Seconds $retryDelay
            }
        }
    }
    
    if (-not $success) {
        Write-Host ""
        Write-Host "[CRITICAL] Failed to pull $image after $maxRetries attempts" -ForegroundColor Red
        Write-Host ""
        Write-Host "Possible solutions:" -ForegroundColor Yellow
        Write-Host "  1. Check your internet connection" -ForegroundColor Gray
        Write-Host "  2. Try using a VPN or different network" -ForegroundColor Gray
        Write-Host "  3. Check if firewall/antivirus is blocking Docker" -ForegroundColor Gray
        Write-Host "  4. Wait a few minutes and try again (Docker Hub might be rate limiting)" -ForegroundColor Gray
        Write-Host "  5. Try increasing Docker memory/CPU in Docker Desktop settings" -ForegroundColor Gray
        Write-Host ""
        
        $continue = Read-Host "Continue with remaining images? (y/N)"
        if ($continue -ne "y") {
            Write-Host ""
            Write-Host "Exiting. You can run this script again later." -ForegroundColor Yellow
            exit 1
        }
    }
    
    Write-Host ""
}

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Image Pull Complete" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "You can now run your deployment script." -ForegroundColor Green
