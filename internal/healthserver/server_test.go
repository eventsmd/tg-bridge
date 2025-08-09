package healthserver

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	s := New(":0")
	ts := httptest.NewServer(s.httpServer.Handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health failed: %v", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "ok" {
		t.Fatalf(`expected body "ok", got %q`, string(body))
	}
}

func TestReadyEndpoint(t *testing.T) {
	s := New(":0")
	ts := httptest.NewServer(s.httpServer.Handler)
	defer ts.Close()

	// Initially not ready
	resp, err := http.Get(ts.URL + "/ready")
	if err != nil {
		t.Fatalf("GET /ready (initial) failed: %v", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503 when not ready, got %d", resp.StatusCode)
	}

	// Set ready
	s.SetReady(true)

	resp2, err := http.Get(ts.URL + "/ready")
	if err != nil {
		t.Fatalf("GET /ready (ready=true) failed: %v", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp2.Body)

	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 when ready, got %d", resp2.StatusCode)
	}
	body, _ := io.ReadAll(resp2.Body)
	if string(body) != "ready" {
		t.Fatalf(`expected body "ready", got %q`, string(body))
	}

	// Set not ready again
	s.SetReady(false)

	resp3, err := http.Get(ts.URL + "/ready")
	if err != nil {
		t.Fatalf("GET /ready (ready=false) failed: %v", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp3.Body)

	if resp3.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503 when ready=false, got %d", resp3.StatusCode)
	}
}
