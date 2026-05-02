package shell

import (
	"strings"
	"testing"
)

func TestRender_ZshContainsCjFunction(t *testing.T) {
	got, err := Render("zsh")
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(got, "cj()") || !strings.Contains(got, "print -z") {
		t.Errorf("zsh output missing cj() or print -z; got:\n%s", got)
	}
}

func TestRender_BashContainsBindWidget(t *testing.T) {
	got, err := Render("bash")
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(got, "bind -x") {
		t.Errorf("bash output missing bind -x; got:\n%s", got)
	}
}

func TestRender_UnknownShellErrors(t *testing.T) {
	if _, err := Render("tcsh"); err == nil {
		t.Errorf("Render(tcsh) should error")
	}
}
