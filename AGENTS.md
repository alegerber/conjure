# AGENTS.md

> Guidance for AI coding agents (Claude Code, Codex, Cursor, Aider, etc.)
> working on this repo. Humans should read [README.md](README.md) first.

## Project

**conjure** is a small, fast CLI that turns natural-language descriptions
into single-line Unix commands by calling the Anthropic Messages API
directly (no Claude Code, no MCP, no plugin discovery — just `curl` and `jq`).

Why it exists: `claude -p` with full plugin/MCP boot took ~19s per call;
direct API + plain text is ~1.0s. The 19× speedup is the entire point.

## Architecture

```
conjure/
├── bin/conjure              # main CLI — bash 3.2+, the only entry point that does work
├── zsh/conjure.plugin.zsh   # optional oh-my-zsh wrapper: `cj` function + completion
├── install.sh               # creates symlinks: ~/.local/bin/conjure + OMZ plugin
├── README.md                # user-facing docs
├── AGENTS.md                # this file
├── LICENSE                  # MIT
└── .gitignore
```

### Key flows

**1. Plain generation (`conjure "..."`)**
- Read API key from macOS Keychain (`security find-generic-password`)
- Build OS-aware system prompt (`build_system_prompt $os_hint`)
- POST to `https://api.anthropic.com/v1/messages` with model + system + user message
- Parse `.content[0].text`, strip markdown fences, run `check_gnu_isms` heuristic
- Print to stdout

**2. Explain mode (`conjure -e "..."`)**
- Same as above, but the request body adds:
  - `tools: [{name: "emit_command", input_schema: {command, explanation}}]`
  - `tool_choice: {type: "tool", name: "emit_command"}` (forced)
- Parse `.content[].input.command` and `.input.explanation`
- Output: `<command>\n# <explanation>\n`

**3. zsh `cj` wrapper**
- Splits stdout: first line → ZLE buffer via `print -z`, rest → stdout
- Subcommand pass-through (`--setup`, `--version`, `--help`) skips the split

**4. Symlink layout (after `install.sh`)**
- `~/.local/bin/conjure` → `~/github/conjure/bin/conjure`
- `~/.oh-my-zsh/custom/plugins/conjure/conjure.plugin.zsh` → `~/github/conjure/zsh/conjure.plugin.zsh`
- The repo is the source of truth; edits go live immediately, no reinstall.

## Setup for development

The repo is already installed on the maintainer's machine. If you're working
in a fresh clone:

```sh
./install.sh                                    # create symlinks
conjure --setup                                 # store API key in Keychain
```

The Keychain entry is keyed `service=anthropic-api-key`, `account=$USER`. Read
it with `security find-generic-password -a "$USER" -s "anthropic-api-key" -w`.

## Testing

There is no automated test suite (yet). Use these manual checks:

```sh
# Syntax check (always do this after editing bash)
bash -n bin/conjure
bash -n install.sh

# Smoke tests
conjure --version
conjure --help
conjure "list .md files recursively"            # plain
conjure -e "find files modified today"          # explain
CONJURE_EXPLAIN=1 conjure "..."                 # env default
conjure --no-explain "..."                      # override env

# OS-awareness regression checks (these used to produce GNU on macOS)
conjure "find duplicate files by size and name"          # expect: stat -f, not -printf
conjure "replace foo with bar in all .txt files in place" # expect: sed -i ''
conjure "show date 7 days ago in YYYY-MM-DD"             # expect: date -v-7d
conjure "show file size of README.md in bytes"           # expect: stat -f
```

Heuristic check (`check_gnu_isms`) can be tested in isolation by extracting
the function into a temp file — see git history for an example pattern.

## Code style

- **Bash 3.2 compatible.** macOS ships bash 3.2 by default; do not use bash 4+
  features (associative arrays, `${var^^}`, `mapfile`, `&>`).
- 2-space indentation.
- Use `local` in functions for all variables that aren't intentionally global.
- Use `err()` (defined at the top of `bin/conjure`) for stderr, not `echo`.
- Prefer `printf '%s\n' "$x"` over `echo "$x"` for portability.
- `[[ ... ]]` is fine inside zsh files; in bash files prefer `[ ... ]` for
  POSIX-ish style unless `[[` is clearly better (regex match, glob match).
- Avoid creating new files unless necessary. Prefer extending `bin/conjure`.
- No comments that restate the code. Only comments that explain *why*.

## API conventions

When modifying the API call:

- **Endpoint**: `https://api.anthropic.com/v1/messages`
- **Headers**: `x-api-key`, `anthropic-version: 2023-06-01`, `content-type: application/json`
- **Default model**: `claude-haiku-4-5` (fastest current Anthropic model)
- **System prompt**: always sent as `system: [{type: "text", text: ..., cache_control: {type: "ephemeral"}}]`.
  Caching is currently a no-op on Haiku (prompt is below the 2048-token
  threshold), but the structure is in place — don't remove it.
- **Structured output**: use Tool Use (`tools` + forced `tool_choice`), never
  `response_format` JSON-Schema. Tool Use is ~5× faster.
- **Build request bodies with `jq -nc`**, not string concatenation. Keeps
  escaping correct.

When adding a new flag:

1. Document it in `usage()` (the heredoc at the top of `bin/conjure`)
2. Document it in README.md
3. If it influences generation, parse it in `main()` and pass through to `cmd_generate`
4. If it's a new subcommand (like `--setup`), add a case branch before
   the generation-flag-parsing loop

## Working agreements

### Git

- **GitHub identity**: `alegerber` (the maintainer has two; this repo belongs
  to the personal one). The remote uses the SSH host alias `github-private`
  which is already set up in `~/.ssh/config`.
- **Commit messages**: imperative subject line ≤72 chars, body wraps at 72,
  explains why. English. Conventional-Commits-ish but not strict.
- **Branches**: work on `main` for now (small repo, single maintainer).
- Always `bash -n` before commit on bash file edits.

### Updating docs

- New CLI flag → update `usage()` heredoc in `bin/conjure` AND README.md.
- New env var → update `usage()`, README's "Environment overrides" table.
- Behavior change → mention in README and possibly bump version (`CONJURE_VERSION`).

### Telling humans things

- Output user-facing text via `err()` to stderr, never to stdout — stdout is
  reserved for the generated command(s) so piping works.
- Successful generations: just print the command. No banner, no "Generated:".

## Backlog & roadmap

The backlog lives in [GitHub Issues](https://github.com/alegerber/conjure/issues).
Each issue has: context, why, implementation sketch with code, acceptance
criteria, and notes. Pick the highest-priority open issue when starting fresh.

Current priority order (see issues for details):

1. **#2** — `--copy` flag for clipboard via `pbcopy` *(small, complementary to `cj`)*
2. **#3** — `--run` flag for `eval` with confirm prompt *(safety-sensitive)*
3. **#1** — Bash bind -x widget for `cj`-equivalent in bash *(if you don't use bash interactively, defer)*
4. **#4** — Brew formula for public install *(requires going public first)*

When you finish an item, close the issue with `Closes #N` in the commit body.

## When in doubt

- Re-read [README.md](README.md) for user-facing semantics.
- Check `git log --oneline` for recent decisions and their rationale.
- The maintainer prefers small, focused commits over big ones; ask before
  large refactors.
