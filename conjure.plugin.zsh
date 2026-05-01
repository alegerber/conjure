#
# conjure — natural-language → Unix command
#
# Calls the Anthropic Messages API directly (no Claude Code overhead).
# API key is read on demand from the macOS Keychain.
#
# Usage:
#   conjure "list all .md files recursively"
#   cj      "find the 5 largest files"          # short alias
#
# First-time setup:
#   conjure-setup                                # prompts for API key, stores in Keychain
#
# Environment overrides:
#   CONJURE_MODEL=claude-sonnet-4-6  conjure "..."   # default: claude-haiku-4-5
#   CONJURE_MAX_TOKENS=512           conjure "..."   # default: 256
#
# Note: this function shadows ImageMagick's `conjure` (MSL interpreter).
# If you ever need that, invoke it explicitly: \conjure  or  /opt/homebrew/bin/conjure
#

# Read the API key from Keychain. Echoes the key on stdout, nothing on failure.
_conjure_get_key() {
  security find-generic-password -a "$USER" -s "anthropic-api-key" -w 2>/dev/null
}

# Strip markdown fences and surrounding whitespace from a single-line command.
_conjure_clean() {
  local s=$1
  s=${s#\`\`\`*$'\n'}
  s=${s%$'\n'\`\`\`}
  s=${s#\`}
  s=${s%\`}
  s="${s##[[:space:]]##}"
  s="${s%%[[:space:]]##}"
  print -r -- "$s"
}

conjure() {
  if [[ $# -eq 0 ]]; then
    print -u2 "Usage: conjure <description>"
    print -u2 "Run 'conjure-setup' once to store your Anthropic API key."
    return 1
  fi

  local key
  key=$(_conjure_get_key) || true
  if [[ -z "$key" ]]; then
    print -u2 "conjure: API key not found in Keychain. Run: conjure-setup"
    return 1
  fi

  local model="${CONJURE_MODEL:-claude-haiku-4-5}"
  local max_tokens="${CONJURE_MAX_TOKENS:-256}"
  local prompt="Output ONLY a single-line Unix command for macOS/zsh. No markdown fences, no explanation, no leading/trailing whitespace. Task: $*"

  local body
  body=$(jq -nc \
    --arg model "$model" \
    --arg prompt "$prompt" \
    --argjson max_tokens "$max_tokens" \
    '{model: $model, max_tokens: $max_tokens, messages: [{role: "user", content: $prompt}]}')

  local response
  response=$(curl -sS https://api.anthropic.com/v1/messages \
    -H "x-api-key: $key" \
    -H "anthropic-version: 2023-06-01" \
    -H "content-type: application/json" \
    -d "$body") || {
    print -u2 "conjure: API call failed (network or curl error)"
    return 1
  }

  # Surface API errors in a readable way
  local err
  err=$(print -r -- "$response" | jq -r '.error.message // empty')
  if [[ -n "$err" ]]; then
    print -u2 "conjure: API error: $err"
    return 1
  fi

  local cmd
  cmd=$(print -r -- "$response" | jq -r '.content[0].text // empty')
  if [[ -z "$cmd" ]]; then
    print -u2 "conjure: empty response"
    print -u2 "raw: $response"
    return 1
  fi

  _conjure_clean "$cmd"
}

# Interactive helper: prompts for the API key (hidden input) and stores it in
# the macOS Keychain under service=anthropic-api-key.
conjure-setup() {
  if ! command -v security >/dev/null; then
    print -u2 "conjure-setup: 'security' command not found (macOS-only)"
    return 1
  fi
  if ! command -v jq >/dev/null; then
    print -u2 "conjure-setup: 'jq' is required. Install with: brew install jq"
    return 1
  fi
  if ! command -v curl >/dev/null; then
    print -u2 "conjure-setup: 'curl' is required"
    return 1
  fi

  local existing
  existing=$(_conjure_get_key) || true
  if [[ -n "$existing" ]]; then
    print -n "conjure-setup: A key is already stored. Overwrite? [y/N] "
    local ans
    read -r ans
    [[ "$ans" == "y" || "$ans" == "Y" ]] || { print "Aborted."; return 0; }
  fi

  local key
  read -s "?Anthropic API key (hidden, get one at https://console.anthropic.com/settings/keys): " key
  print
  if [[ -z "$key" ]]; then
    print -u2 "conjure-setup: empty input, nothing stored."
    return 1
  fi

  if security add-generic-password -a "$USER" -s "anthropic-api-key" -w "$key" -U; then
    print "✓ API key stored in Keychain (service=anthropic-api-key)."
  else
    print -u2 "conjure-setup: failed to write Keychain entry."
    return 1
  fi
  unset key
}

alias cj=conjure
