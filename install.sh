#!/usr/bin/env bash
#
# conjure installer
#
# Creates a symlink for the CLI in $BIN_DIR (default ~/.local/bin) and,
# if oh-my-zsh is installed, a symlink for the optional zsh plugin.
#
# Usage:
#   ./install.sh                 # default: install
#   ./install.sh --uninstall     # remove symlinks
#   BIN_DIR=~/bin ./install.sh   # custom CLI install dir
#

set -uo pipefail

REPO_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
BIN_DIR="${BIN_DIR:-$HOME/.local/bin}"
CLI_TARGET="$BIN_DIR/conjure"
OMZ_PLUGIN_DIR="${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/plugins/conjure"
OMZ_PLUGIN_TARGET="$OMZ_PLUGIN_DIR/conjure.plugin.zsh"

green() { printf '\033[32m%s\033[0m\n' "$*"; }
yellow() { printf '\033[33m%s\033[0m\n' "$*"; }
err() { printf '\033[31m%s\033[0m\n' "$*" >&2; }

action_install() {
  if [ ! -x "$REPO_DIR/bin/conjure" ]; then
    err "Could not find executable bin/conjure at $REPO_DIR/bin/conjure"
    exit 1
  fi

  mkdir -p "$BIN_DIR"
  ln -sf "$REPO_DIR/bin/conjure" "$CLI_TARGET"
  green "  ✓ Linked $CLI_TARGET → $REPO_DIR/bin/conjure"

  case ":$PATH:" in
    *":$BIN_DIR:"*) ;;
    *) yellow "  ! $BIN_DIR is not on \$PATH. Add it in your shell config:"
       printf '      export PATH="%s:$PATH"\n' "$BIN_DIR" ;;
  esac

  if [ -d "${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}" ]; then
    mkdir -p "$OMZ_PLUGIN_DIR"
    ln -sf "$REPO_DIR/zsh/conjure.plugin.zsh" "$OMZ_PLUGIN_TARGET"
    green "  ✓ Linked $OMZ_PLUGIN_TARGET → $REPO_DIR/zsh/conjure.plugin.zsh"
    yellow "  ! Add 'conjure' to your plugins=(...) array in ~/.zshrc, then 'exec zsh'."
  else
    yellow "  ! oh-my-zsh not detected — skipping zsh plugin. The CLI works without it."
  fi

  printf '\n'
  green "Done. Next steps:"
  printf '  1. Run: conjure --setup       (store your Anthropic API key)\n'
  printf '  2. Try: conjure "list .md files recursively"\n'
}

action_uninstall() {
  [ -L "$CLI_TARGET" ] && rm "$CLI_TARGET" && green "  ✓ Removed $CLI_TARGET"
  [ -L "$OMZ_PLUGIN_TARGET" ] && rm "$OMZ_PLUGIN_TARGET" && green "  ✓ Removed $OMZ_PLUGIN_TARGET"
  yellow "  ! Your API key in Keychain was kept. Remove with:"
  printf '      security delete-generic-password -a "$USER" -s "anthropic-api-key"\n'
}

case "${1:-install}" in
  install)   action_install ;;
  --uninstall|uninstall) action_uninstall ;;
  *) err "Unknown argument: $1"; exit 1 ;;
esac
