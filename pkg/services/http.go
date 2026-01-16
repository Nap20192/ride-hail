package services

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"ride-hail/pkg/server"
)

// HTTPService wraps an HTTP server as a Service
type HTTPService struct {
	server *server.Server
	name   string
}

// NewHTTPService creates a new HTTP service wrapper
func NewHTTPService(name string, srv *server.Server) *HTTPService {
	return &HTTPService{
		name:   name,
		server: srv,
	}
}

func (s *HTTPService) Start(ctx context.Context) error {
	slog.Info("HTTP service starting", "service", s.name)
	if err := s.server.Start(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http server error: %w", err)
	}
	return nil
}

func (s *HTTPService) Stop(ctx context.Context) error {
	slog.Info("HTTP service stopping", "service", s.name)
	return s.server.Shutdown(ctx)
}

func (s *HTTPService) Name() string {
	return s.name
}
