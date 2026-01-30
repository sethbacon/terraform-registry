package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/terraform-registry/terraform-registry/internal/config"
)

// NewRouter creates and configures the Gin router
func NewRouter(cfg *config.Config, db *sql.DB) *gin.Engine {
	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(LoggerMiddleware(cfg))
	router.Use(CORSMiddleware(cfg))

	// Health check endpoint
	router.GET("/health", healthCheckHandler(db))

	// Readiness check endpoint
	router.GET("/ready", readinessHandler(db))

	// Service discovery endpoint (Terraform protocol)
	router.GET("/.well-known/terraform.json", serviceDiscoveryHandler(cfg))

	// API version
	router.GET("/version", versionHandler())

	// Module Registry endpoints (v1)
	v1Modules := router.Group("/v1/modules")
	{
		v1Modules.GET("/:namespace/:name/:system/versions", func(c *gin.Context) {
			c.JSON(http.StatusNotImplemented, gin.H{
				"error": "Module Registry endpoints coming in Phase 2",
			})
		})
		v1Modules.GET("/:namespace/:name/:system/:version/download", func(c *gin.Context) {
			c.JSON(http.StatusNotImplemented, gin.H{
				"error": "Module Registry endpoints coming in Phase 2",
			})
		})
	}

	// Provider Registry endpoints (v1)
	// These are for the standard Provider Registry Protocol
	v1Providers := router.Group("/v1/providers")
	{
		v1Providers.GET("/:namespace/:type/versions", func(c *gin.Context) {
			c.JSON(http.StatusNotImplemented, gin.H{
				"error": "Provider Registry endpoints coming in Phase 3",
			})
		})
		v1Providers.GET("/:namespace/:type/:version/download/:os/:arch", func(c *gin.Context) {
			c.JSON(http.StatusNotImplemented, gin.H{
				"error": "Provider Registry endpoints coming in Phase 3",
			})
		})
	}

	// Network Mirror endpoints (separate from Provider Registry to avoid routing conflicts)
	// These endpoints include the hostname of the origin registry as per the Network Mirror Protocol
	// They use a different path structure: /terraform/providers/:hostname/:namespace/:type/...
	v1Mirror := router.Group("/terraform/providers")
	{
		v1Mirror.GET("/:hostname/:namespace/:type/index.json", func(c *gin.Context) {
			c.JSON(http.StatusNotImplemented, gin.H{
				"error": "Network Mirror endpoints coming in Phase 3",
			})
		})
		v1Mirror.GET("/:hostname/:namespace/:type/:version.json", func(c *gin.Context) {
			c.JSON(http.StatusNotImplemented, gin.H{
				"error": "Network Mirror endpoints coming in Phase 3",
			})
		})
	}

	// Admin API endpoints
	apiV1 := router.Group("/api/v1")
	{
		// Modules admin endpoints
		apiV1.GET("/modules", func(c *gin.Context) {
			c.JSON(http.StatusNotImplemented, gin.H{
				"error": "Admin API endpoints coming in Phase 5",
			})
		})

		// Providers admin endpoints
		apiV1.GET("/providers", func(c *gin.Context) {
			c.JSON(http.StatusNotImplemented, gin.H{
				"error": "Admin API endpoints coming in Phase 5",
			})
		})

		// Users admin endpoints
		apiV1.GET("/users", func(c *gin.Context) {
			c.JSON(http.StatusNotImplemented, gin.H{
				"error": "Admin API endpoints coming in Phase 4",
			})
		})

		// Organizations admin endpoints
		apiV1.GET("/organizations", func(c *gin.Context) {
			c.JSON(http.StatusNotImplemented, gin.H{
				"error": "Admin API endpoints coming in Phase 4",
			})
		})
	}

	return router
}

// healthCheckHandler returns the health status of the service
func healthCheckHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check database connection
		if err := db.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  "database connection failed",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now().UTC().Format(time.RFC3339),
		})
	}
}

// readinessHandler returns the readiness status of the service
func readinessHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check database connection
		if err := db.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"ready": false,
				"error": "database not ready",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"ready": true,
			"time":  time.Now().UTC().Format(time.RFC3339),
		})
	}
}

// serviceDiscoveryHandler implements Terraform service discovery
func serviceDiscoveryHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"modules.v1":   cfg.Server.BaseURL + "/v1/modules/",
			"providers.v1": cfg.Server.BaseURL + "/v1/providers/",
		})
	}
}

// versionHandler returns the API version
func versionHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version": "0.1.0",
			"api_version": "v1",
			"protocols": gin.H{
				"modules":  "v1",
				"providers": "v1",
				"mirror":    "v1",
			},
		})
	}
}

// LoggerMiddleware provides structured logging
func LoggerMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)

		// Log the request
		if cfg.Logging.Format == "json" {
			logJSON(c, latency, path, query)
		} else {
			logText(c, latency, path, query)
		}
	}
}

// logJSON logs in JSON format
func logJSON(c *gin.Context, latency time.Duration, path, query string) {
	// This is a simple JSON logger. In production, use a proper structured logging library like zap or zerolog
	// For now, we'll use gin's default logger format
	// In Phase 9, we'll implement proper structured logging with zap/zerolog
	_ = latency
	_ = path
	_ = query
}

// logText logs in plain text format
func logText(c *gin.Context, latency time.Duration, path, query string) {
	// Use gin's default logger for now
	_ = latency
	_ = path
	_ = query
}

// CORSMiddleware handles CORS
func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range cfg.Security.CORS.AllowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}

		if allowed {
			if origin == "" {
				c.Header("Access-Control-Allow-Origin", "*")
			} else {
				c.Header("Access-Control-Allow-Origin", origin)
			}
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
			c.Header("Access-Control-Max-Age", "3600")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
