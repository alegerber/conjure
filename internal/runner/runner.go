package runner

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var ErrAborted = errors.New("aborted by user")

// ConfirmAndRun prints cmd, reads y/N from in, and on confirmation runs
// the command via the platform's default shell. The executed command's
// stdout is wired to out; its stderr always goes to os.Stderr. If a caller
// passes os.Stderr as out, writes from the two streams may interleave at
// arbitrary byte boundaries because the Go runtime does not synchronise
// independent file handles to the same underlying descriptor.
// On Windows it shells out via "cmd /C"; elsewhere via "sh -c".
func ConfirmAndRun(cmd string, in io.Reader, out io.Writer) error {
	_, _ = fmt.Fprintf(out, "Run this command? [y/N]\n  %s\n> ", cmd)
	reader := bufio.NewReader(in)
	ans, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("read confirmation: %w", err)
	}
	ans = strings.TrimSpace(strings.ToLower(ans))
	if ans != "y" && ans != "yes" {
		return ErrAborted
	}

	var c *exec.Cmd
	if runtime.GOOS == "windows" {
		c = exec.Command("cmd", "/C", cmd)
	} else {
		c = exec.Command("sh", "-c", cmd)
	}
	c.Stdout = out
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}
	return nil
}
