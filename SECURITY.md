# Security Policy

## Supported Versions

| Version | Supported          |
|---------|--------------------|
| 0.1.x   | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability in this project, please report it responsibly.

**Do not open a public GitHub issue for security vulnerabilities.**

Instead, please report vulnerabilities by emailing the maintainers directly or by using [GitHub's private vulnerability reporting](https://github.com/alegerber/conjure/security/advisories/new).

When reporting, please include:

- A description of the vulnerability
- Affected version, image tag, or commit SHA
- Deployment mode: Docker, local Node, or AWS Lambda
- Clear reproduction steps or a proof of concept
- Expected impact and any assumptions about attacker access
- Relevant logs, request samples, or configuration details with secrets redacted
- Any suggested fix (if applicable)

## Response Timeline

- **Acknowledgement**: We will acknowledge receipt of your report within 72 hours.
- **Assessment**: We aim to assess and validate the vulnerability within 7 days.
- **Fix**: Critical vulnerabilities will be prioritized and patched as soon as possible.

## Operational Security Notes

conjure translates natural-language descriptions into shell commands by
sending them to the Anthropic API and printing the result on stdout. The
relevant security surface is therefore *not* the network library — it is
the boundary between **untrusted model output** and **your shell**, plus
the handling of your **API credentials**.

- **API key storage.** `conjure --setup` stores the key in the macOS
  Keychain under service `anthropic-api-key` (account = `$USER`). The key
  is read on demand via `security find-generic-password` and never written
  to disk in plain text by conjure itself. Treat the Keychain entry like
  any other long-lived secret: lock your screen, enable FileVault, do not
  share the user account.
- **`ANTHROPIC_API_KEY` env-var fallback.** If set, the env var takes
  precedence over the Keychain entry. This is convenient for CI but exposes
  the key to every child process and to `ps`/`/proc/<pid>/environ` on
  shared hosts. Prefer the Keychain on workstations; on CI, scope the
  variable to a single job and never echo it.
- **Executing generated commands is the user's responsibility.** Output is
  plain text by design so you can read it before running. **Never pipe
  conjure output into `eval`, `sh -c`, or `bash -c` without reviewing it
  first** — a malformed description, a prompt-injection payload, or a
  model hallucination can produce destructive commands (`rm -rf`, `curl |
  sh`, credential exfiltration, …). The README's `eval "$(conjure …)"`
  example is a footgun, not a recommended pattern.
- **`cj` (zsh) requires explicit confirmation.** The zsh helper uses
  `print -z` to push the command into the next prompt's editor buffer.
  Nothing executes until you press Enter. Read the buffer before hitting
  return — especially when using `--explain`, since the explanation prints
  *above* the prompt and is easy to skim past.
- **Prompt-injection via the description.** Whatever you pass as the
  description is forwarded verbatim to the model with no sanitization. If
  the description is constructed from untrusted input (e.g. a Git commit
  message, a webhook payload, a colleague's chat snippet), an attacker can
  steer the model toward malicious commands. Treat conjure's output as
  *exactly as trustworthy as the description that produced it*.
- **Data sent to Anthropic.** Descriptions and the OS-aware system prompt
  are sent over TLS to `api.anthropic.com`. Anything you type — file paths,
  hostnames, internal project names, error messages pasted from logs — is
  visible to Anthropic and subject to their data-handling policy. Do not
  paste secrets, customer data, or regulated information into descriptions.
- **The BSD/GNU heuristic is not a security boundary.** The check in
  `internal/heuristic` warns on stderr when known-bad GNU flags slip
  through on macOS. It is a UX guard, not a sandbox: it does not block
  execution, does not enumerate dangerous commands, and is trivially
  bypassed by paraphrased output.
- **Uninstall does not remove the Keychain entry.** `./install.sh
  --uninstall` only removes the symlinks. If you are wiping a machine,
  hand it off, or rotating keys, delete the credential explicitly:
  `security delete-generic-password -a "$USER" -s "anthropic-api-key"`.

### Threat Model

**In scope** (please report):

- Extraction of the API key by another local user, another process, or via
  any conjure code path that logs, copies, or transmits it outside the
  Anthropic request.
- Code paths that cause command output to be executed automatically
  (without an explicit user confirmation step).
- TLS/transport flaws in how conjure talks to `api.anthropic.com`.
- Injection vectors where data conjure controls (system prompt, request
  framing) lets a malicious description bypass the read-before-run
  expectation — e.g. a generated command that hides a payload behind ANSI
  escapes, terminal-resetting sequences, or zero-width characters.
- Privilege escalation in `install.sh` (writes under `~/.local/bin` and the
  oh-my-zsh plugins directory).

**Out of scope** (acknowledged risks, not vulnerabilities):

- A user choosing to `eval` unreviewed output. Documented footgun.
- Compromise of the local machine, the Anthropic account, or the user's
  shell rc files.
- The model returning incorrect, unsafe, or low-quality commands when
  given a benign description — this is a model-quality issue, not a
  conjure vulnerability. Report it as a regular bug.
- The Anthropic API itself, its TLS endpoint, or its data-handling
  practices — please report those to Anthropic directly.
- Supply-chain compromise of Go modules in `go.sum` — report upstream and
  open a regular issue here so we can pin/replace.


