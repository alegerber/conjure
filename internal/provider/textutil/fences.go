// Package textutil holds small text helpers shared across provider backends.
package textutil

import "strings"

// StripFences removes a leading/trailing markdown code fence and surrounding
// whitespace, returning the first non-empty line. If that line is wrapped in
// single backticks they are stripped too. Mirrors the historical bash helper.
func StripFences(s string) string {
	s = strings.TrimSpace(s)
	lines := strings.Split(s, "\n")
	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[0]), "```") {
		lines = lines[1:]
	}
	if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "```" {
		lines = lines[:len(lines)-1]
	}
	for _, line := range lines {
		if t := strings.TrimSpace(line); t != "" {
			if len(t) >= 2 && strings.HasPrefix(t, "`") && strings.HasSuffix(t, "`") {
				return t[1 : len(t)-1]
			}
			return t
		}
	}
	return ""
}
