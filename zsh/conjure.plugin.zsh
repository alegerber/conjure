#
# conjure — zsh integration (oh-my-zsh plugin)
#
# This is a thin wrapper. The actual logic lives in `bin/conjure` (a
# portable bash script). This file adds zsh-specific niceties:
#   - the `cj` function: pushes the generated command into the editor buffer
#     so you can review/edit before pressing Enter
#   - tab completion
#
# The CLI must be on PATH (e.g. via a symlink in ~/.local/bin/conjure
# created by install.sh). If it isn't, the function will fail loudly.
#

if (( ! $+commands[conjure] )); then
  print -u2 "conjure: 'conjure' not found on PATH. Run install.sh from the conjure repo."
  return 1
fi

# `cj <description>` — generate a command and place it in the prompt buffer
# (via `print -z`). The next prompt shows the command, ready to edit or run.
# Setup commands like `--setup` are passed through and printed normally.
cj() {
  if [[ "$1" == --* ]]; then
    conjure "$@"
    return $?
  fi
  local cmd
  cmd=$(conjure "$@") || return $?
  [[ -n "$cmd" ]] && print -z -- "$cmd"
}

# Minimal completion: suggest the documented flags after `conjure` / `cj`.
_conjure() {
  _arguments \
    '--setup[Store API key in macOS Keychain]' \
    '--version[Print version]' \
    '--help[Show help]' \
    '*::description:'
}
compdef _conjure conjure cj 2>/dev/null
