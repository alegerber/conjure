package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGeneratePlain_SendsExpectedRequestAndStripsFences(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer sk-test" {
			t.Errorf("Authorization = %q, want Bearer sk-test", got)
		}
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if req.Model != "gpt-4o-mini" {
			t.Errorf("model = %q, want gpt-4o-mini", req.Model)
		}
		if len(req.Messages) != 2 || req.Messages[0].Role != "system" || req.Messages[1].Role != "user" {
			t.Errorf("messages = %+v", req.Messages)
		}
		if !strings.HasPrefix(req.Messages[1].Content, "Task: ") {
			t.Errorf("user content missing Task: prefix: %q", req.Messages[1].Content)
		}
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"` + "```\\nls -la\\n```" + `"}}]}`))
	}))
	defer srv.Close()

	c := New("sk-test", "gpt-4o-mini", 256)
	c.endpoint = srv.URL
	got, err := c.GeneratePlain(context.Background(), "sys", "list files")
	if err != nil {
		t.Fatalf("GeneratePlain: %v", err)
	}
	if got != "ls -la" {
		t.Errorf("got %q, want ls -la", got)
	}
}

func TestGenerateExplain_ReturnsToolCallArgs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(req.Tools) != 1 || req.Tools[0].Function.Name != "emit_command" {
			t.Errorf("tools = %+v", req.Tools)
		}
		if req.ToolChoice == nil || req.ToolChoice.Type != "function" || req.ToolChoice.Function.Name != "emit_command" {
			t.Errorf("tool_choice = %+v", req.ToolChoice)
		}
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"","tool_calls":[{"id":"x","type":"function","function":{"name":"emit_command","arguments":"{\"command\":\"ls -la\",\"explanation\":\"Lists files.\"}"}}]}}]}`))
	}))
	defer srv.Close()

	c := New("k", "m", 256)
	c.endpoint = srv.URL
	out, err := c.GenerateExplain(context.Background(), "sys", "task")
	if err != nil {
		t.Fatalf("GenerateExplain: %v", err)
	}
	if out.Command != "ls -la" || out.Explanation != "Lists files." {
		t.Errorf("out = %+v", out)
	}
}

func TestGeneratePlain_APIErrorIsReturned(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"message":"bad model","type":"invalid_request_error"}}`))
	}))
	defer srv.Close()
	c := New("k", "m", 256)
	c.endpoint = srv.URL
	_, err := c.GeneratePlain(context.Background(), "s", "t")
	if err == nil || !strings.Contains(err.Error(), "bad model") {
		t.Errorf("err = %v", err)
	}
}

func TestNewWithBase_BuildsEndpoint(t *testing.T) {
	c := NewWithBase("k", "m", 0, "https://example.com/")
	if c.endpoint != "https://example.com/v1/chat/completions" {
		t.Errorf("endpoint = %q", c.endpoint)
	}
}

func TestName_Defaults(t *testing.T) {
	c := New("k", "m", 0)
	if c.Name() != "openai" {
		t.Errorf("Name = %q", c.Name())
	}
	c.providerName = "codex"
	if c.Name() != "codex" {
		t.Errorf("Name = %q after override", c.Name())
	}
}
