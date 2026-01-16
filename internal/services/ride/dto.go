package ride

import (
	"fmt"
	"time"

	"ride-hail/internal/shared/core"
	"ride-hail/pkg/uuid"
)

// POST /rides
type CreateRideRequest struct {
	PassengerID   uuid.UUID `json:"passenger_id,omitempty"` // Set from JWT context, not request body
	PickupLat     float64   `json:"pickup_latitude"`
	PickupLng     float64   `json:"pickup_longitude"`
	PickupAddress string    `json:"pickup_address"`
	DestLat       float64   `json:"destination_latitude"`
	DestLng       float64   `json:"destination_longitude"`
	DestAddress   string    `json:"destination_address"`
	VehicleType   string    `json:"ride_type"`
}

func (r *CreateRideRequest) Validate() error {
	validTypes := map[string]bool{
		"ECONOMY": true,
		"PREMIUM": true,
		"XL":      true,
	}
	if !validTypes[r.VehicleType] {
		return fmt.Errorf("invalid vehicle_type: must be ECONOMY, PREMIUM, or XL")
	}

	if r.PickupLat < -90 || r.PickupLat > 90 {
		return fmt.Errorf("pickup_latitude must be between -90 and 90")
	}
	if r.PickupLng < -180 || r.PickupLng > 180 {
		return fmt.Errorf("pickup_longitude must be between -180 and 180")
	}
	if r.DestLat < -90 || r.DestLat > 90 {
		return fmt.Errorf("destination_latitude must be between -90 and 90")
	}
	if r.DestLng < -180 || r.DestLng > 180 {
		return fmt.Errorf("destination_longitude must be between -180 and 180")
	}

	if r.PickupAddress == "" {
		return fmt.Errorf("pickup_address is required")
	}
	if r.DestAddress == "" {
		return fmt.Errorf("destination_address is required")
	}

	if r.PickupLat == r.DestLat && r.PickupLng == r.DestLng {
		return fmt.Errorf("pickup and destination must be different")
	}

	return nil
}

// POST /rides/{id}/cancel
type CancelRideRequest struct {
	Reason string `json:"reason"`
}

func (r *CancelRideRequest) Validate() error {
	if r.Reason == "" {
		return fmt.Errorf("cancellation reason is required")
	}
	if len(r.Reason) > 500 {
		return fmt.Errorf("reason too long (max 500 characters)")
	}
	return nil
}

type RideResponse struct {
	ID                string    `json:"id"`
	RideNumber        string    `json:"ride_number"`
	Status            string    `json:"status"`
	VehicleType       string    `json:"vehicle_type,omitempty"`
	EstimatedFare     float64   `json:"estimated_fare,omitempty"`
	EstimatedDuration int       `json:"estimated_duration_minutes,omitempty"`
	EstimatedDistance float64   `json:"estimated_distance_km,omitempty"`
	RequestedAt       time.Time `json:"requested_at"`
	PickupLocation    Location  `json:"pickup_location,omitempty"`
	DestLocation      Location  `json:"destination_location,omitempty"`
}

type CancelRideResponse struct {
	RideID      uuid.UUID `json:"ride_id"`
	Status      string    `json:"status"`
	CancelledAt time.Time `json:"cancelled_at"`
	Message     string    `json:"message"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type CreateRideInput struct {
	PassengerID   uuid.UUID
	VehicleType   string
	PickupLat     float64
	PickupLng     float64
	PickupAddress string
	DestLat       float64
	DestLng       float64
	DestAddress   string
}

type CancelRideInput struct {
	RideID uuid.UUID
	Reason string
}

type RideOutput struct {
	ID                string
	RideNumber        string
	Status            core.RideStatus
	VehicleType       string
	EstimatedFare     float64
	EstimatedDuration int
	EstimatedDistance float64
	RequestedAt       time.Time
	PickupLat         float64
	PickupLng         float64
	PickupAddress     string
	DestLat           float64
	DestLng           float64
	DestAddress       string
}
