package mq

import "time"

type RideRequestMessage struct {
	RequestedAt         time.Time           `json:"requested_at"`
	RideID              string              `json:"ride_id"`
	RideNumber          string              `json:"ride_number"`
	PassengerID         string              `json:"passenger_id"`
	VehicleType         string              `json:"vehicle_type"`
	PickupLocation      LocationCoordinates `json:"pickup_location"`
	DestinationLocation LocationCoordinates `json:"destination_location"`
	MaxDistanceKm       float64             `json:"max_distance_km"`
	TimeoutSeconds      int                 `json:"timeout_seconds"`
	CorrelationID       string              `json:"correlation_id"`
	EstimatedFare       float64             `json:"estimated_fare"`
}

type RideStatusMessage struct {
	UpdatedAt     time.Time              `json:"updated_at"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	RideID        string                 `json:"ride_id"`
	RideNumber    string                 `json:"ride_number"`
	PassengerID   string                 `json:"passenger_id"`
	DriverID      string                 `json:"driver_id,omitempty"`
	OldStatus     string                 `json:"old_status"`
	NewStatus     string                 `json:"new_status"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
}

type DriverResponseMessage struct {
	RespondedAt             time.Time            `json:"responded_at"`
	RideID                  string               `json:"ride_id"`
	DriverID                string               `json:"driver_id"`
	Reason                  string               `json:"reason,omitempty"`
	EstimatedArrivalMinutes int                  `json:"estimated_arrival_minutes,omitempty"`
	EstimatedArrival        time.Time            `json:"estimated_arrival,omitempty"`
	DriverLocation          *LocationCoordinates `json:"driver_location,omitempty"`
	DriverInfo              *DriverInfo          `json:"driver_info,omitempty"`
	CorrelationID           string               `json:"correlation_id,omitempty"`
	Accepted                bool                 `json:"accepted"`
}


type DriverStatusMessage struct {
	UpdatedAt time.Time              `json:"updated_at"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	DriverID  string                 `json:"driver_id"`
	OldStatus string                 `json:"old_status"`
	NewStatus string                 `json:"new_status"`
}

type LocationUpdateMessage struct {
	Timestamp  time.Time            `json:"timestamp"`
	EntityID   string               `json:"entity_id"`
	EntityType string               `json:"entity_type"` // "driver" or "passenger"
	RideID     string               `json:"ride_id,omitempty"`
	Location   *LocationCoordinates `json:"location"`
	Accuracy   float64              `json:"accuracy_meters,omitempty"`
	Speed      float64              `json:"speed_kmh,omitempty"`
	Heading    float64              `json:"heading_degrees,omitempty"`
}

type RideMatchedMessage struct {
	EstimatedArrival time.Time            `json:"estimated_arrival"`
	MatchedAt        time.Time            `json:"matched_at"`
	RideID           string               `json:"ride_id"`
	RideNumber       string               `json:"ride_number"`
	PassengerID      string               `json:"passenger_id"`
	DriverID         string               `json:"driver_id"`
	DriverName       string               `json:"driver_name"`
	VehicleType      string               `json:"vehicle_type"`
	VehicleMake      string               `json:"vehicle_make"`
	VehicleModel     string               `json:"vehicle_model"`
	VehicleColor     string               `json:"vehicle_color"`
	VehiclePlate     string               `json:"vehicle_plate"`
	DriverRating     float64              `json:"driver_rating"`
	DriverLocation   *LocationCoordinates `json:"driver_location,omitempty"`
	CorrelationID    string               `json:"correlation_id,omitempty"`
}

type RideCompletedMessage struct {
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	CompletedAt   time.Time `json:"completed_at"`
	RideID        string    `json:"ride_id"`
	RideNumber    string    `json:"ride_number"`
	PassengerID   string    `json:"passenger_id"`
	DriverID      string    `json:"driver_id"`
	Duration      int       `json:"duration_minutes"`
	Distance      float64   `json:"distance_km"`
	EstimatedFare float64   `json:"estimated_fare"`
	FinalFare     float64   `json:"final_fare"`
}

type RideCancelledMessage struct {
	CancelledAt     time.Time `json:"cancelled_at"`
	RideID          string    `json:"ride_id"`
	RideNumber      string    `json:"ride_number"`
	PassengerID     string    `json:"passenger_id"`
	DriverID        string    `json:"driver_id,omitempty"`
	CancelledBy     string    `json:"cancelled_by"`
	Reason          string    `json:"reason"`
	CancellationFee float64   `json:"cancellation_fee,omitempty"`
}

type LocationCoordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type DriverInfo struct {
	Name         string  `json:"name"`
	Rating       float64 `json:"rating"`
}
