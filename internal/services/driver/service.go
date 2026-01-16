package driver

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ride-hail/internal/services/driver/models"
	"ride-hail/internal/services/driver/specification"
	"ride-hail/internal/shared/core"
	appErrors "ride-hail/internal/shared/errors"
	"ride-hail/pkg/mq"
	"ride-hail/pkg/sqlc"
	"ride-hail/pkg/uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultMatchRadiusKm   = 5.0
	defaultOfferTimeoutSec = 30
	defaultSpeedKmh        = 30.0
	locationRateLimit      = 3 * time.Second
	driverEarningsRate     = 0.8
)

type DriverService struct {
	db       *pgxpool.Pool
	queries  *sqlc.Queries
	mqClient *mq.Client
	spec     *specification.DriverSpecification
}

func NewDriverService(db *pgxpool.Pool, queries *sqlc.Queries, mqClient *mq.Client) *DriverService {
	return &DriverService{
		db:       db,
		queries:  queries,
		mqClient: mqClient,
		spec:     specification.NewDriverSpecification(queries),
	}
}

func (s *DriverService) Online(ctx context.Context, arg models.OnlineRequest) (session uuid.UUID, err error) {
	_, specErr := s.spec.Online(ctx, arg)

	if specErr != nil {
		return uuid.Nil, fmt.Errorf("driver cannot go online: %w", specErr)
	}

	tx, err := s.db.Begin(ctx)

	if err != nil {
		return uuid.Nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()
	qtx := s.queries.WithTx(tx)

	err = qtx.UpdateDriverStatus(ctx, sqlc.UpdateDriverStatusParams{
		Status: core.DriverStatusAvailable.String(),
		ID:     arg.DriverID,
	})

	if err != nil {
		return uuid.Nil, err
	}

	err = statusOnline(ctx, qtx, arg)

	if err != nil {
		return uuid.Nil, err
	}
	session, err = createSession(ctx, qtx, arg.DriverID)

	if err != nil {
		return uuid.Nil, err
	}

	return session, nil
}

func (s *DriverService) Offline(ctx context.Context, driverID uuid.UUID) error {
	err := s.spec.Offline(ctx, driverID)
	if err != nil {
		return err
	}
	err = statusOffline(ctx, s.queries, driverID)
	return err
}

func (s *DriverService) Location(ctx context.Context, args models.LocationUpdateRequest) error {
	if s.spec != nil {
		if err := s.spec.UpdateLocation(ctx, args); err != nil {
			return err
		}
	}

	current, err := s.queries.GetDriverCurrentLocation(ctx, args.DriverID)
	if err == nil {
		if time.Since(current.UpdatedAt) < locationRateLimit {
			return appErrors.NewConflictError("location updates are too frequent")
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	qtx := s.queries.WithTx(tx)

	if err = qtx.MarkDriverCoordinatesAsOld(ctx, args.DriverID); err != nil {
		return err
	}

	_, err = qtx.CreateCoordinateForDriver(ctx, sqlc.CreateCoordinateForDriverParams{
		EntityID:  args.DriverID,
		Address:   "",
		Latitude:  sqlc.NumericFromFloat(args.Latitude),
		Longitude: sqlc.NumericFromFloat(args.Longitude),
	})
	if err != nil {
		return err
	}

	if args.RideID != nil {
		err = qtx.CreateLocationHistory(ctx, sqlc.CreateLocationHistoryParams{
			DriverID:       args.DriverID,
			Latitude:       sqlc.NumericFromFloat(args.Latitude),
			Longitude:      sqlc.NumericFromFloat(args.Longitude),
			AccuracyMeters: sqlc.NumericFromFloat(args.AccuracyMeters),
			SpeedKmh:       sqlc.NumericFromFloat(args.SpeedKmh),
			HeadingDegrees: sqlc.NumericFromFloat(args.HeadingDegrees),
			RideID:         *args.RideID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *DriverService) Start(ctx context.Context, driverID uuid.UUID) error {
	return nil

}

func (s *DriverService) Complete(ctx context.Context, driverID uuid.UUID) error {

	return nil
}
