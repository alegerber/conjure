# Providers

As of 0.3.0 conjure supports five backends — three pay-per-token APIs, one
local model, and two subscription-mode adapters that reuse credentials from
the official Codex and Claude CLIs.

| Key          | Backend                                            | Auth                                                       |
|--------------|----------------------------------------------------|------------------------------------------------------------|
| `anthropic`  | `https://api.anthropic.com/v1/messages`            | API key in keyring (`anthropic-api-key` / `ANTHROPIC_API_KEY`) |
| `openai`     | `https://api.openai.com/v1/chat/completions`       | API key in keyring (`openai-api-key` / `OPENAI_API_KEY`)   |
| `ollama`     | `<ollama_host>/api/chat` (default `localhost:11434`)| None — runs against your local Ollama                      |
| `codex`      | OpenAI API, key from `~/.codex/auth.json`          | Reuses Codex CLI credentials (`codex login` first)         |
| `claude-cli` | Shells out to `claude -p`                          | Reuses Claude Code session (install + log into `claude`)   |

## Per-call override

```sh
conjure --provider openai "list md files"
conjure --provider ollama --model llama3.1 "show top 3 dirs"
```

## Subscription-mode latency

Subscription modes are slower than the pay-per-token APIs:

- `claude-cli` adds ~3.7s per call for plugin/MCP boot.
- `codex` is just OpenAI under the hood, so latency matches `openai`.
