package deps

import (
	"context"
	"fmt"
	"time"

	"ride-hail/internal/shared/config"
	"ride-hail/pkg/mq"

	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
)

type InfraDeps struct {
	Pool     *pgxpool.Pool
	RabbitMQ *mq.Client
}

type infraOption func(*InfraDeps) error

func NewInfraDeps(opts ...infraOption) (*InfraDeps, error) {
	deps := &InfraDeps{}

	for _, opt := range opts {
		if err := opt(deps); err != nil {
			return nil, err
		}
	}

	return deps, nil
}

func WithPostgres(ctx context.Context, config config.Config) infraOption {
	return func(deps *InfraDeps) error {
		poolConfig, err := pgxpool.ParseConfig(
			config.DatabaseConnString(),
		)
		if err != nil {
			return err
		}

		poolConfig.MaxConns = 20
		poolConfig.MinConns = 5

		poolConfig.MaxConnLifetime = time.Hour
		poolConfig.MaxConnIdleTime = 30 * time.Minute

		poolConfig.HealthCheckPeriod = time.Minute
		poolConfig.ConnConfig.ConnectTimeout = 5 * time.Second

		pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
		if err != nil {
			return err
		}
		if err := pool.Ping(ctx); err != nil {
			return err
		}

		deps.Pool = pool

		return nil
	}
}

func WithRabbit(ctx context.Context, config config.Config) infraOption {
	return func(deps *InfraDeps) error {
		conn, err := amqp.DialConfig(

			config.RabbitMQURL(),

			amqp.Config{
				Heartbeat: 10 * time.Second,
				Locale:    "en_US",
				Dial:      amqp.DefaultDial(5 * time.Second),
			},
		)
		if err != nil {
			return fmt.Errorf("rabbit connect: %w", err)
		}

		// Wrap connection in mq.Client
		mqClient, err := mq.NewClient(conn)
		if err != nil {
			conn.Close()
			return fmt.Errorf("failed to create mq client: %w", err)
		}
		deps.RabbitMQ = mqClient

		return nil
	}
}

func CloseInfraDeps(deps *InfraDeps) error {
	if deps.Pool != nil {
		deps.Pool.Close()
	}

	if deps.RabbitMQ != nil {
		if err := deps.RabbitMQ.Close(); err != nil {
			return err
		}
	}

	return nil
}
