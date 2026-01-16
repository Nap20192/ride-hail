package shutdown

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// GracefulShutdown handles graceful shutdown of services
type GracefulShutdown struct {
	handlers []func(context.Context) error
	signals  []os.Signal
	timeout  time.Duration
}

// New creates a new GracefulShutdown instance
func New(timeout time.Duration) *GracefulShutdown {
	return &GracefulShutdown{
		timeout:  timeout,
		handlers: make([]func(context.Context) error, 0),
		signals:  []os.Signal{syscall.SIGINT, syscall.SIGTERM},
	}
}

// Register adds a shutdown handler
func (g *GracefulShutdown) Register(handler func(context.Context) error) {
	g.handlers = append(g.handlers, handler)
}

// WithSignals sets custom signals to listen for
func (g *GracefulShutdown) WithSignals(signals ...os.Signal) *GracefulShutdown {
	g.signals = signals
	return g
}

// Wait blocks until a shutdown signal is received, then executes all handlers
func (g *GracefulShutdown) Wait() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, g.signals...)

	sig := <-sigChan
	slog.Info("Received shutdown signal", "signal", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()

	// Execute all shutdown handlers
	for i, handler := range g.handlers {
		if err := handler(ctx); err != nil {
			slog.Error("Shutdown handler failed", "index", i, "error", err)
		}
	}

	slog.Info("Graceful shutdown completed")
}

// WaitWithContext blocks until context is canceled or shutdown signal is received
func (g *GracefulShutdown) WaitWithContext(parentCtx context.Context) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, g.signals...)

	select {
	case sig := <-sigChan:
		slog.Info("Received shutdown signal", "signal", sig.String())
	case <-parentCtx.Done():
		slog.Info("Context canceled, initiating shutdown")
	}

	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()

	// Execute all shutdown handlers
	for i, handler := range g.handlers {
		if err := handler(ctx); err != nil {
			slog.Error("Shutdown handler failed", "index", i, "error", err)
		}
	}

	slog.Info("Graceful shutdown completed")
}
