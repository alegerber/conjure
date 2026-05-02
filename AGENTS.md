# AGENTS.md

> Guidance for AI coding agents (Claude Code, Codex, Cursor, Aider, etc.)
> working on this repo. Humans should read [README.md](README.md) first.

## Project

**conjure** is a small, fast CLI that turns natural-language descriptions
into single-line Unix commands by calling the Anthropic Messages API
directly. As of 0.2.0 it ships as a single static Go binary; no `curl`,
`jq`, or shell-runtime dependencies on the user side.

Why it exists: `claude -p` with full plugin/MCP boot took ~19s per call;
direct API + plain text is ~1.0s. The 19× speedup is the entire point.

## Architecture

```
conjure/
├── cmd/conjure/                # main package: cobra root + subcommands
│   ├── main.go                 # entrypoint, Version constant
│   ├── root.go                 # newRootCmd(), generate flags
│   ├── generate.go             # plain + --explain generate path
│   ├── setup.go                # `conjure setup` (write key to keyring)
│   └── shell_init.go           # `conjure shell-init [zsh|bash|fish|powershell]`
├── internal/
│   ├── api/                    # Anthropic Messages client (mockable transport)
│   ├── prompt/                 # OS-aware system-prompt builder
│   ├── heuristic/              # GNU-vs-BSD warning heuristic
│   ├── keyring/                # system keyring + env-var fallback (Resolve())
│   ├── clipboard/              # cross-platform copy (pbcopy/xclip/wl-copy/clip.exe)
│   ├── runner/                 # `--run` confirm prompt + exec
│   └── shell/                  # embedded shell-init templates + Render()
├── docs/
│   └── MIGRATION.md            # 0.1.x bash → 0.2.0 Go upgrade notes
├── install.sh                  # bootstrap installer (downloads release binary)
├── README.md                   # user-facing docs
├── AGENTS.md                   # this file
└── LICENSE
```

### Key flows

**1. Plain generation (`conjure "..."`)**
- Resolve API key via `keyring.Resolve()` (system keyring → `ANTHROPIC_API_KEY` fallback)
- Build OS-aware system prompt (`prompt.BuildSystem(runtime.GOOS)`)
- POST to `https://api.anthropic.com/v1/messages` via `internal/api`
- Strip markdown fences, run `heuristic.CheckGNU(cmd, os)` warning
- Print to stdout; optionally pipe to clipboard (`--copy`) or runner (`--run`)

**2. Explain mode (`conjure -e "..."`)**
- Same as above, but the request body adds:
  - `tools: [{name: "emit_command", input_schema: {command, explanation}}]`
  - `tool_choice: {type: "tool", name: "emit_command"}` (forced)
- Output: `<command>\n# <explanation>\n`

**3. Shell integration (`eval "$(conjure shell-init zsh)"`)**
- `internal/shell` embeds four templates (zsh/bash/fish/powershell) via `//go:embed`
- The emitted script defines `cj()`, which splits stdout: first line into the
  shell's input buffer (zsh `print -z`, bash `READLINE_LINE`, fish
  `commandline -r`, PowerShell `PSConsoleReadLine.Insert`), rest to stdout.
- Subcommand pass-through (`setup`, `--version`, `--help`) skips the split.

**4. Subcommands**
- `conjure setup` — interactive prompt (no echo, via `golang.org/x/term`),
  writes to keyring under service `anthropic-api-key`, account `$USER`.
- `conjure shell-init <shell>` — emits the embedded integration script.

## Setup for development

```sh
# Build and run locally
go build -o /tmp/conjure ./cmd/conjure
/tmp/conjure --version

# Or install into $GOBIN
go install ./cmd/conjure

# Store API key (or export ANTHROPIC_API_KEY)
conjure setup
```

The keyring entry is `service=anthropic-api-key`, `account=$USER` on every
platform (macOS Keychain, freedesktop Secret Service on Linux, Credential
Manager on Windows). Override at runtime with `ANTHROPIC_API_KEY=...`.

## Testing

Real test suite, no manual-only checks:

```sh
go vet ./...
go test ./... -race -count=1
```

Per-package highlights:
- `internal/api` — mocks `http.RoundTripper`; covers plain + tool-use + error paths.
- `internal/prompt` — string-content assertions per `runtime.GOOS`.
- `internal/heuristic` — known-positive and known-negative GNU-isms cases.
- `internal/keyring` — env fallback path, source-tracking, in-memory `Store` mock.
- `internal/clipboard` — branch coverage for Wayland/X11 selection on Linux.
- `internal/runner` — `tea`-style confirm prompt with mocked stdin.
- `internal/shell` — `Render()` returns expected substrings per shell.

API smoke checks (require live key; not run in CI):

```sh
conjure "list .md files recursively"
conjure -e "find files modified today"
CONJURE_EXPLAIN=1 conjure "..."
conjure --no-explain "..."
conjure --copy "list files"
conjure --run "echo hello"
```

OS-awareness regression checks (still useful manually):

```sh
conjure "find duplicate files by size and name"           # expect: stat -f, not -printf
conjure "replace foo with bar in all .txt files in place" # expect: sed -i ''
conjure "show date 7 days ago in YYYY-MM-DD"              # expect: date -v-7d
```

## Code style

- **Go 1.25+.** Module is pinned to `go 1.25.0` in `go.mod`. Use generics,
  `slices`/`maps` packages, and `errors.Join` freely.
- **Standard library first.** External deps must earn their place. Current set:
  `cobra` (CLI), `go-keyring` (cross-platform keyring), `x/term` (no-echo input).
  Prefer one external dep per package max; if you need more, talk to the
  maintainer first.
- **Errors:** wrap with `fmt.Errorf("...: %w", err)`. No `errors.New` for
  wrapped values. No `panic` outside `main`.
- **No comments that restate the code.** Comments explain *why*, not *what*.
  Public symbols get a doc comment starting with the symbol name.
- **Tests live next to code** (`foo.go` + `foo_test.go`, same package).
  Internal-only helpers use `_internal_test.go` only when they cross packages.
- **Tables over loops of `if`** for option/case dispatch (see
  `internal/shell/shell.go:supported`).
- **`embed` for static assets** (templates, prompts) — never `os.ReadFile`
  at runtime for things shipped with the binary.

## API conventions

When modifying the Anthropic call (in `internal/api`):

- **Endpoint**: `https://api.anthropic.com/v1/messages`
- **Headers**: `x-api-key`, `anthropic-version: 2023-06-01`, `content-type: application/json`
- **Default model**: `claude-haiku-4-5` (fastest current Anthropic model).
- **System prompt**: always sent as `system: [{type: "text", text: ...,
  cache_control: {type: "ephemeral"}}]`. Caching is currently a no-op on
  Haiku (prompt is below the 2048-token threshold), but the structure is in
  place — don't remove it.
- **Structured output**: use Tool Use (`tools` + forced `tool_choice`), never
  `response_format` JSON-Schema. Tool Use is ~5× faster.
- **Transport**: `api.Client` exposes `HTTPClient *http.Client` so tests
  inject a mocked `RoundTripper` via `&http.Client{Transport: mock}`. Do
  not call `http.DefaultClient` directly inside the package.

When adding a new flag:

1. Add it to `generateOpts` (or the relevant subcommand's opts struct).
2. Wire it in `attachGenerateFlags` (`cmd/conjure/root.go`).
3. Document it in `cmd.Long` / `cmd.Short` and README.md.
4. If it influences generation, plumb it through to the right pipeline stage
   (api request, post-processing, output).

## Working agreements

### Git

- **GitHub identity**: `alegerber` (the maintainer has two; this repo belongs
  to the personal one). The remote uses the SSH host alias `github-private`
  which is already set up in `~/.ssh/config`.
- **Commit messages**: imperative subject line ≤72 chars, body wraps at 72,
  explains why. English. Conventional-Commits-ish but not strict.
- **Branches**: feature work on a topic branch (e.g. `rewrite/go`,
  `feat/<thing>`). PRs into `main`. Worktrees under `.worktrees/` are fine.
- Always run `go vet ./...` and `go test ./... -race` before commit.

### Updating docs

- New CLI flag → update README.md (Usage / Flags table) and `cmd.Long`.
- New env var → README's "Environment overrides" table.
- Behavior change → mention in README and bump `Version` in `cmd/conjure/main.go`.

### Telling humans things

- User-facing diagnostics go to **stderr**. Stdout is reserved for the
  generated command(s) so piping (`cj`, `--copy`, `--run`) works.
- Successful generations: just print the command. No banner, no "Generated:".

## Backlog & roadmap

The backlog lives in [GitHub Issues](https://github.com/alegerber/conjure/issues).

Status as of 0.2.0:

- ✅ **#1** — bash `bind -x` widget for `cj`-equivalent (bash template ships
  the widget bound to `Ctrl-X Ctrl-J`).
- ✅ **#2** — `--copy` flag for clipboard.
- ✅ **#3** — `--run` flag with confirm prompt.
- ⏳ **#4** — Brew/Scoop distribution via goreleaser (configuration deferred
  until Brew tap + Scoop bucket repos exist; install today via `go install` or
  release-binary `install.sh`).

When you finish an item, close the issue with `Closes #N` in the commit body.

## When in doubt

- Re-read [README.md](README.md) for user-facing semantics.
- Check `git log --oneline` for recent decisions and their rationale.
- The maintainer prefers small, focused commits over big ones; ask before
  large refactors.
