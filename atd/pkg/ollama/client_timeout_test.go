package ollama

// Regression tests for failures/20260916_atd_audit_workspace_no_return.md and
// failures/20260917_atd_audit_docs_exits_zero_with_no_report.md: generateHTTP
// and embedHTTP used to call http.Post with no deadline at all, so a slow or
// stalled Ollama backend blocked `atd audit` forever instead of failing
// within a bounded time. These tests pin the fix: a small configured timeout
// against a deliberately slow backend must return promptly with a clearly
// distinguishable timeout error, never hang.

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGenerateHTTP_BoundedByTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Write([]byte(`{"response":"too late"}`))
	}))
	defer srv.Close()

	start := time.Now()
	_, err := generateHTTP(srv.URL, "fake-model", "prompt", nil, nil, 100)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected a timeout error, got nil")
	}
	if elapsed > time.Second {
		t.Fatalf("generateHTTP did not respect the configured timeout: took %v, want well under 1s", elapsed)
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("error should clearly indicate a timeout, got: %v", err)
	}
}

func TestEmbedHTTP_BoundedByTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Write([]byte(`{"embedding":[0.1]}`))
	}))
	defer srv.Close()

	start := time.Now()
	_, err := embedHTTP(srv.URL, "fake-embed", "text", 100)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected a timeout error, got nil")
	}
	if elapsed > time.Second {
		t.Fatalf("embedHTTP did not respect the configured timeout: took %v, want well under 1s", elapsed)
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("error should clearly indicate a timeout, got: %v", err)
	}
}

func TestGenerateHTTP_NonTimeoutErrorIsNotMislabeled(t *testing.T) {
	// A backend that actively refuses (closed port, real srv.Close()'d
	// server) is a different failure mode than a timeout, and the error
	// message must not claim "timed out" for it.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()

	_, err := generateHTTP(srv.URL, "fake-model", "prompt", nil, nil, 5000)
	if err == nil {
		t.Fatal("expected a connection error, got nil")
	}
	if strings.Contains(err.Error(), "timed out") {
		t.Errorf("connection-refused error should not be mislabeled as a timeout: %v", err)
	}
}
