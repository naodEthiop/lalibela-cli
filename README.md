# lalibela-cli
![Project Logo](/logo.png)

`lalibela-cli` is the repository for **Lalibela**, a keyboard-first Go CLI (`lalibela`) that scaffolds backend projects.

## Highlights

- `flag`-based CLI 
- Interactive and fast modes
- Framework scaffolds: `gin`, `echo`, `fiber`, `net/http`
- Optional features: `Clean`, `Logger`, `PostgreSQL`, `JWT`, `Docker`
- Rollback on generation failure
- Build metadata support (`Version`, `GitCommit`, `BuildDate`)
- Template catalog output via `-template-list`

## Installation

### Install latest with Go

```bash
go install github.com/naodEthiop/lalibela-cli/cmd/lalibela@latest
```

### Download prebuilt binaries

1. Open GitHub Releases: https://github.com/naodEthiop/lalibela-cli/releases
2. Download the archive for your OS/arch.
3. Add the extracted binary to your `PATH`.

## Usage

### Interactive mode

```bash
lalibela
```

### Fast mode

```bash
lalibela -fast
```

Fast defaults:

- Project name: `myapp`
- Framework: `gin`
- Features: `Logger,PostgreSQL,JWT,Docker`

### Non-interactive feature flags

```bash
lalibela -name myapi -framework gin -features "Clean,Logger,JWT"
```

### Version output

```bash
lalibela -version
```

Expected format:

```text
lalibela X.Y.Z
build date: 2026-02-19T12:00:00Z
commit: a1b2c3d
```

### Template listing

```bash
lalibela -template-list
```

## Optional HOME Config

Lalibela can read defaults from `~/.lalibela.json` (or `%USERPROFILE%\\.lalibela.json` on Windows).

Example:

```json
{
  "project_name": "starter-api",
  "framework": "echo",
  "features": ["Logger", "Docker"],
  "fast": false
}
```

Use a custom config path:

```bash
lalibela -config ./lalibela.json
```

## Frameworks and Feature Support

### Frameworks

- `gin`
- `echo`
- `fiber`
- `nethttp` (`net/http` scaffold)

### Features

- `Clean` (Clean Architecture layer)
- `Logger` (zap logger template)
- `PostgreSQL` (`database/sql` + `lib/pq` template)
- `JWT` middleware stub
- `Docker` Dockerfile template

## Project Layout

```text
cmd/lalibela/main.go
internal/cli/options.go
internal/generator/generate.go
internal/utils/exec.go
templates/main.go.tmpl
templates/env.tmpl
templates/logger.go.tmpl
templates/database.go.tmpl
templates/jwt.go.tmpl
templates/Dockerfile.tmpl
templates/routes/*.tmpl
templates/clean/**.tmpl
.goreleaser.yml
.github/workflows/release.yml
scripts/build-cross.ps1
```

## Development

Run tests:

```bash
go test ./...
```

Run locally:

```bash
go run ./cmd/lalibela
```

## Cross-Platform Builds

### Local PowerShell build matrix

```powershell
./scripts/build-cross.ps1 -Version v0.1.1
```

Artifacts are generated in `./dist`.

### GoReleaser

```bash
goreleaser release --clean
```

## Semantic Versioning and Tags

Use `vX.Y.Z` tags:

```bash
git tag v0.1.1
git push origin v0.1.1
```

Tag push triggers `.github/workflows/release.yml`, which runs GoReleaser and publishes release binaries.

## Build Metadata via ldflags

Use ldflags to stamp binaries:

```bash
go build -ldflags "-X main.Version=v0.1.1 -X main.GitCommit=$(git rev-parse --short HEAD) -X main.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o lalibela ./cmd/lalibela
```

## Contribution Guidelines

1. Fork and create a branch.
2. Add/adjust templates and generator logic with tests.
3. Run `go test ./...`.
4. Open a PR with a clear summary and sample command output.
5. Keep CLI flag behavior backward compatible where possible.

## License

MIT 
