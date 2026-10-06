#Requires -Version 5.1
$ErrorActionPreference = 'Stop'
$RepoDir = Split-Path -Parent $PSScriptRoot
$Version = (Get-Content (Join-Path $RepoDir 'VERSION') -Raw).Trim()
$Commit = try { git -C $RepoDir rev-parse --short HEAD 2>$null } catch { 'unknown' }
if (-not $Commit) { $Commit = 'unknown' }
$DistDir = Join-Path $RepoDir 'dist'
New-Item -ItemType Directory -Force -Path $DistDir | Out-Null
$OldOS, $OldArch, $OldCGO = $env:GOOS, $env:GOARCH, $env:CGO_ENABLED
Push-Location (Join-Path $RepoDir 'src')
try {
    $env:CGO_ENABLED = '0'
    foreach ($OS in @('linux', 'darwin', 'windows')) {
        foreach ($Arch in @('amd64', 'arm64')) {
            $env:GOOS = $OS
            $env:GOARCH = $Arch
            $Name = "scriptor-$OS-$Arch"
            if ($OS -eq 'windows') { $Name += '.exe' }
            Write-Host "Building $Name..."
            & go build -ldflags "-s -w -X scriptor/cmd.version=$Version -X scriptor/cmd.commit=$Commit" -o (Join-Path $DistDir $Name) .
            if ($LASTEXITCODE -ne 0) { throw "Go build failed with exit code $LASTEXITCODE" }
        }
    }
} finally {
    $env:GOOS = $OldOS
    $env:GOARCH = $OldArch
    $env:CGO_ENABLED = $OldCGO
    Pop-Location
}
