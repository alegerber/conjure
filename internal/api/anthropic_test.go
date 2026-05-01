package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGeneratePlain_StripsFencesAndReturnsCommand(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("x-api-key"); got != "test-key" {
			t.Errorf("x-api-key = %q, want test-key", got)
		}
		if got := r.Header.Get("anthropic-version"); got != "2023-06-01" {
			t.Errorf("anthropic-version = %q", got)
		}
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if !strings.HasPrefix(req.Messages[0].Content, "Task: ") {
			t.Errorf("user message = %q, want Task: prefix", req.Messages[0].Content)
		}
		if len(req.System) != 1 {
			t.Fatalf("len(req.System) = %d, want 1", len(req.System))
		}
		if req.System[0].CacheControl == nil || req.System[0].CacheControl.Type != "ephemeral" {
			t.Errorf("system[0].cache_control = %+v, want ephemeral", req.System[0].CacheControl)
		}
		if req.Model != "claude-haiku-4-5" {
			t.Errorf("model = %q, want claude-haiku-4-5", req.Model)
		}
		if req.MaxTokens != 256 {
			t.Errorf("max_tokens = %d, want 256", req.MaxTokens)
		}
		w.Header().Set("content-type", "application/json")
		// Return a fenced code block so that stripFences is exercised.
		// Wire JSON: {"content":[{"type":"text","text":"```\nfind . -name '*.md'\n```\n"}]}
		body := fmt.Sprintf(
			`{"content":[{"type":"text","text":"%s\nfind . -name '*.md'\n%s\n"}]}`,
			"```", "```",
		)
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	c := New("test-key", "claude-haiku-4-5", 256)
	c.Endpoint = srv.URL

	got, err := c.GeneratePlain(context.Background(), "system text", "list md files")
	if err != nil {
		t.Fatalf("GeneratePlain: %v", err)
	}
	want := "find . -name '*.md'"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStripFences_PreservesInnerBackticks(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{"plain", "ls -la", "ls -la"},
		{"fenced", "```\nls -la\n```", "ls -la"},
		{"fenced with lang", "```sh\nls -la\n```", "ls -la"},
		{"inline backticks", "`ls -la`", "ls -la"},
		{"command with inner backticks", "echo `date`", "echo `date`"},
		{"fenced with inner backticks", "```\necho `date`\n```", "echo `date`"},
		{"surrounding whitespace", "  \nls\n  ", "ls"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := stripFences(tc.in); got != tc.want {
				t.Errorf("stripFences(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestGeneratePlain_APIErrorIsReturned(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"type":"invalid_request_error","message":"bad model"}}`))
	}))
	defer srv.Close()

	c := New("k", "m", 256)
	c.Endpoint = srv.URL
	_, err := c.GeneratePlain(context.Background(), "s", "t")
	if err == nil || !strings.Contains(err.Error(), "bad model") {
		t.Errorf("err = %v, want contain 'bad model'", err)
	}
}

func TestGenerateExplain_ReturnsCommandAndExplanation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(req.Tools) != 1 || req.Tools[0].Name != "emit_command" {
			t.Errorf("tools = %+v, want one tool named emit_command", req.Tools)
		}
		if req.ToolChoice == nil || req.ToolChoice.Type != "tool" || req.ToolChoice.Name != "emit_command" {
			t.Errorf("tool_choice = %+v", req.ToolChoice)
		}
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write([]byte(`{"content":[{"type":"tool_use","name":"emit_command","input":{"command":"ls -la","explanation":"Lists files."}}]}`))
	}))
	defer srv.Close()

	c := New("k", "m", 256)
	c.Endpoint = srv.URL
	out, err := c.GenerateExplain(context.Background(), "sys", "list files")
	if err != nil {
		t.Fatalf("GenerateExplain: %v", err)
	}
	if out.Command != "ls -la" {
		t.Errorf("Command = %q", out.Command)
	}
	if out.Explanation != "Lists files." {
		t.Errorf("Explanation = %q", out.Explanation)
	}
}
