package sanitize

import "testing"

func TestCommandLine(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{"plain", "ls -la", "ls -la"},
		{"trim_whitespace", "  ls -la  ", "ls -la"},
		{"strip_carriage_return_overwrite", "rm -rf $HOME\rls", "rm -rf $HOMEls"},
		{"truncate_at_newline", "ls\nrm -rf /", "ls"},
		{"strip_ansi_clear_line", "\x1b[2K\rsafe-cmd", "[2Ksafe-cmd"},
		{"strip_nul_byte", "echo\x00 hi", "echo hi"},
		{"strip_del", "ls\x7f-la", "ls-la"},
		{"keep_tab", "awk '{print $1\t$2}'", "awk '{print $1\t$2}'"},
		{"keep_unicode", "echo äöü", "echo äöü"},
		{"empty", "", ""},
		{"only_controls", "\r\n\x00", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := CommandLine(tc.in); got != tc.want {
				t.Errorf("CommandLine(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
