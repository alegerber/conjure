package main

import (
	"testing"

	"github.com/alegerber/conjure/internal/provider"
)

func TestResolveSetupChoice(t *testing.T) {
	kinds := provider.Kinds()
	cases := []struct {
		name      string
		choice    string
		want      provider.Kind
		wantError bool
	}{
		{name: "first_index", choice: "1", want: kinds[0]},
		{name: "last_index", choice: "5", want: kinds[len(kinds)-1]},
		{name: "by_canonical_name", choice: "anthropic", want: provider.KindAnthropic},
		{name: "by_canonical_claude_cli", choice: "claude-cli", want: provider.KindClaudeCLI},
		{name: "rejects_legacy_claude_alias", choice: "claude", wantError: true},
		{name: "rejects_zero_index", choice: "0", wantError: true},
		{name: "rejects_overflow_index", choice: "99", wantError: true},
		{name: "rejects_unknown_string", choice: "gemini", wantError: true},
		{name: "rejects_empty", choice: "", wantError: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveSetupChoice(tc.choice, kinds)
			if tc.wantError {
				if err == nil {
					t.Errorf("resolveSetupChoice(%q) returned %q, want error", tc.choice, got)
				}
				return
			}
			if err != nil {
				t.Errorf("resolveSetupChoice(%q) errored: %v", tc.choice, err)
				return
			}
			if got != tc.want {
				t.Errorf("resolveSetupChoice(%q) = %q, want %q", tc.choice, got, tc.want)
			}
		})
	}
}

// kindDescriptions must keep one entry per provider.Kinds() entry, otherwise
// the menu silently shows blank descriptions for new providers.
func TestKindDescriptions_CoverAllKinds(t *testing.T) {
	for _, k := range provider.Kinds() {
		if _, ok := kindDescriptions[k]; !ok {
			t.Errorf("kindDescriptions missing entry for %q", k)
		}
	}
}
