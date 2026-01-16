-- name: UpdateDriverStatus :exec
UPDATE drivers
SET status = $1, updated_at = NOW()
WHERE id = $2;

-- name: CreateDriverSession :one
INSERT INTO driver_sessions (driver_id, started_at)
VALUES ($1, NOW())
RETURNING *;

-- name: EndDriverSession :one
UPDATE driver_sessions
SET ended_at = NOW(), total_rides = $2, total_earnings = $3
WHERE id = $1
RETURNING *;

-- name: GetCurrentDriverSession :one
SELECT * FROM driver_sessions
WHERE driver_id = $1 AND ended_at IS NULL
ORDER BY started_at DESC
LIMIT 1;

-- name: CreateCoordinateForDriver :one
INSERT INTO coordinates (
  entity_id, entity_type, address,
  latitude, longitude, is_current
) VALUES ($1, 'driver', $2, $3, $4, true)
RETURNING *;

-- name: MarkDriverCoordinatesAsOld :exec
UPDATE coordinates
SET is_current = false, updated_at = NOW()
WHERE entity_id = $1 AND entity_type = 'driver' AND is_current = true;

-- name: GetDriverCurrentLocation :one
SELECT * FROM coordinates
WHERE entity_id = $1 AND entity_type = 'driver' AND is_current = true
LIMIT 1;

-- name: CreateLocationHistory :exec
INSERT INTO location_history (
  driver_id, latitude, longitude,
  accuracy_meters, speed_kmh, heading_degrees,
  ride_id, recorded_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW());

-- name: FindNearbyDrivers :many
SELECT d.id, d.vehicle_type, d.rating, d.status, u.email,
       d.vehicle_attrs,
       c.latitude, c.longitude,
       ST_Distance(
         ST_MakePoint(c.longitude, c.latitude)::geography,
         ST_MakePoint($1, $2)::geography
       ) / 1000 as distance_km
FROM drivers d
JOIN users u ON d.id = u.id
JOIN coordinates c ON c.entity_id = d.id
  AND c.entity_type = 'driver'
  AND c.is_current = true
WHERE d.status = 'AVAILABLE'
  AND d.vehicle_type = $3
  AND d.is_verified = true
  AND ST_DWithin(
    ST_MakePoint(c.longitude, c.latitude)::geography,
    ST_MakePoint($1, $2)::geography,
    $4 * 1000
  )
ORDER BY distance_km, d.rating DESC
LIMIT $5;

-- name: UpdateDriverRide :exec
UPDATE drivers
SET updated_at = NOW()
WHERE id = $1;


-- name: UpdateRideStatus :exec
UPDATE rides
SET status = $1, updated_at = NOW()
WHERE id = $2;

-- name: UpdateRideMatched :exec
UPDATE rides
SET status = 'MATCHED',
    driver_id = $1,
    matched_at = NOW(),
    updated_at = NOW()
WHERE id = $2;

-- name: UpdateRideStarted :exec
UPDATE rides
SET status = 'IN_PROGRESS',
    started_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateRideCompleted :one
UPDATE rides
SET status = 'COMPLETED',
    completed_at = NOW(),
    final_fare = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateDriverStats :exec
UPDATE drivers
SET total_rides = total_rides + 1,
    total_earnings = total_earnings + $2,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateSessionStats :exec
UPDATE driver_sessions
SET total_rides = total_rides + 1,
    total_earnings = total_earnings + $2
WHERE id = $1;
