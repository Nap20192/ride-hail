package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MessageHandler func(ctx context.Context, message Message) error

type Message struct {
	CorrelationID string
	MessageID     string
	Timestamp     time.Time
	Body          []byte
	Delivery      amqp.Delivery
}

func (m *Message) ParseJSON(target interface{}) error {
	return json.Unmarshal(m.Body, target)
}

type Consumer interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type MessageConsumer struct {
	client        *Client
	handler       MessageHandler
	stopChan      chan struct{}
	queue         string
	consumerTag   string
	wg            sync.WaitGroup
	prefetchCount int
	maxRetries    int
	workers       int
	once          sync.Once
}

type ConsumerConfig struct {
	Handler       MessageHandler
	Queue         string
	ConsumerTag   string
	PrefetchCount int // Number of messages to prefetch (default: 10)
	MaxRetries    int // Maximum number of retry attempts (default: 3)
	Workers       int // Number of concurrent workers (default: 5)
}

func NewConsumer(client *Client, config ConsumerConfig) *MessageConsumer {
	if config.PrefetchCount == 0 {
		config.PrefetchCount = 10
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.Workers == 0 {
		config.Workers = 5
	}

	return &MessageConsumer{
		client:        client,
		queue:         config.Queue,
		consumerTag:   config.ConsumerTag,
		handler:       config.Handler,
		prefetchCount: config.PrefetchCount,
		maxRetries:    config.MaxRetries,
		workers:       config.Workers,
		stopChan:      make(chan struct{}),
	}
}

func (c *MessageConsumer) Start(ctx context.Context) error {
	if err := c.client.SetQos(c.prefetchCount); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}
	deliveries, err := c.client.Consume(c.queue, c.consumerTag, false)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	slog.Info("Consumer started",
		"queue", c.queue,
		"consumer_tag", c.consumerTag,
		"prefetch_count", c.prefetchCount,
		"workers", c.workers,
	)

	for i := 0; i < c.workers; i++ {
		c.wg.Add(1)
		go c.worker(ctx, i, deliveries)
	}

	return nil
}

// processes messages from the delivery channel
func (c *MessageConsumer) worker(ctx context.Context, workerID int, deliveries <-chan amqp.Delivery) {
	defer c.wg.Done()

	slog.Debug("Consumer worker started", "worker_id", workerID, "queue", c.queue)

	for {
		select {
		case <-ctx.Done():
			slog.Debug("Consumer worker stopped (context done)", "worker_id", workerID)
			return
		case <-c.stopChan:
			slog.Debug("Consumer worker stopped (stop signal)", "worker_id", workerID)
			return
		case delivery, ok := <-deliveries:
			if !ok {
				slog.Warn("Delivery channel closed", "worker_id", workerID, "queue", c.queue)
				return
			}

			c.processMessage(ctx, workerID, delivery)
		}
	}
}

func (c *MessageConsumer) processMessage(ctx context.Context, workerID int, delivery amqp.Delivery) {
	message := Message{
		CorrelationID: delivery.CorrelationId,
		MessageID:     delivery.MessageId,
		Timestamp:     delivery.Timestamp,
		Body:          delivery.Body,
		Delivery:      delivery,
	}

	startTime := time.Now()
	retryCount := 0

	if delivery.Headers != nil {
		if val, ok := delivery.Headers["x-retry-count"]; ok {
			if count, ok := val.(int32); ok {
				retryCount = int(count)
			}
		}
	}

	err := c.handler(ctx, message)

	duration := time.Since(startTime)
	logAttrs := []interface{}{
		"worker_id", workerID,
		"queue", c.queue,
		"correlation_id", message.CorrelationID,
		"message_id", message.MessageID,
		"duration_ms", duration.Milliseconds(),
		"retry_count", retryCount,
	}

	if err != nil {
		slog.Error("Message processing failed", append(logAttrs, "error", err)...)

		if retryCount < c.maxRetries {
			// retry by republishing with incremented retry count
			if republishErr := c.republishWithRetry(ctx, delivery, retryCount+1); republishErr != nil {
				slog.Error("Failed to republish message for retry", "error", republishErr)
				// fallback: nack without requeue
				if nackErr := c.client.Nack(delivery.DeliveryTag, false, false); nackErr != nil {
					slog.Error("Failed to nack message", "error", nackErr)
				}
			} else {
				// ack the original message
				if ackErr := c.client.Ack(delivery.DeliveryTag, false); ackErr != nil {
					slog.Error("Failed to ack message after republish", "error", ackErr)
				} else {
					slog.Info("Message republished for retry",
						"correlation_id", message.CorrelationID,
						"retry_count", retryCount+1,
						"max_retries", c.maxRetries,
					)
				}
			}
		} else {
			// max retries exceeded, send to dead letter queue
			if nackErr := c.client.Nack(delivery.DeliveryTag, false, false); nackErr != nil {
				slog.Error("Failed to nack message (send to DLQ)", "error", nackErr)
			} else {
				slog.Warn("Message sent to dead letter queue (max retries exceeded)",
					"correlation_id", message.CorrelationID,
					"retry_count", retryCount,
				)
			}
		}
	} else {
		slog.Debug("Message processed successfully", logAttrs...)

		if ackErr := c.client.Ack(delivery.DeliveryTag, false); ackErr != nil {
			slog.Error("Failed to ack message", "error", ackErr)
		}
	}
}

// republishWithRetry republishes a failed message with an incremented retry count
func (c *MessageConsumer) republishWithRetry(ctx context.Context, delivery amqp.Delivery, newRetryCount int) error {
	headers := amqp.Table{}
	if delivery.Headers != nil {
		for k, v := range delivery.Headers {
			headers[k] = v
		}
	}

	headers["x-retry-count"] = int32(newRetryCount)

	// add delay before retry
	delayMs := int32(100 * (1 << uint(newRetryCount-1)))
	if delayMs > 5000 {
		delayMs = 5000
	}

	publishing := amqp.Publishing{
		Headers:       headers,
		ContentType:   delivery.ContentType,
		CorrelationId: delivery.CorrelationId,
		MessageId:     delivery.MessageId,
		Timestamp:     delivery.Timestamp,
		DeliveryMode:  delivery.DeliveryMode,
		Body:          delivery.Body,
	}

	time.Sleep(time.Duration(delayMs) * time.Millisecond)

	// republish to the same queue via default exchange
	return c.client.Send(ctx, "", c.queue, publishing)
}

func (c *MessageConsumer) Stop(ctx context.Context) error {
	var err error
	c.once.Do(func() {
		slog.Info("Stopping consumer", "queue", c.queue, "consumer_tag", c.consumerTag)
		close(c.stopChan)

		done := make(chan struct{})
		go func() {
			c.wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			slog.Info("Consumer stopped gracefully", "queue", c.queue)
		case <-ctx.Done():
			slog.Warn("Consumer stop timeout exceeded", "queue", c.queue)
			err = ctx.Err()
		}
	})

	return err
}

// manages multiple consumers
type ConsumerGroup struct {
	consumers []Consumer
	wg        sync.WaitGroup
}

func NewConsumerGroup() *ConsumerGroup {
	return &ConsumerGroup{
		consumers: make([]Consumer, 0),
	}
}

func (g *ConsumerGroup) Add(consumer Consumer) {
	g.consumers = append(g.consumers, consumer)
}

func (g *ConsumerGroup) StartAll(ctx context.Context) error {
	for i, consumer := range g.consumers {
		g.wg.Add(1)
		go func(idx int, c Consumer) {
			defer g.wg.Done()
			if err := c.Start(ctx); err != nil {
				slog.Error("Consumer failed to start", "index", idx, "error", err)
			}
		}(i, consumer)
	}

	return nil
}

func (g *ConsumerGroup) StopAll(ctx context.Context) error {
	slog.Info("Stopping consumer group", "count", len(g.consumers))

	for i, consumer := range g.consumers {
		if err := consumer.Stop(ctx); err != nil {
			slog.Error("Failed to stop consumer", "index", i, "error", err)
		}
	}

	done := make(chan struct{})
	go func() {
		g.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		slog.Info("All consumers stopped")
		return nil
	case <-ctx.Done():
		slog.Warn("Consumer group stop timeout exceeded")
		return ctx.Err()
	}
}
