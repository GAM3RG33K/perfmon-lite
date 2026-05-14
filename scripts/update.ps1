# perfmon — update.ps1
# Checks the currently installed version against the latest GitHub release
# and upgrades if a newer version is available.
# Usage: .\scripts\update.ps1

$Repo = "GAM3RG33K/perfmon-lite"
$BinName = "perfmon.exe"

# ── Locate installed binary ────────────────────────────────────────────
$BinPath = Get-Command $BinName -ErrorAction SilentlyContinue
if (-not $BinPath) {
  # Check common locations
  $Paths = @(
    "$env:LOCALAPPDATA\perfmon\$BinName",
    "$env:ProgramFiles\perfmon\$BinName",
    "$env:SystemRoot\system32\$BinName"
  )
  foreach ($p in $Paths) {
    if (Test-Path $p) { $BinPath = $p; break }
  }
}

if (-not $BinPath) {
  Write-Host "ERROR: perfmon not found. Run scripts\install.ps1 first."
  exit 1
}

$BinFullPath = $BinPath.Source -or $BinPath

# ── Detect current version ─────────────────────────────────────────────
try {
  $Output = & $BinFullPath --version 2>&1
  $Current = $Output -replace '.*v', ''
  if (-not $Current) { throw "no version" }
} catch {
  Write-Host "ERROR: could not detect installed version from $BinFullPath"
  exit 1
}
Write-Host "  installed: v${Current}"

# ── Fetch latest release ───────────────────────────────────────────────
Write-Host "  checking GitHub..."
$ApiUrl = "https://api.github.com/repos/$Repo/releases/latest"
try {
  $Latest = (Invoke-RestMethod -Uri $ApiUrl -UseBasicParsing).tag_name
} catch {
  Write-Host "ERROR: could not fetch latest release from GitHub"
  exit 1
}
Write-Host "  latest:     ${Latest}"

# ── Compare versions ───────────────────────────────────────────────────
$LatestStr = $Latest.TrimStart("v")
if ($Current -eq $LatestStr) {
  Write-Host ""
  Write-Host "  ✓ perfmon is already up to date (v${Current})"
  exit 0
}

Write-Host ""
Write-Host "  New version available: ${Latest} (current: v${Current})"

# ── Detect platform ────────────────────────────────────────────────────
$Arch = $env:PROCESSOR_ARCHITECTURE
if ($Arch -eq "AMD64") { $GoArch = "amd64" }
elseif ($Arch -eq "ARM64") {
  Write-Host "  Windows on ARM — using x64 binary (via emulation)"
  $GoArch = "amd64"
}
else {
  Write-Host "ERROR: unsupported architecture: $Arch"
  exit 1
}

# ── Download and replace ───────────────────────────────────────────────
$Version = $Latest.TrimStart("v")
$Asset = "perfmon_${Version}_windows_${GoArch}"
$Url = "https://github.com/${Repo}/releases/download/${Latest}/${Asset}"

$TmpFile = "$env:TEMP\perfmon_update.exe"
Write-Host "  downloading: $Url"
try {
  Invoke-WebRequest -Uri $Url -OutFile $TmpFile -UseBasicParsing
} catch {
  Write-Host "ERROR: download failed: $_"
  exit 1
}

Write-Host "  upgrading:  $BinFullPath"
try {
  Copy-Item $TmpFile $BinFullPath -Force
  Remove-Item $TmpFile -Force
} catch {
  Write-Host "ERROR: could not replace binary (try running as Administrator): $_"
  exit 1
}

# ── Verify ─────────────────────────────────────────────────────────────
try {
  $NewVer = & $BinFullPath --version 2>&1 | ForEach-Object { $_ -replace '.*v', '' }
} catch {
  Write-Host "ERROR: updated binary failed --version check"
  exit 1
}

Write-Host ""
Write-Host "  ─────────────────────────────────────"
Write-Host "   perfmon updated: v${Current} → v${Latest}"
Write-Host "   Binary: $BinFullPath"
Write-Host "  ─────────────────────────────────────"
