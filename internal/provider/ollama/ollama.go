// Package ollama is the local-Ollama backend.
package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/alegerber/conjure/internal/provider"
	"github.com/alegerber/conjure/internal/provider/textutil"
)

const DefaultHost = "http://localhost:11434"

type Client struct {
	Host       string
	Model      string
	HTTPClient *http.Client
	Endpoint   string // override for tests; takes precedence over Host
}

func New(host, model string) *Client {
	if host == "" {
		host = DefaultHost
	}
	return &Client{
		Host:       host,
		Model:      model,
		HTTPClient: http.DefaultClient,
		Endpoint:   strings.TrimRight(host, "/") + "/api/chat",
	}
}

func (c *Client) Name() string { return string(provider.KindOllama) }

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

type chatRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
	Stream   bool      `json:"stream"`
	Tools    []tool    `json:"tools,omitempty"`
}

type respFunctionCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type respToolCall struct {
	Function respFunctionCall `json:"function"`
}

type respMessage struct {
	Role      string         `json:"role"`
	Content   string         `json:"content"`
	ToolCalls []respToolCall `json:"tool_calls,omitempty"`
}

type chatResponse struct {
	Message respMessage `json:"message"`
	Error   string      `json:"error,omitempty"`
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
	if c.Model == "" {
		return "", fmt.Errorf("ollama: model not set (configure via `conjure setup` or --model)")
	}
	req := chatRequest{
		Model: c.Model,
		Messages: []message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: "Task: " + task},
		},
		Stream: false,
	}
	resp, err := c.do(ctx, req)
	if err != nil {
		return "", err
	}
	if resp.Message.Content == "" {
		return "", fmt.Errorf("empty response")
	}
	return textutil.StripFences(resp.Message.Content), nil
}

func (c *Client) GenerateExplain(ctx context.Context, systemPrompt, task string) (*provider.EmitCommand, error) {
	if c.Model == "" {
		return nil, fmt.Errorf("ollama: model not set (configure via `conjure setup` or --model)")
	}
	req := chatRequest{
		Model: c.Model,
		Messages: []message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: "Task: " + task},
		},
		Stream: false,
		Tools: []tool{{
			Type: "function",
			Function: toolFunctionDef{
				Name:        "emit_command",
				Description: "Emit the generated Unix command and a brief explanation",
				Parameters:  json.RawMessage(emitCommandSchema),
			},
		}},
	}
	resp, err := c.do(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(resp.Message.ToolCalls) == 0 {
		return nil, fmt.Errorf("no tool_call in response (model may not support tool use)")
	}
	args := resp.Message.ToolCalls[0].Function.Arguments
	var out provider.EmitCommand
	if err := json.Unmarshal(args, &out); err != nil {
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
	httpReq.Header.Set("content-type", "application/json")

	httpResp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama call failed (is it running at %s?): %w", c.Host, err)
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
	if resp.Error != "" {
		return nil, fmt.Errorf("ollama error: %s", resp.Error)
	}
	return &resp, nil
}

