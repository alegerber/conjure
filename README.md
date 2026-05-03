# conjure

> Natural-language → Unix command, in under a second.

```sh
$ conjure "list all .md files recursively"
find . -type f -name "*.md"

$ conjure -e "find files modified in the last 7 days"
find . -type f -mtime -7
# Finds all files modified within the last 7 days, recursively from the current directory.

$ cj "show top 3 largest directories"
# command lands editable in the next prompt, ready to edit or run
```

A small CLI that turns natural-language descriptions into single-line Unix
commands. Five backends are supported — see [docs/PROVIDERS.md](docs/PROVIDERS.md).

## Why?

| Approach | Latency |
|---|---|
| `claude -p` with `--json-schema` and full plugin/MCP boot | ~19s |
| `claude -p` with minimal flags + plain text | ~3.7s |
| **conjure** (direct API + plain text) | **~1.0s** |
| conjure with `--explain` (tool use) | ~1.3s |

That's a **~19× speedup** for the same kind of result, with materially better
output quality on macOS (BSD coreutils).

## Requirements

- Go 1.25+ (`brew install go`) to build from source
- macOS, Linux, or Windows — credentials use the OS keyring
  (Keychain / Secret Service / Credential Manager)
- An API key for one of the supported providers (or a local Ollama / a
  signed-in `codex` / `claude` CLI)

## Install

```sh
git clone https://github.com/alegerber/conjure ~/github/conjure
cd ~/github/conjure
./install.sh
```

The installer runs `go install ./cmd/conjure` (binary lands in
`${GOBIN:-$GOPATH/bin}`, typically `~/go/bin/conjure`) and symlinks it into
`~/.local/bin/conjure` so the binary stays on `$PATH` at a stable location.
Override the symlink dir with `BIN_DIR=`. Future `go install ./cmd/conjure`
runs keep both paths in sync — no need to re-run the installer for upgrades.

Make sure `~/.local/bin` is on your `$PATH`. For the `cj()` shell helper
(drops the generated command into your input buffer instead of printing it),
append this to your shell rc:

```sh
eval "$(conjure shell-init zsh)"     # or: bash | fish | powershell
```

## First-time setup

```sh
conjure setup
```

This walks you through picking a provider and stashing credentials. API keys
go into the OS keyring (Keychain on macOS, Secret Service on Linux,
Credential Manager on Windows); preferences are written to
`~/.config/conjure/config.json` with mode 0600.

See [docs/CONFIGURATION.md](docs/CONFIGURATION.md) for all config fields and
the OS-aware system prompt details.

## Usage

```sh
conjure "<description>"          # plain command, ~0.85s
conjure -e "<description>"       # command + brief explanation, ~1.3s
conjure --no-explain "..."       # force plain output (overrides config explain=true)
cj "<description>"               # pushes command into the shell input buffer (see Install)
```

### Piping the result

```sh
conjure "list .md files" | pbcopy             # copy to clipboard (macOS)
eval "$(conjure 'list .md files')"            # run directly — review first!
```

> **Safety note:** `eval` on AI-generated commands is dangerous. Always read
> the command before running.

## Cost

Each call uses ~200 input + ~80 output tokens with Haiku 4.5 → about
**$0.0003 per call**. 100 calls/day ≈ $1/month.

## Uninstall

```sh
cd ~/github/conjure && ./install.sh --uninstall
```

This removes the `~/.local/bin/conjure` symlink. The binary in
`${GOBIN:-$GOPATH/bin}` and your API key in the keyring are kept by default;
remove them with:

```sh
rm "$(go env GOBIN || go env GOPATH)/bin/conjure"        # or: ~/go/bin/conjure
security delete-generic-password -a "$USER" -s "anthropic-api-key"
```

## Name conflict

The `conjure` binary shadows ImageMagick's `conjure` (an MSL script
interpreter, rarely used). To invoke ImageMagick's version explicitly, use
`\conjure` or the full path (e.g. `/opt/homebrew/bin/conjure`).

## License

MIT
