package metricsserver

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetricsEndpoint_ExposesStandardMetrics(t *testing.T) {
	s := New(":0")
	ts := httptest.NewServer(s.httpServer.Handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics failed: %v", err)
	}
	defer func(Body io.ReadCloser) { _ = Body.Close() }(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	text := string(body)

	// Verify that at least one of well-known default collectors' metrics is present.
	if !(strings.Contains(text, "go_goroutines") || strings.Contains(text, "process_cpu_seconds_total") || strings.Contains(text, "go_memstats_alloc_bytes")) {
		t.Fatalf("expected standard Prometheus metrics in body, got:\n%s", text)
	}

	// Basic sanity on content type (Prometheus text exposition format)
	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		t.Fatalf("expected Content-Type header to be set")
	}
}

func TestTelegramChannelMessagesCounter(t *testing.T) {
	s := New(":0")

	// Simulate receiving 2 messages for channel "testchannel"
	s.AddTelegramChannelMessages("testchannel", 2)

	ts := httptest.NewServer(s.httpServer.Handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics failed: %v", err)
	}
	defer func(Body io.ReadCloser) { _ = Body.Close() }(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	text := string(body)

	if !strings.Contains(text, "telegram_channel_messages_total{channel=\"testchannel\"} 2") {
		t.Fatalf("expected telegram_channel_messages_total counter with value 2 for channel label, got:\n%s", text)
	}
}
