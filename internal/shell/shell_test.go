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

func TestRender_FishContainsCommandlineWidget(t *testing.T) {
	got, err := Render("fish")
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(got, "function cj") || !strings.Contains(got, "commandline -r") {
		t.Errorf("fish output missing function cj or commandline -r; got:\n%s", got)
	}
}

func TestRender_PowerShellContainsPSReadLineInsert(t *testing.T) {
	got, err := Render("powershell")
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(got, "function cj") || !strings.Contains(got, "PSConsoleReadLine") {
		t.Errorf("powershell output missing function cj or PSConsoleReadLine; got:\n%s", got)
	}
}

func TestRender_UnknownShellErrors(t *testing.T) {
	if _, err := Render("tcsh"); err == nil {
		t.Errorf("Render(tcsh) should error")
	}
}
