param(
    [string]$Version = "dev"
)

$ErrorActionPreference = "Stop"

$commit = (git rev-parse --short HEAD)
$date = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
$targets = @(
    @{ GOOS = "linux"; GOARCH = "amd64" },
    @{ GOOS = "linux"; GOARCH = "arm64" },
    @{ GOOS = "darwin"; GOARCH = "amd64" },
    @{ GOOS = "darwin"; GOARCH = "arm64" },
    @{ GOOS = "windows"; GOARCH = "amd64" },
    @{ GOOS = "windows"; GOARCH = "arm64" }
)

New-Item -ItemType Directory -Force dist | Out-Null

foreach ($target in $targets) {
    $goos = $target.GOOS
    $goarch = $target.GOARCH
    $ext = ""
    if ($goos -eq "windows") { $ext = ".exe" }
    $out = "dist/lalibela_${goos}_${goarch}${ext}"

    Write-Host "Building $out"
    $env:GOOS = $goos
    $env:GOARCH = $goarch
    $env:CGO_ENABLED = "0"

    go build -ldflags "-s -w -X main.Version=$Version -X main.GitCommit=$commit -X main.BuildDate=$date" -o $out ./cmd/lalibela
}

Write-Host "Cross-platform binaries generated in ./dist"
