# perfmon — install.ps1
# Downloads the latest release binary for Windows and installs it.
# Usage: iwr https://perfmon.qzz.io/windows -useb | iex
#   or:  .\scripts\install.ps1

param(
  [string]$InstallDir = ""
)

$Repo = "GAM3RG33K/perfmon-lite"
$BinName = "perfmon.exe"

# ── Detect architecture ─────────────────────────────────────────────────
$Arch = $env:PROCESSOR_ARCHITECTURE
if ($Arch -eq "AMD64") { $GoArch = "amd64" }
elseif ($Arch -eq "ARM64") {
  Write-Host "  Windows on ARM detected — using x64 binary (via emulation)"
  $GoArch = "amd64"
}
else {
  Write-Host "ERROR: unsupported architecture: $Arch"
  exit 1
}

# ── Resolve install directory ──────────────────────────────────────────
if (-not $InstallDir) {
  $LocalDir = "$env:LOCALAPPDATA\perfmon"
  if (-not (Test-Path $LocalDir)) { New-Item -Type Directory -Force $LocalDir | Out-Null }
  $InstallDir = $LocalDir

  # Check if it's in PATH already
  $InPath = [Environment]::GetEnvironmentVariable("PATH", "User") -match [regex]::Escape($InstallDir)
  if (-not $InPath) {
    [Environment]::SetEnvironmentVariable("PATH", "$InstallDir;$env:PATH", "User")
    $env:PATH = "$InstallDir;$env:PATH"
    Write-Host "  added $InstallDir to user PATH"
  }
}

if (-not (Test-Path $InstallDir)) { New-Item -Type Directory -Force $InstallDir | Out-Null }

# ── Fetch latest release ───────────────────────────────────────────────
Write-Host "  Checking latest release..."
$ApiUrl = "https://api.github.com/repos/$Repo/releases/latest"
try {
  $Latest = (Invoke-RestMethod -Uri $ApiUrl -UseBasicParsing).tag_name
} catch {
  Write-Host "ERROR: could not fetch latest release from GitHub"
  exit 1
}
Write-Host "  latest release: $Latest"

# ── Download binary ────────────────────────────────────────────────────
$Version = $Latest.TrimStart("v")
$Asset = "perfmon_${Version}_windows_${GoArch}.exe"
$Url = "https://github.com/${Repo}/releases/download/${Latest}/${Asset}"

$OutFile = Join-Path $InstallDir $BinName
Write-Host "  downloading: $Url"
try {
  Invoke-WebRequest -Uri $Url -OutFile $OutFile -UseBasicParsing
} catch {
  Write-Host "ERROR: download failed: $_"
  exit 1
}

# ── Verify ─────────────────────────────────────────────────────────────
$Size = (Get-Item $OutFile).Length
Write-Host "  installed: $OutFile ($([math]::Round($Size / 1KB)) KB)"

try {
  $Ver = & $OutFile --version 2>$null
  if (-not $Ver) { throw "no output" }
} catch {
  Write-Host "WARNING: installed binary failed --version check"
  exit 1
}

Write-Host ""
Write-Host "  ─────────────────────────────────────"
Write-Host "   perfmon ${Latest} installed!"
Write-Host "   Binary: $OutFile"
Write-Host "   PATH:   $InstallDir"
Write-Host ""
Write-Host "   Open a new terminal and run:"
Write-Host "     ${BinName} --mock"
Write-Host "  ─────────────────────────────────────"
