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
	if !strings.HasSuffix(got, "\n") {
		t.Error("Darwin prompt must end with a newline")
	}
}

func TestBuildSystem_LinuxMentionsGNU(t *testing.T) {
	got := BuildSystem("linux")
	for _, want := range []string{"GNU", "stat -c", "date -d"} {
		if !strings.Contains(got, want) {
			t.Errorf("Linux prompt missing %q", want)
		}
	}
	if !strings.HasSuffix(got, "\n") {
		t.Error("Linux prompt must end with a newline")
	}
}

func TestBuildSystem_OtherIsPOSIXFallback(t *testing.T) {
	got := BuildSystem("freebsd")
	if !strings.Contains(got, "POSIX-portable") {
		t.Errorf("Other-OS prompt should mention POSIX-portable; got: %q", got)
	}
	if !strings.Contains(got, "freebsd") {
		t.Errorf("Other-OS prompt should contain the os name; got: %q", got)
	}
	if !strings.HasSuffix(got, "\n") {
		t.Error("Fallback prompt must end with a newline")
	}
}

func TestBuildSystem_CaseInsensitive(t *testing.T) {
	// The `os` config field may carry mixed-case values like "Darwin".
	if got := BuildSystem("Darwin"); !strings.Contains(got, "BSD") {
		t.Errorf(`BuildSystem("Darwin") should route to darwin branch; got: %q`, got)
	}
	if got := BuildSystem("LINUX"); !strings.Contains(got, "GNU") {
		t.Errorf(`BuildSystem("LINUX") should route to linux branch; got: %q`, got)
	}
}
