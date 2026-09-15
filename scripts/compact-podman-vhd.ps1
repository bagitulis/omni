# Compact the podman WSL VM disk.
#
# Why this exists: Windows sees VHDX as fully allocated even after we
# delete images inside the VM. Rebuilding via `build.py smart` grows
# the VHDX but never shrinks it, so a day of iteration accumulates
# 20-30 GB of dead space.
#
# Usage (MUST run as Administrator):
#   1. Right-click PowerShell 7 → Run as Administrator
#   2. cd C:\Users\PC\Documents\Project\omni
#   3. .\scripts\compact-podman-vhd.ps1
#
# Time: ~2-5 minutes depending on disk speed.
# Safe: WSL is shut down first + no data is touched inside the VM;
# only the VHDX physical size is reclaimed.

param(
    [string]$VhdPath = "$env:USERPROFILE\.local\share\containers\podman\machine\wsl\wsldist\podman-machine-default\ext4.vhdx"
)

# ---- Admin gate ----
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Host "[FAIL] Run this from an ELEVATED PowerShell (Run as Administrator)." -ForegroundColor Red
    exit 1
}

if (-not (Test-Path $VhdPath)) {
    Write-Host "[FAIL] VHDX not found: $VhdPath" -ForegroundColor Red
    exit 1
}

$before = (Get-Item $VhdPath).Length / 1GB
$freeBefore = (Get-PSDrive C).Free / 1GB
Write-Host "=== Compact podman VHDX ===" -ForegroundColor Cyan
Write-Host "  VHDX:  $VhdPath"
Write-Host "  Size before:      $([math]::Round($before, 2)) GB"
Write-Host "  C: free before:   $([math]::Round($freeBefore, 1)) GB"
Write-Host ""

# ---- Stop everything cleanly ----
Write-Host "Stopping omni containers..."
$containers = @("omni-cloudflared","omni-nginx","omni-frontend","omni-backend","omni-pgbouncer","omni-redis","omni-postgres")
foreach ($c in $containers) {
    podman stop $c --time 10 2>&1 | Out-Null
}

Write-Host "Stopping podman machine..."
podman machine stop 2>&1 | Out-Null

Write-Host "Shutting down WSL..."
wsl.exe --shutdown 2>&1 | Out-Null
Start-Sleep -Seconds 5

# ---- Compact ----
Write-Host ""
Write-Host "Running Optimize-VHD (2-5 minutes)..." -ForegroundColor Yellow
$t0 = Get-Date
try {
    Import-Module Hyper-V -ErrorAction Stop
    Optimize-VHD -Path $VhdPath -Mode Full -ErrorAction Stop
    $dur = ((Get-Date) - $t0).TotalSeconds
    Write-Host "  Done in $([math]::Round($dur, 1))s" -ForegroundColor Green
} catch {
    Write-Host "  [FAIL] $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "  Falling back to diskpart..."
    $diskpartScript = @"
select vdisk file="$VhdPath"
attach vdisk readonly
compact vdisk
detach vdisk
exit
"@
    $tmp = [System.IO.Path]::GetTempFileName()
    Set-Content -Path $tmp -Value $diskpartScript
    diskpart /s $tmp
    Remove-Item $tmp -Force
}

# ---- Report ----
$after = (Get-Item $VhdPath).Length / 1GB
$freeAfter = (Get-PSDrive C).Free / 1GB
Write-Host ""
Write-Host "=== Results ===" -ForegroundColor Cyan
Write-Host "  Size before:      $([math]::Round($before, 2)) GB"
Write-Host "  Size after:       $([math]::Round($after, 2)) GB"
Write-Host "  Reclaimed:        $([math]::Round(($before - $after), 2)) GB" -ForegroundColor Green
Write-Host "  C: free before:   $([math]::Round($freeBefore, 1)) GB"
Write-Host "  C: free after:    $([math]::Round($freeAfter, 1)) GB"
Write-Host ""

# ---- Restart ----
Write-Host "Restarting podman machine + containers..."
podman machine start 2>&1 | Out-Null
Start-Sleep -Seconds 8
foreach ($c in $containers) {
    podman start $c 2>&1 | Out-Null
}
Start-Sleep -Seconds 5
Write-Host ""
Write-Host "=== Container health ===" -ForegroundColor Cyan
podman ps --format "{{.Names}}`t{{.Status}}" | Where-Object { $_ -match 'omni-' }
