package models

import (
	"errors"
	"ride-hail/pkg/uuid"
	"time"
)

type OnlineRequest struct {
	DriverID  uuid.UUID
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func (r *OnlineRequest) Validate() error {
	if r.Latitude < -90 || r.Latitude > 90 {
		return errors.New("invalid latitude: must be between -90 and 90")
	}
	if r.Longitude < -180 || r.Longitude > 180 {
		return errors.New("invalid longitude: must be betwee -180 and 180")
	}
	return nil
}

type LocationUpdateRequest struct {
	DriverID       uuid.UUID  `json:"-"`
	RideID         *uuid.UUID `json:"ride_id,omitempty"`
	Latitude       float64    `json:"latitude"`
	Longitude      float64    `json:"longitude"`
	AccuracyMeters float64    `json:"accuracy_meters"`
	SpeedKmh       float64    `json:"speed_kmh"`
	HeadingDegrees float64    `json:"heading_degrees"`
}

func (r *LocationUpdateRequest) Validate() error {
	if r.Latitude < -90 || r.Latitude > 90 {
		return errors.New("invalid latitude: must be between -90 and 90")
	}
	if r.Longitude < -180 || r.Longitude > 180 {
		return errors.New("invalid longitude: must be between -180 and 180")
	}
	if r.AccuracyMeters < 0 {
		return errors.New("accuracy_meters cannot be negative")
	}
	if r.SpeedKmh < 0 {
		return errors.New("speed_kmh cannot be negative")
	}
	if r.HeadingDegrees < 0 || r.HeadingDegrees > 360 {
		return errors.New("heading_degrees must be between 0 and 360")
	}
	return nil
}

type StartRideRequest struct {
	RideID    string  `json:"ride_id"`
	Latitude  float64 `json:"driver_location.latitude"`
	Longitude float64 `json:"driver_location.longitude"`
}

func (r *StartRideRequest) Validate() error {
	if r.RideID == "" {
		return errors.New("ride_id is required")
	}
	if r.Latitude < -90 || r.Latitude > 90 {
		return errors.New("invalid latitude")
	}
	if r.Longitude < -180 || r.Longitude > 180 {
		return errors.New("invalid longitude")
	}
	return nil
}

type CompleteRideRequest struct {
	RideID                string  `json:"ride_id"`
	FinalLatitude         float64 `json:"final_location.latitude"`
	FinalLongitude        float64 `json:"final_location.longitude"`
	ActualDistanceKm      float64 `json:"actual_distance_km"`
	ActualDurationMinutes int     `json:"actual_duration_minutes"`
}

func (r *CompleteRideRequest) Validate() error {
	if r.RideID == "" {
		return errors.New("ride_id is required")
	}
	if r.FinalLatitude < -90 || r.FinalLatitude > 90 {
		return errors.New("invalid final latitude")
	}
	if r.FinalLongitude < -180 || r.FinalLongitude > 180 {
		return errors.New("invalid final longitude")
	}
	if r.ActualDistanceKm < 0 {
		return errors.New("actual_distance_km cannot be negative")
	}
	if r.ActualDurationMinutes < 0 {
		return errors.New("actual_duration_minutes cannot be negative")
	}
	return nil
}

type OnlineResponse struct {
	Status    string `json:"status"`
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

type OfflineResponse struct {
	Status         string          `json:"status"`
	SessionID      string          `json:"session_id"`
	SessionSummary *SessionSummary `json:"session_summary,omitempty"`
	Message        string          `json:"message"`
}

type SessionSummary struct {
	DurationHours  float64 `json:"duration_hours"`
	RidesCompleted int     `json:"rides_completed"`
	Earnings       float64 `json:"earnings"`
}

type LocationUpdateResponse struct {
	CoordinateID string    `json:"coordinate_id"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type StartRideResponse struct {
	RideID    string    `json:"ride_id"`
	Status    string    `json:"status"`
	StartedAt time.Time `json:"started_at"`
	Message   string    `json:"message"`
}

type CompleteRideResponse struct {
	RideID         string    `json:"ride_id"`
	Status         string    `json:"status"`
	CompletedAt    time.Time `json:"completed_at"`
	DriverEarnings float64   `json:"driver_earnings"`
	Message        string    `json:"message"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type OnlineInput struct {
	DriverID  string
	Latitude  float64
	Longitude float64
}

type LocationUpdateInput struct {
	DriverID       string
	Latitude       float64
	Longitude      float64
	AccuracyMeters float64
	SpeedKmh       float64
	HeadingDegrees float64
	RideID         *string
}

type StartRideInput struct {
	DriverID  string
	RideID    string
	Latitude  float64
	Longitude float64
}

type CompleteRideInput struct {
	DriverID              string
	RideID                string
	FinalLatitude         float64
	FinalLongitude        float64
	ActualDistanceKm      float64
	ActualDurationMinutes int
}

type OnlineOutput struct {
	SessionID string
	Status    string
}

type OfflineOutput struct {
	SessionID      string
	DurationHours  float64
	RidesCompleted int
	Earnings       float64
}

type LocationOutput struct {
	CoordinateID string
	UpdatedAt    time.Time
}

type StartRideOutput struct {
	RideID    string
	Status    string
	StartedAt time.Time
}

type CompleteRideOutput struct {
	RideID         string
	Status         string
	CompletedAt    time.Time
	DriverEarnings float64
}
