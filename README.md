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

- macOS (uses Keychain via `security`) — Linux support is on the roadmap
- `curl`, `jq` (`brew install jq`)
- An Anthropic API key — get one at
  [console.anthropic.com/settings/keys](https://console.anthropic.com/settings/keys)

## Install

```sh
git clone https://github.com/alegerber/conjure ~/github/conjure
cd ~/github/conjure
./install.sh
```

The installer creates a symlink at `~/.local/bin/conjure` (override with
`BIN_DIR=`) and, if oh-my-zsh is detected, links the optional zsh plugin
(adds the `cj` short alias and tab completion).

Make sure `~/.local/bin` is on your `$PATH`. If you use the zsh plugin, add
`conjure` to your `plugins=(...)` array in `~/.zshrc`.

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
cj "<description>"               # zsh-only: pushes command into the editor buffer
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

This removes the symlinks. Your API key in Keychain is kept by default;
remove it with:

```sh
security delete-generic-password -a "$USER" -s "anthropic-api-key"
```

## Name conflict

The `conjure` binary shadows ImageMagick's `conjure` (an MSL script
interpreter, rarely used). To invoke ImageMagick's version explicitly, use
`\conjure` or the full path (e.g. `/opt/homebrew/bin/conjure`).

## License

MIT
