# Lalibela CLI — v1 Stabilization + Roadmap Spec

**Author:** Naod (Fkremariam Fentahun Masresha)
**Repo:** github.com/naodEthiop/lalibela-cli
**Status:** Draft — based on a full codebase audit (2026-09-19)
**Purpose:** Fix what's broken and misleading today, then build toward a "Vite for Go" ecosystem position.

---

## 0. Why this spec exists

Lalibela is a Go project scaffolder (Gin/Echo/Fiber/net-http) with real installers (winget, Scoop, install scripts, `go install`), 45+ commits, and genuinely solid pieces — `internal/generator` and `internal/cli` are well tested, gin/echo scaffolds compile cleanly, and `internal/updater`'s design (HTTPS + SHA256 + GPG + atomic replace) is sound.

But an audit surfaced a credibility gap between what the README promises and what the code does, plus one confirmed installer bug a real user hit in production. This spec sequences the fix, in order: **stop misleading users → stop being broken → then build the "popular like Vite" roadmap.**

Everything in Phase 2+ is pointless if Phase 0 isn't done — a nicer plugin system on top of a feature system that silently does nothing is worse, not better.

---

## Phase 0 — Stop the Bleeding (do this first, before anything else)

### 0.1 Fix `install.sh` 404 on macOS — **[GitHub Issue #2](https://github.com/naodEthiop/lalibela-cli/issues/2)**
**Confirmed bug, filed by a real user (`manfouothierno`, opened Mar 10, 2026):**
```
curl -fsSL https://raw.githubusercontent.com/naodEthiop/lalibela-cli/main/scripts/install.sh | sh
curl: (56) The requested URL returned error: 404
```
`go install` also failed for them ("not working too").

**Root cause to verify and fix:**
- Confirm `scripts/install.sh` actually exists on `main` at that raw URL today (it may have been moved, renamed, or the release asset path inside the script points at a tag/asset that no longer exists).
- Check whether the script's *internal* logic 404s (e.g. it constructs a release download URL using a tag naming convention that doesn't match actual GitHub release tags, or assumes an asset filename pattern that changed).
- Check why `go install github.com/naodEthiop/lalibela-cli/cmd/lalibela@latest` failed for this user — likely candidates: no tagged release resolvable as `@latest` via the Go module proxy, a `go.mod` module path mismatch, or the module not being properly tagged/pushed.
- Fix must be verified end-to-end on macOS (or a Linux equivalent if macOS isn't available) — actually run the curl command against the fixed script, not just read the diff.

### 0.2 Fix `lalibela update` — non-functional for every real install
`internal/updater.resolvePublicKeyPath()` requires `~/.lalibela/public.gpg` or `LALIBELA_UPDATE_PUBLIC_KEY`. **No install path provisions either.** Every `lalibela update` fails with "missing update public key" out of the box.
- Embed the public key in the binary at build time (preferred — no provisioning step needed on any platform), **or**
- Ship the key via every install path (install.sh, install.ps1, winget manifest, scoop manifest) — more surface area to keep in sync, only pick this if there's a reason not to embed.
- Add an integration test that runs `lalibela update --dry-run` (or equivalent) right after a fresh install simulation and asserts it does NOT fail on key resolution.

### 0.3 Remove `.gocache/` from git
109MB, 1,916 files of raw Go build cache blobs, committed. This is the single worst repo-hygiene issue.
- `git rm -r --cached .gocache/`, add to `.gitignore`.
- Because this is 109MB across history, consider whether to also purge it from git history (`git filter-repo` or BFG) — decide based on whether repo clone size is actually a pain point yet; at minimum stop it from growing further.

### 0.4 Delete confirmed dead files
- `.goreleaser.yml` — dead duplicate of `.goreleaser.yaml` (CI uses `--config .goreleaser.yaml` explicitly). The `.yml` version is missing the GPG `signs:` block and has a stray BOM on line 1. Keeping it risks a future editor changing "the" config and silently not affecting releases.
- `test.txt` — leftover UTF-16 garbage ("signed test") from a commit-signing test.
- Confirm with the repo owner (you) before removing `google790fa80a4bea91a1.html` — only needed if this repo is ever served as a website.
- Add a `.gitignore` convention for test-run scaffolds (this audit session alone produced `b/`, `fgd/`, `gf/`, `el/` as untracked leftovers from local CLI testing) — either a wildcard pattern or a documented convention of always testing in a scratch dir outside the repo.

### 0.5 Wire the "production features," or stop calling them that
**This is the core credibility issue.** Every feature — the 6 always-on "default production features" (`internal/server/health.go`, `cors.go`, `error_handler.go`, `graceful_shutdown.go`, `internal/config/config.go`, `internal/logger/logger.go`) and every opt-in feature (JWT, PostgreSQL, Clean Architecture) — generates files that the scaffolded `main.go`/`startup.go`/`routes.go` **never imports or calls**. Everything compiles because each package is self-contained, but functionally these features are decorative.

Specifically:
- `routes.go.tmpl` hand-rolls its own inline `/health` handler duplicating `internal/server/health.go`, which is never referenced.
- `config/database.go` (PostgreSQL) and `internal/middleware/jwt.go` (JWT) are generated but never imported anywhere.
- The entire Clean Architecture layer (`internal/domain`, `usecase`, `repository`, `delivery/httptransport`, `internal/app/bootstrap.go`) — `app.NewContainer()` is never called from `main.go`.

**Fix, in order of preference:**
1. **Best:** Wire each feature into the generated `main.go`/`startup.go`/`routes.go` at generation time, so selecting a feature actually changes runtime behavior. This is template work (`internal/generator` + the `.tmpl` files), not a redesign — the packages themselves are already correctly written, they're just orphaned.
2. **If full wiring is too large for one pass:** cut scope. Mark unwired features as experimental/manual in `--help` output and the README ("writes the files, you wire them in — full auto-wiring coming in vX") rather than presenting them as functional. Do not ship a fix that silently continues to claim "production feature system" for features that don't do anything.
3. Whichever path: add a codegen-level test that asserts, for at least one feature (start with JWT or PostgreSQL), that the generated project's `main.go` actually references the feature's package — not just that the file exists on disk.

---

## Phase 1 — Test the Dangerous Paths

Coverage today:
```
cmd/lalibela            0.0%   ← entire CLI entrypoint
internal/cli            83.7%
internal/features       51.2%  (orchestration only)
internal/features/*     0.0%   ← all 12 individual feature installers
internal/generator       61.8%
internal/modules         0.0%
internal/ui              0.0%
internal/updater         0.0%
internal/utils           0.0%
```

Priority order (highest risk first):

1. **`internal/updater`** — downloads and replaces the running binary, verifies GPG signatures, does OS-specific atomic replace (delayed-rename trick on Windows since a process can't overwrite itself there). Zero tests on supply-chain-security-relevant code. Write tests for: signature verification rejection on tampered payload, checksum mismatch handling, atomic replace on each OS path, and failure/rollback behavior if the download is interrupted.
2. **`cmd/lalibela` destructive handlers** — `uninstall` (deletes the running binary, optionally `rm -rf ~/.lalibela`), `run` (shells out to `go run .`), `update`. Zero tests on code that deletes things. Write tests for: `uninstall` without `--force` prompts/aborts correctly, `uninstall --force` only touches paths it should, no accidental deletion outside the intended directory.
3. **The 12 feature installers** (`internal/features/*`) — only `engine.go` orchestration is tested; actual file-writing logic per feature (auth, config, cors, docker, errorhandler, gracefulshutdown, health, logger, postgres, ratelimit, redis, swagger) has zero coverage. Write generation-output tests per feature: given a feature selection, assert the expected files exist with expected content/imports.
4. **`internal/modules`** — writes marker files to `~/.lalibela/modules/*/.installed` that nothing reads back except itself. Either add tests for what exists, or fold this into the wiring fix in 0.5 if it turns out to be more dead weight.

---

## Phase 2 — Consistency & Documentation Fixes (smaller, do alongside Phase 1)

- **Two different state-file write strategies**, inconsistently: per-project `.lalibela/features.json` uses plain `os.WriteFile` (not atomic — a crash mid-write can corrupt it), while `~/.lalibela/modules.json` correctly does atomic temp-file-then-rename. Standardize on the atomic pattern everywhere state is written.
- **README overstates/misdocuments `~/.lalibela/`**: the README says local state (e.g. installed feature metadata) lives under `~/.lalibela/`, but per-project feature state actually lives at `<project>/.lalibela/features.json` (relative to the generated project, not home). `~/.lalibela/` is only used for inert module markers and the update GPG key. Fix the README to describe these as the two separate things they are.
- **`internal/modules`** is entirely undocumented in the README — either document it or fold it away if 0.5's fix makes it redundant.
- Confirm fiber/nethttp scaffolds actually `go build ./...` cleanly end-to-end (gin/echo were confirmed in the audit; fiber/nethttp were reviewed statically as correct but not build-confirmed due to a sandboxed `GOPROXY` issue during the audit — this needs a real run in an unrestricted environment).

---

## Phase 3 — Extensibility Engine ("Vite's plugin API" equivalent)

Today `internal/features/registry.go` is a **compiled Go map** (`Registry = map[string]Feature{...}`), wired at build time, with each feature implementing a 3-method interface (`Name`, `Compatible`, `Install`). Adding a feature means writing a Go package and rebuilding the whole binary. There is no dynamic loading, no external manifest, no discovery — **0% of the "plugin ecosystem" roadmap item exists today**, contrary to what the roadmap section implies.

To build a real plugin system:
- Define an external template/feature manifest format (e.g. a directory of `.tmpl` files + a metadata YAML describing name, compatibility, injection points).
- Build a loader that reads user- or community-supplied feature bundles from disk or a registry URL (a Git repo, similar to how Homebrew taps or Vite template repos work).
- Drop the requirement that `Install()` be compiled Go code — templates + declarative injection points should cover the majority of real features (env vars, middleware registration, route additions, config keys).
- `lalibela.yaml` project/org config: pins template/feature-bundle versions so a whole team gets identical `lalibela add x` behavior.
- Versioned templates with semver so `lalibela upgrade`/diff tooling (also currently 0% implemented per the roadmap audit) has something real to compare against.

**Do not start this phase until Phase 0 is done** — a plugin system for a feature model that doesn't wire into the generated app just multiplies the same bug across more features.

---

## Phase 4 — Depth Features (the moat vs. "AI can just generate this")

These are multi-file, stateful generations where LLM one-shot output is still unreliable — this is where a maintained, tested template genuinely beats prompting an agent:

- Database migrations: `lalibela migrate up/down/create` (wrap `golang-migrate` or similar) — today the postgres feature only drops one static empty `db/migrations/0001_init.sql`; there is no migration runner or CLI command at all.
- DB layer choice at scaffold time: raw SQL / sqlc / GORM / ent.
- CI/CD starter profiles: `--ci github|gitlab` with real, working pipelines (currently 0% implemented).
- Shell completion (bash/zsh/fish/pwsh) — currently 0% implemented.
- Observability feature flags: Prometheus metrics endpoint, OpenTelemetry tracing scaffold.
- Auth beyond JWT: OAuth2/OIDC, API keys, session auth.
- "Framework-aware feature patching" — today this is only a basic per-framework allow/deny gate (`shared.IsFeatureCompatible`), not real patching. Decide what "patching" should mean concretely (e.g. actually modifying an existing route file to add middleware) before building it.

---

## Phase 5 — AI-Agent Compatibility Layer (the actual differentiator)

The strongest long-term answer to "AI can just do this now" is not to out-generate AI, but to become the thing an AI coding agent calls so it doesn't reinvent your project's conventions every time:

- A machine-readable manifest (`lalibela.schema.json`) describing project structure/conventions, so agents (Claude Code, Copilot, Cursor) can read it and extend the project correctly instead of guessing.
- `lalibela explain` — dumps project conventions as text an agent or new hire can ingest.
- Optional MCP server exposing `add feature`, `run`, `migrate` as callable tools for an agent.

---

## Sequencing Summary

| Phase | Focus | Blocking? |
|---|---|---|
| 0 | Fix install.sh (#2), fix `update`, remove `.gocache/`, delete dead files, wire-or-disclose features | **Yes — do first** |
| 1 | Tests on updater, destructive commands, feature installers | Do alongside/right after 0 |
| 2 | State-write consistency, README accuracy, confirm fiber/nethttp builds | Do alongside 1 |
| 3 | Plugin/template ecosystem | After 0 is fully done |
| 4 | Migrations, CI profiles, shell completion, deeper features | After 3 |
| 5 | AI-agent manifest/MCP layer | After 3, can parallel 4 |

**Definition of done for Phase 0** (ship this as v0.2.0 before touching anything else):
- [ ] `curl ... install.sh | sh` works on a clean macOS machine (verified, not assumed)
- [ ] `go install .../cmd/lalibela@latest` works on a clean machine
- [ ] `lalibela update` succeeds on a fresh install with no manual key setup
- [ ] `.gocache/` no longer tracked
- [ ] `.goreleaser.yml` and `test.txt` removed
- [ ] At least one feature (JWT or PostgreSQL) demonstrably wired into generated `main.go`, with a test proving it
- [ ] README no longer claims things the code doesn't do
