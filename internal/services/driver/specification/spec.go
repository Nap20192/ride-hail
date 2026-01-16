package specification

import (
	"context"
	"errors"

	"ride-hail/internal/services/driver/models"
	appErrors "ride-hail/internal/shared/errors"
	"ride-hail/pkg/sqlc"
	"ride-hail/pkg/uuid"

	"github.com/jackc/pgx/v5"
)

var (
	ErrDriverAlreadyOnline  = appErrors.NewConflictError("driver already online")
	ErrDriverAlreadyOffline = appErrors.NewConflictError("driver already offline")
)

type DriverSpecification struct {
	queries *sqlc.Queries
}

func NewDriverSpecification(queries *sqlc.Queries) *DriverSpecification {
	return &DriverSpecification{
		queries: queries,
	}
}

func (s *DriverSpecification) Online(ctx context.Context, arg models.OnlineRequest) (uuid.UUID, error) {

	if err := arg.Validate(); err != nil {
		return uuid.Nil, appErrors.NewInvalidInputError(err.Error())
	}

	session, err := s.queries.GetCurrentDriverSession(ctx, arg.DriverID)

	if err == nil {
		return session.ID, ErrDriverAlreadyOnline
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, nil
	}

	return uuid.Nil, err
}

func (s *DriverSpecification) Offline(ctx context.Context, driverID uuid.UUID) error {
	if driverID.IsZero() {
		return appErrors.NewInvalidInputError("driver_id is required")
	}

	_, err := s.queries.GetCurrentDriverSession(ctx, driverID)

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrDriverAlreadyOffline
	}

	if err != nil {
		return err
	}

	return err
}

func (s *DriverSpecification) UpdateLocation(ctx context.Context, arg models.LocationUpdateRequest) error {
	if arg.DriverID.IsZero() {
		return appErrors.NewInvalidInputError("driver_id is required")
	}

	if err := arg.Validate(); err != nil {
		return appErrors.NewInvalidInputError(err.Error())
	}

	_, err := s.queries.GetCurrentDriverSession(ctx, arg.DriverID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrDriverAlreadyOffline
	}
	if err != nil {
		return err
	}

	return nil
}
