#
# conjure — zsh integration (oh-my-zsh plugin)
#
# This is a thin wrapper. The actual logic lives in `bin/conjure` (a
# portable bash script). This file just adds zsh-specific niceties:
#   - the `cj` short alias
#   - tab completion for env-var hints
#
# The CLI must be on PATH (e.g. via a symlink in ~/.local/bin/conjure
# created by install.sh). If it isn't, the alias will fail loudly.
#

if (( ! $+commands[conjure] )); then
  print -u2 "conjure: 'conjure' not found on PATH. Run install.sh from the conjure repo."
  return 1
fi

alias cj=conjure

# Minimal completion: suggest the documented flags after `conjure`.
_conjure() {
  _arguments \
    '--setup[Store API key in macOS Keychain]' \
    '--version[Print version]' \
    '--help[Show help]' \
    '*::description:'
}
compdef _conjure conjure cj 2>/dev/null
