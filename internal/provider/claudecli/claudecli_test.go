package claudecli

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestHelperProcess is invoked as a fake `claude` binary by the tests below.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	mode := os.Getenv("HELPER_MODE")
	switch mode {
	case "plain":
		_, _ = os.Stdout.Write([]byte("```\nls -la\n```\n"))
	case "explain":
		_, _ = os.Stdout.Write([]byte("ls -la\n# Lists files.\n"))
	case "fail":
		_, _ = os.Stderr.Write([]byte("nope\n"))
		os.Exit(1)
	}
	os.Exit(0)
}

func fakeExec(mode string) func(ctx context.Context, name string, args ...string) *exec.Cmd {
	return func(ctx context.Context, name string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--"}
		cs = append(cs, name)
		cs = append(cs, args...)
		cmd := exec.CommandContext(ctx, os.Args[0], cs...)
		cmd.Env = append(os.Environ(),
			"GO_WANT_HELPER_PROCESS=1",
			"HELPER_MODE="+mode,
		)
		return cmd
	}
}

func TestGeneratePlain(t *testing.T) {
	orig := execCommand
	t.Cleanup(func() { execCommand = orig })
	execCommand = fakeExec("plain")

	c := New()
	got, err := c.GeneratePlain(context.Background(), "sys", "list")
	if err != nil {
		t.Fatalf("GeneratePlain: %v", err)
	}
	if got != "ls -la" {
		t.Errorf("got %q", got)
	}
}

func TestGenerateExplain(t *testing.T) {
	orig := execCommand
	t.Cleanup(func() { execCommand = orig })
	execCommand = fakeExec("explain")

	c := New()
	out, err := c.GenerateExplain(context.Background(), "sys", "list")
	if err != nil {
		t.Fatalf("GenerateExplain: %v", err)
	}
	if out.Command != "ls -la" {
		t.Errorf("Command = %q", out.Command)
	}
	if out.Explanation != "Lists files." {
		t.Errorf("Explanation = %q", out.Explanation)
	}
}

func TestGeneratePlain_BinaryFailureSurfacesStderr(t *testing.T) {
	orig := execCommand
	t.Cleanup(func() { execCommand = orig })
	execCommand = fakeExec("fail")

	c := New()
	_, err := c.GeneratePlain(context.Background(), "sys", "list")
	if err == nil || !strings.Contains(err.Error(), "nope") {
		t.Errorf("err = %v", err)
	}
}

func TestSplitCommandAndExplanation(t *testing.T) {
	cases := []struct {
		name, in, wantCmd, wantExpl string
	}{
		{"two lines", "ls -la\n# Lists files.", "ls -la", "Lists files."},
		{"fenced", "```\nls -la\n# Lists files.\n```", "ls -la", "Lists files."},
		{"no explanation", "ls -la", "ls -la", ""},
		{"inline backticks", "`ls -la`\n# x", "ls -la", "x"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotCmd, gotExpl := splitCommandAndExplanation(tc.in)
			if gotCmd != tc.wantCmd || gotExpl != tc.wantExpl {
				t.Errorf("got (%q, %q), want (%q, %q)", gotCmd, gotExpl, tc.wantCmd, tc.wantExpl)
			}
		})
	}
}
