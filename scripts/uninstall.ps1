# perfmon — uninstall.ps1
# Removes perfmon-tool binary from common install locations.
$BinName = "perfmon-tool.exe"
$Removed = $false

Write-Host "Uninstalling perfmon..."

# Common locations
$Paths = @(
    "$env:LOCALAPPDATA\perfmon\$BinName",
    "$env:LOCALAPPDATA\perfmon",
    "${env:ProgramFiles}\perfmon\$BinName"
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
    Write-Host "  perfmon not found in common locations."
    Write-Host "  You may have installed it in a custom path — delete it manually."
    exit 0
}

Write-Host "  perfmon uninstalled successfully!"
