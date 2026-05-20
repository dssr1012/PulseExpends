package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server        ServerConfig        `json:"server"`
	Storage       StorageConfig       `json:"storage"`
	OBS           OBSConfig           `json:"obs"`
	PostgreSQL    PostgreSQLConfig    `json:"postgresql"`
	PythonService PythonServiceConfig `json:"python_service"`
	JWT           JWTConfig           `json:"jwt"`
	Logging       LoggingConfig       `json:"logging"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Address        string        `json:"address"`
	TLS            TLSConfig     `json:"tls"`
	AllowedOrigins []string      `json:"allowed_origins"`
	ReadTimeout    time.Duration `json:"read_timeout"`
	WriteTimeout   time.Duration `json:"write_timeout"`
	IdleTimeout    time.Duration `json:"idle_timeout"`
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	Enabled  bool   `json:"enabled"`
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"key_file"`
}

// StorageConfig holds storage configuration
type StorageConfig struct {
	Type string `json:"type"` // "obs" or "postgres"
}

// OBSConfig holds Huawei Cloud OBS configuration
type OBSConfig struct {
	Endpoint        string `json:"endpoint"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	BucketName      string `json:"bucket_name"`
	Region          string `json:"region"`
	EncryptionKey   string `json:"encryption_key"`
}

// PostgreSQLConfig holds PostgreSQL configuration
type PostgreSQLConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	SSLMode  string `json:"ssl_mode"`
	MaxConns int    `json:"max_conns"`
	MaxIdle  int    `json:"max_idle"`
}

// PythonServiceConfig holds Python microservice configuration
type PythonServiceConfig struct {
	URL string `json:"url"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret                  string        `json:"secret"`
	ExpirationHours         int           `json:"expiration_hours"`
	RefreshExpirationHours  int           `json:"refresh_expiration_hours"`
	Issuer                  string        `json:"issuer"`
	Audience                string        `json:"audience"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	Output     string `json:"output"`
	StackTrace bool   `json:"stack_trace"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Address: getEnv("SERVER_ADDRESS", ":8080"),
			TLS: TLSConfig{
				Enabled:  getEnvAsBool("SERVER_TLS_ENABLED", false),
				CertFile: getEnv("SERVER_TLS_CERT_FILE", ""),
				KeyFile:  getEnv("SERVER_TLS_KEY_FILE", ""),
			},
			AllowedOrigins: getEnvAsSlice("SERVER_ALLOWED_ORIGINS", []string{"*"}),
			ReadTimeout:    getEnvAsDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:   getEnvAsDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:    getEnvAsDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
		Storage: StorageConfig{
			Type: getEnv("STORAGE_TYPE", "obs"), // Default to OBS for Phase 1
		},
		OBS: OBSConfig{
			Endpoint:        getEnv("OBS_ENDPOINT", "https://obs.la-south-2.myhuaweicloud.com"),
			AccessKeyID:     getEnv("OBS_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("OBS_SECRET_ACCESS_KEY", ""),
			BucketName:      getEnv("OBS_BUCKET_NAME", "pulse-expends"),
			Region:          getEnv("OBS_REGION", "la-south-2"),
			EncryptionKey:   getEnv("OBS_ENCRYPTION_KEY", ""),
		},
		PostgreSQL: PostgreSQLConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnvAsInt("POSTGRES_PORT", 5432),
			User:     getEnv("POSTGRES_USER", "pulseexpends"),
			Password: getEnv("POSTGRES_PASSWORD", ""),
			Database: getEnv("POSTGRES_DB", "pulseexpends"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
			MaxConns: getEnvAsInt("POSTGRES_MAX_CONNS", 25),
			MaxIdle:  getEnvAsInt("POSTGRES_MAX_IDLE", 25),
		},
		PythonService: PythonServiceConfig{
			URL: getEnv("PYTHON_SERVICE_URL", "http://localhost:8000"),
		},
		JWT: JWTConfig{
			Secret:                 getEnv("JWT_SECRET", "change-this-in-production"),
			ExpirationHours:        getEnvAsInt("JWT_EXPIRATION_HOURS", 24),
			RefreshExpirationHours: getEnvAsInt("JWT_REFRESH_EXPIRATION_HOURS", 168),
			Issuer:                 getEnv("JWT_ISSUER", "pulse-expends"),
			Audience:               getEnv("JWT_AUDIENCE", "pulse-expends-users"),
		},
		Logging: LoggingConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Format:     getEnv("LOG_FORMAT", "json"),
			Output:     getEnv("LOG_OUTPUT", "stdout"),
			StackTrace: getEnvAsBool("LOG_STACK_TRACE", false),
		},
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate server configuration
	if c.Server.Address == "" {
		return fmt.Errorf("server address is required")
	}

	if c.Server.TLS.Enabled {
		if c.Server.TLS.CertFile == "" {
			return fmt.Errorf("TLS cert file is required when TLS is enabled")
		}
		if c.Server.TLS.KeyFile == "" {
			return fmt.Errorf("TLS key file is required when TLS is enabled")
		}
	}

	// Validate storage configuration
	if c.Storage.Type != "obs" && c.Storage.Type != "postgres" {
		return fmt.Errorf("storage type must be either 'obs' or 'postgres'")
	}

	// Validate OBS configuration if using OBS
	if c.Storage.Type == "obs" {
		if c.OBS.Endpoint == "" {
			return fmt.Errorf("OBS endpoint is required")
		}
		if c.OBS.AccessKeyID == "" {
			return fmt.Errorf("OBS access key ID is required")
		}
		if c.OBS.SecretAccessKey == "" {
			return fmt.Errorf("OBS secret access key is required")
		}
		if c.OBS.BucketName == "" {
			return fmt.Errorf("OBS bucket name is required")
		}
		if c.OBS.Region == "" {
			return fmt.Errorf("OBS region is required")
		}
	}

	// Validate PostgreSQL configuration if using PostgreSQL
	if c.Storage.Type == "postgres" {
		if c.PostgreSQL.Host == "" {
			return fmt.Errorf("PostgreSQL host is required")
		}
		if c.PostgreSQL.Port <= 0 || c.PostgreSQL.Port > 65535 {
			return fmt.Errorf("PostgreSQL port must be between 1 and 65535")
		}
		if c.PostgreSQL.User == "" {
			return fmt.Errorf("PostgreSQL user is required")
		}
		if c.PostgreSQL.Password == "" {
			return fmt.Errorf("PostgreSQL password is required")
		}
		if c.PostgreSQL.Database == "" {
			return fmt.Errorf("PostgreSQL database name is required")
		}
	}

	// Validate Python service configuration
	if c.PythonService.URL == "" {
		return fmt.Errorf("Python service URL is required")
	}

	// Validate JWT configuration
	if c.JWT.Secret == "" || c.JWT.Secret == "change-this-in-production" {
		return fmt.Errorf("JWT secret must be set and not use default value")
	}
	if c.JWT.ExpirationHours <= 0 {
		return fmt.Errorf("JWT expiration hours must be positive")
	}
	if c.JWT.RefreshExpirationHours <= 0 {
		return fmt.Errorf("JWT refresh expiration hours must be positive")
	}

	return nil
}

// GetStorageType returns the storage type
func (c *Config) GetStorageType() string {
	return c.Storage.Type
}

// IsOBSStorage returns true if using OBS storage
func (c *Config) IsOBSStorage() bool {
	return c.Storage.Type == "obs"
}

// IsPostgreSQLStorage returns true if using PostgreSQL storage
func (c *Config) IsPostgreSQLStorage() bool {
	return c.Storage.Type == "postgres"
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value, exists := os.LookupEnv(key); exists {
		// Simple comma-separated parsing
		var result []string
		start := 0
		for i, char := range value {
			if char == ',' {
				if i > start {
					result = append(result, value[start:i])
				}
				start = i + 1
			}
		}
		if start < len(value) {
			result = append(result, value[start:])
		}
		return result
	}
	return defaultValue
}

// GetDSN returns the PostgreSQL DSN
func (c *PostgreSQLConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode)
}

// GetOBSConfig returns OBS configuration as a map
func (c *OBSConfig) ToMap() map[string]string {
	return map[string]string{
		"endpoint":         c.Endpoint,
		"access_key_id":    c.AccessKeyID,
		"secret_access_key": c.SecretAccessKey,
		"bucket_name":      c.BucketName,
		"region":           c.Region,
		"encryption_key":   c.EncryptionKey,
	}
}