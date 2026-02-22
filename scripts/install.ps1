param(
    [string]$RepoOwner = "naodEthiop",
    [string]$RepoName = "lalibela-cli",
    [string]$BinaryName = "lalibela",
    [string]$Version = "latest",
    [string]$InstallDir = ""
)

$ErrorActionPreference = "Stop"

function Get-LatestVersion {
    param(
        [string]$Owner,
        [string]$Name
    )

    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Owner/$Name/releases/latest"
    if (-not $release.tag_name) {
        throw "Could not determine latest release tag."
    }
    return [string]$release.tag_name
}

function Resolve-Arch {
    switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLowerInvariant()) {
        "x64" { return "amd64" }
        "arm64" { return "arm64" }
        default { throw "Unsupported architecture." }
    }
}

function Ensure-UserPathContains {
    param(
        [string]$PathEntry
    )

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if (-not $userPath) {
        $userPath = ""
    }

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

    if (-not ($env:Path.Split(";", [System.StringSplitOptions]::RemoveEmptyEntries) -contains $PathEntry)) {
        $env:Path = "$env:Path;$PathEntry"
    }
}

if ([string]::IsNullOrWhiteSpace($InstallDir)) {
    $InstallDir = Join-Path $env:LOCALAPPDATA "Programs\Lalibela\bin"
}

$arch = Resolve-Arch
if ($Version -eq "latest") {
    $Version = Get-LatestVersion -Owner $RepoOwner -Name $RepoName
}

$archive = "${BinaryName}_${Version}_windows_${arch}.zip"
$baseUrl = "https://github.com/$RepoOwner/$RepoName/releases/download/$Version"
$tmpDir = Join-Path $env:TEMP ("lalibela-install-" + [guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $tmpDir | Out-Null

try {
    $archivePath = Join-Path $tmpDir $archive
    $checksumPath = Join-Path $tmpDir "checksums.txt"
    $extractDir = Join-Path $tmpDir "extract"
    New-Item -ItemType Directory -Path $extractDir | Out-Null

    Invoke-WebRequest -Uri "$baseUrl/$archive" -OutFile $archivePath
    Invoke-WebRequest -Uri "$baseUrl/checksums.txt" -OutFile $checksumPath

    $expectedLine = Select-String -Path $checksumPath -Pattern (" " + [regex]::Escape($archive) + "$") | Select-Object -First 1
    if (-not $expectedLine) {
        throw "Checksum not found for $archive."
    }
    $expected = ($expectedLine.Line -split "\s+")[0]
    $actual = (Get-FileHash -Path $archivePath -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($actual -ne $expected.ToLowerInvariant()) {
        throw "Checksum verification failed for $archive."
    }

    Expand-Archive -Path $archivePath -DestinationPath $extractDir -Force

    $binary = Get-ChildItem -Path $extractDir -Filter "$BinaryName.exe" -File -Recurse | Select-Object -First 1
    if (-not $binary) {
        throw "Binary not found in archive."
    }

    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    $target = Join-Path $InstallDir "$BinaryName.exe"
    Copy-Item -Path $binary.FullName -Destination $target -Force

    Ensure-UserPathContains -PathEntry $InstallDir

    Write-Host "Installed $BinaryName.exe to $InstallDir"
    Write-Host "Open a new terminal and run: $BinaryName --version"
}
finally {
    if (Test-Path $tmpDir) {
        Remove-Item -Path $tmpDir -Recurse -Force
    }
}
