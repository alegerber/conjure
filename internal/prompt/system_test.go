package prompt

import (
	"strings"
	"testing"
)

func TestBuildSystem_DarwinMentionsBSDExamples(t *testing.T) {
	got := BuildSystem("darwin")
	for _, want := range []string{"BSD", "stat -f", "sed -i ''", "date -v", "ls -G", "realpath"} {
		if !strings.Contains(got, want) {
			t.Errorf("Darwin prompt missing %q", want)
		}
	}
}

func TestBuildSystem_LinuxMentionsGNU(t *testing.T) {
	got := BuildSystem("linux")
	for _, want := range []string{"GNU", "stat -c", "date -d"} {
		if !strings.Contains(got, want) {
			t.Errorf("Linux prompt missing %q", want)
		}
	}
}

func TestBuildSystem_OtherIsPOSIXFallback(t *testing.T) {
	got := BuildSystem("freebsd")
	if !strings.Contains(got, "POSIX-portable") {
		t.Errorf("Other-OS prompt should mention POSIX-portable; got: %q", got)
	}
}
