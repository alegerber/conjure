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
#
# Behavior:
#   - First line of `conjure` output is treated as the command and pushed
#     into the ZLE editor buffer.
#   - Any remaining lines (e.g. the `# explanation` from --explain) are
#     printed to stdout, so you see them above the next prompt.
#   - Subcommands (--setup, --version, --help) pass through and print
#     normally without buffer manipulation.
cj() {
  case "$1" in
    --setup|--version|-v|--help|-h)
      conjure "$@"
      return $?
      ;;
  esac
  local output
  output=$(conjure "$@") || return $?
  [[ -z "$output" ]] && return 0

  local cmd=${output%%$'\n'*}
  local rest=""
  if [[ "$output" == *$'\n'* ]]; then
    rest=${output#*$'\n'}
    rest=${rest%$'\n'}
  fi

  [[ -n "$rest" ]] && print -r -- "$rest"
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
