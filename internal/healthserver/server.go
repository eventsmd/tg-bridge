package healthserver

import (
	"context"
	"net/http"
	"sync/atomic"
)

type Server struct {
	addr       string
	httpServer *http.Server
	ready      atomic.Bool
}

func New(addr string) *Server {
	s := &Server{addr: addr}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) {
		if s.ready.Load() {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ready"))
			return
		}
		http.Error(w, "not ready", http.StatusServiceUnavailable)
	})

	s.httpServer = &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}
	return s
}

func (s *Server) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) SetReady(v bool) {
	s.ready.Store(v)
}

func (s *Server) Addr() string {
	return s.addr
}
