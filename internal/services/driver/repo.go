package driver

import (
	"context"
	"ride-hail/internal/services/driver/models"
	"ride-hail/internal/shared/core"
	"ride-hail/pkg/sqlc"
	"ride-hail/pkg/uuid"
)

func statusOnline(ctx context.Context, qtx sqlc.Querier, args models.OnlineRequest) error {
	err := qtx.UpdateDriverStatus(ctx, sqlc.UpdateDriverStatusParams{
		Status: core.DriverStatusAvailable.String(),
		ID:     args.DriverID,
	})

	return err
}
func statusOffline(ctx context.Context, qtx sqlc.Querier, driverID uuid.UUID) error {
	err := qtx.UpdateDriverStatus(ctx, sqlc.UpdateDriverStatusParams{
		Status: core.DriverStatusOffline.String(),
		ID:     driverID,
	})

	return err
}

func createSession(ctx context.Context, qtx sqlc.Querier, driverID uuid.UUID) (uuid.UUID, error) {
	

	session, err := qtx.CreateDriverSession(ctx, driverID)
	if err != nil {
		return uuid.Nil, err
	}

	return session.ID, nil
}
