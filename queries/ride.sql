-- name: CreateRide :one
insert into rides (
    ride_number,
    passenger_id,
    vehicle_type,
    status,
    estimated_fare,
    pickup_coordinate_id,
    destination_coordinate_id
) values (
    $1,
    $2,
    $3,
    'REQUESTED',
    $4,
    $5,
    $6
)
returning *;

-- name: CreateCoordinate :one
insert into coordinates (
    entity_id,
    entity_type,
    address,
    latitude,
    longitude,
    distance_km,
    duration_minutes,
    fare_amount,
    is_current
) values (
    $1, $2, $3, $4, $5, $6, $7, $8, false
)
returning *;


-- name: CreateRideEvent :exec
insert into ride_events (
    ride_id,
    event_type,
    event_data
) values (
    $1, $2, $3
);

-- name: GetRideByID :one
select * from rides
where id = $1
limit 1;

-- name: CancelRide :one
update rides
set
    status = 'CANCELLED',
    cancellation_reason = $2,
    cancelled_at = now(),
    updated_at = now()
where id = $1
  and status != 'COMPLETED'
  and status != 'CANCELLED'
returning *;

-- name: IncrementRideCounter :one
INSERT INTO ride_counters (day, counter)
VALUES (@date::date, 1)
ON CONFLICT (day) DO UPDATE
  SET counter = ride_counters.counter + 1
RETURNING day,counter;
