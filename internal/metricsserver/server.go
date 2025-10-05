package metricsserver

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	addr       string
	httpServer *http.Server
	registry   *prometheus.Registry

	telegramMessages *prometheus.CounterVec
}

func New(addr string) *Server {
	mux := http.NewServeMux()

	// Build a registry with standard collectors to expose default Go/process metrics
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	// Business metric: count of telegram messages received per channel (username)
	telegramMessages := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telegram_channel_messages_total",
			Help: "Total number of messages received from Telegram channels, labeled by channel username.",
		},
		[]string{"channel"},
	)
	reg.MustRegister(telegramMessages)

	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	s := &Server{
		addr:             addr,
		registry:         reg,
		telegramMessages: telegramMessages,
		httpServer: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}
	return s
}

// AddTelegramChannelMessages increases the counter for a given channel by n.
// If n <= 0, the call is a no-op.
func (s *Server) AddTelegramChannelMessages(channel string, n int) {
	if n <= 0 {
		return
	}
	s.telegramMessages.WithLabelValues(channel).Add(float64(n))
}

func (s *Server) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) Addr() string {
	return s.addr
}
