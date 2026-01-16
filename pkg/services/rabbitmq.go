package services

import (
	"context"
	"log/slog"

	"ride-hail/pkg/mq"
)

// RabbitMQService wraps RabbitMQ client as a Service with recconnection support
type RabbitMQService struct {
	name   string
	client *mq.Client
	config mq.Config
}

// NewRabbitMQService creates a new RabbitMQ service wrapper
func NewRabbitMQService(name string, client *mq.Client) *RabbitMQService {
	return &RabbitMQService{
		name:   name,
		client: client,
	}
}

// service with config for reconnection
func NewRabbitMQServiceWithConfig(name string, client *mq.Client, config mq.Config) *RabbitMQService {
	return &RabbitMQService{
		name:   name,
		client: client,
		config: config,
	}
}

func (s *RabbitMQService) Start(ctx context.Context) error {
	slog.Info("RabbitMQ service ready",
		"service", s.name,
		"connected", s.client.IsConnected(),
	)
	return nil
}

func (s *RabbitMQService) Stop(ctx context.Context) error {
	slog.Info("RabbitMQ service stopping", "service", s.name)
	return s.client.Close()
}

func (s *RabbitMQService) Name() string {
	return s.name
}

func (s *RabbitMQService) IsHealthy() bool {
	return s.client.IsConnected()
}
