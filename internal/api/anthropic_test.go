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
