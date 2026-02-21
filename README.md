# Lalibela CLI

Production-grade backend scaffolding for Go teams.  
**Lalibela** gives you a fast, modern developer experience (inspired by Vite), but built as a lightweight Go CLI with zero runtime dependencies.

---

## Why this tool?

Frontend tooling set the bar for great DX: instant setup, clear output, and sensible defaults. Backend tooling often still feels slow and fragmented.

Lalibela closes that gap for Go backend development:

- Fast project generation from templates
- Clean CLI UX with actionable output
- Modern startup workflow (`go run .`)
- Template-driven and framework-aware
- Single binary, no runtime dependency chain

---

## Features

- Scaffolds new Go web projects from templates
- Supports multiple server frameworks (`gin`, `echo`, `fiber`, `net/http`)
- Auto-configures `templates/index.html` welcome page
- Starts local development server with `lalibela run`
- Optional browser auto-open (`--open`)
- Interactive mode and non-interactive mode (`--yes`)
- Actionable error messages and command help
- Built-in feature installation system (`lalibela add <feature>`)
- Embedded templates in the binary
- Cross-platform support: Windows, macOS, Linux

---

## Installation

### Option 1: Install with Go

```bash
go install github.com/naodEthiop/lalibela-cli/cmd/lalibela@latest
```

### Option 2: Download prebuilt binaries

1. Open Releases: `https://github.com/naodEthiop/lalibela-cli/releases`
2. Download your OS/architecture archive
3. Extract and add `lalibela` to your `PATH`

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
lalibela help add
lalibela help run
```

---

## Example Generated Project Structure

```text
myapi/
+- .env
+- go.mod
+- main.go
+- startup.go
+- templates/
¦  +- index.html
¦  +- lalibela2.webp
+- internal/
¦  +- routes/
¦  ¦  +- routes.go
¦  +- middleware/
¦  ¦  +- jwt.go            (if selected)
¦  +- config/
¦  ¦  +- config.go         (default production feature)
¦  +- logger/
¦  ¦  +- logger.go         (default production feature)
¦  +- server/
¦     +- health.go
¦     +- cors.go
¦     +- error_handler.go
¦     +- graceful_shutdown.go
+- .lalibela/
   +- features.json
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
