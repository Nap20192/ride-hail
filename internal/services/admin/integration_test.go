package admin_test

import (
	"context"
	"testing"
	"time"

	"ride-hail/internal/deps"
	"ride-hail/internal/services/admin"
	"ride-hail/internal/shared/config"
	"ride-hail/pkg/sqlc"
)

// TestIntegration_GetSystemMetrics tests the GetSystemMetrics method with a real database
// Run with: go test -tags=integration ./internal/services/admin/...
func TestIntegration_GetSystemMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Setup database connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	infra, err := deps.NewInfraDeps(deps.WithPostgres(ctx, *cfg))
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer infra.Pool.Close()

	// Create service
	queries := sqlc.New(infra.Pool)
	service := admin.NewAdminService(queries)

	// Test GetSystemMetrics
	metrics, err := service.GetSystemMetrics(ctx)
	if err != nil {
		t.Fatalf("GetSystemMetrics failed: %v", err)
	}

	// Validate response structure
	if metrics == nil {
		t.Fatal("metrics should not be nil")
	}

	// Validate metrics are non-negative
	if metrics.ActiveRides < 0 {
		t.Errorf("ActiveRides should not be negative, got %d", metrics.ActiveRides)
	}
	if metrics.AvailableDrivers < 0 {
		t.Errorf("AvailableDrivers should not be negative, got %d", metrics.AvailableDrivers)
	}
	if metrics.BusyDrivers < 0 {
		t.Errorf("BusyDrivers should not be negative, got %d", metrics.BusyDrivers)
	}
	if metrics.TotalRidesToday < 0 {
		t.Errorf("TotalRidesToday should not be negative, got %d", metrics.TotalRidesToday)
	}
	if metrics.TotalRevenueToday < 0 {
		t.Errorf("TotalRevenueToday should not be negative, got %f", metrics.TotalRevenueToday)
	}
	if metrics.CancellationRate < 0 || metrics.CancellationRate > 100 {
		t.Errorf("CancellationRate should be between 0 and 100, got %f", metrics.CancellationRate)
	}

	t.Logf("System Metrics: ActiveRides=%d, AvailableDrivers=%d, BusyDrivers=%d, TodayRides=%d, Revenue=%.2f",
		metrics.ActiveRides, metrics.AvailableDrivers, metrics.BusyDrivers, metrics.TotalRidesToday, metrics.TotalRevenueToday)
}

// TestIntegration_GetDriverDistribution tests the GetDriverDistribution method with a real database
func TestIntegration_GetDriverDistribution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	infra, err := deps.NewInfraDeps(deps.WithPostgres(ctx, *cfg))
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer infra.Pool.Close()

	queries := sqlc.New(infra.Pool)
	service := admin.NewAdminService(queries)

	// Test GetDriverDistribution
	distribution, err := service.GetDriverDistribution(ctx)
	if err != nil {
		t.Fatalf("GetDriverDistribution failed: %v", err)
	}

	if distribution == nil {
		t.Fatal("distribution should not be nil")
	}

	// Should have all vehicle types initialized
	requiredTypes := []string{"ECONOMY", "PREMIUM", "XL"}
	for _, vtype := range requiredTypes {
		if _, exists := distribution[vtype]; !exists {
			t.Errorf("vehicle type %s missing from distribution", vtype)
		}
		if distribution[vtype] < 0 {
			t.Errorf("driver count for %s should not be negative, got %d", vtype, distribution[vtype])
		}
	}

	t.Logf("Driver Distribution: ECONOMY=%d, PREMIUM=%d, XL=%d",
		distribution["ECONOMY"], distribution["PREMIUM"], distribution["XL"])
}

// TestIntegration_GetActiveRides tests the GetActiveRides method with a real database
func TestIntegration_GetActiveRides(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	infra, err := deps.NewInfraDeps(deps.WithPostgres(ctx, *cfg))
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer infra.Pool.Close()

	queries := sqlc.New(infra.Pool)
	service := admin.NewAdminService(queries)

	// Test GetActiveRides with different pagination
	tests := []struct {
		name     string
		page     int
		pageSize int
	}{
		{
			name:     "first page default size",
			page:     1,
			pageSize: 20,
		},
		{
			name:     "first page small size",
			page:     1,
			pageSize: 5,
		},
		{
			name:     "second page",
			page:     2,
			pageSize: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rides, totalCount, err := service.GetActiveRides(ctx, tt.page, tt.pageSize)
			if err != nil {
				t.Fatalf("GetActiveRides failed: %v", err)
			}

			if rides == nil {
				t.Fatal("rides should not be nil")
			}

			if totalCount < 0 {
				t.Errorf("totalCount should not be negative, got %d", totalCount)
			}

			if len(rides) > tt.pageSize {
				t.Errorf("returned rides count (%d) exceeds page size (%d)", len(rides), tt.pageSize)
			}

			// Validate ride structure if any rides exist
			for i, ride := range rides {
				if ride.RideID == "" {
					t.Errorf("ride %d: RideID should not be empty", i)
				}
				if ride.RideNumber == "" {
					t.Errorf("ride %d: RideNumber should not be empty", i)
				}
				if ride.Status == "" {
					t.Errorf("ride %d: Status should not be empty", i)
				}
				if ride.PassengerID == "" {
					t.Errorf("ride %d: PassengerID should not be empty", i)
				}
			}

			t.Logf("Page %d (size %d): Found %d rides, total count: %d",
				tt.page, tt.pageSize, len(rides), totalCount)
		})
	}
}

// TestIntegration_Pagination_Boundaries tests edge cases for pagination
func TestIntegration_Pagination_Boundaries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	infra, err := deps.NewInfraDeps(deps.WithPostgres(ctx, *cfg))
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer infra.Pool.Close()

	queries := sqlc.New(infra.Pool)
	service := admin.NewAdminService(queries)

	// Test invalid page (should default to 1)
	rides, totalCount, err := service.GetActiveRides(ctx, 0, 10)
	if err != nil {
		t.Fatalf("GetActiveRides with page=0 failed: %v", err)
	}
	t.Logf("Page 0 (defaults to 1): Found %d rides, total: %d", len(rides), totalCount)

	// Test invalid page size (should default to 20)
	rides, totalCount, err = service.GetActiveRides(ctx, 1, 0)
	if err != nil {
		t.Fatalf("GetActiveRides with pageSize=0 failed: %v", err)
	}
	t.Logf("PageSize 0 (defaults to 20): Found %d rides, total: %d", len(rides), totalCount)

	// Test very large page size (should cap at 100)
	rides, totalCount, err = service.GetActiveRides(ctx, 1, 999)
	if err != nil {
		t.Fatalf("GetActiveRides with pageSize=999 failed: %v", err)
	}
	if len(rides) > 100 {
		t.Errorf("rides count (%d) exceeds maximum page size of 100", len(rides))
	}
	t.Logf("PageSize 999 (capped at 100): Found %d rides, total: %d", len(rides), totalCount)
}

// TestIntegration_ContextCancellation tests that operations respect context cancellation
func TestIntegration_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	infra, err := deps.NewInfraDeps(deps.WithPostgres(ctx, *cfg))
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer infra.Pool.Close()

	queries := sqlc.New(infra.Pool)
	service := admin.NewAdminService(queries)

	// Create a context that's immediately cancelled
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Operations should fail with context cancelled error
	_, err = service.GetSystemMetrics(cancelledCtx)
	if err == nil {
		t.Error("expected error with cancelled context, got nil")
	}
	if err != nil {
		t.Logf("correctly got error with cancelled context: %v", err)
	}
}
