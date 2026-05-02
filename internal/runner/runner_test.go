package runner

import (
	"bytes"
	"strings"
	"testing"
)

func TestConfirmAndRun_DeniedReturnsErrAborted(t *testing.T) {
	in := strings.NewReader("n\n")
	var out bytes.Buffer
	err := ConfirmAndRun("echo skip", in, &out)
	if err != ErrAborted {
		t.Errorf("err = %v, want ErrAborted", err)
	}
}

func TestConfirmAndRun_AcceptedExecutes(t *testing.T) {
	in := strings.NewReader("y\n")
	var out bytes.Buffer
	if err := ConfirmAndRun("printf hello-runner", in, &out); err != nil {
		t.Fatalf("ConfirmAndRun: %v", err)
	}
	if !strings.Contains(out.String(), "hello-runner") {
		t.Errorf("stdout = %q, want hello-runner", out.String())
	}
}
