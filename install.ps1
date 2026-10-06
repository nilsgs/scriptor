#Requires -Version 5.1
$ErrorActionPreference = 'Stop'
& (Join-Path $PSScriptRoot 'scripts\build.ps1')
& (Join-Path $PSScriptRoot 'scripts\install.ps1')
$InstallDir = Join-Path $env:USERPROFILE '.scriptor\bin'
$UserPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if (-not ($UserPath -split ';' | Where-Object { $_.TrimEnd('\') -eq $InstallDir })) {
    [Environment]::SetEnvironmentVariable('Path', "$InstallDir;$UserPath", 'User')
}
if (-not ($env:PATH -split ';' | Where-Object { $_.TrimEnd('\') -eq $InstallDir })) {
    $env:PATH = "$InstallDir;$env:PATH"
}
Write-Host 'Done. Restart your terminal, then run: scriptor'
