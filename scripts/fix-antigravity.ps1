# ============================================
# Fix Antigravity Version Script
# Run this if you get "This version of Antigravity is no longer supported"
# ============================================

Write-Host "🔧 Fixing Antigravity Version..." -ForegroundColor Cyan

# 1. Fix fingerprint.js
$fingerprintPath = Get-ChildItem -Path "$env:USERPROFILE\.config\opencode\node_modules" -Recurse -Filter "fingerprint.js" | 
    Where-Object { $_.FullName -match "opencode-antigravity-auth" } | 
    Select-Object -First 1 -ExpandProperty FullName

if ($fingerprintPath) {
    $content = Get-Content $fingerprintPath -Raw
    
    # Check if already patched
    if ($content -match 'ANTIGRAVITY_VERSIONS = \["1\.15\.8"\]') {
        Write-Host "✅ fingerprint.js already patched" -ForegroundColor Green
    } else {
        # Patch it
        $newContent = $content -replace 'const ANTIGRAVITY_VERSIONS = \[.*?\];', 'const ANTIGRAVITY_VERSIONS = ["1.15.8"];'
        $newContent | Set-Content $fingerprintPath -Encoding UTF8
        Write-Host "✅ fingerprint.js patched to 1.15.8" -ForegroundColor Green
    }
} else {
    Write-Host "❌ fingerprint.js not found" -ForegroundColor Red
}

# 2. Fix accounts file - update all userAgent versions
$accountsPath = "$env:APPDATA\opencode\antigravity-accounts.json"
if (Test-Path $accountsPath) {
    $accounts = Get-Content $accountsPath | ConvertFrom-Json
    $fixed = 0
    
    foreach ($acc in $accounts.accounts) {
        if ($acc.fingerprint -and $acc.fingerprint.userAgent -notmatch "1\.15\.8") {
            $acc.fingerprint.userAgent = $acc.fingerprint.userAgent -replace "antigravity/[\d\.]+", "antigravity/1.15.8"
            $fixed++
        }
    }
    
    if ($fixed -gt 0) {
        $accounts | ConvertTo-Json -Depth 10 | Set-Content $accountsPath -Encoding UTF8
        Write-Host "✅ Fixed $fixed account(s) to version 1.15.8" -ForegroundColor Green
    } else {
        Write-Host "✅ All accounts already on 1.15.8" -ForegroundColor Green
    }
} else {
    Write-Host "⚠️ Accounts file not found" -ForegroundColor Yellow
}

Write-Host "`n🎉 Done! You can now run OpenCode." -ForegroundColor Cyan
