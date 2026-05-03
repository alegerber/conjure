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

- macOS, Linux, or Windows — credentials live in the OS keyring
  (Keychain on macOS, Secret Service / libsecret on Linux,
  Credential Manager on Windows). Linux desktops without a Secret
  Service provider can fall back to the `ANTHROPIC_API_KEY` /
  `OPENAI_API_KEY` env vars.
- An API key for one of the supported providers (or a local Ollama / a
  signed-in `codex` / `claude` CLI).
- Go 1.25+ (`brew install go`) only required when building from source.

## Install

### Homebrew (macOS / Linux)

```sh
brew tap alegerber/conjure
brew install conjure
```

### Install script (macOS / Linux / WSL)

```sh
curl -sSL https://raw.githubusercontent.com/alegerber/conjure/main/install.sh | bash
```

Downloads the matching release archive from GitHub and drops the binary
into `~/.local/bin/conjure`. Override with `BIN_DIR=~/bin` and pin a
specific version with `VERSION=v1.0.0`.

### Manual download

Grab a prebuilt archive for your OS/arch from the
[releases page](https://github.com/alegerber/conjure/releases) and place
the `conjure` binary somewhere on your `$PATH`.

### From source

```sh
git clone https://github.com/alegerber/conjure
cd conjure
go install ./cmd/conjure
```

Make sure `~/.local/bin` (or `${GOBIN:-$GOPATH/bin}`) is on your `$PATH`.
For the `cj()` shell helper (drops the generated command into your input
buffer instead of printing it), append this to your shell rc:

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

### Flags

| Flag             | Default | Description                                                                          |
|------------------|---------|--------------------------------------------------------------------------------------|
| `-e`, `--explain`| `false` | Generate command + brief explanation (uses tool-use; ~1.3s).                         |
| `--no-explain`   | `false` | Force plain output (overrides `explain=true` from config).                           |
| `--copy`         | `false` | Copy the generated command to the system clipboard.                                  |
| `--run`          | `false` | After printing, prompt to execute the command.                                       |
| `--provider`     | config  | Override configured provider (`anthropic`, `openai`, `ollama`, `codex`, `claude-cli`). |
| `--model`        | config  | Override configured model name.                                                      |
| `--ollama-host`  | config  | Override Ollama host URL (precedence: flag > `$OLLAMA_HOST` > config).               |

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
brew uninstall conjure                         # if installed via Homebrew
./install.sh --uninstall                       # if installed via install.sh
go clean -i github.com/alegerber/conjure/...   # if installed from source
```

API keys stored in the OS keyring are kept by default. Remove them with:

```sh
# macOS
security delete-generic-password -a "$USER" -s "anthropic-api-key"
# Linux (Secret Service)
secret-tool clear service anthropic-api-key
# Windows (PowerShell)
cmdkey /delete:anthropic-api-key
```

## Name conflict

The `conjure` binary shadows ImageMagick's `conjure` (an MSL script
interpreter, rarely used). To invoke ImageMagick's version explicitly, use
`\conjure` or the full path (e.g. `/opt/homebrew/bin/conjure`).

## License

MIT
