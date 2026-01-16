-- Admin Service Queries

-- name: GetActiveRidesCount :one
SELECT COUNT(*) as count
FROM rides 
WHERE status IN ('MATCHED', 'EN_ROUTE', 'ARRIVED', 'IN_PROGRESS');

-- name: GetAvailableDriversCount :one
SELECT COUNT(*) as count
FROM users 
WHERE role = 'DRIVER' 
  AND attrs->>'driver_status' = 'AVAILABLE';

-- name: GetBusyDriversCount :one
SELECT COUNT(*) as count
FROM users 
WHERE role = 'DRIVER' 
  AND attrs->>'driver_status' = 'BUSY';

-- name: GetTodayRidesCount :one
SELECT COUNT(*) as count
FROM rides 
WHERE DATE(requested_at) = CURRENT_DATE;

-- name: GetTodayRevenue :one
SELECT COALESCE(SUM(final_fare), 0) as total
FROM rides 
WHERE DATE(completed_at) = CURRENT_DATE 
  AND status = 'COMPLETED';

-- name: GetAverageWaitTime :one
SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (matched_at - requested_at)) / 60), 0) as avg_minutes
FROM rides
WHERE matched_at IS NOT NULL
  AND DATE(requested_at) = CURRENT_DATE;

-- name: GetAverageRideDuration :one
SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (completed_at - started_at)) / 60), 0) as avg_minutes
FROM rides
WHERE completed_at IS NOT NULL
  AND started_at IS NOT NULL
  AND DATE(completed_at) = CURRENT_DATE;

-- name: GetCancellationRate :one
SELECT 
  CASE 
    WHEN COUNT(*) = 0 THEN 0
    ELSE (COUNT(*) FILTER (WHERE status = 'CANCELLED')::float / COUNT(*)::float) * 100
  END as rate
FROM rides
WHERE DATE(requested_at) = CURRENT_DATE;

-- name: GetDriverDistributionByVehicleType :many
SELECT 
  COALESCE(attrs->>'vehicle_type', 'UNKNOWN') as vehicle_type,
  COUNT(*) as count
FROM users 
WHERE role = 'DRIVER' 
  AND attrs->>'driver_status' != 'OFFLINE'
GROUP BY attrs->>'vehicle_type';

-- name: GetActiveRidesPaginated :many
SELECT 
  r.id,
  r.ride_number,
  r.status,
  r.passenger_id,
  r.driver_id,
  r.started_at,
  r.requested_at,
  p.email as passenger_email,
  d.email as driver_email,
  r.estimated_fare,
  pickup_coord.address as pickup_address,
  dest_coord.address as destination_address,
  pickup_coord.latitude as pickup_latitude,
  pickup_coord.longitude as pickup_longitude,
  dest_coord.latitude as destination_latitude,
  dest_coord.longitude as destination_longitude
FROM rides r
LEFT JOIN users p ON r.passenger_id = p.id
LEFT JOIN users d ON r.driver_id = d.id
LEFT JOIN coordinates pickup_coord ON r.pickup_coordinate_id = pickup_coord.id
LEFT JOIN coordinates dest_coord ON r.destination_coordinate_id = dest_coord.id
WHERE r.status IN ('MATCHED', 'EN_ROUTE', 'ARRIVED', 'IN_PROGRESS')
ORDER BY r.requested_at DESC
LIMIT $1 OFFSET $2;

-- name: GetActiveRidesTotalCount :one
SELECT COUNT(*) as count
FROM rides
WHERE status IN ('MATCHED', 'EN_ROUTE', 'ARRIVED', 'IN_PROGRESS');
