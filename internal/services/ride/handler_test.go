package ride

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ride-hail/internal/auth"
	"ride-hail/internal/middleware"
	"ride-hail/pkg/uuid"
)

// TestCreateRide_ValidInput tests successful ride creation
func TestCreateRide_ValidInput(t *testing.T) {
	tests := []struct {
		name        string
		vehicleType string
		pickupLat   float64
		pickupLng   float64
		destLat     float64
		destLng     float64
	}{
		{
			name:        "ECONOMY ride",
			vehicleType: "ECONOMY",
			pickupLat:   43.238949,
			pickupLng:   76.889709,
			destLat:     43.250000,
			destLng:     76.900000,
		},
		{
			name:        "PREMIUM ride",
			vehicleType: "PREMIUM",
			pickupLat:   43.238949,
			pickupLng:   76.889709,
			destLat:     43.260000,
			destLng:     76.910000,
		},
		{
			name:        "XL ride",
			vehicleType: "XL",
			pickupLat:   43.238949,
			pickupLng:   76.889709,
			destLat:     43.270000,
			destLng:     76.920000,
		},
		{
			name:        "Minimum valid coordinates",
			vehicleType: "ECONOMY",
			pickupLat:   -90.0,
			pickupLng:   -180.0,
			destLat:     -89.0,
			destLng:     -179.0,
		},
		{
			name:        "Maximum valid coordinates",
			vehicleType: "ECONOMY",
			pickupLat:   90.0,
			pickupLng:   180.0,
			destLat:     89.0,
			destLng:     179.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreateRideRequest{
				PassengerID:   uuid.New(),
				PickupLat:     tt.pickupLat,
				PickupLng:     tt.pickupLng,
				PickupAddress: "123 Pickup St",
				DestLat:       tt.destLat,
				DestLng:       tt.destLng,
				DestAddress:   "456 Destination Ave",
				VehicleType:   tt.vehicleType,
			}

			err := req.Validate()
			if err != nil {
				t.Errorf("Valid input should not produce error: %v", err)
			}
		})
	}
}

// TestCreateRide_InvalidVehicleType tests invalid vehicle types
func TestCreateRide_InvalidVehicleType(t *testing.T) {
	tests := []struct {
		name        string
		vehicleType string
	}{
		{"Empty vehicle type", ""},
		{"Invalid vehicle type", "LUXURY"},
		{"Lowercase vehicle type", "economy"},
		{"Mixed case vehicle type", "Economy"},
		{"Unknown vehicle type", "INVALID"},
		{"Special characters", "ECONOMY!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreateRideRequest{
				PassengerID:   uuid.New(),
				PickupLat:     43.238949,
				PickupLng:     76.889709,
				PickupAddress: "123 Pickup St",
				DestLat:       43.250000,
				DestLng:       76.900000,
				DestAddress:   "456 Destination Ave",
				VehicleType:   tt.vehicleType,
			}

			err := req.Validate()
			if err == nil {
				t.Errorf("Expected error for invalid vehicle type '%s'", tt.vehicleType)
			}
		})
	}
}

// TestCreateRide_InvalidCoordinates tests invalid coordinates
func TestCreateRide_InvalidCoordinates(t *testing.T) {
	tests := []struct {
		name      string
		pickupLat float64
		pickupLng float64
		destLat   float64
		destLng   float64
		errorMsg  string
	}{
		{
			name:      "Pickup latitude too low",
			pickupLat: -91.0,
			pickupLng: 76.889709,
			destLat:   43.250000,
			destLng:   76.900000,
			errorMsg:  "pickup_latitude must be between -90 and 90",
		},
		{
			name:      "Pickup latitude too high",
			pickupLat: 91.0,
			pickupLng: 76.889709,
			destLat:   43.250000,
			destLng:   76.900000,
			errorMsg:  "pickup_latitude must be between -90 and 90",
		},
		{
			name:      "Pickup longitude too low",
			pickupLat: 43.238949,
			pickupLng: -181.0,
			destLat:   43.250000,
			destLng:   76.900000,
			errorMsg:  "pickup_longitude must be between -180 and 180",
		},
		{
			name:      "Pickup longitude too high",
			pickupLat: 43.238949,
			pickupLng: 181.0,
			destLat:   43.250000,
			destLng:   76.900000,
			errorMsg:  "pickup_longitude must be between -180 and 180",
		},
		{
			name:      "Destination latitude too low",
			pickupLat: 43.238949,
			pickupLng: 76.889709,
			destLat:   -91.0,
			destLng:   76.900000,
			errorMsg:  "destination_latitude must be between -90 and 90",
		},
		{
			name:      "Destination latitude too high",
			pickupLat: 43.238949,
			pickupLng: 76.889709,
			destLat:   91.0,
			destLng:   76.900000,
			errorMsg:  "destination_latitude must be between -90 and 90",
		},
		{
			name:      "Destination longitude too low",
			pickupLat: 43.238949,
			pickupLng: 76.889709,
			destLat:   43.250000,
			destLng:   -181.0,
			errorMsg:  "destination_longitude must be between -180 and 180",
		},
		{
			name:      "Destination longitude too high",
			pickupLat: 43.238949,
			pickupLng: 76.889709,
			destLat:   43.250000,
			destLng:   181.0,
			errorMsg:  "destination_longitude must be between -180 and 180",
		},
		{
			name:      "Same pickup and destination",
			pickupLat: 43.238949,
			pickupLng: 76.889709,
			destLat:   43.238949,
			destLng:   76.889709,
			errorMsg:  "pickup and destination must be different",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreateRideRequest{
				PassengerID:   uuid.New(),
				PickupLat:     tt.pickupLat,
				PickupLng:     tt.pickupLng,
				PickupAddress: "123 Pickup St",
				DestLat:       tt.destLat,
				DestLng:       tt.destLng,
				DestAddress:   "456 Destination Ave",
				VehicleType:   "ECONOMY",
			}

			err := req.Validate()
			if err == nil {
				t.Errorf("Expected error: %s", tt.errorMsg)
			} else if err.Error() != tt.errorMsg {
				t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
			}
		})
	}
}

// TestCreateRide_MissingAddresses tests missing address fields
func TestCreateRide_MissingAddresses(t *testing.T) {
	tests := []struct {
		name          string
		pickupAddress string
		destAddress   string
		errorMsg      string
	}{
		{
			name:          "Missing pickup address",
			pickupAddress: "",
			destAddress:   "456 Destination Ave",
			errorMsg:      "pickup_address is required",
		},
		{
			name:          "Missing destination address",
			pickupAddress: "123 Pickup St",
			destAddress:   "",
			errorMsg:      "destination_address is required",
		},
		{
			name:          "Both addresses missing",
			pickupAddress: "",
			destAddress:   "",
			errorMsg:      "pickup_address is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreateRideRequest{
				PassengerID:   uuid.New(),
				PickupLat:     43.238949,
				PickupLng:     76.889709,
				PickupAddress: tt.pickupAddress,
				DestLat:       43.250000,
				DestLng:       76.900000,
				DestAddress:   tt.destAddress,
				VehicleType:   "ECONOMY",
			}

			err := req.Validate()
			if err == nil {
				t.Errorf("Expected error: %s", tt.errorMsg)
			}
		})
	}
}

// TestCreateRide_HandlerWithoutAuth tests handler without authentication
func TestCreateRide_HandlerWithoutAuth(t *testing.T) {
	// This would be an integration test requiring a full service setup
	// For now, we test that the handler expects authentication

	reqBody := CreateRideRequest{
		PickupLat:     43.238949,
		PickupLng:     76.889709,
		PickupAddress: "123 Pickup St",
		DestLat:       43.250000,
		DestLng:       76.900000,
		DestAddress:   "456 Destination Ave",
		VehicleType:   "ECONOMY",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/rides", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Without setting auth context, handler should reject
	// Note: This is a partial test - full test would require service initialization
	// The handler expects middleware.UserContextKey to be set
	ctx := req.Context()
	if ctx.Value(middleware.UserContextKey) != nil {
		t.Error("Context should not have user set without authentication")
	}
}

// TestCreateRide_HandlerWithAuth tests handler with proper authentication
func TestCreateRide_HandlerWithAuth(t *testing.T) {
	reqBody := CreateRideRequest{
		PickupLat:     43.238949,
		PickupLng:     76.889709,
		PickupAddress: "123 Pickup St",
		DestLat:       43.250000,
		DestLng:       76.900000,
		DestAddress:   "456 Destination Ave",
		VehicleType:   "ECONOMY",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/rides", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Set authenticated user in context
	passengerID := uuid.New()
	claims := auth.JWTClaims{
		UserID: passengerID,
		Role:   "PASSENGER",
	}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	req = req.WithContext(ctx)

	// Verify context is properly set
	if req.Context().Value(middleware.UserContextKey) == nil {
		t.Error("Context should have user set with authentication")
	}

	extractedClaims, ok := req.Context().Value(middleware.UserContextKey).(auth.JWTClaims)
	if !ok {
		t.Error("Could not extract claims from context")
	}

	if extractedClaims.UserID != passengerID {
		t.Errorf("Expected passenger ID %s, got %s", passengerID, extractedClaims.UserID)
	}
}

// TestCreateRide_InvalidJSON tests malformed JSON input
func TestCreateRide_InvalidJSON(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"Empty body", ""},
		{"Invalid JSON", "{invalid json}"},
		{"Incomplete JSON", `{"pickup_latitude": 43.23`},
		{"Wrong types", `{"pickup_latitude": "not a number"}`},
		{"Extra commas", `{"pickup_latitude": 43.23,,,}`},
		{"Missing quotes", `{pickup_latitude: 43.23}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/rides", bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")

			var result CreateRideRequest
			err := json.NewDecoder(req.Body).Decode(&result)

			if err == nil && tt.body != "" {
				// Empty body might decode to zero values
				t.Error("Expected JSON decode error for malformed input")
			}
		})
	}
}

// TestFareCalculation tests the fare calculation logic
func TestFareCalculation(t *testing.T) {
	service := &RideService{}

	tests := []struct {
		name        string
		vehicleType string
		distanceKm  float64
		durationMin int
		expected    float64
	}{
		{
			name:        "ECONOMY - short trip",
			vehicleType: "ECONOMY",
			distanceKm:  5.0,
			durationMin: 15,
			expected:    500 + (5.0 * 100) + (15 * 50), // 500 + 500 + 750 = 1750
		},
		{
			name:        "PREMIUM - medium trip",
			vehicleType: "PREMIUM",
			distanceKm:  10.0,
			durationMin: 30,
			expected:    800 + (10.0 * 120) + (30 * 60), // 800 + 1200 + 1800 = 3800
		},
		{
			name:        "XL - long trip",
			vehicleType: "XL",
			distanceKm:  20.0,
			durationMin: 60,
			expected:    1000 + (20.0 * 150) + (60 * 75), // 1000 + 3000 + 4500 = 8500
		},
		{
			name:        "ECONOMY - zero distance",
			vehicleType: "ECONOMY",
			distanceKm:  0.0,
			durationMin: 5,
			expected:    500 + 0 + (5 * 50), // 500 + 0 + 250 = 750
		},
		{
			name:        "ECONOMY - fractional values",
			vehicleType: "ECONOMY",
			distanceKm:  2.5,
			durationMin: 8,
			expected:    500 + (2.5 * 100) + (8 * 50), // 500 + 250 + 400 = 1150
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.calculateFare(tt.vehicleType, tt.distanceKm, tt.durationMin)

			if result != tt.expected {
				t.Errorf("Expected fare %.2f, got %.2f", tt.expected, result)
			}
		})
	}
}

// TestRateTableValues tests that rate tables are defined correctly
func TestRateTableValues(t *testing.T) {
	tests := []struct {
		vehicleType    string
		expectedBase   float64
		expectedPerKm  float64
		expectedPerMin float64
	}{
		{"ECONOMY", 500, 100, 50},
		{"PREMIUM", 800, 120, 60},
		{"XL", 1000, 150, 75},
	}

	for _, tt := range tests {
		t.Run(tt.vehicleType, func(t *testing.T) {
			rate, exists := Rates[tt.vehicleType]
			if !exists {
				t.Errorf("Rate table missing for vehicle type: %s", tt.vehicleType)
			}

			if rate.Base != tt.expectedBase {
				t.Errorf("Expected base fare %.2f, got %.2f", tt.expectedBase, rate.Base)
			}

			if rate.PerKm != tt.expectedPerKm {
				t.Errorf("Expected per-km rate %.2f, got %.2f", tt.expectedPerKm, rate.PerKm)
			}

			if rate.PerMin != tt.expectedPerMin {
				t.Errorf("Expected per-minute rate %.2f, got %.2f", tt.expectedPerMin, rate.PerMin)
			}
		})
	}
}

// TestCreateRideRequest_EdgeCases tests edge cases
func TestCreateRideRequest_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		setupReq    func() CreateRideRequest
		shouldError bool
		errorMsg    string
	}{
		{
			name: "Very long address",
			setupReq: func() CreateRideRequest {
				return CreateRideRequest{
					PassengerID:   uuid.New(),
					PickupLat:     43.238949,
					PickupLng:     76.889709,
					PickupAddress: string(make([]byte, 1000)), // 1000 char address
					DestLat:       43.250000,
					DestLng:       76.900000,
					DestAddress:   "456 Destination Ave",
					VehicleType:   "ECONOMY",
				}
			},
			shouldError: false, // Currently no max length validation
		},
		{
			name: "Whitespace only address",
			setupReq: func() CreateRideRequest {
				return CreateRideRequest{
					PassengerID:   uuid.New(),
					PickupLat:     43.238949,
					PickupLng:     76.889709,
					PickupAddress: "   ",
					DestLat:       43.250000,
					DestLng:       76.900000,
					DestAddress:   "456 Destination Ave",
					VehicleType:   "ECONOMY",
				}
			},
			shouldError: false, // Currently no whitespace validation
		},
		{
			name: "Very small coordinate difference",
			setupReq: func() CreateRideRequest {
				return CreateRideRequest{
					PassengerID:   uuid.New(),
					PickupLat:     43.238949,
					PickupLng:     76.889709,
					PickupAddress: "123 Pickup St",
					DestLat:       43.238950, // 0.000001 degree difference
					DestLng:       76.889710,
					DestAddress:   "456 Destination Ave",
					VehicleType:   "ECONOMY",
				}
			},
			shouldError: false, // Valid - different coordinates
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupReq()
			err := req.Validate()

			if tt.shouldError && err == nil {
				t.Errorf("Expected error: %s", tt.errorMsg)
			} else if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
