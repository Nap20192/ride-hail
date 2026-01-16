begin;

-- Drop tables in reverse order (respecting foreign key dependencies)
drop table if exists location_history;
drop table if exists driver_sessions;
drop table if exists drivers;
drop table if exists "driver_status";

commit;
