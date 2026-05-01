package clipboard

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

type execCopier struct {
	cmd  string
	args []string
}

func (e execCopier) Copy(s string) error {
	c := exec.Command(e.cmd, e.args...)
	c.Stdin = bytes.NewBufferString(s)
	if out, err := c.CombinedOutput(); err != nil {
		return fmt.Errorf("%s failed: %w (%s)", e.cmd, err, string(out))
	}
	return nil
}

func (e execCopier) Source() string { return e.cmd }

func waylandSession() bool { return os.Getenv("WAYLAND_DISPLAY") != "" }
