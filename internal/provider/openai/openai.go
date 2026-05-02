// Package openai is the OpenAI Chat Completions backend.
package openai

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
	defaultEndpoint = "https://api.openai.com/v1/chat/completions"

	// KeyringService is the system-keyring service name for the OpenAI API key.
	KeyringService = "openai-api-key"
	// EnvVar is the env-var fallback for the OpenAI API key.
	EnvVar = "OPENAI_API_KEY"
)

// Client is the OpenAI Chat Completions client. The Endpoint field is exposed
// so tests can point it at httptest.Server.
type Client struct {
	APIKey       string
	Model        string
	MaxTokens    int
	HTTPClient   *http.Client
	Endpoint     string
	ProviderName string // overridden by codex wrapper; defaults to "openai".
}

func New(apiKey, model string, maxTokens int) *Client {
	return NewWithBase(apiKey, model, maxTokens, "")
}

// NewWithBase constructs a client with a base URL override (e.g. an
// OpenAI-compatible proxy). Empty baseURL uses the default endpoint.
func NewWithBase(apiKey, model string, maxTokens int, baseURL string) *Client {
	ep := defaultEndpoint
	if baseURL != "" {
		ep = strings.TrimRight(baseURL, "/") + "/v1/chat/completions"
	}
	return &Client{
		APIKey:       apiKey,
		Model:        model,
		MaxTokens:    maxTokens,
		HTTPClient:   http.DefaultClient,
		Endpoint:     ep,
		ProviderName: string(provider.KindOpenAI),
	}
}

func (c *Client) Name() string {
	if c.ProviderName == "" {
		return string(provider.KindOpenAI)
	}
	return c.ProviderName
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type toolFunctionDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters"`
}

type tool struct {
	Type     string          `json:"type"`
	Function toolFunctionDef `json:"function"`
}

type toolChoiceFunc struct {
	Name string `json:"name"`
}

type toolChoice struct {
	Type     string         `json:"type"`
	Function toolChoiceFunc `json:"function"`
}

type chatRequest struct {
	Model      string      `json:"model"`
	Messages   []message   `json:"messages"`
	MaxTokens  int         `json:"max_tokens,omitempty"`
	Tools      []tool      `json:"tools,omitempty"`
	ToolChoice *toolChoice `json:"tool_choice,omitempty"`
}

type respFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type respToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function respFunctionCall `json:"function"`
}

type respMessage struct {
	Role      string         `json:"role"`
	Content   string         `json:"content"`
	ToolCalls []respToolCall `json:"tool_calls,omitempty"`
}

type respChoice struct {
	Message respMessage `json:"message"`
}

type apiError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

type chatResponse struct {
	Choices []respChoice `json:"choices"`
	Error   *apiError    `json:"error,omitempty"`
}

const emitCommandSchema = `{
  "type": "object",
  "properties": {
    "command":     {"type":"string","description":"Single-line Unix command, no markdown fences"},
    "explanation": {"type":"string","description":"Brief one-sentence explanation"}
  },
  "required": ["command","explanation"]
}`

func (c *Client) GeneratePlain(ctx context.Context, systemPrompt, task string) (string, error) {
	req := chatRequest{
		Model:     c.Model,
		MaxTokens: c.MaxTokens,
		Messages: []message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: "Task: " + task},
		},
	}
	resp, err := c.do(ctx, req)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("empty response")
	}
	return stripFences(resp.Choices[0].Message.Content), nil
}

func (c *Client) GenerateExplain(ctx context.Context, systemPrompt, task string) (*provider.EmitCommand, error) {
	req := chatRequest{
		Model:     c.Model,
		MaxTokens: c.MaxTokens,
		Messages: []message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: "Task: " + task},
		},
		Tools: []tool{{
			Type: "function",
			Function: toolFunctionDef{
				Name:        "emit_command",
				Description: "Emit the generated Unix command and a brief explanation",
				Parameters:  json.RawMessage(emitCommandSchema),
			},
		}},
		ToolChoice: &toolChoice{Type: "function", Function: toolChoiceFunc{Name: "emit_command"}},
	}
	resp, err := c.do(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 || len(resp.Choices[0].Message.ToolCalls) == 0 {
		return nil, fmt.Errorf("no tool_call in response")
	}
	args := resp.Choices[0].Message.ToolCalls[0].Function.Arguments
	var out provider.EmitCommand
	if err := json.Unmarshal([]byte(args), &out); err != nil {
		return nil, fmt.Errorf("decode tool_call arguments: %w", err)
	}
	if out.Command == "" {
		return nil, fmt.Errorf("tool_call returned empty command")
	}
	return &out, nil
}

func (c *Client) do(ctx context.Context, body chatRequest) (*chatResponse, error) {
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
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

	var resp chatResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, fmt.Errorf("decode response: %w (body: %s)", err, string(respBytes))
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("API error: %s", resp.Error.Message)
	}
	return &resp, nil
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
