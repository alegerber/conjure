# Migrating from conjure 0.1.x (bash) to 0.2.0 (Go)

## What changed
- conjure is now a single Go binary; no more `jq` or `curl` dependencies on your side.
- `conjure --setup` → `conjure setup`
- The zsh plugin file is gone. Generate the integration on demand:
  `eval "$(conjure shell-init zsh)"`  (or `bash`, `fish`, `powershell`).
- New flags: `--copy`, `--run`.

## Migration steps
1. Uninstall the old bash version:
   - Remove `~/.local/bin/conjure` if it was a symlink to `bin/conjure`.
   - Remove `~/.oh-my-zsh/custom/plugins/conjure/` if it exists.
   - Remove `conjure` from your `plugins=(...)` array.
2. Install the new binary (see README → Install).
3. Replace the plugin source line in your shell config with the `eval` form
   shown above.
4. Your API key in macOS Keychain (service `anthropic-api-key`) is reused
   automatically; no need to run `conjure setup` again.
   On Linux/Windows, the binary uses Secret Service / Credential Manager.
   If your environment lacks one, set the `ANTHROPIC_API_KEY` env var.
