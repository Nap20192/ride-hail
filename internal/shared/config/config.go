package config

import (
	"fmt"
	"os"
	"strconv"

	"ride-hail/pkg/utils"
)

// Config holds all service configurations
type Config struct {
	Database  DatabaseConfig
	RabbitMQ  RabbitMQConfig
	LogLevel  string
	Env       string
	Ports     Ports
	WebSocket WebSocketConfig
}

// DatabaseConfig holds database connection parameters
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

// RabbitMQConfig holds RabbitMQ connection parameters
type RabbitMQConfig struct {
	Host     string
	Port     string
	User     string
	Password string
}

// WebSocketConfig holds WebSocket server configuration
type WebSocketConfig struct {
	Port int
}

// Ports holds service port configurations
type Ports struct {
	RideService           int
	DriverLocationService int
	AdminService          int
}

func (s Ports) Ride() string {
	return fmt.Sprintf("%d", s.RideService)
}

func (s Ports) DriverLocation() string {
	return fmt.Sprintf("%d", s.DriverLocationService)
}

func (s Ports) Admin() string {
	return fmt.Sprintf("%d", s.AdminService)
}

// LoadConfig loads configuration from environment variables or config.yaml
func LoadConfig() (*Config, error) {
	// Try to load from config.yaml first
	if _, err := os.Stat("config.yaml"); err == nil {
		return LoadConfigFromYAML("config.yaml")
	}

	// Fall back to environment variables
	return LoadConfigFromEnv()
}

// LoadConfigFromYAML loads configuration from a YAML file
func LoadConfigFromYAML(path string) (*Config, error) {
	data, err := utils.ParseYAML(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	cfg := &Config{}

	// Parse database config
	if db, ok := data["database"].(map[string]interface{}); ok {
		cfg.Database.Host = getStringFromMap(db, "host", "localhost")
		cfg.Database.Port = getStringFromMap(db, "port", "5432")
		cfg.Database.User = getStringFromMap(db, "user", "postgres")
		cfg.Database.Password = getStringFromMap(db, "password", "postgres")
		cfg.Database.Database = getStringFromMap(db, "database", "ride-hail")
	}

	// Parse RabbitMQ config
	if mq, ok := data["rabbitmq"].(map[string]interface{}); ok {
		cfg.RabbitMQ.Host = getStringFromMap(mq, "host", "localhost")
		cfg.RabbitMQ.Port = getStringFromMap(mq, "port", "5672")
		cfg.RabbitMQ.User = getStringFromMap(mq, "user", "guest")
		cfg.RabbitMQ.Password = getStringFromMap(mq, "password", "guest")
	}

	// Parse WebSocket config
	if ws, ok := data["websocket"].(map[string]interface{}); ok {
		cfg.WebSocket.Port = getIntFromMap(ws, "port", 8080)
	}

	// Parse services config
	if services, ok := data["services"].(map[string]interface{}); ok {
		cfg.Ports.RideService = getIntFromMap(services, "ride_service", 3000)
		cfg.Ports.DriverLocationService = getIntFromMap(services, "driver_location_service", 3001)
		cfg.Ports.AdminService = getIntFromMap(services, "admin_service", 3004)
	}

	// Parse application config
	cfg.LogLevel = getStringFromMap(data, "log_level", "INFO")
	cfg.Env = getStringFromMap(data, "environment", "development")

	return cfg, nil
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv() (*Config, error) {
	// Parse integer ports
	wsPort, err := strconv.Atoi(utils.GetEnv("WS_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid WS_PORT: %w", err)
	}

	ridePort, err := strconv.Atoi(utils.GetEnv("RIDE_SERVICE_PORT", "3000"))
	if err != nil {
		return nil, fmt.Errorf("invalid RIDE_SERVICE_PORT: %w", err)
	}

	driverPort, err := strconv.Atoi(utils.GetEnv("DRIVER_LOCATION_SERVICE_PORT", "3001"))
	if err != nil {
		return nil, fmt.Errorf("invalid DRIVER_LOCATION_SERVICE_PORT: %w", err)
	}

	adminPort, err := strconv.Atoi(utils.GetEnv("ADMIN_SERVICE_PORT", "3004"))
	if err != nil {
		return nil, fmt.Errorf("invalid ADMIN_SERVICE_PORT: %w", err)
	}

	return &Config{
		Database: DatabaseConfig{
			Host:     utils.GetEnv("DB_HOST", "localhost"),
			Port:     utils.GetEnv("DB_PORT", "5432"),
			User:     utils.GetEnv("DB_USER", "postgres"),
			Password: utils.GetEnv("DB_PASSWORD", "postgres"),
			Database: utils.GetEnv("DB_NAME", "ride_hail"),
		},
		RabbitMQ: RabbitMQConfig{
			Host:     utils.GetEnv("RABBITMQ_HOST", "localhost"),
			Port:     utils.GetEnv("RABBITMQ_PORT", "5672"),
			User:     utils.GetEnv("RABBITMQ_USER", "guest"),
			Password: utils.GetEnv("RABBITMQ_PASSWORD", "guest"),
		},
		WebSocket: WebSocketConfig{
			Port: wsPort,
		},
		Ports: Ports{
			RideService:           ridePort,
			DriverLocationService: driverPort,
			AdminService:          adminPort,
		},
		LogLevel: utils.GetEnv("LOG_LEVEL", "INFO"),
		Env:      utils.GetEnv("ENVIRONMENT", "development"),
	}, nil
}

// Helper functions to safely extract values from maps
func getStringFromMap(m map[string]interface{}, key string, defaultVal string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultVal
}

func getIntFromMap(m map[string]interface{}, key string, defaultVal int) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return defaultVal
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.Database == "" {
		return fmt.Errorf("database name is required")
	}
	if c.RabbitMQ.Host == "" {
		return fmt.Errorf("rabbitmq host is required")
	}
	if c.WebSocket.Port < 1 || c.WebSocket.Port > 65535 {
		return fmt.Errorf("websocket port must be between 1 and 65535")
	}
	if c.Ports.RideService < 1 || c.Ports.RideService > 65535 {
		return fmt.Errorf("ride service port must be between 1 and 65535")
	}
	if c.Ports.DriverLocationService < 1 || c.Ports.DriverLocationService > 65535 {
		return fmt.Errorf("driver location service port must be between 1 and 65535")
	}
	if c.Ports.AdminService < 1 || c.Ports.AdminService > 65535 {
		return fmt.Errorf("admin service port must be between 1 and 65535")
	}
	return nil
}

// DatabaseConnString returns the PostgreSQL connection string
func (c *Config) DatabaseConnString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Database,
	)
}

// RabbitMQURL returns the RabbitMQ connection URL
func (c *Config) RabbitMQURL() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/",
		c.RabbitMQ.User,
		c.RabbitMQ.Password,
		c.RabbitMQ.Host,
		c.RabbitMQ.Port,
	)
}
