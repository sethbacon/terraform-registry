package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/terraform-registry/terraform-registry/internal/api"
	"github.com/terraform-registry/terraform-registry/internal/config"
	"github.com/terraform-registry/terraform-registry/internal/db"
)

const (
	version = "0.1.0"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Error: %v\n", err)
	}
}

func run() error {
	// Parse command from args
	command := "serve"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Execute command
	switch command {
	case "serve":
		return serve(cfg)
	case "migrate":
		if len(os.Args) < 3 {
			return fmt.Errorf("usage: %s migrate <up|down>", os.Args[0])
		}
		return runMigrations(cfg, os.Args[2])
	case "version":
		fmt.Printf("Terraform Registry v%s\n", version)
		return nil
	default:
		return fmt.Errorf("unknown command: %s\nAvailable commands: serve, migrate, version", command)
	}
}

func serve(cfg *config.Config) error {
	// Set Gin mode
	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Debug: Print database configuration (mask password)
	maskedPassword := "****"
	if cfg.Database.Password != "" {
		maskedPassword = cfg.Database.Password[:1] + "****"
	}
	log.Printf("Database config: host=%s, port=%d, user=%s, password=%s, dbname=%s, sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, maskedPassword,
		cfg.Database.Name, cfg.Database.SSLMode)
	log.Printf("Full DSN (masked): %s", cfg.Database.GetDSN())

	// Connect to database
	database, err := db.Connect(cfg.Database.GetDSN(), cfg.Database.MaxConnections)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer database.Close()

	log.Println("Connected to database successfully")

	// Run migrations automatically on startup
	log.Println("Running database migrations...")
	if err := db.RunMigrations(database, "up"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	log.Println("Database migrations completed successfully")

	// Get migration version
	version, dirty, err := db.GetMigrationVersion(database)
	if err != nil {
		log.Printf("Warning: failed to get migration version: %v", err)
	} else {
		log.Printf("Database schema version: %d (dirty: %v)", version, dirty)
	}

	// Create router
	router := api.NewRouter(cfg, database)

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.Server.GetAddress(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on %s", cfg.Server.GetAddress())
		log.Printf("Base URL: %s", cfg.Server.BaseURL)
		log.Printf("Storage backend: %s", cfg.Storage.DefaultBackend)
		log.Printf("Multi-tenancy: %v", cfg.MultiTenancy.Enabled)
		log.Println("Server is ready to accept connections")

		var err error
		if cfg.Security.TLS.Enabled {
			log.Printf("TLS enabled: cert=%s, key=%s", cfg.Security.TLS.CertFile, cfg.Security.TLS.KeyFile)
			err = server.ListenAndServeTLS(cfg.Security.TLS.CertFile, cfg.Security.TLS.KeyFile)
		} else {
			err = server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Println("Server stopped gracefully")
	return nil
}

func runMigrations(cfg *config.Config, direction string) error {
	// Connect to database
	database, err := db.Connect(cfg.Database.GetDSN(), cfg.Database.MaxConnections)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer database.Close()

	log.Printf("Running migrations: %s", direction)

	// Run migrations
	if err := db.RunMigrations(database, direction); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	// Get current version
	version, dirty, err := db.GetMigrationVersion(database)
	if err != nil {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	log.Printf("Migration completed successfully. Current version: %d (dirty: %v)", version, dirty)
	return nil
}
