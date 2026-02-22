# Lalibela CLI

Production-grade backend scaffolding for Go teams.  
**Lalibela** gives you a fast, modern developer experience (inspired by Vite), but built as a lightweight Go CLI with zero runtime dependencies.

---

## Why Lalibela?

Lalibela was built to improve developer productivity by reducing repetitive setup work and helping teams ship faster with confidence.

It automates common backend workflows, simplifies complex project bootstrapping, and enforces secure defaults from the start so new services begin with a strong baseline.

### The Name and Its Symbolism

The name "Lalibela" is inspired by architectural precision and craftsmanship. It reflects the belief that strong software, like great architecture, is shaped with structure, intention, and durability.

This philosophy guides the CLI design: practical foundations, clear structure, and reliable outcomes for real-world engineering.

---

## Features

- Scaffolds new Go web projects from templates
- Supports multiple server frameworks (`gin`, `echo`, `fiber`, `net/http`)
- Auto-configures `templates/index.html` welcome page
- Starts local development server with `lalibela run`
- Optional browser auto-open (`--open`)
- Interactive and non-interactive modes (`--yes`)
- Actionable errors with command-specific help
- Colorized help/version output for better terminal UX
- Built-in feature installation system (`lalibela add <feature>`)
- Safe self-uninstall command (`lalibela uninstall` with optional `--force`)
- Embedded templates in the binary
- Cross-platform support: Windows, macOS, Linux

---

## Installation

### Option 1: Install with Go

```bash
go install github.com/naodEthiop/lalibela-cli/cmd/lalibela@latest
```

Windows (auto-configure PATH after `go install`):

PowerShell:
```powershell
Invoke-Expression ((Invoke-WebRequest -UseBasicParsing "https://raw.githubusercontent.com/naodEthiop/lalibela-cli/main/scripts/install-go.ps1").Content)
```

Command Prompt (`cmd.exe`):
```cmd
powershell -NoProfile -ExecutionPolicy Bypass -Command "Invoke-Expression ((Invoke-WebRequest -UseBasicParsing 'https://raw.githubusercontent.com/naodEthiop/lalibela-cli/main/scripts/install-go.ps1').Content)"
```

### Option 2: Download prebuilt binaries

1. Open Releases: `https://github.com/naodEthiop/lalibela-cli/releases`
2. Download your OS/architecture archive
3. Extract and add `lalibela` to your `PATH`

### Option 3: One-command installer

Windows (PowerShell):

```powershell
Invoke-Expression ((Invoke-WebRequest -UseBasicParsing "https://raw.githubusercontent.com/naodEthiop/lalibela-cli/main/scripts/install.ps1").Content)
```

macOS/Linux:

```bash
curl -fsSL https://raw.githubusercontent.com/naodEthiop/lalibela-cli/main/scripts/install.sh | sh
```

### Option 4: Package managers (short commands)

Windows Winget (after package approval):

```powershell
winget install NaodEthiop.Lalibela
```

Windows Scoop:

```powershell
scoop install https://raw.githubusercontent.com/naodEthiop/lalibela-cli/main/packaging/scoop/lalibela.json
```

---

## Quick Start

```bash
# 1) Scaffold a new project
lalibela

# 2) Move into the generated folder
cd myapp

# 3) Start development server
go run .
```

Non-interactive:

```bash
lalibela --yes -name myapi -framework gin
cd myapi
go run .
```

---

## CLI Usage

### Root

```bash
lalibela [flags]
lalibela help [command]
```

### Commands

```bash
lalibela add <feature>
lalibela run [--open]
lalibela uninstall [--force]
```

### Common Flags

- `-h, --help` show help
- `-v, --version` print version
- `-y, --yes` auto-accept prompts / non-interactive mode
- `-fast` scaffold with defaults
- `-name <project>` set project name
- `-framework <gin|echo|fiber|nethttp>` select framework
- `-features "Clean,Logger,PostgreSQL,JWT,Docker"` select legacy scaffold features
- `-template-list` print template catalog
- `-config <path>` custom config file path

### Examples

```bash
lalibela
lalibela --yes -name billing-api -framework echo
lalibela -name auth-api -framework gin -features "Logger,JWT,Docker"
lalibela add postgres
lalibela add redis
lalibela run
lalibela run --open
lalibela uninstall
lalibela uninstall --force
lalibela help add
lalibela help run
lalibela help uninstall
```

---

## Example Generated Project Structure

```text
myapi/
|- .env
|- go.mod
|- main.go
|- startup.go
|- templates/
|  |- index.html
|  |- lalibela2.webp
|- internal/
|  |- routes/
|  |  |- routes.go
|  |- middleware/
|  |  |- jwt.go            (if selected)
|  |- config/
|  |  |- config.go         (default production feature)
|  |- logger/
|  |  |- logger.go         (default production feature)
|  |- server/
|     |- health.go
|     |- cors.go
|     |- error_handler.go
|     |- graceful_shutdown.go
|- .lalibela/
   |- features.json
```

---

## Configuration (`~/.lalibela.json`)

```json
{
  "project_name": "starter-api",
  "framework": "gin",
  "features": ["Logger", "Docker"],
  "fast": false
}
```

Use a custom config file:

```bash
lalibela -config ./lalibela.json
```

---

## Roadmap

- Improved plugin ecosystem for third-party templates
- Expanded framework-aware feature patching
- Database migration command workflow
- Optional CI/CD starter profiles
- Shell completion support
- Better project upgrade/diff tooling

---

## Contributing

Contributions are welcome.

1. Fork the repository
2. Create a feature branch
3. Run tests: `go test ./...`
4. Open a PR with:
   - clear summary
   - rationale
   - before/after CLI output when relevant

Please keep changes backward compatible and UX-focused.

---

## License

MIT License.
