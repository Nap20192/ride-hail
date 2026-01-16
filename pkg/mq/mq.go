package mq

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"ride-hail/pkg/utils"
)

type Config struct {
	Port     string
	Host     string
	UserName string
	Password string
}

func LoadMqConfig() (Config, error) {
	return Config{
		Port:     utils.GetEnv("RABBITMQ_PORT", "5672"),
		Host:     utils.GetEnv("RABBITMQ_HOST", "localhost"),
		UserName: utils.GetEnv("RABBITMQ_USER", "guest"),
		Password: utils.GetEnv("RABBITMQ_PASSWORD", "guest"),
	}, nil
}

func (c Config) Url() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/", c.UserName, c.Password, c.Host, c.Port)
}

type Client struct {
	conn            *amqp.Connection
	ch              *amqp.Channel
	connectionReady chan struct{}
	closeChan       chan *amqp.Error
	config          Config
	closeMutex      sync.RWMutex
	reconnectMutex  sync.Mutex
	isReconnecting  bool
	closed          bool
}

func Connect(config Config) (*amqp.Connection, error) {
	return amqp.Dial(config.Url())
}

func NewClient(conn *amqp.Connection) (*Client, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	client := &Client{
		conn:            conn,
		ch:              ch,
		connectionReady: make(chan struct{}, 1),
		closeChan:       make(chan *amqp.Error, 1),
	}
	// Signal that connection is ready
	client.connectionReady <- struct{}{}
	// Start monitoring connection
	go client.monitorConnection()
	return client, nil
}

// NewClientWithReconnect creates a client with automatic reconnection
func NewClientWithReconnect(config Config) (*Client, error) {
	conn, err := Connect(config)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	client := &Client{
		conn:            conn,
		ch:              ch,
		config:          config,
		connectionReady: make(chan struct{}, 1),
		closeChan:       make(chan *amqp.Error, 1),
	}
	// Signal that connection is ready
	client.connectionReady <- struct{}{}
	// Start monitoring connection
	go client.monitorConnection()
	return client, nil
}

// monitorConnection watches for connection failures and triggers reconnection
func (c *Client) monitorConnection() {
	c.closeChan = c.conn.NotifyClose(make(chan *amqp.Error, 1))
	for {
		err, ok := <-c.closeChan
		if !ok {
			return
		}
		c.closeMutex.RLock()
		if c.closed {
			c.closeMutex.RUnlock()
			return
		}
		c.closeMutex.RUnlock()
		slog.Warn("RabbitMQ connection lost", "error", err)
		c.reconnect()
	}
}

// reconnect attempts to reconnect with exponential backoff
func (c *Client) reconnect() {
	c.reconnectMutex.Lock()
	if c.isReconnecting {
		c.reconnectMutex.Unlock()
		return
	}
	c.isReconnecting = true
	c.reconnectMutex.Unlock()
	defer func() {
		c.reconnectMutex.Lock()
		c.isReconnecting = false
		c.reconnectMutex.Unlock()
	}()
	maxRetries := 10
	baseDelay := 1 * time.Second
	maxDelay := 30 * time.Second
	for attempt := 1; attempt <= maxRetries; attempt++ {
		c.closeMutex.RLock()
		if c.closed {
			c.closeMutex.RUnlock()
			return
		}
		c.closeMutex.RUnlock()
		slog.Info("Attempting to reconnect to RabbitMQ", "attempt", attempt, "max", maxRetries)
		conn, err := Connect(c.config)
		if err != nil {
			delay := time.Duration(math.Min(float64(baseDelay)*math.Pow(2, float64(attempt-1)), float64(maxDelay)))
			slog.Warn("Reconnection failed, retrying", "attempt", attempt, "delay", delay, "error", err)
			time.Sleep(delay)
			continue
		}
		ch, err := conn.Channel()
		if err != nil {
			conn.Close()
			delay := time.Duration(math.Min(float64(baseDelay)*math.Pow(2, float64(attempt-1)), float64(maxDelay)))
			slog.Warn("Failed to create channel, retrying", "attempt", attempt, "delay", delay, "error", err)
			time.Sleep(delay)
			continue
		}
		// Update client state
		c.conn = conn
		c.ch = ch
		// Signal reconnection success
		select {
		case c.connectionReady <- struct{}{}:
		default:
		}
		slog.Info("Successfully reconnected to RabbitMQ")
		// Restart connection monitoring
		go c.monitorConnection()
		return
	}
	slog.Error("Failed to reconnect to RabbitMQ after maximum retries")
}

// waitForConnection waits until connection is ready
func (c *Client) waitForConnection(ctx context.Context) error {
	select {
	case <-c.connectionReady:
		c.connectionReady <- struct{}{} // Put it back for next call
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Client) Close() error {
	c.closeMutex.Lock()
	c.closed = true
	c.closeMutex.Unlock()
	if err := c.ch.Close(); err != nil {
		return err
	}
	return c.conn.Close()
}

func (c *Client) CreateQueue(name string, durable, autoDelete bool) error {
	_, err := c.ch.QueueDeclare(name, durable, autoDelete, false, false, nil)
	return err
}

func (c *Client) CreateQueueWithArgs(name string, durable, autoDelete bool, args amqp.Table) error {
	_, err := c.ch.QueueDeclare(name, durable, autoDelete, false, false, args)
	return err
}

func (c *Client) CreateBinding(name, binding, exchange string) error {
	return c.ch.QueueBind(name, binding, exchange, false, nil)
}

func (c *Client) Send(ctx context.Context, exchange, routingKey string, optinions amqp.Publishing) error {
	if err := c.waitForConnection(ctx); err != nil {
		return fmt.Errorf("connection not ready: %w", err)
	}
	return c.ch.PublishWithContext(ctx, exchange, routingKey, true, false, optinions)
}

func (c *Client) Consume(queue, consumer string, autoAck bool) (<-chan amqp.Delivery, error) {
	return c.ch.Consume(queue, consumer, autoAck, false, false, false, nil)
}

func (c *Client) CreateExchange(name, kind string, durable, autoDelete bool) error {
	return c.ch.ExchangeDeclare(
		name,
		kind,
		durable,
		autoDelete,
		false,
		false,
		nil,
	)
}

func (c *Client) SetQos(prefetchCount int) error {
	return c.ch.Qos(prefetchCount, 0, false)
}

// Ack acknowledges a message
func (c *Client) Ack(tag uint64, multiple bool) error {
	return c.ch.Ack(tag, multiple)
}

// Nack negatively acknowledges a message
func (c *Client) Nack(tag uint64, multiple, requeue bool) error {
	return c.ch.Nack(tag, multiple, requeue)
}

// Reject rejects a message
func (c *Client) Reject(tag uint64, requeue bool) error {
	return c.ch.Reject(tag, requeue)
}

// IsConnected returns whether the client is connected
func (c *Client) IsConnected() bool {
	c.closeMutex.RLock()
	defer c.closeMutex.RUnlock()
	return !c.closed && c.conn != nil && !c.conn.IsClosed()
}
