package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dssr1012/pulse-expends/internal/config"
	"github.com/dssr1012/pulse-expends/internal/mcp"
	"github.com/dssr1012/pulse-expends/internal/repository"
	"github.com/dssr1012/pulse-expends/pkg/obs"
	"github.com/dssr1012/pulse-expends/pkg/postgres"
	"github.com/rs/zerolog"
)

func main() {
	// Initialize logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load configuration")
	}

	logger.Info().Msg("Starting PulseExpends MCP Server")
	logger.Info().Str("version", "1.0.0").Str("phase", "1").Msg("Initializing Phase 1: OBS Storage")

	// Initialize repository manager based on configuration
	var repoManager *repository.RepositoryManager
	
	if cfg.Storage.Type == "postgres" {
		// Phase 2: PostgreSQL storage
		logger.Info().Msg("Initializing PostgreSQL repository")
		
		// Create PostgreSQL connection string
		connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			cfg.PostgreSQL.Host,
			cfg.PostgreSQL.Port,
			cfg.PostgreSQL.User,
			cfg.PostgreSQL.Password,
			cfg.PostgreSQL.Database,
		)
		
		postgresRepo, err := postgres.NewPostgresRepository(connStr)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to initialize PostgreSQL repository")
		}
		
		repoManager = repository.NewRepositoryManager(
			postgres.NewPostgresTransactionRepository(postgresRepo.db),
			postgres.NewPostgresCircleRepository(postgresRepo.db),
			postgres.NewPostgresUserRepository(postgresRepo.db),
			postgres.NewPostgresAuthRepository(postgresRepo.db),
		)
		
		logger.Info().Msg("PostgreSQL repository initialized successfully")
		
	} else {
		// Phase 1: OBS storage (default)
		logger.Info().Msg("Initializing OBS repository")
		
		obsRepo, err := obs.NewOBSRepository(
			cfg.OBS.Endpoint,
			cfg.OBS.AccessKeyID,
			cfg.OBS.SecretAccessKey,
			cfg.OBS.BucketName,
			cfg.OBS.EncryptionKey,
		)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to initialize OBS repository")
		}
		
		repoManager = repository.NewRepositoryManager(
			obs.NewOBSTransactionRepository(obsRepo),
			obs.NewOBSCircleRepository(obsRepo),
			nil, // User repository not implemented for OBS yet
			nil, // Auth repository not implemented for OBS yet
		)
		
		logger.Info().Str("bucket", cfg.OBS.BucketName).Str("region", cfg.OBS.Region).Msg("OBS repository initialized successfully")
	}

	// Initialize MCP server
	mcpServer := mcp.NewMCPServer(repoManager)
	
	// Start HTTP server
	server := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      mcpServer,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start server in goroutine
	go func() {
		logger.Info().Str("address", cfg.Server.Address).Msg("Starting MCP server")
		
		if cfg.Server.TLS.Enabled {
			logger.Info().Msg("Starting server with TLS")
			if err := server.ListenAndServeTLS(cfg.Server.TLS.CertFile, cfg.Server.TLS.KeyFile); err != nil && err != http.ErrServerClosed {
				logger.Fatal().Err(err).Msg("Server failed to start with TLS")
			}
		} else {
			logger.Info().Msg("Starting server without TLS (HTTP)")
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatal().Err(err).Msg("Server failed to start")
			}
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()
	logger.Info().Msg("Shutting down server...")

	// Give outstanding requests 5 seconds to complete
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("Server forced to shutdown")
	}

	// Close repository connections
	if err := repoManager.CloseAll(); err != nil {
		logger.Error().Err(err).Msg("Failed to close repository connections")
	}

	logger.Info().Msg("Server stopped gracefully")
}

// Helper function to get configuration
func getConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			Address: ":8080",
			TLS: config.TLSConfig{
				Enabled:  false,
				CertFile: "",
				KeyFile:  "",
			},
			AllowedOrigins: []string{"*"},
		},
		Storage: config.StorageConfig{
			Type: "obs", // Default to OBS for Phase 1
		},
		OBS: config.OBSConfig{
			Endpoint:        getEnv("OBS_ENDPOINT", "https://obs.la-south-2.myhuaweicloud.com"),
			AccessKeyID:     getEnv("OBS_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("OBS_SECRET_ACCESS_KEY", ""),
			BucketName:      getEnv("OBS_BUCKET_NAME", "pulse-expends"),
			Region:          getEnv("OBS_REGION", "la-south-2"),
			EncryptionKey:   getEnv("OBS_ENCRYPTION_KEY", ""),
		},
		PostgreSQL: config.PostgreSQLConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnvAsInt("POSTGRES_PORT", 5432),
			User:     getEnv("POSTGRES_USER", "pulseexpends"),
			Password: getEnv("POSTGRES_PASSWORD", ""),
			Database: getEnv("POSTGRES_DB", "pulseexpends"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
		},
		PythonService: config.PythonServiceConfig{
			URL: getEnv("PYTHON_SERVICE_URL", "http://localhost:8000"),
		},
		JWT: config.JWTConfig{
			Secret:           getEnv("JWT_SECRET", "change-this-in-production"),
			ExpirationHours:  getEnvAsInt("JWT_EXPIRATION_HOURS", 24),
			RefreshExpirationHours: getEnvAsInt("JWT_REFRESH_EXPIRATION_HOURS", 168),
		},
	}
}

// Helper function to get environment variable with default
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// Helper function to get environment variable as int with default
func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// Import fmt and strconv
import (
	"fmt"
	"strconv"
)