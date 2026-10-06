#Requires -Version 5.1
$ErrorActionPreference = 'Stop'
$RepoDir = Split-Path -Parent $PSScriptRoot
$InstallDir = Join-Path $env:USERPROFILE '.scriptor\bin'
$Artifact = Join-Path $RepoDir 'dist\scriptor.exe'
if (-not (Test-Path -LiteralPath $Artifact)) { throw 'Run scripts/build.ps1 first.' }
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
Copy-Item -LiteralPath $Artifact -Destination (Join-Path $InstallDir 'scriptor.exe') -Force
Write-Host "Installed scriptor to $InstallDir"
