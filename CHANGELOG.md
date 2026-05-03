# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2026-05-03

First stable release. The Go rewrite has been in production use since 0.2.0
and the public CLI surface (subcommands, flags, config schema) is now
considered stable; future breaking changes will bump the major version.

### Added
- Cross-platform release pipeline via [GoReleaser](https://goreleaser.com/):
  signed checksums plus `tar.gz` (darwin/linux) and `zip` (windows) archives
  for `amd64` and `arm64`.
- GitHub Actions `release.yml` workflow triggered by `v*` tags.
- Homebrew tap distribution: `brew tap alegerber/conjure && brew install conjure`.
- `install.sh` now downloads the matching release asset from GitHub instead
  of requiring a Go toolchain. `VERSION=v1.0.0` and `BIN_DIR=` overrides are
  supported.
- Tests for `internal/provider/factory.New` covering all five backends and
  their error paths.
- Cross-platform CI matrix: `lint` and `test` now run on
  `ubuntu-latest`, `macos-latest`, and `windows-latest`. A separate
  cross-compile smoke job exercises all six release `GOOS`/`GOARCH`
  combinations on every PR.
- `CHANGELOG.md` (this file), following Keep a Changelog.

### Changed
- README documents Linux Secret Service and Windows Credential Manager as
  first-class keyring backends; the previous "Linux on the roadmap" line is
  gone. The flags table now lists every flag wired in
  `attachGenerateFlags` (`-e/--explain`, `--no-explain`, `--copy`, `--run`,
  `--provider`, `--model`, `--ollama-host`).
- `SECURITY.md` describes credential storage on macOS, Linux, and Windows
  and tracks the supported-versions table for 1.x.
- Version bumped to `1.0.0` in `cmd/conjure/main.go`.

### Notes
- No code or behavior changes to the generation pipeline. Existing
  `~/.config/conjure/config.json` files and keyring entries are reused as-is.

## [0.3.0] - 2026-04

### Added
- Multi-provider support behind a `provider.Provider` interface: Anthropic,
  OpenAI, Ollama, Codex CLI credentials, and the `claude` CLI.
- `--copy` flag (clipboard) and `--run` flag (confirm + exec).
- `--ollama-host` flag and `OLLAMA_HOST` environment variable.
- LLM-output sanitization before display and execution.
- bash `bind -x` widget (`Ctrl-X Ctrl-J`) shipped with the bash shell-init
  template.

### Changed
- README slimmed down; provider- and configuration-specific docs moved to
  `docs/PROVIDERS.md` and `docs/CONFIGURATION.md`.
- Setup menu driven from `provider.Kinds()` for consistent ordering.
- BSD/GNU heuristic tightened; runner I/O documented.

## [0.2.0] - 2026-02

### Changed
- Complete rewrite from bash to Go. Single static binary; no `curl`/`jq`
  runtime dependency. See [docs/MIGRATION.md](docs/MIGRATION.md).
- `conjure --setup` becomes `conjure setup`; the oh-my-zsh plugin is
  replaced by `eval "$(conjure shell-init <shell>)"`.

## [0.1.x]

Initial bash implementation. Superseded by 0.2.0; see
[docs/MIGRATION.md](docs/MIGRATION.md) for upgrade steps.

[Unreleased]: https://github.com/alegerber/conjure/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/alegerber/conjure/releases/tag/v1.0.0
[0.3.0]: https://github.com/alegerber/conjure/releases/tag/v0.3.0
[0.2.0]: https://github.com/alegerber/conjure/releases/tag/v0.2.0
