# Compact the podman WSL VM disk.
#
# Why this exists: Windows sees VHDX as fully allocated even after we
# delete images inside the VM. Rebuilding via `build.py smart` grows
# the VHDX but never shrinks it, so a day of iteration accumulates
# 20-30 GB of dead space.
#
# Usage (MUST run as Administrator):
#   powershell -File <abs-path>\scripts\compact-podman-vhd.ps1
#   .\scripts\compact-podman-vhd.ps1 -DryRun      # preview, change nothing
#   .\scripts\compact-podman-vhd.ps1 -NoRestart   # leave VM stopped after
#
# Restart policy: this script restores the state it FOUND, not a fixed one.
# If the machine was running beforehand it is started again (with the same
# containers that were running); if it was stopped it STAYS stopped. Pass
# -NoRestart to force it to stay down regardless. Earlier versions always
# restarted the VM and all 7 omni-* containers, which resurrected a machine
# the operator had deliberately stopped (observed 2026-09-21).
#
# Time: ~2-5 minutes depending on disk speed.
# Safe: WSL is shut down first + no data is touched inside the VM;
# only the VHDX physical size is reclaimed.

param(
    [string]$VhdPath = "$env:USERPROFILE\.local\share\containers\podman\machine\wsl\wsldist\podman-machine-default\ext4.vhdx",
    [switch]$NoRestart,
    [switch]$DryRun
)

# ---- Path check (before the admin gate, so -DryRun works unelevated) ----
if (-not (Test-Path $VhdPath)) {
    Write-Host "[FAIL] VHDX not found: $VhdPath" -ForegroundColor Red
    exit 1
}

$before = (Get-Item $VhdPath).Length / 1GB
$freeBefore = (Get-PSDrive C).Free / 1GB

# ---- Dry run: read-only preview, deliberately BEFORE the admin gate ----
# Previewing changes nothing, so requiring elevation would only stop the
# operator from checking the plan first.
if ($DryRun) {
    Write-Host "=== Compact podman VHDX (DRY RUN) ===" -ForegroundColor Cyan
    Write-Host "  VHDX:  $VhdPath"
    Write-Host "  Size now:         $([math]::Round($before, 2)) GB"
    Write-Host "  C: free now:      $([math]::Round($freeBefore, 1)) GB"
    Write-Host ""

    # Report the state that a real run would restore.
    $dryMachineState = podman machine inspect podman-machine-default --format "{{.State}}" 2>$null
    $dryWasRunning = ($dryMachineState -eq "running")
    Write-Host "  Machine running now: $dryWasRunning"
    if (-not $dryWasRunning) {
        Write-Host "  -> a real run would STOP then leave it STOPPED" -ForegroundColor Yellow
    } elseif ($NoRestart) {
        Write-Host "  -> a real run would STOP it and leave it STOPPED (-NoRestart)" -ForegroundColor Yellow
    } else {
        Write-Host "  -> a real run would STOP it and START it again" -ForegroundColor Yellow
    }
    Write-Host ""
    Write-Host "Would perform:" -ForegroundColor Yellow
    Write-Host "  1. record current machine/container state"
    Write-Host "  2. stop running containers, stop podman machine, wsl --shutdown"
    Write-Host "  3. Optimize-VHD -Mode Full (fallback: diskpart compact vdisk)"
    Write-Host "  4. restore the state found in step 1 (unless -NoRestart)"
    Write-Host ""
    Write-Host "[DRY RUN] Nothing was changed." -ForegroundColor Green
    exit 0
}

# ---- Admin gate (a real compaction rewrites the VHDX) ----
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Host "[FAIL] Run this from an ELEVATED PowerShell (Run as Administrator)." -ForegroundColor Red
    Write-Host "       To preview without elevation, use: -DryRun" -ForegroundColor Yellow
    exit 1
}

Write-Host "=== Compact podman VHDX ===" -ForegroundColor Cyan
Write-Host "  VHDX:  $VhdPath"
Write-Host "  Size before:      $([math]::Round($before, 2)) GB"
Write-Host "  C: free before:   $([math]::Round($freeBefore, 1)) GB"
Write-Host ""

# ---- Record the state we must restore ----
# `podman ps` only works when the machine is up; a stopped machine has no
# running containers by definition, so an empty list is the correct default.
$wasRunning = $false
$runningContainers = @()

$machineState = podman machine inspect podman-machine-default --format "{{.State}}" 2>$null
if ($machineState -eq "running") {
    $wasRunning = $true
    $runningContainers = @(podman ps --format "{{.Names}}" 2>$null)
}

Write-Host "State found: machine running=$wasRunning, containers=$($runningContainers.Count)"
if ($runningContainers.Count -gt 0) {
    Write-Host "  ($($runningContainers -join ', '))"
}
Write-Host ""

# ---- Stop everything cleanly ----
# Stop in reverse dependency order so Postgres gets a graceful shutdown.
$shutdownOrder = @(
    "omni-cloudflared","omni-nginx","omni-frontend","omni-backend",
    "omni-pgbouncer","omni-redis","omni-postgres"
)
Write-Host "Stopping running omni containers gracefully (--time 30)..."
foreach ($c in $shutdownOrder) {
    if ($runningContainers -contains $c) {
        podman stop $c --time 30 2>&1 | Out-Null
    }
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

# ---- Restart (only if it was running before, and not -NoRestart) ----
$shouldRestart = $wasRunning -and (-not $NoRestart)

if (-not $shouldRestart) {
    if ($NoRestart) {
        Write-Host "Machine left STOPPED (-NoRestart)." -ForegroundColor Yellow
    } else {
        Write-Host "Machine was already stopped before this run - leaving it STOPPED." -ForegroundColor Yellow
    }
    Write-Host ""
    Write-Host "To start it later:  podman machine start" -ForegroundColor Cyan
    Write-Host "Then start containers:  cd <repo>; python build.py quick" -ForegroundColor Cyan
    exit 0
}

Write-Host "Restoring state: machine was running, starting it back..."

$null = podman machine start 2>&1
Start-Sleep -Seconds 8
foreach ($c in $runningContainers) {
    podman start $c 2>&1 | Out-Null
}
Start-Sleep -Seconds 5

Write-Host ""
Write-Host "=== Container health ===" -ForegroundColor Cyan
podman ps --format "{{.Names}}`t{{.Status}}" | Where-Object { $_ -match 'omni-' }
