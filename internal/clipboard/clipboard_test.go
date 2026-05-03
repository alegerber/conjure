package clipboard

import (
	"errors"
	"os/exec"
	"runtime"
	"testing"
)

func TestNew_PicksOSAppropriateBackend(t *testing.T) {
	c, err := New()
	if err != nil {
		t.Skipf("no clipboard backend on this platform: %v", err)
	}
	switch runtime.GOOS {
	case "darwin":
		if c.Source() != "pbcopy" {
			t.Errorf("darwin source = %q", c.Source())
		}
	case "windows":
		if c.Source() != "clip.exe" {
			t.Errorf("windows source = %q", c.Source())
		}
	case "linux":
		switch c.Source() {
		case "wl-copy", "xclip", "xsel":
		default:
			t.Errorf("linux source = %q", c.Source())
		}
	}
}

func TestNew_LinuxPicksWaylandWhenDisplaySet(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux only")
	}
	t.Setenv("WAYLAND_DISPLAY", ":wayland-0")
	c, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if c.Source() != "wl-copy" {
		t.Errorf("with WAYLAND_DISPLAY set, want wl-copy, got %q", c.Source())
	}
}

func TestNew_LinuxX11PrefersXclipOverXsel(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux only")
	}
	t.Setenv("WAYLAND_DISPLAY", "")

	_, xclipErr := exec.LookPath("xclip")
	_, xselErr := exec.LookPath("xsel")
	if xclipErr != nil && xselErr != nil {
		t.Skip("neither xclip nor xsel on PATH")
	}

	c, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	wantSource := "xsel"
	if xclipErr == nil {
		wantSource = "xclip" // xclip wins when both present
	}
	if c.Source() != wantSource {
		t.Errorf("source = %q, want %q (xclip available=%v, xsel available=%v)",
			c.Source(), wantSource, xclipErr == nil, xselErr == nil)
	}
}

func TestNew_LinuxX11ErrorsWhenNoBackendInstalled(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux only")
	}
	t.Setenv("WAYLAND_DISPLAY", "")
	t.Setenv("PATH", "") // empty PATH makes exec.LookPath fail for every command

	_, err := New()
	if err == nil {
		t.Fatal("New: want error when no clipboard backend on PATH")
	}
	if !errors.Is(err, ErrUnavailable) {
		t.Errorf("err = %v, want errors.Is(err, ErrUnavailable)", err)
	}
}
