# conjure

> Natural-language → Unix command, in under a second.

```sh
$ conjure "list all .md files recursively"
find . -type f -name "*.md"

$ conjure "find files modified in the last 7 days"
find . -type f -mtime -7

$ cj "show top 3 largest directories"
du -sh */ | sort -rh | head -3
```

A small bash CLI that calls the Anthropic Messages API directly — no Claude
Code, no MCP, no plugin discovery. Just `curl` and `jq`.

## Why?

| Approach | Latency |
|---|---|
| `claude -p` with `--json-schema` and full plugin/MCP boot | ~19s |
| `claude -p` with minimal flags + plain text | ~3.7s |
| **conjure** (direct API + plain text) | **~0.85s** |

That's a **~22× speedup** for the same kind of result.

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
conjure --setup
```

This prompts for your API key (input hidden) and stores it in the macOS
Keychain under service `anthropic-api-key`. The key is read on demand and
never written to disk in plain text.

## Usage

```sh
conjure "<description>"
cj      "<description>"     # zsh-only short alias
```

### Environment overrides

| Variable | Default | Purpose |
|---|---|---|
| `CONJURE_MODEL` | `claude-haiku-4-5` | Anthropic model id (e.g. `claude-sonnet-4-6`) |
| `CONJURE_MAX_TOKENS` | `256` | Max output tokens |
| `CONJURE_OS` | auto-detected | OS hint shown to the model (`Darwin` / `Linux`) |

```sh
CONJURE_MODEL=claude-sonnet-4-6 conjure "complicated multi-step pipeline"
```

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
