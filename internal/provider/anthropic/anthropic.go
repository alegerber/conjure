// Package anthropic is the Anthropic Messages API backend.
package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/alegerber/conjure/internal/provider"
)

const (
	endpoint        = "https://api.anthropic.com/v1/messages"
	anthropicAPIVer = "2023-06-01"

	// KeyringService is the system-keyring service name for the Anthropic API key.
	KeyringService = "anthropic-api-key"
	// EnvVar is the env-var fallback for the Anthropic API key.
	EnvVar = "ANTHROPIC_API_KEY"
)

type Client struct {
	APIKey     string
	Model      string
	MaxTokens  int
	HTTPClient *http.Client
	Endpoint   string // override for tests
}

func New(apiKey, model string, maxTokens int) *Client {
	return &Client{
		APIKey:     apiKey,
		Model:      model,
		MaxTokens:  maxTokens,
		HTTPClient: http.DefaultClient,
		Endpoint:   endpoint,
	}
}

func (c *Client) Name() string { return string(provider.KindAnthropic) }

func (c *Client) GeneratePlain(ctx context.Context, systemPrompt, task string) (string, error) {
	req := Request{
		Model:     c.Model,
		MaxTokens: c.MaxTokens,
		System: []SystemBlock{{
			Type:         "text",
			Text:         systemPrompt,
			CacheControl: &CacheControl{Type: "ephemeral"},
		}},
		Messages: []Message{{Role: "user", Content: "Task: " + task}},
	}
	resp, err := c.do(ctx, req)
	if err != nil {
		return "", err
	}
	if len(resp.Content) == 0 || resp.Content[0].Text == "" {
		return "", fmt.Errorf("empty response")
	}
	return stripFences(resp.Content[0].Text), nil
}

func (c *Client) do(ctx context.Context, body Request) (*Response, error) {
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("x-api-key", c.APIKey)
	httpReq.Header.Set("anthropic-version", anthropicAPIVer)
	httpReq.Header.Set("content-type", "application/json")

	httpResp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("API call failed: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	respBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var resp Response
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, fmt.Errorf("decode response: %w (body: %s)", err, string(respBytes))
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("API error: %s", resp.Error.Message)
	}
	return &resp, nil
}

const emitCommandSchema = `{
  "type": "object",
  "properties": {
    "command":     {"type":"string","description":"Single-line Unix command, no markdown fences"},
    "explanation": {"type":"string","description":"Brief one-sentence explanation"}
  },
  "required": ["command","explanation"]
}`

func (c *Client) GenerateExplain(ctx context.Context, systemPrompt, task string) (*provider.EmitCommand, error) {
	req := Request{
		Model:     c.Model,
		MaxTokens: c.MaxTokens,
		System: []SystemBlock{{
			Type:         "text",
			Text:         systemPrompt,
			CacheControl: &CacheControl{Type: "ephemeral"},
		}},
		Messages: []Message{{Role: "user", Content: "Task: " + task}},
		Tools: []Tool{{
			Name:        "emit_command",
			Description: "Emit the generated Unix command and a brief explanation",
			InputSchema: json.RawMessage(emitCommandSchema),
		}},
		ToolChoice: &ToolChoice{Type: "tool", Name: "emit_command"},
	}
	resp, err := c.do(ctx, req)
	if err != nil {
		return nil, err
	}
	for _, block := range resp.Content {
		if block.Type == "tool_use" && block.Name == "emit_command" {
			var out provider.EmitCommand
			if err := json.Unmarshal(block.Input, &out); err != nil {
				return nil, fmt.Errorf("decode tool_use input: %w", err)
			}
			if out.Command == "" {
				return nil, fmt.Errorf("tool_use returned empty command")
			}
			return &out, nil
		}
	}
	return nil, fmt.Errorf("no tool_use block in response")
}

// stripFences removes a leading/trailing markdown code fence and surrounding
// whitespace, returning the first non-empty line. Mirrors the bash version.
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
