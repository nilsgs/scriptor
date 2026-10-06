#Requires -Version 5.1
$ErrorActionPreference = 'Stop'
$RepoDir = Split-Path -Parent $PSScriptRoot
$Version = (Get-Content (Join-Path $RepoDir 'VERSION') -Raw).Trim()
$Commit = try { git -C $RepoDir rev-parse --short HEAD 2>$null } catch { 'unknown' }
if (-not $Commit) { $Commit = 'unknown' }
$DistDir = Join-Path $RepoDir 'dist'
New-Item -ItemType Directory -Force -Path $DistDir | Out-Null
Write-Host "Building scriptor v${Version}+${Commit}..."
Push-Location (Join-Path $RepoDir 'src')
try {
    & go build -ldflags "-s -w -X scriptor/cmd.version=$Version -X scriptor/cmd.commit=$Commit" -o (Join-Path $DistDir 'scriptor.exe') .
    if ($LASTEXITCODE -ne 0) { throw "Go build failed with exit code $LASTEXITCODE" }
} finally { Pop-Location }
