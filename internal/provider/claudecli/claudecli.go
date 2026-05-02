// Package claudecli shells out to the official `claude` CLI so users on a
// Claude Pro/Max subscription can reuse their existing session.
//
// Latency caveat: this path adds ~3.7s of plugin/MCP boot per call, vs ~1s
// for the direct Anthropic API. Use it only when you can't pay per-token.
package claudecli

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/alegerber/conjure/internal/provider"
)

// execCommand is exec.CommandContext, exposed for tests.
var execCommand = exec.CommandContext

type Client struct {
	Binary string // typically "claude"
}

func New() *Client { return &Client{Binary: "claude"} }

func (c *Client) Name() string { return string(provider.KindClaudeCLI) }

func (c *Client) bin() string {
	if c.Binary == "" {
		return "claude"
	}
	return c.Binary
}

func (c *Client) GeneratePlain(ctx context.Context, systemPrompt, task string) (string, error) {
	cmd := execCommand(ctx, c.bin(),
		"-p",
		"--output-format", "text",
		"--allowedTools", "",
		"--append-system-prompt", systemPrompt,
		"Task: "+task,
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("claude -p failed: %w (stderr: %s)", err, strings.TrimSpace(stderr.String()))
	}
	out := stripFences(stdout.String())
	if out == "" {
		return "", fmt.Errorf("empty response from claude -p")
	}
	return out, nil
}

func (c *Client) GenerateExplain(ctx context.Context, systemPrompt, task string) (*provider.EmitCommand, error) {
	// claude -p has no structured-output mode that matches our schema cleanly,
	// so we ask the model to emit "<command>\n# <explanation>" and parse.
	augmented := systemPrompt + "\n\nFor this request, output exactly two lines:\n" +
		"1. The single-line command.\n" +
		"2. A line starting with '# ' containing a brief one-sentence explanation.\n" +
		"No other text."
	cmd := execCommand(ctx, c.bin(),
		"-p",
		"--output-format", "text",
		"--allowedTools", "",
		"--append-system-prompt", augmented,
		"Task: "+task,
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("claude -p failed: %w (stderr: %s)", err, strings.TrimSpace(stderr.String()))
	}
	cmdLine, expl := splitCommandAndExplanation(stdout.String())
	if cmdLine == "" {
		return nil, fmt.Errorf("empty command from claude -p")
	}
	return &provider.EmitCommand{Command: cmdLine, Explanation: expl}, nil
}

func splitCommandAndExplanation(raw string) (string, string) {
	raw = strings.TrimSpace(raw)
	lines := strings.Split(raw, "\n")
	// Strip surrounding code fences.
	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[0]), "```") {
		lines = lines[1:]
	}
	if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "```" {
		lines = lines[:len(lines)-1]
	}
	var cmdLine, expl string
	for _, ln := range lines {
		t := strings.TrimSpace(ln)
		if t == "" {
			continue
		}
		if strings.HasPrefix(t, "#") {
			expl = strings.TrimSpace(strings.TrimPrefix(t, "#"))
			continue
		}
		if cmdLine == "" {
			if len(t) >= 2 && strings.HasPrefix(t, "`") && strings.HasSuffix(t, "`") {
				t = t[1 : len(t)-1]
			}
			cmdLine = t
		}
	}
	return cmdLine, expl
}

func stripFences(s string) string {
	s = strings.TrimSpace(s)
	lines := strings.Split(s, "\n")
	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[0]), "```") {
		lines = lines[1:]
	}
	if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "```" {
		lines = lines[:len(lines)-1]
	}
	for _, line := range lines {
		if t := strings.TrimSpace(line); t != "" {
			if len(t) >= 2 && strings.HasPrefix(t, "`") && strings.HasSuffix(t, "`") {
				return t[1 : len(t)-1]
			}
			return t
		}
	}
	return ""
}

// LookPath returns nil if `claude` is on PATH, an error otherwise. Used by
// `conjure setup` to fail fast when picking this provider.
func LookPath() error {
	if _, err := exec.LookPath("claude"); err != nil {
		return fmt.Errorf("`claude` not found on PATH — install Claude Code CLI first")
	}
	return nil
}
