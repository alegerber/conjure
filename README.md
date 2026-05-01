# conjure

> Natural-language → Unix command, in under a second.

A zsh plugin that turns sentences like *"find the 5 largest files"* into actual
shell commands by calling the Anthropic Messages API directly. No Claude Code
overhead, no MCP boot, no plugin discovery — just `curl` and `jq`.

```zsh
$ conjure "list all .md files recursively"
find . -name "*.md" -type f

$ cj "5 largest files in this tree"
find . -type f -exec ls -lh {} \; | sort -k5 -hr | head -5
```

## Why direct API instead of `claude -p`?

Benchmarked locally on macOS:

| Approach | Latency |
|---|---|
| `claude -p` with `--json-schema` and full plugin/MCP boot | ~19s |
| `claude -p` with minimal flags + plain text | ~3.7s |
| **Direct Anthropic API + plain text** (this plugin) | **~0.85s** |

That's a **~22× speedup** for the same kind of result.

## Requirements

- macOS (uses Keychain via `security`)
- `zsh` + [oh-my-zsh](https://ohmyz.sh/)
- `curl` (preinstalled)
- `jq` — `brew install jq`
- An Anthropic API key — get one at
  [console.anthropic.com/settings/keys](https://console.anthropic.com/settings/keys)

## Install

```zsh
git clone https://github.com/alegerber/zsh-conjure \
  ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/conjure
```

Then add `conjure` to your `plugins=(...)` array in `~/.zshrc`:

```zsh
plugins=(
  git
  # ...
  conjure
)
```

Reload your shell: `source ~/.zshrc` (or open a new terminal).

## First-time setup

Run once to store your API key in the macOS Keychain:

```zsh
conjure-setup
```

The key is stored under the service name `anthropic-api-key` for your user. It
is read on demand and never written to disk in plain text.

## Usage

```zsh
conjure "<description>"
cj      "<description>"     # short alias
```

### Environment overrides

| Variable | Default | Purpose |
|---|---|---|
| `CONJURE_MODEL` | `claude-haiku-4-5` | Anthropic model id (e.g. `claude-sonnet-4-6`) |
| `CONJURE_MAX_TOKENS` | `256` | Max output tokens |

```zsh
CONJURE_MODEL=claude-sonnet-4-6 conjure "complicated multi-step pipeline"
```

### Piping the result

Since output is plain text, you can pipe it:

```zsh
conjure "list .md files" | pbcopy             # copy to clipboard
eval "$(conjure 'list .md files')"            # ⚠️ run directly — review first!
```

> **Safety note:** `eval` on AI-generated commands is dangerous. Always read the
> command before running. Consider a confirmation wrapper if you do this often.

## Cost

Each call uses ~200 input + ~80 output tokens with Haiku 4.5 → about
**$0.0003 per call**. 100 calls/day ≈ $1/month.

## Name conflict

This plugin shadows ImageMagick's `conjure` (MSL script interpreter). If you
need ImageMagick's `conjure`, invoke it explicitly:

```zsh
\conjure              # bypass shell function lookup
/opt/homebrew/bin/conjure
```

## Uninstall

```zsh
# remove plugin from plugins=(...) in ~/.zshrc, then:
rm -rf ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/conjure
security delete-generic-password -a "$USER" -s "anthropic-api-key"
```

## License

MIT
