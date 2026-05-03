package textutil

import "testing"

func TestStripFences(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{"plain", "ls -la", "ls -la"},
		{"fenced", "```\nls -la\n```", "ls -la"},
		{"fenced_with_lang", "```sh\nls -la\n```", "ls -la"},
		{"inline_backticks", "`ls -la`", "ls -la"},
		{"command_with_inner_backticks", "echo `date`", "echo `date`"},
		{"fenced_with_inner_backticks", "```\necho `date`\n```", "echo `date`"},
		{"surrounding_whitespace", "  \nls\n  ", "ls"},
		{"empty", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := StripFences(tc.in); got != tc.want {
				t.Errorf("StripFences(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
