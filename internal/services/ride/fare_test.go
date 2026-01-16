package ride

import (
	"testing"
)

func TestNewFareCalculator(t *testing.T) {
	fc := NewFareCalculator()

	if fc == nil {
		t.Fatal("NewFareCalculator() returned nil")
	}

	rates := fc.GetRates()
	if len(rates) != 3 {
		t.Errorf("Expected 3 vehicle types, got %d", len(rates))
	}

	// Verify ECONOMY rates
	economy, err := fc.GetRate("ECONOMY")
	if err != nil {
		t.Errorf("Failed to get ECONOMY rate: %v", err)
	}
	if economy.BaseFare != 500.0 {
		t.Errorf("ECONOMY base fare: expected 500, got %.2f", economy.BaseFare)
	}
	if economy.RatePerKm != 100.0 {
		t.Errorf("ECONOMY rate per km: expected 100, got %.2f", economy.RatePerKm)
	}
	if economy.RatePerMinute != 50.0 {
		t.Errorf("ECONOMY rate per minute: expected 50, got %.2f", economy.RatePerMinute)
	}

	// Verify PREMIUM rates
	premium, err := fc.GetRate("PREMIUM")
	if err != nil {
		t.Errorf("Failed to get PREMIUM rate: %v", err)
	}
	if premium.BaseFare != 800.0 {
		t.Errorf("PREMIUM base fare: expected 800, got %.2f", premium.BaseFare)
	}
	if premium.RatePerKm != 120.0 {
		t.Errorf("PREMIUM rate per km: expected 120, got %.2f", premium.RatePerKm)
	}
	if premium.RatePerMinute != 60.0 {
		t.Errorf("PREMIUM rate per minute: expected 60, got %.2f", premium.RatePerMinute)
	}

	// Verify XL rates
	xl, err := fc.GetRate("XL")
	if err != nil {
		t.Errorf("Failed to get XL rate: %v", err)
	}
	if xl.BaseFare != 1000.0 {
		t.Errorf("XL base fare: expected 1000, got %.2f", xl.BaseFare)
	}
	if xl.RatePerKm != 150.0 {
		t.Errorf("XL rate per km: expected 150, got %.2f", xl.RatePerKm)
	}
	if xl.RatePerMinute != 75.0 {
		t.Errorf("XL rate per minute: expected 75, got %.2f", xl.RatePerMinute)
	}
}

func TestCalculateFare_Economy(t *testing.T) {
	fc := NewFareCalculator()

	// Test: 10 km, 20 minutes, no surge
	// Expected: 500 + (10 * 100) + (20 * 50) = 500 + 1000 + 1000 = 2500
	fare, err := fc.CalculateFare("ECONOMY", 10.0, 20.0, 1.0)
	if err != nil {
		t.Errorf("CalculateFare failed: %v", err)
	}
	if fare != 2500.0 {
		t.Errorf("Expected fare 2500, got %.2f", fare)
	}
}

func TestCalculateFare_Premium(t *testing.T) {
	fc := NewFareCalculator()

	// Test: 5 km, 15 minutes, no surge
	// Expected: 800 + (5 * 120) + (15 * 60) = 800 + 600 + 900 = 2300
	fare, err := fc.CalculateFare("PREMIUM", 5.0, 15.0, 1.0)
	if err != nil {
		t.Errorf("CalculateFare failed: %v", err)
	}
	if fare != 2300.0 {
		t.Errorf("Expected fare 2300, got %.2f", fare)
	}
}

func TestCalculateFare_XL(t *testing.T) {
	fc := NewFareCalculator()

	// Test: 8 km, 10 minutes, no surge
	// Expected: 1000 + (8 * 150) + (10 * 75) = 1000 + 1200 + 750 = 2950
	fare, err := fc.CalculateFare("XL", 8.0, 10.0, 1.0)
	if err != nil {
		t.Errorf("CalculateFare failed: %v", err)
	}
	if fare != 2950.0 {
		t.Errorf("Expected fare 2950, got %.2f", fare)
	}
}

func TestCalculateFare_WithSurge(t *testing.T) {
	fc := NewFareCalculator()

	// Test: 10 km, 20 minutes, 1.5x surge
	// Expected: (500 + 1000 + 1000) * 1.5 = 2500 * 1.5 = 3750
	fare, err := fc.CalculateFare("ECONOMY", 10.0, 20.0, 1.5)
	if err != nil {
		t.Errorf("CalculateFare failed: %v", err)
	}
	if fare != 3750.0 {
		t.Errorf("Expected fare 3750, got %.2f", fare)
	}
}

func TestCalculateFare_RoundingToNearest10(t *testing.T) {
	fc := NewFareCalculator()

	tests := []struct {
		name     string
		distance float64
		duration float64
		expected float64
	}{
		{
			name:     "rounds up from 5",
			distance: 1.0,  // 100
			duration: 2.45, // 122.5
			// Total: 500 + 100 + 122.5 = 722.5 → rounds to 720
			expected: 720.0,
		},
		{
			name:     "rounds up from 6",
			distance: 1.0, // 100
			duration: 2.5, // 125
			// Total: 500 + 100 + 125 = 725 → rounds to 730
			expected: 730.0,
		},
		{
			name:     "rounds down from 4",
			distance: 1.0,  // 100
			duration: 2.48, // 124
			// Total: 500 + 100 + 124 = 724 → rounds to 720
			expected: 720.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fare, err := fc.CalculateFare("ECONOMY", tt.distance, tt.duration, 1.0)
			if err != nil {
				t.Errorf("CalculateFare failed: %v", err)
			}
			if fare != tt.expected {
				t.Errorf("Expected fare %.2f, got %.2f", tt.expected, fare)
			}
		})
	}
}

func TestCalculateFare_InvalidVehicleType(t *testing.T) {
	fc := NewFareCalculator()

	_, err := fc.CalculateFare("INVALID", 10.0, 20.0, 1.0)
	if err == nil {
		t.Error("Expected error for invalid vehicle type, got nil")
	}
}

func TestCalculateFare_SurgeMinimum(t *testing.T) {
	fc := NewFareCalculator()

	// Test with surge < 1.0, should default to 1.0
	fare1, _ := fc.CalculateFare("ECONOMY", 10.0, 20.0, 0.5)
	fare2, _ := fc.CalculateFare("ECONOMY", 10.0, 20.0, 1.0)

	if fare1 != fare2 {
		t.Errorf("Surge multiplier < 1.0 should be treated as 1.0: got %.2f and %.2f", fare1, fare2)
	}
}

func TestEstimateFare(t *testing.T) {
	fc := NewFareCalculator()

	// Test: 10 km distance
	// Estimated duration: (10 / 30) * 60 = 20 minutes
	// Expected: 500 + (10 * 100) + (20 * 50) = 2500
	fare, err := fc.EstimateFare("ECONOMY", 10.0, 1.0)
	if err != nil {
		t.Errorf("EstimateFare failed: %v", err)
	}
	if fare != 2500.0 {
		t.Errorf("Expected estimated fare 2500, got %.2f", fare)
	}
}

func TestEstimateFare_WithSurge(t *testing.T) {
	fc := NewFareCalculator()

	// Test: 10 km distance, 1.5x surge
	// Estimated duration: 20 minutes
	// Expected: (500 + 1000 + 1000) * 1.5 = 3750
	fare, err := fc.EstimateFare("ECONOMY", 10.0, 1.5)
	if err != nil {
		t.Errorf("EstimateFare failed: %v", err)
	}
	if fare != 3750.0 {
		t.Errorf("Expected estimated fare 3750, got %.2f", fare)
	}
}

func TestGetRate_InvalidType(t *testing.T) {
	fc := NewFareCalculator()

	_, err := fc.GetRate("INVALID")
	if err == nil {
		t.Error("Expected error for invalid vehicle type, got nil")
	}
}

func TestCalculateFare_ZeroDistance(t *testing.T) {
	fc := NewFareCalculator()

	// Test: 0 km, 5 minutes - should still charge base fare + time
	// Expected: 500 + 0 + (5 * 50) = 750
	fare, err := fc.CalculateFare("ECONOMY", 0.0, 5.0, 1.0)
	if err != nil {
		t.Errorf("CalculateFare failed: %v", err)
	}
	if fare != 750.0 {
		t.Errorf("Expected fare 750, got %.2f", fare)
	}
}

func TestCalculateFare_LongDistance(t *testing.T) {
	fc := NewFareCalculator()

	// Test: 100 km, 120 minutes (2 hours)
	// Expected: 500 + (100 * 100) + (120 * 50) = 500 + 10000 + 6000 = 16500
	fare, err := fc.CalculateFare("ECONOMY", 100.0, 120.0, 1.0)
	if err != nil {
		t.Errorf("CalculateFare failed: %v", err)
	}
	if fare != 16500.0 {
		t.Errorf("Expected fare 16500, got %.2f", fare)
	}
}
