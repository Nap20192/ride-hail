package admin

import (
	"context"
	"fmt"

	"ride-hail/pkg/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

type AdminService struct {
	queries *sqlc.Queries
}

func NewAdminService(queries *sqlc.Queries) *AdminService {
	return &AdminService{
		queries: queries,
	}
}

// GetSystemMetrics retrieves overall system statistics
func (s *AdminService) GetSystemMetrics(ctx context.Context) (*SystemMetrics, error) {
	// Query active rides
	activeRides, err := s.queries.GetActiveRidesCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active rides count: %w", err)
	}

	// Query available drivers
	availableDrivers, err := s.queries.GetAvailableDriversCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get available drivers count: %w", err)
	}

	// Query busy drivers
	busyDrivers, err := s.queries.GetBusyDriversCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get busy drivers count: %w", err)
	}

	// Query today's rides count
	todayRides, err := s.queries.GetTodayRidesCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get today rides count: %w", err)
	}

	// Query today's revenue
	todayRevenue, err := s.queries.GetTodayRevenue(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get today revenue: %w", err)
	}

	// Query average wait time
	avgWaitTime, err := s.queries.GetAverageWaitTime(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get average wait time: %w", err)
	}

	// Query average ride duration
	avgRideDuration, err := s.queries.GetAverageRideDuration(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get average ride duration: %w", err)
	}

	// Query cancellation rate
	cancellationRate, err := s.queries.GetCancellationRate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cancellation rate: %w", err)
	}

	// Convert types
	revenueFloat := numericToFloat64(todayRevenue)
	waitTimeFloat := numericToFloat64(avgWaitTime)
	durationFloat := numericToFloat64(avgRideDuration)

	var cancellationFloat float64
	if cancellationRate != nil {
		if val, ok := cancellationRate.(float64); ok {
			cancellationFloat = val
		}
	}

	metrics := &SystemMetrics{
		ActiveRides:            int(activeRides),
		AvailableDrivers:       int(availableDrivers),
		BusyDrivers:            int(busyDrivers),
		TotalRidesToday:        int(todayRides),
		TotalRevenueToday:      revenueFloat,
		AverageWaitTimeMinutes: waitTimeFloat,
		AverageRideDurationMin: durationFloat,
		CancellationRate:       cancellationFloat,
	}

	return metrics, nil
}

// Helper function to convert pgtype.Numeric or interface{} to float64
func numericToFloat64(val interface{}) float64 {
	if val == nil {
		return 0.0
	}

	// Check if it's a pgtype.Numeric
	if num, ok := val.(pgtype.Numeric); ok {
		f8, err := num.Float64Value()
		if err != nil || !f8.Valid {
			return 0.0
		}
		return f8.Float64
	}

	// Check if it's already a float64
	if f64, ok := val.(float64); ok {
		return f64
	}

	return 0.0
}

// GetDriverDistribution retrieves driver count by vehicle type
func (s *AdminService) GetDriverDistribution(ctx context.Context) (map[string]int, error) {
	rows, err := s.queries.GetDriverDistributionByVehicleType(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get driver distribution: %w", err)
	}

	distribution := map[string]int{
		"ECONOMY": 0,
		"PREMIUM": 0,
		"XL":      0,
	}

	for _, row := range rows {
		vehicleType := row.VehicleType.(string)
		count := int(row.Count)
		distribution[vehicleType] = count
	}

	return distribution, nil
}

// GetActiveRides retrieves paginated list of active rides
func (s *AdminService) GetActiveRides(ctx context.Context, page, pageSize int) ([]ActiveRide, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	limit := pageSize

	// Get total count
	totalCount, err := s.queries.GetActiveRidesTotalCount(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	// Get paginated rides
	rows, err := s.queries.GetActiveRidesPaginated(ctx, sqlc.GetActiveRidesPaginatedParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get active rides: %w", err)
	}

	rides := make([]ActiveRide, 0, len(rows))
	for _, row := range rows {
		ride := ActiveRide{
			RideID:             row.ID.String(),
			RideNumber:         row.RideNumber,
			Status:             ptrToString(row.Status),
			PassengerID:        row.PassengerID.String(),
			PickupAddress:      ptrToString(row.PickupAddress),
			DestinationAddress: ptrToString(row.DestinationAddress),
		}

		// Check if DriverID is not zero UUID
		if row.DriverID.String() != "00000000-0000-0000-0000-000000000000" {
			driverID := row.DriverID.String()
			ride.DriverID = &driverID
		}

		if row.StartedAt != nil {
			ride.StartedAt = row.StartedAt
		}

		rides = append(rides, ride)
	}

	return rides, int(totalCount), nil
}

func ptrToString(val *string) string {
	if val == nil {
		return ""
	}
	return *val
}
