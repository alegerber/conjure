package anthropic

import "encoding/json"

type SystemBlock struct {
	Type         string        `json:"type"`
	Text         string        `json:"text"`
	CacheControl *CacheControl `json:"cache_control,omitempty"`
}

type CacheControl struct {
	Type string `json:"type"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"input_schema"`
}

type ToolChoice struct {
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
}

type Request struct {
	Model      string        `json:"model"`
	MaxTokens  int           `json:"max_tokens"`
	System     []SystemBlock `json:"system"`
	Messages   []Message     `json:"messages"`
	Tools      []Tool        `json:"tools,omitempty"`
	ToolChoice *ToolChoice   `json:"tool_choice,omitempty"`
}

type ContentBlock struct {
	Type  string          `json:"type"`
	Text  string          `json:"text,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
}

type Response struct {
	Content []ContentBlock `json:"content"`
	Error   *APIError      `json:"error,omitempty"`
}

type APIError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
