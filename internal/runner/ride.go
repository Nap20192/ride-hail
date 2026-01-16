package runner

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ride-hail/internal/deps"
	"ride-hail/internal/shared/config"
	"ride-hail/pkg/group"
)

// RideRun initializes and runs the Ride service.
func RideRun(ctx context.Context, config config.Config) error {
	infra, err := deps.NewInfraDeps(
		deps.WithRabbit(ctx, config),
		deps.WithPostgres(ctx, config),
	)
	if err != nil {
		return err
	}
	app, err := deps.NewAppDeps(
		deps.WithAuthService(infra),
		deps.WithRideService(infra),
	)
	if err != nil {
		return err
	}

	api, err := deps.NewApiDeps(
		deps.WithRideApi(app, config),
	)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	g, gCtx := group.WithContext(ctx)

	g.Go(func() error {
		if err := api.RideApi.Start(); err != nil && err != http.ErrServerClosed {
			slog.Error("Ride API server error", slog.String("error", err.Error()))
			return err
		}
		return nil
	})

	g.Go(func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(c)

		select {
		case <-gCtx.Done():
			return gCtx.Err()
		case sig := <-c:
			slog.Info("shutdown signal received", slog.String("signal", sig.String()))
			cancel()
			return nil
		}
	})

	g.Go(func() error {
		<-gCtx.Done()

		ctxWithTimeout, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelShutdown()
		err := api.RideApi.StopServer(ctxWithTimeout)
		if err != nil {
			slog.Error("failed to stop Ride API server", slog.String("error", err.Error()))
			return err
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}
