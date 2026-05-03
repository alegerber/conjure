// Package sanitize scrubs LLM-produced text before it reaches the user's
// terminal or shell. The threat model: a model response could contain ANSI
// escapes, carriage returns, or NUL bytes that hide what is actually shown
// or executed (e.g. "rm -rf $HOME\rls" displays as "ls"). Every byte that
// could move the cursor, clear lines, or terminate a string is dropped.
package sanitize

import "strings"

// CommandLine returns the first line of s with control bytes removed.
//
// Removed: 0x00-0x1F (except 0x09 horizontal tab), 0x7F DEL.
// Truncation: anything from the first '\n' onward is dropped — a command
// from the model is required to be a single line, and the rest of the
// pipeline (runner, shell glue) makes the same assumption.
func CommandLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r < 0x20 && r != '\t' {
			continue
		}
		if r == 0x7F {
			continue
		}
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}
