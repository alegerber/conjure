# Configuration

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

## Environment overrides

API keys are the only values still read from the environment; everything
else lives in the config file above.

| Variable | Purpose |
|---|---|
| `ANTHROPIC_API_KEY` | Fallback Anthropic key when keyring is empty |
| `OPENAI_API_KEY` | Fallback OpenAI key when keyring is empty |

## Explain mode

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

## OS-aware system prompt

conjure ships an OS-specific system prompt that explicitly forbids GNU-only
flags on macOS and shows GNU↔BSD examples (`find -printf` →
`find -exec stat -f`, `sed -i 'X'` → `sed -i '' 'X'`, `date -d` → `date -v`,
etc.). A small heuristic check warns on stderr if a known-bad GNU pattern
slips through anyway.

The system prompt is tagged `cache_control: ephemeral` for Anthropic prompt
caching, but the prompt is currently ~350 tokens — below Haiku's 2048-token
cache threshold, so caching is effectively a no-op on the default model.
It starts working free-of-charge if you switch to Sonnet (1024-token
threshold) by setting `"model": "claude-sonnet-4-6"` in the config file.
