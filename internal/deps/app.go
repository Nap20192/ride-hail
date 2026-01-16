package deps

import (
	"fmt"

	"ride-hail/internal/auth"
	"ride-hail/internal/services/admin"
	"ride-hail/internal/services/driver"
	"ride-hail/internal/services/ride"
	"ride-hail/pkg/mq"
	"ride-hail/pkg/sqlc"
)

type AppDeps struct {
	AuthService   *auth.AuthService
	RideService   *ride.RideService
	DriverService *driver.DriverService
	AdminService  *admin.AdminService
}

type appOption func(*AppDeps) error

func NewAppDeps(opts ...appOption) (*AppDeps, error) {
	deps := &AppDeps{}

	for _, opt := range opts {
		if err := opt(deps); err != nil {
			return nil, err
		}
	}

	return deps, nil
}

func WithAuthService(infra *InfraDeps) appOption {
	return func(deps *AppDeps) error {
		if infra.Pool == nil {
			return fmt.Errorf("missing dependencies for AuthService")
		}
		deps.AuthService = auth.NewAuthService(*sqlc.New(infra.Pool))
		return nil
	}
}

func WithRideService(infra *InfraDeps) appOption {
	return func(deps *AppDeps) error {
		if infra.Pool == nil || infra.RabbitMQ == nil {
			return fmt.Errorf("missing dependencies for RideService")
		}

		queries := sqlc.New(infra.Pool)
		publisher := mq.NewRideEventPublisher(infra.RabbitMQ)
		deps.RideService = ride.NewRideService(infra.Pool, queries, publisher)
		return nil
	}
}

func WithDriverService(infra *InfraDeps) appOption {
	return func(deps *AppDeps) error {
		if infra.Pool == nil || infra.RabbitMQ == nil || deps.AuthService == nil {
			return fmt.Errorf("missing dependencies for DriverService")
		}
		queries := sqlc.New(infra.Pool)
		deps.DriverService = driver.NewDriverService(infra.Pool, queries, infra.RabbitMQ)
		return nil
	}
}

func WithAdminService(infra *InfraDeps) appOption {
	return func(deps *AppDeps) error {
		if infra.Pool == nil || deps.AuthService == nil {
			return fmt.Errorf("missing dependencies for AdminService")
		}
		deps.AdminService = admin.NewAdminService(sqlc.New(infra.Pool))
		return nil
	}
}
