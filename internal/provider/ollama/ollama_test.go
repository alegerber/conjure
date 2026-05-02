package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGeneratePlain_RoundTrip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Errorf("path = %q", r.URL.Path)
		}
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if req.Model != "llama3.1" {
			t.Errorf("model = %q", req.Model)
		}
		if req.Stream {
			t.Errorf("stream = true, want false")
		}
		if len(req.Messages) != 2 {
			t.Errorf("len(messages) = %d", len(req.Messages))
		}
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write([]byte(`{"message":{"role":"assistant","content":"ls -la"}}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "llama3.1")
	c.Endpoint = srv.URL + "/api/chat"
	got, err := c.GeneratePlain(context.Background(), "sys", "list")
	if err != nil {
		t.Fatalf("GeneratePlain: %v", err)
	}
	if got != "ls -la" {
		t.Errorf("got %q", got)
	}
}

func TestGenerateExplain_ParsesToolCall(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req chatRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if len(req.Tools) != 1 {
			t.Errorf("tools = %+v", req.Tools)
		}
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write([]byte(`{"message":{"role":"assistant","content":"","tool_calls":[{"function":{"name":"emit_command","arguments":{"command":"ls","explanation":"Lists files."}}}]}}`))
	}))
	defer srv.Close()
	c := New(srv.URL, "llama3.1")
	c.Endpoint = srv.URL + "/api/chat"
	out, err := c.GenerateExplain(context.Background(), "sys", "list")
	if err != nil {
		t.Fatalf("GenerateExplain: %v", err)
	}
	if out.Command != "ls" || out.Explanation != "Lists files." {
		t.Errorf("out = %+v", out)
	}
}

func TestGeneratePlain_ErrorPropagated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"error":"model not found"}`))
	}))
	defer srv.Close()
	c := New(srv.URL, "missing")
	c.Endpoint = srv.URL + "/api/chat"
	_, err := c.GeneratePlain(context.Background(), "s", "t")
	if err == nil || !strings.Contains(err.Error(), "model not found") {
		t.Errorf("err = %v", err)
	}
}

func TestGeneratePlain_RequiresModel(t *testing.T) {
	c := New("http://localhost:11434", "")
	_, err := c.GeneratePlain(context.Background(), "s", "t")
	if err == nil || !strings.Contains(err.Error(), "model") {
		t.Errorf("err = %v, want model-not-set error", err)
	}
}
