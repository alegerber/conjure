package heuristic

import "strings"

// CheckGNU returns warning strings (each one human-readable) when known
// GNU-only flags appear in cmd while os is "darwin". Returns nil for non-Darwin
// or when no patterns match.
//
// Pattern matching is substring-based (intentionally simple). It can produce
// false positives on quoted strings or composite words ("xfind -printf",
// echo "stat -c is bad"). The output is labelled as a "heuristic warning",
// so occasional noise is acceptable.
func CheckGNU(cmd, os string) []string {
	if strings.ToLower(os) != "darwin" {
		return nil
	}
	var hits []string
	if strings.Contains(cmd, "find ") && strings.Contains(cmd, "-printf") {
		hits = append(hits, "'find -printf' is GNU-only; use 'find -exec stat -f ... {} +'")
	}
	if strings.Contains(cmd, "--color=auto") {
		hits = append(hits, "'--color=auto' is GNU-only on macOS; ls accepts -G instead")
	}
	if strings.Contains(cmd, "stat -c") {
		hits = append(hits, "'stat -c' is GNU-only; macOS uses 'stat -f'")
	}
	if strings.Contains(cmd, "date -d ") {
		hits = append(hits, "'date -d' is GNU-only; macOS uses 'date -v' (e.g. 'date -v-1d')")
	}
	if strings.Contains(cmd, "readlink -f") {
		hits = append(hits, "'readlink -f' is GNU-only; use 'realpath' or 'cd ... && pwd'")
	}
	// Catch GNU "sed -i 'PATTERN'" without the BSD-required empty backup arg.
	// Tokenising on whitespace avoids the substring false positive on quoted
	// strings like `echo "sed -i 'foo'"`. Remaining limitation: BSD's named-
	// backup form `sed -i .bak 'PATTERN'` still triggers a warning, since
	// distinguishing it from a GNU-style pattern requires real shell parsing.
	if hasGNUSedIArg(cmd) {
		hits = append(hits, "'sed -i' on macOS needs an empty backup arg: sed -i '' 'PATTERN'")
	}
	return hits
}

func hasGNUSedIArg(cmd string) bool {
	fields := strings.Fields(cmd)
	for i, f := range fields {
		if f != "sed" {
			continue
		}
		for j := i + 1; j+1 < len(fields); j++ {
			if fields[j] != "-i" {
				continue
			}
			next := fields[j+1]
			if next != "''" && next != `""` {
				return true
			}
			break
		}
	}
	return false
}
