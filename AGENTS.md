# AGENTS.md

> Guidance for AI coding agents (Claude Code, Codex, Cursor, Aider, etc.)
> working on this repo. Humans should read [README.md](README.md) first.

## Project

**conjure** is a small, fast CLI that turns natural-language descriptions
into single-line Unix commands. As of 1.0.0 it speaks to five backends
(Anthropic / OpenAI / Ollama / Codex CLI credentials / Claude CLI) behind a
common `provider.Provider` interface, and ships as a single static Go binary
— no `curl`, `jq`, or shell-runtime dependencies on the user side.

Why it exists: `claude -p` with full plugin/MCP boot took ~19s per call;
direct API + plain text is ~1.0s. The 19× speedup is the entire point on
the pay-per-token providers (anthropic / openai). The subscription modes
(`codex`, `claude-cli`) trade speed for "use the plan I'm already paying
for" — they're explicitly out-of-scope for the speed budget.

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
│   ├── provider/               # Provider interface + Spec + Kind enum
│   │   ├── anthropic/          # Anthropic Messages client (mockable transport)
│   │   ├── openai/             # OpenAI Chat Completions client
│   │   ├── ollama/             # local Ollama (/api/chat) client
│   │   ├── codex/              # reads ~/.codex/auth.json, wraps openai client
│   │   ├── claudecli/          # exec.Command wrapper around `claude -p`
│   │   └── factory/            # New(spec, key) → concrete Provider
│   ├── config/                 # ~/.config/conjure/config.json (Load/Save)
│   ├── prompt/                 # OS-aware system-prompt builder
│   ├── heuristic/              # GNU-vs-BSD warning heuristic
│   ├── keyring/                # system keyring + env-var fallback (ResolveFor())
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
- Resolve provider via `cmd/conjure/generate.go:resolveProvider` —
  `--provider` flag > `config.Provider` > legacy default (anthropic, if a
  keyring entry exists).
- For API-key providers, resolve credentials via
  `keyring.ResolveFor(service, envVar)`.
- Build OS-aware system prompt (`prompt.BuildSystem(runtime.GOOS)`).
- Hand off to the chosen `provider.Provider`; strip markdown fences (the
  HTTP backends do this internally), run `heuristic.CheckGNU(cmd, os)`.
- Print to stdout; optionally pipe to clipboard (`--copy`) or runner (`--run`).

**2. Explain mode (`conjure -e "..."`)**
- Same resolution path. Each provider implements `GenerateExplain`:
  - `anthropic`: forced Tool Use (`tools` + `tool_choice: {type:"tool"}`).
  - `openai` / `ollama`: forced function call
    (`tool_choice: {type:"function"}`).
  - `codex`: delegates to the `openai` client.
  - `claude-cli`: instructs `claude -p` to emit `<command>\n# <expl>` and
    parses on `\n#` (no native tool-use mode through the CLI).
- Output: `<command>\n# <explanation>\n`.

**3. Shell integration (`eval "$(conjure shell-init zsh)"`)**
- `internal/shell` embeds four templates (zsh/bash/fish/powershell) via `//go:embed`
- The emitted script defines `cj()`, which splits stdout: first line into the
  shell's input buffer (zsh `print -z`, bash `READLINE_LINE`, fish
  `commandline -r`, PowerShell `PSConsoleReadLine.Insert`), rest to stdout.
- Subcommand pass-through (`setup`, `--version`, `--help`) skips the split.

**4. Subcommands**
- `conjure setup` — interactive picker: choose provider, then either
  prompt for an API key (no echo, via `golang.org/x/term`) and write to a
  per-provider keyring service, point at an Ollama host, verify the Codex
  auth file, or check `claude` is on `PATH`. Persists the choice to
  `~/.config/conjure/config.json`.
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

Keyring entries are stored per-provider:
`service=anthropic-api-key` for the Anthropic key, `service=openai-api-key`
for the OpenAI key, both under `account=$USER`. Resolution uses
`keyring.ResolveFor(service, envVar)`: system keyring first, then env-var
fallback (`ANTHROPIC_API_KEY` / `OPENAI_API_KEY`). The `codex` provider
ignores the keyring entirely and reads `~/.codex/auth.json`; `ollama` and
`claude-cli` need no credentials at all.

## Testing

Real test suite, no manual-only checks:

```sh
go vet ./...
go test ./... -race -count=1
```

Per-package highlights:
- `internal/provider/anthropic|openai|ollama` — each uses `httptest.Server`
  to assert request shape and decode forced tool-use responses.
- `internal/provider/codex` — fixture-driven (auth file path passed
  directly to `LoadAPIKey`).
- `internal/provider/claudecli` — fakes `exec.Command` via the `os.Args[0]`
  re-exec trick (see `TestHelperProcess`).
- `internal/config` — round-trips Save/Load through a temp `XDG_CONFIG_HOME`.
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
conjure --no-explain "..."   # override config explain=true
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

## Provider conventions

Every backend implements `provider.Provider` (in `internal/provider/provider.go`):

```go
type Provider interface {
    Name() string
    GeneratePlain(ctx, system, task) (string, error)
    GenerateExplain(ctx, system, task) (*EmitCommand, error)
}
```

The factory (`internal/provider/factory`) is the only place that knows about
all backends; everywhere else depends on the interface.

**`anthropic`** — `https://api.anthropic.com/v1/messages`. Headers `x-api-key`,
`anthropic-version: 2023-06-01`. System prompt is sent as
`system: [{type:"text", text:..., cache_control:{type:"ephemeral"}}]`.
Structured output uses Tool Use (`tools` + forced `tool_choice`), never
`response_format` — Tool Use is ~5× faster on Haiku. Default model
`claude-haiku-4-5`. Keyring service `anthropic-api-key`, env fallback
`ANTHROPIC_API_KEY`.

**`openai`** — `https://api.openai.com/v1/chat/completions`. Header
`Authorization: Bearer <key>`. Structured output uses
`tools: [{type:"function",...}]` + `tool_choice: {type:"function",...}`.
Default model `gpt-4o-mini`. Keyring service `openai-api-key`, env
fallback `OPENAI_API_KEY`. Override base URL via `config.openai_base_url`.

**`ollama`** — `<ollama_host>/api/chat` (default `http://localhost:11434`,
configured via `ollama_host` in the config file).
Body uses OpenAI-style `tools`. Note that tool-call `arguments` arrive as a
JSON object (not a stringified one, unlike OpenAI). No default model — must
be set in config or via `--model`.

**`codex`** — wraps the `openai` client but reads the API key from
`~/.codex/auth.json` (override via `codex_auth_file` in the config file).
Reports `Name() == "codex"` so logs disambiguate. Out-of-scope for the
speed budget — same latency as `openai` though.

**`claude-cli`** — shells out to `claude -p --output-format text
--allowedTools "" --append-system-prompt <sys>`. Mockable via the
`execCommand` package var. Explain mode uses a "<command>\n# <expl>"
two-line convention parsed back. ~3.7s per call from plugin/MCP boot;
explicitly out-of-scope for the speed budget.

**Transport in tests**: HTTP-based clients (`anthropic`, `openai`, `ollama`)
expose an `Endpoint` field tests point at `httptest.Server`. The
`claudecli` package exposes `var execCommand = exec.CommandContext` so
tests can inject a fake binary.

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
- New config field → README's "Configuration file" table.
- Behavior change → mention in README and bump `Version` in `cmd/conjure/main.go`.

### Telling humans things

- User-facing diagnostics go to **stderr**. Stdout is reserved for the
  generated command(s) so piping (`cj`, `--copy`, `--run`) works.
- Successful generations: just print the command. No banner, no "Generated:".

## Backlog & roadmap

The backlog lives in [GitHub Issues](https://github.com/alegerber/conjure/issues).

Status as of 1.0.0:

- ✅ **#1** — bash `bind -x` widget for `cj`-equivalent (bash template ships
  the widget bound to `Ctrl-X Ctrl-J`).
- ✅ **#2** — `--copy` flag for clipboard.
- ✅ **#3** — `--run` flag with confirm prompt.
- ✅ **#4** — Brew distribution via goreleaser (released in 1.0.0:
  `.goreleaser.yml`, `.github/workflows/release.yml`, tap repo
  `alegerber/homebrew-conjure`). Scoop bucket still pending.

When you finish an item, close the issue with `Closes #N` in the commit body.

## When in doubt

- Re-read [README.md](README.md) for user-facing semantics.
- Check `git log --oneline` for recent decisions and their rationale.
- The maintainer prefers small, focused commits over big ones; ask before
  large refactors.
