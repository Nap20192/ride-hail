package initmq

import (
	"fmt"
	"log/slog"
	"os"

	"ride-hail/pkg/logger"
	"ride-hail/pkg/mq"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	if _, err := logger.InitLogger("INFO", false); err != nil {
		slog.Error("Failed to initialize logger", "error", err)
		os.Exit(1)
	}

	slog.Info("Starting RabbitMQ")

	mqConfig, err := mq.LoadMqConfig()
	if err != nil {
		slog.Error("Failed to load MQ config", "error", err)
		os.Exit(1)
	}

	conn, err := mq.Connect(mqConfig)
	if err != nil {
		slog.Error("Failed to connect to RabbitMQ", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	client, err := mq.NewClient(conn)
	if err != nil {
		slog.Error("Failed to create MQ client", "error", err)
		os.Exit(1)
	}
	defer client.Close()

	if err := initializeTopology(client); err != nil {
		slog.Error("Failed to initialize topology", "error", err)
		os.Exit(1)
	}

	slog.Info("RabbitMQ topology initialized successfully")
}

func initializeTopology(client *mq.Client) error {
	slog.Info("Creating exchanges...")

	if err := client.CreateExchange("ride_topic", "topic", true, false); err != nil {
		return fmt.Errorf("failed to create ride_topic exchange: %w", err)
	}
	slog.Info("Created exchange", "name", "ride_topic", "type", "topic")

	if err := client.CreateExchange("driver_topic", "topic", true, false); err != nil {
		return fmt.Errorf("failed to create driver_topic exchange: %w", err)
	}
	slog.Info("Created exchange", "name", "driver_topic", "type", "topic")

	if err := client.CreateExchange("location_fanout", "fanout", true, false); err != nil {
		return fmt.Errorf("failed to create location_fanout exchange: %w", err)
	}
	slog.Info("Created exchange", "name", "location_fanout", "type", "fanout")

	if err := client.CreateExchange("dlx", "topic", true, false); err != nil {
		return fmt.Errorf("failed to create dlx exchange: %w", err)
	}
	slog.Info("Created exchange", "name", "dlx", "type", "topic")

	slog.Info("Creating queues...")

	dlxArgs := amqp.Table{
		"x-dead-letter-exchange": "dlx",
	}

	if err := client.CreateQueueWithArgs("ride_requests", true, false, dlxArgs); err != nil {
		return fmt.Errorf("failed to create ride_requests queue: %w", err)
	}
	slog.Info("Created queue", "name", "ride_requests")

	if err := client.CreateQueueWithArgs("ride_status", true, false, dlxArgs); err != nil {
		return fmt.Errorf("failed to create ride_status queue: %w", err)
	}
	slog.Info("Created queue", "name", "ride_status")

	if err := client.CreateQueueWithArgs("driver_matching", true, false, dlxArgs); err != nil {
		return fmt.Errorf("failed to create driver_matching queue: %w", err)
	}
	slog.Info("Created queue", "name", "driver_matching")

	if err := client.CreateQueueWithArgs("driver_responses", true, false, dlxArgs); err != nil {
		return fmt.Errorf("failed to create driver_responses queue: %w", err)
	}
	slog.Info("Created queue", "name", "driver_responses")

	if err := client.CreateQueueWithArgs("driver_status", true, false, dlxArgs); err != nil {
		return fmt.Errorf("failed to create driver_status queue: %w", err)
	}
	slog.Info("Created queue", "name", "driver_status")

	if err := client.CreateQueueWithArgs("location_updates_ride", true, false, dlxArgs); err != nil {
		return fmt.Errorf("failed to create location_updates_ride queue: %w", err)
	}
	slog.Info("Created queue", "name", "location_updates_ride")

	if err := client.CreateQueueWithArgs("location_updates_admin", true, false, dlxArgs); err != nil {
		return fmt.Errorf("failed to create location_updates_admin queue: %w", err)
	}
	slog.Info("Created queue", "name", "location_updates_admin")

	if err := client.CreateQueue("dead_letters", true, false); err != nil {
		return fmt.Errorf("failed to create dead_letters queue: %w", err)
	}
	slog.Info("Created queue", "name", "dead_letters")

	slog.Info("Creating queue bindings...")

	if err := client.CreateBinding("ride_requests", "ride.request.*", "ride_topic"); err != nil {
		return fmt.Errorf("failed to bind ride_requests: %w", err)
	}
	slog.Info("Created binding", "queue", "ride_requests", "exchange", "ride_topic", "key", "ride.request.*")

	if err := client.CreateBinding("ride_status", "ride.status.*", "ride_topic"); err != nil {
		return fmt.Errorf("failed to bind ride_status: %w", err)
	}
	slog.Info("Created binding", "queue", "ride_status", "exchange", "ride_topic", "key", "ride.status.*")

	if err := client.CreateBinding("driver_matching", "ride.request.*", "ride_topic"); err != nil {
		return fmt.Errorf("failed to bind driver_matching: %w", err)
	}
	slog.Info("Created binding", "queue", "driver_matching", "exchange", "ride_topic", "key", "ride.request.*")

	if err := client.CreateBinding("driver_responses", "driver.response.*", "driver_topic"); err != nil {
		return fmt.Errorf("failed to bind driver_responses: %w", err)
	}
	slog.Info("Created binding", "queue", "driver_responses", "exchange", "driver_topic", "key", "driver.response.*")

	if err := client.CreateBinding("driver_status", "driver.status.*", "driver_topic"); err != nil {
		return fmt.Errorf("failed to bind driver_status: %w", err)
	}
	slog.Info("Created binding", "queue", "driver_status", "exchange", "driver_topic", "key", "driver.status.*")

	if err := client.CreateBinding("location_updates_ride", "", "location_fanout"); err != nil {
		return fmt.Errorf("failed to bind location_updates_ride: %w", err)
	}
	slog.Info("Created binding", "queue", "location_updates_ride", "exchange", "location_fanout")

	if err := client.CreateBinding("location_updates_admin", "", "location_fanout"); err != nil {
		return fmt.Errorf("failed to bind location_updates_admin: %w", err)
	}
	slog.Info("Created binding", "queue", "location_updates_admin", "exchange", "location_fanout")

	if err := client.CreateBinding("dead_letters", "#", "dlx"); err != nil {
		return fmt.Errorf("failed to bind dead_letters: %w", err)
	}
	slog.Info("Created binding", "queue", "dead_letters", "exchange", "dlx", "key", "#")

	return nil
}
