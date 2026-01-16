package models

import (
	"time"

	"ride-hail/internal/shared/core"
	"ride-hail/pkg/uuid"
)

type Driver struct {
	ID            uuid.UUID
	Status        core.DriverStatus
	VehicleType   string
	IsVerified    bool
	Rating        float64
	TotalRides    int
	CurrentRideID uuid.UUID
	Location      Location
	SessionID     uuid.UUID
}

type Location struct {
	Latitude       float64
	Longitude      float64
	AccuracyMeters float64
	SpeedKmh       float64
	HeadingDegrees float64
	UpdatedAt      time.Time
}

type RideRequest struct {
	RideID               uuid.UUID
	VehicleType          string
	PickupLatitude       float64
	PickupLongitude      float64
	MaxDistanceKm        float64
	RequiredRating       float64
	EstimatedDurationMin int
}

type LocationUpdate struct {
	DriverID       uuid.UUID
	Location       Location
	LastUpdateTime time.Time
	RideID         *uuid.UUID
}

type DriverSession struct {
	ID            uuid.UUID
	DriverID      uuid.UUID
	StartedAt     time.Time
	EndedAt       *time.Time
	TotalRides    int
	TotalEarnings float64
	DurationHours float64
}

type VehicleInfo struct {
	Make  string
	Model string
	Color string
	Plate string
	Year  int
}

func (d *Driver) CanGoOnline() bool {
	return d.IsVerified && d.Status == core.DriverStatusOffline
}

func (d *Driver) CanGoOffline() bool {
	return d.Status == core.DriverStatusAvailable && d.CurrentRideID == uuid.Nil
}

func (d *Driver) CanAcceptRide() bool {
	return d.Status == core.DriverStatusAvailable &&
		d.IsVerified &&
		d.CurrentRideID == uuid.Nil &&
		d.Location != (Location{})
}

func (d *Driver) IsLocationStale() bool {
	if d.Location == (Location{}) {
		return true
	}
	return time.Since(d.Location.UpdatedAt) > 5*time.Minute
}
