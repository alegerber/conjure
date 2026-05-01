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
		{"find . -printf '%s\\n'", "find -printf"},
		{"ls --color=auto", "--color=auto"},
		{"stat -c '%s' file", "stat -c"},
		{"date -d 'yesterday'", "date -d"},
		{"readlink -f /tmp/x", "readlink -f"},
		{"sed -i 's/a/b/' file", "sed -i"},
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
