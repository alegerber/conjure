package heuristic

import "strings"

// CheckGNU returns warning strings (each one human-readable) when known
// GNU-only flags appear in cmd while os is "darwin". Returns nil for non-Darwin
// or when no patterns match.
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
	// "sed -i 'X'" without the BSD "" backup arg right after -i.
	if strings.Contains(cmd, "sed -i '") && !strings.Contains(cmd, "sed -i ''") {
		hits = append(hits, "'sed -i' on macOS needs an empty backup arg: sed -i '' 'PATTERN'")
	}
	return hits
}
