package heuristic

import (
	"strings"
	"testing"
)

func TestCheckGNU_DarwinFlagsKnownPatterns(t *testing.T) {
	cases := []struct {
		cmd  string
		want string
	}{
		{"find . -printf '%s\\n'", "'find -printf' is GNU-only; use 'find -exec stat -f ... {} +'"},
		{"ls --color=auto", "'--color=auto' is GNU-only on macOS; ls accepts -G instead"},
		{"stat -c '%s' file", "'stat -c' is GNU-only; macOS uses 'stat -f'"},
		{"date -d 'yesterday'", "'date -d' is GNU-only; macOS uses 'date -v' (e.g. 'date -v-1d')"},
		{"readlink -f /tmp/x", "'readlink -f' is GNU-only; use 'realpath' or 'cd ... && pwd'"},
		{"sed -i 's/a/b/' file", "'sed -i' on macOS needs an empty backup arg: sed -i '' 'PATTERN'"},
	}
	for _, tc := range cases {
		hits := CheckGNU(tc.cmd, "darwin")
		joined := strings.Join(hits, "\n")
		if !strings.Contains(joined, tc.want) {
			t.Errorf("cmd=%q hits=%v, want contains %q", tc.cmd, hits, tc.want)
		}
	}
}

func TestCheckGNU_DarwinAcceptsBSDSed(t *testing.T) {
	hits := CheckGNU("sed -i '' 's/a/b/' file", "darwin")
	for _, h := range hits {
		if strings.Contains(h, "sed -i") {
			t.Errorf("BSD sed -i '' should not warn; got: %q", h)
		}
	}
}

func TestCheckGNU_LinuxReturnsNoWarnings(t *testing.T) {
	hits := CheckGNU("find . -printf '%s\\n'", "linux")
	if len(hits) != 0 {
		t.Errorf("Linux should produce no warnings; got: %v", hits)
	}
}

// Tokenisation eliminates the false positive where "sed -i" appears inside a
// quoted string and is not actually being executed.
func TestCheckGNU_QuotedSedDoesNotTrigger(t *testing.T) {
	hits := CheckGNU(`echo "sed -i 'foo'"`, "darwin")
	for _, h := range hits {
		if strings.Contains(h, "sed -i") {
			t.Errorf("quoted sed -i in echo should not warn; got: %q", h)
		}
	}
}

func TestCheckGNU_BSDDoubleQuotedEmptyBackupAccepted(t *testing.T) {
	hits := CheckGNU(`sed -i "" 's/a/b/' file`, "darwin")
	for _, h := range hits {
		if strings.Contains(h, "sed -i") {
			t.Errorf("BSD sed -i \"\" should not warn; got: %q", h)
		}
	}
}
