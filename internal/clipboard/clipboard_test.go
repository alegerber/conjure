package clipboard

import (
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
		if c.Source() != "wl-copy" && c.Source() != "xclip" {
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

func TestNew_LinuxFallsBackToXclipWithoutWayland(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux only")
	}
	t.Setenv("WAYLAND_DISPLAY", "")
	c, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if c.Source() != "xclip" {
		t.Errorf("without WAYLAND_DISPLAY, want xclip, got %q", c.Source())
	}
}
