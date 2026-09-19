You are working on the `lalibela-cli` repo (github.com/naodEthiop/lalibela-cli), a Go CLI project generator. A full codebase audit was just done and a spec was written (see SPEC.md in this same directory — read it first, it has full context on all findings). This prompt covers **Phase 0 only**: the fixes that must ship before any roadmap work. Work through them in order. Do not start Phase 1+ items from the spec unless explicitly asked.

## Ground rules
- Work on a new branch, e.g. `fix/phase-0-stabilization`.
- Commit each fix separately with a clear message (don't bundle unrelated fixes into one commit) — this makes the PR reviewable.
- Every fix below has a "verify" step — actually run it, don't just read the diff and assume it works.
- If a fix requires a decision only the repo owner can make (e.g. whether to purge `.gocache/` from git history, or whether `google790fa80a4bea91a1.html` is still needed), stop and ask instead of guessing.

## Task 1 — Fix the macOS install.sh 404 (GitHub Issue #2)
**Reproduce first.** Run exactly what the user ran:
```
curl -fsSL https://raw.githubusercontent.com/naodEthiop/lalibela-cli/main/scripts/install.sh | sh
```
Confirm whether the 404 is on fetching the script itself, or on a URL the script constructs internally (e.g. a GitHub release asset URL built from a tag pattern that doesn't match real tags, or an OS/arch string that doesn't match actual release filenames).

Check in parallel:
- Does `scripts/install.sh` exist on `main` right now? If the raw URL 404s but the file exists, something about branch/path history is wrong (was it renamed, was `main` force-pushed, is there a `.gitattributes` issue).
- Inside `install.sh`, trace how it resolves "latest release" and constructs the download URL. Compare against what `.goreleaser.yaml` actually names its release assets (check `.github/workflows/release.yml` and the `.goreleaser.yaml` `builds`/`archives` sections) — a naming mismatch between what goreleaser produces and what install.sh expects is a likely root cause.
- Separately, figure out why `go install github.com/naodEthiop/lalibela-cli/cmd/lalibela@latest` also failed for this user. Check: is there a resolvable semver-tagged release the Go module proxy can find? Does `go.mod`'s module path match the actual repo path (case sensitivity matters — repo is `lalibela-cli`, but was at one point referenced as `Lalibela-cli` with a capital L; confirm the canonical casing and that go.mod matches it exactly)?

**Fix, then verify end-to-end**: run the install command again (or the closest equivalent in this environment — Linux is fine if macOS isn't available, the bug is almost certainly platform-agnostic since it's a URL/tag issue) and confirm it actually completes and produces a working `lalibela` binary.

**Reply on the issue** (or leave a comment ready for the repo owner to post) explaining the root cause and linking the fix commit.

## Task 2 — Fix `lalibela update`
`internal/updater.resolvePublicKeyPath()` requires `~/.lalibela/public.gpg` or `LALIBELA_UPDATE_PUBLIC_KEY`, and no install path provisions either — so `lalibela update` fails with "missing update public key" for every real user today.

Preferred fix: embed the public key in the binary at build time (e.g. via `//go:embed`), so no provisioning step is needed on any platform. Only fall back to "ship the key via every install path" if there's a concrete reason embedding won't work (ask if unsure).

After fixing, write a test (or a scripted check) that simulates a fresh install and confirms `lalibela update` no longer fails on key resolution — don't just confirm the key is embedded, confirm the actual update flow gets past that step.

## Task 3 — Remove `.gocache/` from git tracking
```
git rm -r --cached .gocache/
```
Add `.gocache/` to `.gitignore`. Confirm the working tree still has the directory (untracked) if needed for local builds, or that it's safe to regenerate.

Do NOT rewrite git history (no `filter-repo`/BFG) without explicit approval — just stop it from growing further for now. Ask if history rewrite should be a follow-up.

## Task 4 — Delete confirmed dead files
- Delete `.goreleaser.yml` (confirmed dead — CI uses `--config .goreleaser.yaml` explicitly; the `.yml` file is missing the GPG `signs:` block and has a stray BOM on line 1). Before deleting, grep the whole repo for any reference to `.goreleaser.yml` to make sure nothing else points at it.
- Delete `test.txt` (leftover UTF-16 "signed test" garbage from a commit-signing test).
- Do NOT delete `google790fa80a4bea91a1.html` — flag it in your summary for the repo owner to confirm, but leave it in place.
- Add `.gitignore` entries for local test-scaffold leftovers (this audit saw `b/`, `fgd/`, `gf/`, `el/` appear from local CLI test runs) — either name them explicitly or add a documented convention/wildcard.

## Task 5 — Wire at least one "production feature" end to end
This is the most important and largest task. Right now every feature (6 always-on "default production features" — health, cors, error_handler, graceful_shutdown, config, logger — plus opt-in JWT, PostgreSQL, Clean Architecture) generates files that the scaffolded app never imports or calls. `routes.go.tmpl` even hand-rolls its own duplicate `/health` handler instead of using the generated `internal/server/health.go`.

**Scope for this task: pick ONE feature and wire it completely as a proof of the pattern** — recommend starting with either:
- **health** (simplest — it's one of the always-on features, and routes.go already has a duplicate handler to replace), or
- **JWT** (higher value — demonstrates middleware wiring, which is the harder/more valuable case)

Steps:
1. Read the existing template pipeline (`internal/generator`, the relevant `.tmpl` files, `internal/features/registry.go`, and the specific feature's package) to understand how generation currently assembles `main.go`/`startup.go`/`routes.go`.
2. Modify the template(s) so that when the feature is selected, the generated `main.go`/`routes.go` actually imports and calls into the feature's package (e.g. `routes.go` calls `server.RegisterHealthCheck(router)` from `internal/server/health.go` instead of hand-rolling its own handler; or for JWT, `routes.go`/`startup.go` actually registers the JWT middleware from `internal/middleware/jwt.go` on the intended routes).
3. Remove the now-redundant hand-rolled logic in the template (e.g. the duplicate inline `/health` handler) so there's exactly one implementation, not two.
4. Add a test: generate a project with the feature enabled, and assert (via source inspection or an actual `go build` + a smoke-test HTTP call) that the feature is live in the output — not just that the file exists on disk.
5. Generate a test project with the feature enabled and confirm `go build ./...` still succeeds, and that the feature is observably working (e.g. hitting `/health` returns what the wired handler returns, or a JWT-protected route actually rejects unauthenticated requests).

**Do not attempt to wire all features in this pass.** After the one you pick is fully wired and tested, write up (in your summary, not as a committed file unless asked) which other features remain unwired, so the repo owner can decide whether to continue the pattern feature-by-feature or scope it down in the README/help text for now ("writes the files, wiring for the rest coming in vX").

## Final steps
1. Run the full test suite and `go vet ./...` — confirm nothing broke.
2. Run `go build ./...` for every framework (gin, echo, fiber, nethttp) with every feature combination you can reasonably script, and confirm they all still compile.
3. Write a summary of what was fixed, what's still open from Phase 0's checklist (see SPEC.md's "Definition of done for Phase 0"), and any decisions you had to punt back to the repo owner.
4. Open a PR (don't merge) titled something like "Phase 0 stabilization: fix install.sh 404 (#2), fix update key, remove build cache from git, wire health feature" with the summary as the description, and reference issue #2 so it links/closes on merge.
