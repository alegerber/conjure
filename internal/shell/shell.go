package shell

import (
	"embed"
	"fmt"
	"sort"
	"strings"
)

//go:embed templates/*
var fs embed.FS

var supported = map[string]string{
	"zsh":        "templates/zsh.sh.tmpl",
	"bash":       "templates/bash.sh.tmpl",
	"fish":       "templates/fish.fish.tmpl",
	"powershell": "templates/powershell.ps1.tmpl",
}

// Render returns the shell-integration script for the given shell name.
func Render(name string) (string, error) {
	path, ok := supported[strings.ToLower(name)]
	if !ok {
		return "", fmt.Errorf("unsupported shell %q (supported: zsh, bash, fish, powershell)", name)
	}
	b, err := fs.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read embedded template: %w", err)
	}
	return string(b), nil
}

// SupportedShells returns the list of supported shell names, sorted.
func SupportedShells() []string {
	names := make([]string, 0, len(supported))
	for n := range supported {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
