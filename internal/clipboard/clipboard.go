package clipboard

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
)

var ErrUnavailable = errors.New("no clipboard backend available")

type Copier interface {
	Copy(s string) error
	Source() string
}

func New() (Copier, error) {
	switch runtime.GOOS {
	case "darwin":
		return execCopier{cmd: "pbcopy"}, nil
	case "linux":
		// Wayland session takes precedence if WAYLAND_DISPLAY is set; else X11.
		if waylandSession() {
			return execCopier{cmd: "wl-copy"}, nil
		}
		if _, err := exec.LookPath("xclip"); err == nil {
			return execCopier{cmd: "xclip", args: []string{"-selection", "clipboard"}}, nil
		}
		if _, err := exec.LookPath("xsel"); err == nil {
			return execCopier{cmd: "xsel", args: []string{"--clipboard", "--input"}}, nil
		}
		return nil, fmt.Errorf("%w: install xclip or xsel", ErrUnavailable)
	case "windows":
		return execCopier{cmd: "clip.exe"}, nil
	default:
		return nil, ErrUnavailable
	}
}
