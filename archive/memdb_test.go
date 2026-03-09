package archive_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	session "github.com/anatolykoptev/go-session"

	"github.com/anatolykoptev/go-session/archive"
)

func makeSession(key string) *session.Session {
	s := session.NewSession(key)
	s.Summary = "test summary"
	s.AddFacts([]session.Fact{
		{Content: "fact one", ExtractedAt: time.Now()},
		{Content: "fact two", ExtractedAt: time.Now()},
	})
	s.AddMessage(session.Message{Role: "user", Content: "hello"})
	s.AddMessage(session.Message{Role: "assistant", Content: "hi there"})
	s.AddMessage(session.Message{Role: "user", Content: "how are you"})
	return s
}

func TestNewMemDBArchiver_Success(t *testing.T) {
	var receivedBody []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", ct)
		}
		var err error
		receivedBody, err = io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		_ = r.Body.Close()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"abc","status":"ok"}`))
	}))
	defer srv.Close()

	fn := archive.NewMemDBArchiver(archive.Config{
		BaseURL:    srv.URL,
		UserID:     "testuser",
		CubeID:     "testcube",
		KeepLastN:  5,
		HTTPClient: srv.Client(),
	})

	s := makeSession("session-1")
	if err := fn(s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(receivedBody, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if payload["user_id"] != "testuser" {
		t.Errorf("user_id = %v, want testuser", payload["user_id"])
	}
	if payload["cube_id"] != "testcube" {
		t.Errorf("cube_id = %v, want testcube", payload["cube_id"])
	}
}

func TestNewMemDBArchiver_ContentFormat(t *testing.T) {
	var content string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		var payload map[string]any
		_ = json.Unmarshal(body, &payload)
		content, _ = payload["content"].(string)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"1","status":"ok"}`))
	}))
	defer srv.Close()

	fn := archive.NewMemDBArchiver(archive.Config{
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
	})

	s := makeSession("mykey")
	if err := fn(s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		"mykey",
		"test summary",
		"fact one",
		"fact two",
		"user: hello",
		"assistant: hi there",
	}
	for _, want := range checks {
		if !strings.Contains(content, want) {
			t.Errorf("content missing %q\ncontent:\n%s", want, content)
		}
	}
}

func TestNewMemDBArchiver_KeepLastN(t *testing.T) {
	var content string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		var payload map[string]any
		_ = json.Unmarshal(body, &payload)
		content, _ = payload["content"].(string)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"2","status":"ok"}`))
	}))
	defer srv.Close()

	fn := archive.NewMemDBArchiver(archive.Config{
		BaseURL:    srv.URL,
		KeepLastN:  2,
		HTTPClient: srv.Client(),
	})

	s := session.NewSession("k")
	for i := range 6 {
		role := "user"
		if i%2 != 0 {
			role = "assistant"
		}
		s.AddMessage(session.Message{Role: role, Content: strings.Repeat("x", i+1)})
	}

	if err := fn(s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only the last 2 messages should appear.
	if strings.Contains(content, "Last 6 messages") {
		t.Error("content should not include all 6 messages")
	}
	if !strings.Contains(content, "Last 2 messages") {
		t.Errorf("content should say 'Last 2 messages'\ncontent:\n%s", content)
	}
}

func TestNewMemDBArchiver_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	fn := archive.NewMemDBArchiver(archive.Config{
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
	})

	if err := fn(makeSession("k")); err == nil {
		t.Fatal("expected error for non-200 response, got nil")
	}
}

func TestNewMemDBArchiver_ConnectionError(t *testing.T) {
	fn := archive.NewMemDBArchiver(archive.Config{
		BaseURL: "http://127.0.0.1:19999", // nothing listening
	})

	if err := fn(makeSession("k")); err == nil {
		t.Fatal("expected error for connection refused, got nil")
	}
}

func TestNewMemDBArchiver_DefaultConfig(t *testing.T) {
	var payload map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		_ = json.Unmarshal(body, &payload)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"3","status":"ok"}`))
	}))
	defer srv.Close()

	// Provide only BaseURL so we can hit the test server; all other fields default.
	fn := archive.NewMemDBArchiver(archive.Config{
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
	})

	if err := fn(makeSession("default-test")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payload["user_id"] != "memos" {
		t.Errorf("default user_id = %v, want memos", payload["user_id"])
	}
	if payload["cube_id"] != "memos" {
		t.Errorf("default cube_id = %v, want memos", payload["cube_id"])
	}

	meta, _ := payload["metadata"].(map[string]any)
	if meta["source"] != "go-session" {
		t.Errorf("metadata.source = %v, want go-session", meta["source"])
	}
	if meta["session_key"] != "default-test" {
		t.Errorf("metadata.session_key = %v, want default-test", meta["session_key"])
	}
	if _, ok := meta["archived_at"]; !ok {
		t.Error("metadata.archived_at missing")
	}
}
