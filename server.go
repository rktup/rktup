package rktup

import (
	"context"
	"log"
	"net/http"
)

type Server struct {
	config     *ServerConfig
	httpServer *http.Server
}

type ServerConfig struct {
	Addr        string
	Hostname    string
	GithubToken string
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriter{w, http.StatusOK}
		next.ServeHTTP(rw, r)
		log.Printf("%s %d %s?%s", r.Method, rw.status, r.URL.Path, r.URL.RawQuery)
	})
}

func NewServer(config *ServerConfig) *Server {
	handler := NewHTTPHandler(config.Hostname, config.GithubToken)
	httpServer := &http.Server{
		Addr:    config.Addr,
		Handler: middleware(handler),
	}
	return &Server{
		config:     config,
		httpServer: httpServer,
	}
}

func (s *Server) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
