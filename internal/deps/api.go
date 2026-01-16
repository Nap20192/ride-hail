package deps

import (
	"fmt"

	"ride-hail/internal/services/admin"
	driver "ride-hail/internal/services/driver/api"
	"ride-hail/internal/services/ride"
	"ride-hail/internal/shared/config"
)

type ApiDeps struct {
	RideApi   ride.RideApi
	DriverApi driver.DriverApi
	AdminApi  admin.AdminApi
}

type apiOption func(*ApiDeps) error

func NewApiDeps(opts ...apiOption) (*ApiDeps, error) {
	deps := &ApiDeps{}

	for _, opt := range opts {
		if err := opt(deps); err != nil {
			return nil, err
		}
	}

	return deps, nil
}

func WithRideApi(app *AppDeps, config config.Config) apiOption {
	return func(deps *ApiDeps) error {
		if app.RideService == nil || app.AuthService == nil {
			return fmt.Errorf("missing dependencies for RideApi")
		}
		deps.RideApi = *ride.NewRideApi(app.AuthService, app.RideService, config.Ports.Ride())
		return nil
	}
}

func WithDriverApi(app *AppDeps, config config.Config) apiOption {
	return func(deps *ApiDeps) error {
		if app.DriverService == nil || app.AuthService == nil {
			return fmt.Errorf("missing dependencies for DriverApi")
		}
		deps.DriverApi = *driver.NewDriverApi(app.AuthService, app.DriverService, config.Ports.DriverLocation())
		return nil
	}
}

func WithAdminApi(app *AppDeps, config config.Config) apiOption {
	return func(deps *ApiDeps) error {
		if app.AdminService == nil || app.AuthService == nil {
			return fmt.Errorf("missing dependencies for AdminApi")
		}
		deps.AdminApi = *admin.NewAdminApi(app.AuthService, app.AdminService, config.Ports.Admin())
		return nil
	}
}
