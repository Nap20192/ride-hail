package services

import (
	"context"
	"log/slog"

	"ride-hail/pkg/server"
)

// WebSocketService wraps WebSocket manager as a Service
type WebSocketService struct {
	manager *server.Manager
	name    string
}

// NewWebSocketService creates a new WebSocket service wrapper
func NewWebSocketService(name string, manager *server.Manager) *WebSocketService {
	return &WebSocketService{
		name:    name,
		manager: manager,
	}
}

func (s *WebSocketService) Start(ctx context.Context) error {
	slog.Info("WebSocket service ready", "service", s.name)
	return nil
}

func (s *WebSocketService) Stop(ctx context.Context) error {
	slog.Info("WebSocket service stopping", "service", s.name)
	s.manager.Shutdown()
	return nil
}

func (s *WebSocketService) Name() string {
	return s.name
}
