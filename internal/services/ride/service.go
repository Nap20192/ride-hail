package ride

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ride-hail/internal/shared/core"
	"ride-hail/pkg/geo"
	"ride-hail/pkg/mq"
	"ride-hail/pkg/sqlc"
	"ride-hail/pkg/uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RideEventPublisher = mq.RideEventPublisher

type RideService struct {
	queries   *sqlc.Queries
	db        *pgxpool.Pool
	publisher *RideEventPublisher
}

func NewRideService(db *pgxpool.Pool, queries *sqlc.Queries, publisher *RideEventPublisher) *RideService {
	return &RideService{
		db:        db,
		queries:   queries,
		publisher: publisher,
	}
}

type RideRate struct {
	Base   float64
	PerKm  float64
	PerMin float64
}

var Rates = map[string]RideRate{
	"ECONOMY": {500, 100, 50},
	"PREMIUM": {800, 120, 60},
	"XL":      {1000, 150, 75},
}

func (r *RideService) calculateFare(rideType string, km float64, minutes int) float64 {
	rate := Rates[rideType]
	return rate.Base +
		(km * rate.PerKm) +
		(float64(minutes) * rate.PerMin)
}

type CreateRideResponse struct {
	RideID                   uuid.UUID `json:"ride_id"`
	RideNumber               string    `json:"ride_number"`
	Status                   string    `json:"status"`
	EstimatedFare            float64   `json:"estimated_fare"`
	EstimatedDurationMinutes int       `json:"estimated_duration_minutes"`
	EstimatedDistanceKm      float64   `json:"estimated_distance_km"`
}

func (s *RideService) CreateRide(ctx context.Context, req CreateRideRequest) (CreateRideResponse, error) {
	distanceKm := geo.Distance(req.PickupLat, req.PickupLng, req.DestLat, req.DestLng)
	durationMin := int(distanceKm * 3)

	fare := s.calculateFare(req.VehicleType, distanceKm, durationMin)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return CreateRideResponse{}, err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	qTx := s.queries.WithTx(tx)

	counter, err := qTx.IncrementRideCounter(ctx, time.Now().UTC())
	ride_number := s.generateRideNumber(time.Now().UTC(), int(counter.Counter))

	if err != nil {
		return CreateRideResponse{}, fmt.Errorf("failed to increment counter: %w", err)
	}

	pickup, err := qTx.CreateCoordinate(ctx, sqlc.CreateCoordinateParams{
		EntityID:        req.PassengerID,
		EntityType:      core.UserRolePassenger.String(),
		Address:         req.PickupAddress,
		Latitude:        sqlc.NumericFromFloat(req.PickupLat),
		Longitude:       sqlc.NumericFromFloat(req.PickupLng),
		DistanceKm:      sqlc.NumericFromFloat(distanceKm),
		DurationMinutes: sqlc.Int4FromInt32(int32(durationMin)),
		FareAmount:      sqlc.NumericFromFloat(fare),
	})
	if err != nil {
		return CreateRideResponse{}, err
	}

	destination, err := qTx.CreateCoordinate(ctx, sqlc.CreateCoordinateParams{
		EntityID:   req.PassengerID,
		EntityType: core.UserRolePassenger.String(),
		Address:    req.DestAddress,
		Latitude:   sqlc.NumericFromFloat(req.DestLat),
		Longitude:  sqlc.NumericFromFloat(req.DestLng),
	})
	if err != nil {
		return CreateRideResponse{}, err
	}

	ride, err := qTx.CreateRide(ctx, sqlc.CreateRideParams{
		RideNumber:              ride_number,
		PassengerID:             req.PassengerID,
		VehicleType:             &req.VehicleType,
		EstimatedFare:           sqlc.NumericFromFloat(fare),
		PickupCoordinateID:      pickup.ID,
		DestinationCoordinateID: destination.ID,
	})
	if err != nil {
		return CreateRideResponse{}, err
	}

	err = qTx.CreateRideEvent(ctx, sqlc.CreateRideEventParams{
		RideID:    ride.ID,
		EventType: core.RideEventRequested.String(),
		EventData: json.RawMessage(`{"status":"REQUESTED"}`),
	})
	if err != nil {
		return CreateRideResponse{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return CreateRideResponse{}, err
	}

	// Publish ride request event to RabbitMQ for driver matching
	if s.publisher != nil {
		rideRequestMsg := map[string]interface{}{
			"ride_id":        ride.ID.String(),
			"ride_number":    ride.RideNumber,
			"passenger_id":   req.PassengerID.String(),
			"vehicle_type":   req.VehicleType,
			"pickup_lat":     req.PickupLat,
			"pickup_lng":     req.PickupLng,
			"pickup_addr":    req.PickupAddress,
			"dest_lat":       req.DestLat,
			"dest_lng":       req.DestLng,
			"dest_addr":      req.DestAddress,
			"estimated_fare": fare,
			"requested_at":   time.Now().UTC(),
		}
		if pubErr := s.publisher.PublishRideRequest(ctx, req.VehicleType, rideRequestMsg); pubErr != nil {
			// Log error but don't fail the request - ride is already created
			fmt.Printf("Warning: failed to publish ride request event: %v\n", pubErr)
		}
	}

		// Start a 2-minute timeout to cancel the ride if still REQUESTED
		// Runs asynchronously; best-effort cleanup and notification
		go func(rideID uuid.UUID) {
			// wait for 2 minutes
			t := time.NewTimer(2 * time.Minute)
			defer t.Stop()
			select {
			case <-t.C:
				// after timeout, check ride status
				ctx2 := context.Background()
				r, err := s.queries.GetRideByID(ctx2, rideID)
				if err != nil {
					fmt.Printf("Warning: timeout check failed to get ride %s: %v\n", rideID.String(), err)
					return
				}
				if r.Status != nil && *r.Status == "REQUESTED" {
					reason := "NO_DRIVERS_AVAILABLE"
					cancelledRide, cerr := s.queries.CancelRide(ctx2, sqlc.CancelRideParams{ID: rideID, CancellationReason: &reason})
					if cerr != nil {
						fmt.Printf("Warning: failed to cancel ride %s after timeout: %v\n", rideID.String(), cerr)
						return
					}
					// record cancellation event
					_ = s.queries.CreateRideEvent(ctx2, sqlc.CreateRideEventParams{
						RideID:    rideID,
						EventType: core.RideEventCancelled.String(),
						EventData: json.RawMessage(fmt.Sprintf(`{"status":"CANCELLED","reason":"%s"}`, reason)),
					})
					// publish cancelled status
					if s.publisher != nil {
						cancelledMsg := map[string]interface{}{
							"ride_id":      cancelledRide.ID.String(),
							"ride_number":  cancelledRide.RideNumber,
							"passenger_id": cancelledRide.PassengerID.String(),
							"status":       "CANCELLED",
							"reason":       reason,
							"cancelled_at": time.Now().UTC(),
						}
						if pubErr := s.publisher.PublishRideStatus(ctx2, "CANCELLED", cancelledMsg); pubErr != nil {
							fmt.Printf("Warning: failed to publish ride cancelled event after timeout: %v\n", pubErr)
						}
					}
				}
			case <-ctx.Done():
				// request context cancelled; stop timer
				return
			}
		}(ride.ID)

	return CreateRideResponse{
		RideID:                   ride.ID,
		RideNumber:               ride.RideNumber,
		Status:                   *ride.Status,
		EstimatedFare:            fare,
		EstimatedDurationMinutes: durationMin,
		EstimatedDistanceKm:      distanceKm,
	}, nil
}

func (s *RideService) generateRideNumber(date time.Time, counter int) string {
	datePart := date.Format("20060102")
	seq := fmt.Sprintf("%03d", counter)
	return fmt.Sprintf("RIDE_%s_%s", datePart, seq)
}

func (s *RideService) CancelRide(ctx context.Context, rideID uuid.UUID, passengerID uuid.UUID, reason string) (CancelRideResponse, error) {
	// First, validate that the ride exists and belongs to the passenger
	ride, err := s.queries.GetRideByID(ctx, rideID)
	if err != nil {
		return CancelRideResponse{}, fmt.Errorf("ride not found: %w", err)
	}

	// Check if the ride belongs to the passenger
	if ride.PassengerID != passengerID {
		return CancelRideResponse{}, fmt.Errorf("unauthorized: ride does not belong to passenger")
	}

	// Check if ride is already cancelled or completed
	if ride.Status != nil {
		if *ride.Status == "CANCELLED" {
			return CancelRideResponse{}, fmt.Errorf("ride is already cancelled")
		}
		if *ride.Status == "COMPLETED" {
			return CancelRideResponse{}, fmt.Errorf("cannot cancel completed ride")
		}
	}

	// Begin transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return CancelRideResponse{}, err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	qTx := s.queries.WithTx(tx)

	// Update ride status to CANCELLED
	cancelledRide, err := qTx.CancelRide(ctx, sqlc.CancelRideParams{
		ID:                 rideID,
		CancellationReason: &reason,
	})
	if err != nil {
		return CancelRideResponse{}, fmt.Errorf("failed to cancel ride: %w", err)
	}

	// Record cancellation event
	err = qTx.CreateRideEvent(ctx, sqlc.CreateRideEventParams{
		RideID:    rideID,
		EventType: core.RideEventCancelled.String(),
		EventData: json.RawMessage(fmt.Sprintf(`{"status":"CANCELLED","reason":"%s"}`, reason)),
	})
	if err != nil {
		return CancelRideResponse{}, fmt.Errorf("failed to create ride event: %w", err)
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		return CancelRideResponse{}, err
	}

	// Publish ride cancelled event to RabbitMQ
	if s.publisher != nil {
		cancelledMsg := map[string]interface{}{
			"ride_id":         rideID.String(),
			"ride_number":     cancelledRide.RideNumber,
			"passenger_id":    passengerID.String(),
			"status":          "CANCELLED",
			"reason":          reason,
			"cancelled_at":    time.Now().UTC(),
			"previous_status": *ride.Status,
		}
		if pubErr := s.publisher.PublishRideStatus(ctx, "CANCELLED", cancelledMsg); pubErr != nil {
			// Log error but don't fail the request - ride is already cancelled
			fmt.Printf("Warning: failed to publish ride cancelled event: %v\n", pubErr)
		}
	}

	var cancelledAt time.Time
	if cancelledRide.CancelledAt != nil {
		cancelledAt = *cancelledRide.CancelledAt
	} else {
		cancelledAt = time.Now().UTC()
	}

	return CancelRideResponse{
		RideID:      rideID,
		Status:      "CANCELLED",
		CancelledAt: cancelledAt,
		Message:     "Ride cancelled successfully",
	}, nil
}
