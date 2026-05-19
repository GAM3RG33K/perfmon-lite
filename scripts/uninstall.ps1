# perfmon — uninstall.ps1
# Removes perfmon-tool binary from common install locations.
$Removed = $false

Write-Host "Uninstalling perfmon-tool..."

# Common locations — check both old (perfmon) and new (perfmon-tool)
$Paths = @(
    "$env:LOCALAPPDATA\perfmon\perfmon-tool.exe",
    "$env:LOCALAPPDATA\perfmon",
    "${env:ProgramFiles}\perfmon\perfmon-tool.exe"
)

foreach ($p in $Paths) {
    if (Test-Path $p) {
        if (Test-Path -PathType Leaf $p) {
            Remove-Item -Force $p -ErrorAction SilentlyContinue
            Write-Host "  Removed $p"
            $Removed = $true
        } elseif (Test-Path -PathType Container $p) {
            Remove-Item -Recurse -Force $p -ErrorAction SilentlyContinue
            Write-Host "  Removed directory $p"
            $Removed = $true
        }
    }
}

# Clean up PATH if we added it
$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
$LocalDir = "$env:LOCALAPPDATA\perfmon"
if ($UserPath -and $UserPath.Contains($LocalDir)) {
    $NewPath = ($UserPath -split ";" | Where-Object { $_ -ne $LocalDir }) -join ";"
    [Environment]::SetEnvironmentVariable("PATH", $NewPath, "User")
    Write-Host "  Removed $LocalDir from user PATH"
}

if (-not $Removed) {
    Write-Host "  perfmon-tool not found in common locations."
}

Write-Host ""
Write-Host "  ─────────────────────────────────────"
Write-Host "  Goodbye! Thanks for trying perfmon-tool."
Write-Host ""
Write-Host "  To reinstall:"
Write-Host "    iwr https://get.perfmon.qzz.io/windows -useb | iex"
Write-Host "  ─────────────────────────────────────"
