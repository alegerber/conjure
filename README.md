# conjure

> Natural-language → Unix command, in under a second.

```sh
$ conjure "list all .md files recursively"
find . -type f -name "*.md"

$ conjure -e "find files modified in the last 7 days"
find . -type f -mtime -7
# Finds all files modified within the last 7 days, recursively from the current directory.

$ cj "show top 3 largest directories"
# (zsh-only) command lands editable in the next prompt, ready to edit or run
```

A small CLI that turns natural-language descriptions into single-line Unix
commands. As of 0.3.0 it supports five backends — three pay-per-token APIs,
one local model, and two subscription-mode adapters that reuse credentials
from the official Codex and Claude CLIs.

## Why?

| Approach | Latency |
|---|---|
| `claude -p` with `--json-schema` and full plugin/MCP boot | ~19s |
| `claude -p` with minimal flags + plain text | ~3.7s |
| **conjure** (direct API + plain text) | **~1.0s** |
| conjure with `--explain` (tool use) | ~1.3s |

That's a **~19× speedup** for the same kind of result, with materially better
output quality on macOS (BSD coreutils) — see *OS-aware system prompt* below.

## Requirements

- macOS (uses Keychain via `security`) — Linux support is on the roadmap
- `bash` 3.2+ (preinstalled on macOS)
- `curl` (preinstalled)
- `jq` — `brew install jq`
- An Anthropic API key — get one at
  [console.anthropic.com/settings/keys](https://console.anthropic.com/settings/keys)

## Install

```sh
git clone https://github.com/alegerber/conjure ~/github/conjure
cd ~/github/conjure
./install.sh
```

The installer:

- creates a symlink at `~/.local/bin/conjure` (override with `BIN_DIR=`)
- if oh-my-zsh is detected, links the optional zsh plugin (gives you the `cj`
  short alias and tab completion)

Make sure `~/.local/bin` is on your `$PATH`. If you use the zsh plugin, add
`conjure` to your `plugins=(...)` array in `~/.zshrc`.

## First-time setup

```sh
conjure setup
```

This walks you through picking a provider and stashing whatever credentials
that provider needs. API keys go into the OS keyring (Keychain on macOS,
Secret Service on Linux, Credential Manager on Windows); preferences are
written to `~/.config/conjure/config.json` (or the OS equivalent) with
mode 0600.

## Providers

| Key          | Backend                                            | Auth                                                       |
|--------------|----------------------------------------------------|------------------------------------------------------------|
| `anthropic`  | `https://api.anthropic.com/v1/messages`            | API key in keyring (`anthropic-api-key` / `ANTHROPIC_API_KEY`) |
| `openai`     | `https://api.openai.com/v1/chat/completions`       | API key in keyring (`openai-api-key` / `OPENAI_API_KEY`)   |
| `ollama`     | `<ollama_host>/api/chat` (default `localhost:11434`)| None — runs against your local Ollama                      |
| `codex`      | OpenAI API, key from `~/.codex/auth.json`          | Reuses Codex CLI credentials (`codex login` first)         |
| `claude-cli` | Shells out to `claude -p`                          | Reuses Claude Code session (install + log into `claude`)   |

Override the configured provider per-call:

```sh
conjure --provider openai "list md files"
conjure --provider ollama --model llama3.1 "show top 3 dirs"
```

Subscription modes are slower than the pay-per-token APIs:

- `claude-cli` adds ~3.7s per call for plugin/MCP boot.
- `codex` is just OpenAI under the hood, so latency matches `openai`.

## Usage

```sh
conjure "<description>"          # plain command, ~0.85s
conjure -e "<description>"       # command + brief explanation, ~1.3s
conjure --no-explain "..."       # force plain output (overrides config explain=true)
cj "<description>"               # zsh-only: pushes command into the editor buffer
```

### Explain mode

With `-e` / `--explain`, conjure uses the Anthropic [tool use](https://docs.anthropic.com/en/docs/agents-and-tools/tool-use/overview)
API for structured output, returning two lines:

```
<command>
# <one-sentence explanation>
```

This costs roughly +400ms and ~+50 tokens compared to plain mode. To enable
by default, set `"explain": true` in the config file; override per-call with
`--no-explain`.

In the zsh `cj` function, the command goes into the editor buffer (via
`print -z`) and the explanation is printed above the next prompt — so you
read the explanation, then edit/run the command.

### OS-aware system prompt

conjure ships an OS-specific system prompt (visible in `bin/conjure`) that
explicitly forbids GNU-only flags on macOS and shows GNU↔BSD examples
(`find -printf` → `find -exec stat -f`, `sed -i 'X'` → `sed -i '' 'X'`,
`date -d` → `date -v`, etc.). A small heuristic check warns on stderr if a
known-bad GNU pattern slips through anyway.

The system prompt is tagged `cache_control: ephemeral` for Anthropic prompt
caching, but the prompt is currently ~350 tokens — below Haiku's 2048-token
cache threshold, so caching is effectively a no-op on the default model.
It starts working free-of-charge if you switch to Sonnet (1024-token
threshold) by setting `"model": "claude-sonnet-4-6"` in the config file.

### Configuration file

Persistent settings live in `~/.config/conjure/config.json` (or the OS
equivalent via `os.UserConfigDir`). `conjure setup` writes most fields for
you; the rest are hand-edited.

| Field | Default | Purpose |
|---|---|---|
| `provider` | — | One of `anthropic`, `openai`, `ollama`, `codex`, `claude-cli` |
| `model` | per-provider default | Model id (e.g. `claude-sonnet-4-6`, `gpt-4o`, `llama3.1`) |
| `max_tokens` | `256` | Max output tokens (anthropic/openai/codex) |
| `os` | auto-detected | OS hint shown to the model (`Darwin` / `Linux`) |
| `explain` | `false` | Set to `true` to default to `--explain` |
| `ollama_host` | `http://localhost:11434` | Ollama base URL |
| `openai_base_url` | OpenAI default | Override OpenAI-compatible endpoint |
| `codex_auth_file` | `~/.codex/auth.json` | Path to the Codex CLI auth file |

Example:

```json
{
  "provider": "anthropic",
  "model": "claude-sonnet-4-6",
  "explain": true
}
```

### Environment overrides

API keys are the only values still read from the environment; everything
else lives in the config file above.

| Variable | Purpose |
|---|---|
| `ANTHROPIC_API_KEY` | Fallback Anthropic key when keyring is empty |
| `OPENAI_API_KEY` | Fallback OpenAI key when keyring is empty |

### Piping the result

Output is plain text, so pipe it freely:

```sh
conjure "list .md files" | pbcopy             # copy to clipboard (macOS)
eval "$(conjure 'list .md files')"            # run directly — review first!
```

> **Safety note:** `eval` on AI-generated commands is dangerous. Always read
> the command before running. Consider a confirmation wrapper if you do this
> often.

## Cost

Each call uses ~200 input + ~80 output tokens with Haiku 4.5 → about
**$0.0003 per call**. 100 calls/day ≈ $1/month.

## Name conflict

The `conjure` binary shadows ImageMagick's `conjure` (an MSL script
interpreter, rarely used). If you need ImageMagick's version, invoke it
explicitly:

```sh
\conjure                        # bypass shell function/PATH lookup
/opt/homebrew/bin/conjure       # full path
```

## Uninstall

```sh
cd ~/github/conjure && ./install.sh --uninstall
```

This removes the symlinks. Your API key in Keychain is kept by default; remove
it with:

```sh
security delete-generic-password -a "$USER" -s "anthropic-api-key"
```

## Project layout

```
conjure/
├── bin/conjure              # main CLI (portable bash)
├── zsh/conjure.plugin.zsh   # optional oh-my-zsh wrapper (alias + completion)
├── install.sh               # symlink installer
├── README.md
└── LICENSE
```

## License

MIT
