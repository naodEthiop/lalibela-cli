param(
    [string]$Module = "github.com/naodEthiop/lalibela-cli/cmd/lalibela",
    [string]$Version = "latest",
    [switch]$SkipVerify
)

$ErrorActionPreference = "Stop"

function Ensure-UserPathContains {
    param([string]$PathEntry)

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if (-not $userPath) { $userPath = "" }

    $parts = $userPath.Split(";", [System.StringSplitOptions]::RemoveEmptyEntries)
    $exists = $false
    foreach ($part in $parts) {
        if ($part.TrimEnd("\") -ieq $PathEntry.TrimEnd("\")) {
            $exists = $true
            break
        }
    }

    if (-not $exists) {
        $newPath = if ([string]::IsNullOrWhiteSpace($userPath)) { $PathEntry } else { "$userPath;$PathEntry" }
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    }

    $sessionParts = $env:Path.Split(";", [System.StringSplitOptions]::RemoveEmptyEntries)
    if (-not ($sessionParts | Where-Object { $_.TrimEnd("\") -ieq $PathEntry.TrimEnd("\") })) {
        $env:Path = "$env:Path;$PathEntry"
    }
}

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    throw "Go is not installed or not on PATH."
}

Write-Host "Installing lalibela with go install..." -ForegroundColor Cyan
go install "$Module@$Version"

$gobin = (go env GOBIN).Trim()
$gopath = (go env GOPATH).Trim()
if ([string]::IsNullOrWhiteSpace($gobin)) {
    if ([string]::IsNullOrWhiteSpace($gopath)) {
        throw "Could not determine GOPATH/GOBIN."
    }
    $gobin = Join-Path $gopath "bin"
}

$exePath = Join-Path $gobin "lalibela.exe"
if (-not (Test-Path $exePath)) {
    throw "Install completed but lalibela.exe was not found at $exePath"
}

Ensure-UserPathContains -PathEntry $gobin

if (-not $SkipVerify) {
    if (-not (Get-Command lalibela -ErrorAction SilentlyContinue)) {
        throw "PATH update applied, but this shell still cannot resolve 'lalibela'. Open a new terminal and run 'lalibela --version'."
    }
    $versionOutput = lalibela --version 2>&1
    Write-Host $versionOutput
}

Write-Host "Lalibela installed and PATH configured: $gobin" -ForegroundColor Green
Write-Host "If this is a new shell session, run: lalibela --version"
