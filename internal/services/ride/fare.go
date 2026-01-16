package ride

import (
	"fmt"
	"math"
)

// FareCalculator calculates ride fares based on vehicle type, distance, and duration
type FareCalculator struct {
	rates map[string]FareRate
}

// FareRate defines pricing structure for a vehicle type
type FareRate struct {
	BaseFare      float64 // Base fare in tenge (₸)
	RatePerKm     float64 // Rate per kilometer
	RatePerMinute float64 // Rate per minute
}

// NewFareCalculator creates a new fare calculator with predefined rates
func NewFareCalculator() *FareCalculator {
	return &FareCalculator{
		rates: map[string]FareRate{
			"ECONOMY": {
				BaseFare:      500.0,
				RatePerKm:     100.0,
				RatePerMinute: 50.0,
			},
			"PREMIUM": {
				BaseFare:      800.0,
				RatePerKm:     120.0,
				RatePerMinute: 60.0,
			},
			"XL": {
				BaseFare:      1000.0,
				RatePerKm:     150.0,
				RatePerMinute: 75.0,
			},
		},
	}
}

// CalculateFare computes the fare for a ride
// Formula: base_fare + (distance_km * rate_per_km) + (duration_minutes * rate_per_minute)
func (fc *FareCalculator) CalculateFare(vehicleType string, distanceKm float64, durationMinutes float64, surgeMultiplier float64) (float64, error) {
	rate, exists := fc.rates[vehicleType]
	if !exists {
		return 0, fmt.Errorf("invalid vehicle type: %s", vehicleType)
	}

	// Base calculation
	fare := rate.BaseFare + (distanceKm * rate.RatePerKm) + (durationMinutes * rate.RatePerMinute)

	// Apply surge pricing
	if surgeMultiplier < 1.0 {
		surgeMultiplier = 1.0
	}
	fare *= surgeMultiplier

	// Round to nearest 10 tenge
	fare = math.Round(fare/10.0) * 10.0

	return fare, nil
}

// EstimateFare estimates fare without duration (for initial ride request)
// Uses average speed estimation: 30 km/h in city traffic
func (fc *FareCalculator) EstimateFare(vehicleType string, distanceKm float64, surgeMultiplier float64) (float64, error) {
	const avgSpeedKmH = 30.0
	estimatedDurationMinutes := (distanceKm / avgSpeedKmH) * 60.0

	return fc.CalculateFare(vehicleType, distanceKm, estimatedDurationMinutes, surgeMultiplier)
}

// GetRates returns the fare rates for all vehicle types
func (fc *FareCalculator) GetRates() map[string]FareRate {
	return fc.rates
}

// GetRate returns the fare rate for a specific vehicle type
func (fc *FareCalculator) GetRate(vehicleType string) (FareRate, error) {
	rate, exists := fc.rates[vehicleType]
	if !exists {
		return FareRate{}, fmt.Errorf("invalid vehicle type: %s", vehicleType)
	}
	return rate, nil
}
