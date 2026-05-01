package prompt

import (
	"fmt"
	"strings"
)

const darwin = `You translate natural-language descriptions into single-line Unix commands.
Target system: macOS (Darwin) with zsh and BSD coreutils.

CRITICAL RULES:
1. Output exactly ONE command line. No markdown fences. No explanation prose.
2. Prefer BSD/POSIX syntax. NEVER use GNU-only flags. NEVER use:
   --color=auto, find -printf, date -d, stat -c, sed -i 'PATTERN' (without ''),
   xargs -d, ls --quoting-style, --files0-from, readlink -f.
3. Use idiomatic macOS commands: ls -G, sed -i '' 'PATTERN', stat -f, date -v.
4. Prefer ` + "`-exec ... {} +`" + ` over ` + "`-exec ... \\;`" + ` for performance.
5. Quote arguments that may contain spaces, globs, or special chars.

GNU-bad → BSD-good examples:
  find . -printf '%s %f\n'         →  find . -exec stat -f '%z %N' {} +
  sed -i 's/a/b/' file             →  sed -i '' 's/a/b/' file
  date -d "1 day ago" +%Y-%m-%d    →  date -v-1d +%Y-%m-%d
  stat -c '%s' file                →  stat -f '%z' file
  ls --color=auto                  →  ls -G
  readlink -f path                 →  realpath path     (or: cd path && pwd)
  find . -type f | xargs -d '\n'   →  find . -type f -print0 | xargs -0
`

const linux = `You translate natural-language descriptions into single-line Unix commands.
Target system: Linux with bash and GNU coreutils.

CRITICAL RULES:
1. Output exactly ONE command line. No markdown fences. No explanation prose.
2. Use idiomatic GNU syntax: stat -c, date -d, sed -i 'PATTERN' (no empty arg),
   find -printf, readlink -f, xargs -d.
3. Prefer ` + "`-exec ... {} +`" + ` over ` + "`-exec ... \\;`" + ` for performance.
4. Quote arguments that may contain spaces, globs, or special chars.
`

// BuildSystem returns an OS-aware system prompt for the given operating
// system. os should be a runtime.GOOS-style lowercase identifier ("darwin",
// "linux", "windows", ...); mixed-case values such as "Darwin" are accepted.
// For unrecognised values, a POSIX-portable fallback is returned.
func BuildSystem(os string) string {
	switch strings.ToLower(os) {
	case "darwin":
		return darwin
	case "linux":
		return linux
	default:
		return fmt.Sprintf(`You translate natural-language descriptions into single-line POSIX shell commands
for %s. Output exactly ONE command line. No markdown. Prefer POSIX-portable syntax.
`, os)
	}
}
