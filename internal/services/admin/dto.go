package admin

import "time"

type OverviewResponse struct {
	Timestamp          time.Time      `json:"timestamp"`
	Metrics            *SystemMetrics `json:"metrics"`
	DriverDistribution map[string]int `json:"driver_distribution"`
	Hotspots           []Hotspot      `json:"hotspots,omitempty"`
}

type SystemMetrics struct {
	ActiveRides            int     `json:"active_rides"`
	AvailableDrivers       int     `json:"available_drivers"`
	BusyDrivers            int     `json:"busy_drivers"`
	TotalRidesToday        int     `json:"total_rides_today"`
	TotalRevenueToday      float64 `json:"total_revenue_today"`
	AverageWaitTimeMinutes float64 `json:"average_wait_time_minutes"`
	AverageRideDurationMin float64 `json:"average_ride_duration_minutes"`
	CancellationRate       float64 `json:"cancellation_rate"`
}

type Hotspot struct {
	Location       string `json:"location"`
	ActiveRides    int    `json:"active_rides"`
	WaitingDrivers int    `json:"waiting_drivers"`
}

type ActiveRidesResponse struct {
	Rides      []ActiveRide `json:"rides"`
	TotalCount int          `json:"total_count"`
	Page       int          `json:"page"`
	PageSize   int          `json:"page_size"`
}

type ActiveRide struct {
	RideID                string     `json:"ride_id"`
	RideNumber            string     `json:"ride_number"`
	Status                string     `json:"status"`
	PassengerID           string     `json:"passenger_id"`
	DriverID              *string    `json:"driver_id,omitempty"`
	PickupAddress         string     `json:"pickup_address"`
	DestinationAddress    string     `json:"destination_address"`
	StartedAt             *time.Time `json:"started_at,omitempty"`
	EstimatedCompletion   *time.Time `json:"estimated_completion,omitempty"`
	CurrentDriverLocation *Location  `json:"current_driver_location,omitempty"`
	DistanceCompletedKm   float64    `json:"distance_completed_km"`
	DistanceRemainingKm   float64    `json:"distance_remaining_km"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
