package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"ride-hail/pkg/uuid"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher interface {
	Publish(ctx context.Context, exchange, routingKey string, message interface{}) error
	PublishWithCorrelationID(ctx context.Context, exchange, routingKey, correlationID string, message interface{}) error
}

type MessagePublisher struct {
	client *Client
}

func NewPublisher(client *Client) *MessagePublisher {
	return &MessagePublisher{
		client: client,
	}
}

// publishes a message to an exchange with auto-generated correlation ID
func (p *MessagePublisher) Publish(ctx context.Context, exchange, routingKey string, message interface{}) error {
	correlationID := uuid.New().String()
	return p.PublishWithCorrelationID(ctx, exchange, routingKey, correlationID, message)
}

// publishes a message with a specific correlation ID
func (p *MessagePublisher) PublishWithCorrelationID(ctx context.Context, exchange, routingKey, correlationID string, message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	publishing := amqp.Publishing{
		ContentType:   "application/json",
		CorrelationId: correlationID,
		MessageId:     uuid.New().String(),
		Timestamp:     time.Now(),
		DeliveryMode:  amqp.Persistent, // Persist messages to disk
		Body:          body,
	}

	if err := p.client.Send(ctx, exchange, routingKey, publishing); err != nil {
		slog.Error("Failed to publish message",
			"exchange", exchange,
			"routing_key", routingKey,
			"correlation_id", correlationID,
			"error", err,
		)
		return fmt.Errorf("failed to publish message: %w", err)
	}

	slog.Debug("Message published",
		"exchange", exchange,
		"routing_key", routingKey,
		"correlation_id", correlationID,
		"message_id", publishing.MessageId,
	)

	return nil
}

// publishes ride-related events
type RideEventPublisher struct {
	publisher *MessagePublisher
	exchange  string
}

func NewRideEventPublisher(client *Client) *RideEventPublisher {
	return &RideEventPublisher{
		publisher: NewPublisher(client),
		exchange:  "ride_topic",
	}
}

func (p *RideEventPublisher) PublishRideRequest(ctx context.Context, rideType string, message interface{}) error {
	routingKey := fmt.Sprintf("ride.request.%s", rideType)
	return p.publisher.Publish(ctx, p.exchange, routingKey, message)
}

func (p *RideEventPublisher) PublishRideStatus(ctx context.Context, status string, message interface{}) error {
	routingKey := fmt.Sprintf("ride.status.%s", status)
	return p.publisher.Publish(ctx, p.exchange, routingKey, message)
}

func (p *RideEventPublisher) PublishRideStatusWithCorrelation(ctx context.Context, status, correlationID string, message interface{}) error {
	routingKey := fmt.Sprintf("ride.status.%s", status)
	return p.publisher.PublishWithCorrelationID(ctx, p.exchange, routingKey, correlationID, message)
}

// PublishRideStatusUpdate publishes a ride status update with ride and driver information
func (p *RideEventPublisher) PublishRideStatusUpdate(ctx context.Context, rideID, status, driverID string) error {
	message := RideStatusMessage{
		UpdatedAt: time.Now(),
		RideID:    rideID,
		NewStatus: status,
		DriverID:  driverID,
	}
	return p.PublishRideStatus(ctx, status, message)
}

// PublishRideStatusUpdateWithCorrelation publishes a ride status update with correlation ID
func (p *RideEventPublisher) PublishRideStatusUpdateWithCorrelation(ctx context.Context, rideID, status, driverID, correlationID string) error {
	message := RideStatusMessage{
		UpdatedAt:     time.Now(),
		RideID:        rideID,
		NewStatus:     status,
		DriverID:      driverID,
		CorrelationID: correlationID,
	}
	return p.PublishRideStatusWithCorrelation(ctx, status, correlationID, message)
}

// publishes driver-related events
type DriverEventPublisher struct {
	publisher *MessagePublisher
	exchange  string
}

func NewDriverEventPublisher(client *Client) *DriverEventPublisher {
	return &DriverEventPublisher{
		publisher: NewPublisher(client),
		exchange:  "driver_topic",
	}
}

func (p *DriverEventPublisher) PublishDriverResponse(ctx context.Context, rideID string, message interface{}) error {
	routingKey := fmt.Sprintf("driver.response.%s", rideID)
	return p.publisher.Publish(ctx, p.exchange, routingKey, message)
}

func (p *DriverEventPublisher) PublishDriverStatus(ctx context.Context, driverID string, message interface{}) error {
	routingKey := fmt.Sprintf("driver.status.%s", driverID)
	return p.publisher.Publish(ctx, p.exchange, routingKey, message)
}

func (p *DriverEventPublisher) PublishDriverStatusWithCorrelation(ctx context.Context, driverID, correlationID string, message interface{}) error {
	routingKey := fmt.Sprintf("driver.status.%s", driverID)
	return p.publisher.PublishWithCorrelationID(ctx, p.exchange, routingKey, correlationID, message)
}

// publishes location update events
type LocationEventPublisher struct {
	publisher *MessagePublisher
	exchange  string
}

func NewLocationEventPublisher(client *Client) *LocationEventPublisher {
	return &LocationEventPublisher{
		publisher: NewPublisher(client),
		exchange:  "location_fanout",
	}
}

// publishes a location update (fanout exchange doesn't use routing keys)
func (p *LocationEventPublisher) PublishLocationUpdate(ctx context.Context, message interface{}) error {
	return p.publisher.Publish(ctx, p.exchange, "", message)
}

func (p *LocationEventPublisher) PublishLocationUpdateWithCorrelation(ctx context.Context, correlationID string, message interface{}) error {
	return p.publisher.PublishWithCorrelationID(ctx, p.exchange, "", correlationID, message)
}
