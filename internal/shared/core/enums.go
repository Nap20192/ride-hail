package core

type RideStatus int8

const (
	RideStatusRequested RideStatus = iota
	RideStatusMatched
	RideStatusEnRoute
	RideStatusArrived
	RideStatusInProgress
	RideStatusCompleted
	RideStatusCancelled
)

func (rs RideStatus) String() string {
	return []string{
		"REQUESTED",
		"MATCHED",
		"ENROUTE",
		"ARRIVED",
		"INPROGRESS",
		"COMPLETED",
		"CANCELLED",
	}[rs]
}

type RideEventType int8

const (
	RideEventRequested RideEventType = iota
	RideEventMatched
	RideEventArrived
	RideEventStarted
	RideEventCompleted
	RideEventCancelled
	RideEventStatusChanged
	RideEventLocationUpdated
	RideEventFareAdjusted
)

func (ret RideEventType) String() string {
	return []string{
		"RIDE_REQUESTED",
		"RIDE_MATCHED",
		"RIDE_ARRIVED",
		"RIDE_STARTED",
		"RIDE_COMPLETED",
		"RIDE_CANCELLED",
		"STATUS_CHANGED",
		"LOCATION_UPDATED",
		"FARE_ADJUSTED",
	}[ret]
}

type DriverStatus int8

const (
	DriverStatusOffline DriverStatus = iota
	DriverStatusAvailable
	DriverStatusBusy
	DriverStatusEnRoute
)

func (ds DriverStatus) String() string {
	return []string{
		"OFFLINE",
		"AVAILABLE",
		"BUSY",
		"EN_ROUTE",
	}[ds]
}

type UserStatus int8

const (
	UserStatusInactive UserStatus = iota
	UserStatusActive
	UserStatusBanned
)

func (us UserStatus) String() string {
	return []string{
		"INACTIVE",
		"ACTIVE",
		"BANNED",
	}[us]
}

type UserRole int8

const (
	UserRolePassenger UserRole = iota
	UserRoleDriver
	UserRoleAdmin
)

func (ur UserRole) String() string {
	return []string{
		"PASSENGER",
		"DRIVER",
		"ADMIN",
	}[ur]
}
