begin;

-- Drop tables in reverse order (respecting foreign key dependencies)
drop table if exists ride_events;
drop table if exists "ride_event_type";
drop table if exists rides;
drop table if exists coordinates;
drop table if exists "vehicle_type";
drop table if exists "ride_status";
drop table if exists users;
drop table if exists "user_status";
drop table if exists "roles";

commit;
