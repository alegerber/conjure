package main

import "testing"

func TestResolveOllamaHost_Precedence(t *testing.T) {
	cases := []struct {
		name       string
		flag, env  string
		cfg        string
		setEnv     bool
		want       string
	}{
		{name: "flag_wins_over_env_and_cfg", flag: "http://flag", env: "http://env", cfg: "http://cfg", setEnv: true, want: "http://flag"},
		{name: "env_wins_over_cfg", flag: "", env: "http://env", cfg: "http://cfg", setEnv: true, want: "http://env"},
		{name: "cfg_used_when_no_flag_or_env", flag: "", cfg: "http://cfg", setEnv: false, want: "http://cfg"},
		{name: "empty_when_nothing_set", flag: "", cfg: "", setEnv: false, want: ""},
		{name: "empty_env_treated_as_unset", flag: "", env: "", cfg: "http://cfg", setEnv: true, want: "http://cfg"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setEnv {
				t.Setenv("OLLAMA_HOST", tc.env)
			} else {
				t.Setenv("OLLAMA_HOST", "")
			}
			if got := resolveOllamaHost(tc.flag, tc.cfg); got != tc.want {
				t.Errorf("resolveOllamaHost(%q, %q) [env=%q set=%v] = %q, want %q",
					tc.flag, tc.cfg, tc.env, tc.setEnv, got, tc.want)
			}
		})
	}
}
