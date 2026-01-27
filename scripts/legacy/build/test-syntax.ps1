# Test if docker-operations.ps1 has syntax errors
try {
    $null = & {
        . "C:\Users\PC\Desktop\Project\omni\scripts\build\docker-operations.ps1"
    }
    Write-Host "[OK] No syntax errors found" -ForegroundColor Green
    exit 0
} catch {
    Write-Host "[ERROR] Syntax error: $_" -ForegroundColor Red
    Write-Host "Details: $($_.Exception.Message)"
    exit 1
}
