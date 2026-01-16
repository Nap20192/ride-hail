# True Service-Oriented Architecture (SOA) Structure

## Overview
This document describes the current SOA structure of the ride-hail platform, where each service owns its own data models and database queries.

## Architecture Principles

### ✅ Service Independence
- Each service has its own models package
- No shared database models between services
- Services communicate via events/APIs, not shared data structures

### ✅ Service-Specific Models
Each service owns only the database tables it needs:

#### Ride Service (`services/ride/models/`)
- `rides.sql.go` - Ride lifecycle operations
- `passengers.sql.go` - Passenger data (read-only for this service)
- `ride_events.sql.go` - Event sourcing for rides
- `coordinates.sql.go` - Location data for rides

#### Driver Service (`services/driver/models/`)
- `drivers.sql.go` - Driver profile and operations
- `driver_sessions.sql.go` - Active driver sessions
- `location_history.sql.go` - Driver location tracking
- `matching.sql.go` - Driver-ride matching queries
- `coordinates.sql.go` - Location data for drivers

#### Admin Service (`services/admin/models/`)
- `admin.sql.go` - Read-only admin queries
- `analytics.sql.go` - System analytics and reports
- `enums.sql.go` - Enum types for UI/reporting

### ✅ Truly Shared Types
Only authentication and user management are shared (`shared/types/`):
- `auth.sql.go` - Authentication queries
- `users.sql.go` - User management
- Shared because all services need to authenticate users

## Directory Structure

```
services/
├── ride/
│   ├── api/router/      # HTTP endpoints
│   ├── domain/          # Business logic
│   ├── events/          # Event handlers
│   ├── models/          # Database queries (SQLC generated)
│   ├── queries/         # SQL source files for SQLC
│   └── repository/      # Data access layer
├── driver/
│   ├── api/             # HTTP endpoints
│   ├── domain/          # Business logic
│   ├── events/          # Event handlers
│   ├── models/          # Database queries (SQLC generated)
│   ├── queries/         # SQL source files for SQLC
│   └── repository/      # Data access layer
└── admin/
    ├── api/             # HTTP endpoints
    ├── models/          # Database queries (SQLC generated)
    └── queries/         # SQL source files for SQLC

shared/
├── config/              # Configuration loading
├── events/              # Event type definitions
├── queries/             # SQL files for auth/users
└── types/               # Auth/User models (truly shared)

pkg/
├── auth/                # Authentication utilities
├── middleware/          # HTTP middleware
├── mq/                  # RabbitMQ client
├── postgres/            # PostgreSQL client
├── server/              # HTTP/WebSocket server
├── services/            # Service wrappers (DB, HTTP, MQ, WS)
└── ...                  # Other infrastructure
```

## Key Implementation Details

### Service-Specific Queriers
Each service has its own `Querier` interface with only the methods it needs:

**Ride Service** - 35 methods for rides, passengers, events, coordinates
**Driver Service** - 29 methods for drivers, sessions, locations, matching
**Admin Service** - 19 methods for analytics and read-only queries

### No Shared Database Models
- ❌ No `shared/database/` directory
- ❌ No monolithic Querier interface with all methods
- ✅ Each service imports only `services/{service}/models`
- ✅ Services use events/APIs for cross-service data needs

### Cross-Service Communication
When a service needs data from another service:
1. **Events** - Publish/subscribe via RabbitMQ
2. **API calls** - HTTP requests to other services
3. **Event sourcing** - Read from event log if needed

Example: Ride service needs driver location
- ❌ Don't import `services/driver/models`
- ✅ Subscribe to `driver.location.updated` event
- ✅ Or call driver-service API

## Benefits of This Structure

1. **Service Autonomy** - Each service can evolve independently
2. **Clear Boundaries** - No hidden dependencies via shared models
3. **Scalability** - Services can be deployed/scaled separately
4. **Maintainability** - Changes to one service don't break others
5. **Testing** - Services can be tested in isolation

## Compilation Verification

All services compile successfully:
```bash
go build -o bin/ride-service ./cmd/ride-service     ✅
go build -o bin/driver-service ./cmd/driver-service ✅
go build -o bin/admin-service ./cmd/admin-service   ✅
```

## Migration from Previous Structure

### Before (Monolithic Shared Database & Queries)
```
internal/services/        # Service wrappers mixed with business logic
shared/database/          # All SQLC models in one place
├── rides.sql.go
├── drivers.sql.go
├── admin.sql.go
├── querier.go           # Monolithic interface with ALL methods
└── ...
queries/                  # All SQL files in one place
├── rides.sql
├── drivers.sql
├── admin.sql
└── ...
```

### After (True SOA)
```
services/ride/
├── models/              # Only ride-related SQLC models
└── queries/             # Only ride-related SQL files

services/driver/
├── models/              # Only driver-related SQLC models
└── queries/             # Only driver-related SQL files

services/admin/
├── models/              # Only admin-related SQLC models
└── queries/             # Only admin-related SQL files

shared/
├── types/               # Only auth/user models (truly shared)
└── queries/             # Only auth/user SQL files

pkg/services/            # Reusable service wrappers (infrastructure)
```

## Future Enhancements

1. **API Gateway** - Add gateway for external clients
2. **Service Discovery** - Consider consul/etcd for service registry
3. **Distributed Tracing** - Add OpenTelemetry for request tracing
4. **Circuit Breakers** - Add resilience for inter-service calls
5. **Separate Databases** - Each service could have its own database instance

## Compliance

✅ **Service-Oriented Architecture (SOA)** compliance verified
- Independent services with clear boundaries
- Service-specific data models
- Event-driven communication
- Infrastructure as shared utilities
- Configuration as shared resources
