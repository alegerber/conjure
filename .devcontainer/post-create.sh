#!/usr/bin/env bash
#
# Devcontainer post-create hook for conjure.
# Runs once after the container is built. Verifies the Go toolchain works,
# downloads modules, executes the test suite, and warns about missing API key.
#
set -euo pipefail

cd /workspaces/conjure

green() { printf '\033[32m%s\033[0m\n' "$*"; }
yellow() { printf '\033[33m%s\033[0m\n' "$*"; }
red() { printf '\033[31m%s\033[0m\n' "$*" >&2; }

green "→ Tool versions"
go version
goreleaser --version | head -1
golangci-lint --version | head -1

if [ ! -f go.mod ]; then
    yellow "ℹ  No go.mod in this checkout — likely on the bash-era 'main' branch."
    yellow "   Switch with:  git checkout rewrite/go"
    exit 0
fi

green "→ go mod download"
go mod download

green "→ go build ./..."
go build ./...

green "→ go test ./..."
go test ./...

if [ -z "${ANTHROPIC_API_KEY:-}" ]; then
    yellow ""
    yellow "⚠  ANTHROPIC_API_KEY is not set inside the container."
    yellow "   Live API smoke tests will fail with: API key not found."
    yellow ""
    yellow "   To fix: export the key on your HOST shell BEFORE rebuilding the container."
    yellow "   On macOS, pull from Keychain:"
    yellow "     export ANTHROPIC_API_KEY=\$(security find-generic-password -a \"\$USER\" -s anthropic-api-key -w)"
    yellow "   Then VS Code: Command Palette → 'Dev Containers: Rebuild Container'."
fi

green ""
green "✓ Devcontainer ready"
