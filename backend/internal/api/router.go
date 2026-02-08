package api

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/terraform-registry/terraform-registry/internal/api/admin"
	"github.com/terraform-registry/terraform-registry/internal/api/mirror"
	"github.com/terraform-registry/terraform-registry/internal/api/modules"
	"github.com/terraform-registry/terraform-registry/internal/api/providers"
	"github.com/terraform-registry/terraform-registry/internal/api/webhooks"
	"github.com/terraform-registry/terraform-registry/internal/auth"
	"github.com/terraform-registry/terraform-registry/internal/config"
	"github.com/terraform-registry/terraform-registry/internal/crypto"
	"github.com/terraform-registry/terraform-registry/internal/db/repositories"
	"github.com/terraform-registry/terraform-registry/internal/jobs"
	"github.com/terraform-registry/terraform-registry/internal/middleware"
	"github.com/terraform-registry/terraform-registry/internal/services"
	"github.com/terraform-registry/terraform-registry/internal/storage"

	// Import storage backends to register them
	_ "github.com/terraform-registry/terraform-registry/internal/storage/local"
)

// NewRouter creates and configures the Gin router
func NewRouter(cfg *config.Config, db *sql.DB) *gin.Engine {
	router := gin.New()

	// Initialize storage backend
	storageBackend, err := storage.NewStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize storage backend: %v", err)
	}
	log.Printf("Initialized storage backend: %s", cfg.Storage.DefaultBackend)

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	apiKeyRepo := repositories.NewAPIKeyRepository(db)
	moduleRepo := repositories.NewModuleRepository(db)
	providerRepo := repositories.NewProviderRepository(db)
	auditRepo := repositories.NewAuditRepository(db)
	orgRepo := repositories.NewOrganizationRepository(db)

	// Wrap *sql.DB with sqlx for SCM and mirror repositories
	sqlxDB := sqlx.NewDb(db, "postgres")
	scmRepo := repositories.NewSCMRepository(sqlxDB)
	mirrorRepo := repositories.NewMirrorRepository(sqlxDB)

	// Initialize mirror sync job
	mirrorSyncJob := jobs.NewMirrorSyncJob(mirrorRepo, providerRepo, storageBackend)
	// Start background sync job - check every 10 minutes for mirrors that need syncing
	mirrorSyncJob.Start(context.Background(), 10)
	log.Println("Mirror sync job started (checking every 10 minutes)")

	// Get encryption key from environment for OAuth token encryption
	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	if encryptionKey == "" {
		log.Fatal("ENCRYPTION_KEY environment variable must be set for SCM integration")
	}

	// Initialize token cipher for encrypting OAuth tokens
	tokenCipher, err := crypto.NewTokenCipher([]byte(encryptionKey))
	if err != nil {
		log.Fatalf("Failed to initialize token cipher: %v", err)
	}

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(LoggerMiddleware(cfg))
	router.Use(CORSMiddleware(cfg))
	router.Use(middleware.SecurityHeadersMiddleware(middleware.APISecurityHeadersConfig()))

	// Health check endpoint
	router.GET("/health", healthCheckHandler(db))

	// Readiness check endpoint
	router.GET("/ready", readinessHandler(db))

	// Service discovery endpoint (Terraform protocol)
	router.GET("/.well-known/terraform.json", serviceDiscoveryHandler(cfg))

	// API version
	router.GET("/version", versionHandler())

	// Module Registry endpoints (v1) - Terraform Protocol
	// These are public endpoints that support optional authentication
	v1Modules := router.Group("/v1/modules")
	v1Modules.Use(middleware.OptionalAuthMiddleware(cfg, userRepo, apiKeyRepo, orgRepo))
	{
		v1Modules.GET("/:namespace/:name/:system/versions", modules.ListVersionsHandler(db, cfg))
		v1Modules.GET("/:namespace/:name/:system/:version/download", modules.DownloadHandler(db, storageBackend, cfg))
	}

	// File serving endpoint for local storage with ServeDirectly enabled
	router.GET("/v1/files/*filepath", modules.ServeFileHandler(storageBackend, cfg))

	// Provider Registry endpoints (v1)
	// These are for the standard Provider Registry Protocol
	v1Providers := router.Group("/v1/providers")
	v1Providers.Use(middleware.OptionalAuthMiddleware(cfg, userRepo, apiKeyRepo, orgRepo))
	{
		v1Providers.GET("/:namespace/:type/versions", providers.ListVersionsHandler(db, cfg))
		v1Providers.GET("/:namespace/:type/:version/download/:os/:arch", providers.DownloadHandler(db, storageBackend, cfg))
	}

	// Network Mirror endpoints (separate from Provider Registry to avoid routing conflicts)
	// These endpoints include the hostname of the origin registry as per the Network Mirror Protocol
	// They use a different path structure: /terraform/providers/:hostname/:namespace/:type/...
	v1Mirror := router.Group("/terraform/providers")
	{
		v1Mirror.GET("/:hostname/:namespace/:type/index.json", mirror.IndexHandler(db, cfg))
		v1Mirror.GET("/:hostname/:namespace/:type/:versionfile", mirror.PlatformIndexHandler(db, cfg))
	}

	// Initialize admin handlers
	var authHandlers *admin.AuthHandlers
	authHandlers, err = admin.NewAuthHandlers(cfg, db)
	if err != nil {
		log.Fatalf("Failed to initialize auth handlers: %v", err)
	}
	apiKeyHandlers := admin.NewAPIKeyHandlers(cfg, db)
	userHandlers := admin.NewUserHandlers(cfg, db)
	orgHandlers := admin.NewOrganizationHandlers(cfg, db)
	statsHandlers := admin.NewStatsHandler(sqlxDB)
	mirrorHandlers := admin.NewMirrorHandler(mirrorRepo)
	mirrorHandlers.SetSyncJob(mirrorSyncJob) // Connect sync job for manual triggers
	providerAdminHandlers := admin.NewProviderAdminHandlers(db, storageBackend, cfg)
	moduleAdminHandlers := admin.NewModuleAdminHandlers(db, storageBackend, cfg)

	// Initialize RBAC handlers
	rbacRepo := repositories.NewRBACRepository(sqlxDB)
	rbacHandlers := admin.NewRBACHandlers(rbacRepo)

	// Initialize SCM handlers with the already-created repositories and token cipher
	scmProviderHandlers := admin.NewSCMProviderHandlers(cfg, scmRepo, tokenCipher)
	scmOAuthHandlers := admin.NewSCMOAuthHandlers(cfg, scmRepo, userRepo, tokenCipher)
	scmLinkingHandler := modules.NewSCMLinkingHandler(scmRepo, moduleRepo, tokenCipher, cfg.Server.BaseURL)

	// Initialize SCM publisher service
	scmPublisher := services.NewSCMPublisher(scmRepo, moduleRepo, storageBackend, tokenCipher)
	scmWebhookHandler := webhooks.NewSCMWebhookHandler(scmRepo, scmPublisher)

	// Initialize rate limiters
	authRateLimiter := middleware.NewRateLimiter(middleware.AuthRateLimitConfig())
	generalRateLimiter := middleware.NewRateLimiter(middleware.DefaultRateLimitConfig())
	uploadRateLimiter := middleware.NewRateLimiter(middleware.UploadRateLimitConfig())

	// Admin API endpoints
	apiV1 := router.Group("/api/v1")
	{
		// Public authentication endpoints (no auth required, but rate limited)
		authGroup := apiV1.Group("/auth")
		authGroup.Use(middleware.RateLimitMiddleware(authRateLimiter))
		{
			authGroup.GET("/login", authHandlers.LoginHandler())
			authGroup.GET("/callback", authHandlers.CallbackHandler())
		}

		// Authenticated-only endpoints
		authenticatedGroup := apiV1.Group("")
		authenticatedGroup.Use(middleware.AuthMiddleware(cfg, userRepo, apiKeyRepo, orgRepo))
		authenticatedGroup.Use(middleware.RateLimitMiddleware(generalRateLimiter))
		authenticatedGroup.Use(middleware.AuditMiddleware(auditRepo)) // Audit all authenticated actions
		{
			// Auth endpoints (require auth)
			authenticatedGroup.POST("/auth/refresh", authHandlers.RefreshHandler())
			authenticatedGroup.GET("/auth/me", authHandlers.MeHandler())

			// Stats endpoints (require auth)
			authenticatedGroup.GET("/admin/stats/dashboard", statsHandlers.GetDashboardStats)

			// Modules admin endpoints - require write permissions
			authenticatedGroup.POST("/modules",
				middleware.RateLimitMiddleware(uploadRateLimiter), // Stricter rate limit for uploads
				middleware.RequireScope(auth.ScopeModulesWrite),
				modules.UploadHandler(db, storageBackend, cfg))
			authenticatedGroup.GET("/modules/search",
				middleware.RequireScope(auth.ScopeModulesRead),
				modules.SearchHandler(db, cfg))

			// Providers admin endpoints - require write permissions
			authenticatedGroup.POST("/providers",
				middleware.RateLimitMiddleware(uploadRateLimiter), // Stricter rate limit for uploads
				middleware.RequireScope(auth.ScopeProvidersWrite),
				providers.UploadHandler(db, storageBackend, cfg))
			authenticatedGroup.GET("/providers/search",
				middleware.RequireScope(auth.ScopeProvidersRead),
				providers.SearchHandler(db, cfg))
			authenticatedGroup.GET("/providers/:namespace/:type",
				middleware.RequireScope(auth.ScopeProvidersRead),
				providerAdminHandlers.GetProvider)
			authenticatedGroup.DELETE("/providers/:namespace/:type",
				middleware.RequireScope(auth.ScopeProvidersWrite),
				providerAdminHandlers.DeleteProvider)
			authenticatedGroup.DELETE("/providers/:namespace/:type/versions/:version",
				middleware.RequireScope(auth.ScopeProvidersWrite),
				providerAdminHandlers.DeleteVersion)
			authenticatedGroup.POST("/providers/:namespace/:type/versions/:version/deprecate",
				middleware.RequireScope(auth.ScopeProvidersWrite),
				providerAdminHandlers.DeprecateVersion)
			authenticatedGroup.DELETE("/providers/:namespace/:type/versions/:version/deprecate",
				middleware.RequireScope(auth.ScopeProvidersWrite),
				providerAdminHandlers.UndeprecateVersion)

			// Modules admin endpoints - get, delete, deprecate
			authenticatedGroup.GET("/modules/:namespace/:name/:system",
				middleware.RequireScope(auth.ScopeModulesRead),
				moduleAdminHandlers.GetModule)
			authenticatedGroup.DELETE("/modules/:namespace/:name/:system",
				middleware.RequireScope(auth.ScopeModulesWrite),
				moduleAdminHandlers.DeleteModule)
			authenticatedGroup.DELETE("/modules/:namespace/:name/:system/versions/:version",
				middleware.RequireScope(auth.ScopeModulesWrite),
				moduleAdminHandlers.DeleteVersion)
			authenticatedGroup.POST("/modules/:namespace/:name/:system/versions/:version/deprecate",
				middleware.RequireScope(auth.ScopeModulesWrite),
				moduleAdminHandlers.DeprecateVersion)
			authenticatedGroup.DELETE("/modules/:namespace/:name/:system/versions/:version/deprecate",
				middleware.RequireScope(auth.ScopeModulesWrite),
				moduleAdminHandlers.UndeprecateVersion)

			// API Keys management - self-service for own keys
			// Users can manage their own API keys without api_keys:manage scope
			// The handlers verify ownership; api_keys:manage is only needed for managing others' keys
			apiKeysGroup := authenticatedGroup.Group("/apikeys")
			{
				apiKeysGroup.GET("", apiKeyHandlers.ListAPIKeysHandler())
				apiKeysGroup.POST("", apiKeyHandlers.CreateAPIKeyHandler())
				apiKeysGroup.GET("/:id", apiKeyHandlers.GetAPIKeyHandler())
				apiKeysGroup.PUT("/:id", apiKeyHandlers.UpdateAPIKeyHandler())
				apiKeysGroup.DELETE("/:id", apiKeyHandlers.DeleteAPIKeyHandler())
				apiKeysGroup.POST("/:id/rotate", apiKeyHandlers.RotateAPIKeyHandler())
			}

			// Self-service user endpoints (any authenticated user)
			// These endpoints allow users to access their own data without special scopes
			authenticatedGroup.GET("/users/me/memberships", userHandlers.GetCurrentUserMembershipsHandler())

			// Users management (requires users:read scope for viewing others)
			usersGroup := authenticatedGroup.Group("/users")
			usersGroup.Use(middleware.RequireScope(auth.ScopeUsersRead))
			{
				usersGroup.GET("", userHandlers.ListUsersHandler())
				usersGroup.GET("/search", userHandlers.SearchUsersHandler())
				usersGroup.GET("/:id", userHandlers.GetUserHandler())
				usersGroup.GET("/:id/memberships", userHandlers.GetUserMembershipsHandler())
			}

			usersWriteGroup := authenticatedGroup.Group("/users")
			usersWriteGroup.Use(middleware.RequireScope(auth.ScopeUsersWrite))
			{
				usersWriteGroup.POST("", userHandlers.CreateUserHandler())
				usersWriteGroup.PUT("/:id", userHandlers.UpdateUserHandler())
				usersWriteGroup.DELETE("/:id", userHandlers.DeleteUserHandler())
			}

			// Organizations management
			orgsGroup := authenticatedGroup.Group("/organizations")
			{
				// Read operations require organizations:read
				orgsGroup.GET("", middleware.RequireScope(auth.ScopeOrganizationsRead), orgHandlers.ListOrganizationsHandler())
				orgsGroup.GET("/search", middleware.RequireScope(auth.ScopeOrganizationsRead), orgHandlers.SearchOrganizationsHandler())
				orgsGroup.GET("/:id", middleware.RequireScope(auth.ScopeOrganizationsRead), orgHandlers.GetOrganizationHandler())
				orgsGroup.GET("/:id/members", middleware.RequireScope(auth.ScopeOrganizationsRead), orgHandlers.ListMembersHandler())

				// Create/update/delete require organizations:write
				orgsGroup.POST("", middleware.RequireScope(auth.ScopeOrganizationsWrite), orgHandlers.CreateOrganizationHandler())
				orgsGroup.PUT("/:id", middleware.RequireScope(auth.ScopeOrganizationsWrite), orgHandlers.UpdateOrganizationHandler())
				orgsGroup.DELETE("/:id", middleware.RequireScope(auth.ScopeOrganizationsWrite), orgHandlers.DeleteOrganizationHandler())

				// Member management requires organizations:write
				orgsGroup.POST("/:id/members", middleware.RequireScope(auth.ScopeOrganizationsWrite), orgHandlers.AddMemberHandler())
				orgsGroup.PUT("/:id/members/:user_id", middleware.RequireScope(auth.ScopeOrganizationsWrite), orgHandlers.UpdateMemberHandler())
				orgsGroup.DELETE("/:id/members/:user_id", middleware.RequireScope(auth.ScopeOrganizationsWrite), orgHandlers.RemoveMemberHandler())
			}

			// SCM Provider management
			scmProvidersGroup := authenticatedGroup.Group("/scm-providers")
			{
				// Read operations require scm:read
				scmProvidersGroup.GET("", middleware.RequireScope(auth.ScopeSCMRead), scmProviderHandlers.ListProviders)
				scmProvidersGroup.GET("/:id", middleware.RequireScope(auth.ScopeSCMRead), scmProviderHandlers.GetProvider)

				// Management operations require scm:manage
				scmProvidersGroup.POST("", middleware.RequireScope(auth.ScopeSCMManage), scmProviderHandlers.CreateProvider)
				scmProvidersGroup.PUT("/:id", middleware.RequireScope(auth.ScopeSCMManage), scmProviderHandlers.UpdateProvider)
				scmProvidersGroup.DELETE("/:id", middleware.RequireScope(auth.ScopeSCMManage), scmProviderHandlers.DeleteProvider)

				// OAuth flow endpoints require scm:manage
				scmProvidersGroup.GET("/:id/oauth/authorize", middleware.RequireScope(auth.ScopeSCMManage), scmOAuthHandlers.InitiateOAuth)
				scmProvidersGroup.DELETE("/:id/oauth/token", middleware.RequireScope(auth.ScopeSCMManage), scmOAuthHandlers.RevokeOAuth)
				scmProvidersGroup.POST("/:id/oauth/refresh", middleware.RequireScope(auth.ScopeSCMManage), scmOAuthHandlers.RefreshToken)
			}

			// SCM OAuth callback (public endpoint, no auth required)
			apiV1.GET("/scm-providers/:id/oauth/callback", scmOAuthHandlers.HandleOAuthCallback)

			// Module SCM linking endpoints
		moduleSCMGroup := authenticatedGroup.Group("/admin/modules/:id/scm")
			moduleSCMGroup.Use(middleware.RequireScope(auth.ScopeModulesWrite))
			{
				moduleSCMGroup.POST("", scmLinkingHandler.LinkModuleToSCM)
				moduleSCMGroup.GET("", scmLinkingHandler.GetModuleSCMInfo)
				moduleSCMGroup.PUT("", scmLinkingHandler.UpdateSCMLink)
				moduleSCMGroup.DELETE("", scmLinkingHandler.UnlinkModuleFromSCM)
				moduleSCMGroup.POST("/sync", scmLinkingHandler.TriggerManualSync)
				moduleSCMGroup.GET("/events", scmLinkingHandler.GetWebhookEvents)
			}

			// Mirror management endpoints with granular RBAC
			// Read operations require mirrors:read scope
			// Management operations require mirrors:manage scope
			mirrorsGroup := authenticatedGroup.Group("/admin/mirrors")
			{
				// Read operations - require mirrors:read (or mirrors:manage or admin)
				mirrorsGroup.GET("", middleware.RequireScope(auth.ScopeMirrorsRead), mirrorHandlers.ListMirrorConfigs)
				mirrorsGroup.GET("/:id", middleware.RequireScope(auth.ScopeMirrorsRead), mirrorHandlers.GetMirrorConfig)
				mirrorsGroup.GET("/:id/status", middleware.RequireScope(auth.ScopeMirrorsRead), mirrorHandlers.GetMirrorStatus)

				// Management operations - require mirrors:manage (or admin)
				mirrorsGroup.POST("", middleware.RequireScope(auth.ScopeMirrorsManage), mirrorHandlers.CreateMirrorConfig)
				mirrorsGroup.PUT("/:id", middleware.RequireScope(auth.ScopeMirrorsManage), mirrorHandlers.UpdateMirrorConfig)
				mirrorsGroup.DELETE("/:id", middleware.RequireScope(auth.ScopeMirrorsManage), mirrorHandlers.DeleteMirrorConfig)
				mirrorsGroup.POST("/:id/sync", middleware.RequireScope(auth.ScopeMirrorsManage), mirrorHandlers.TriggerSync)
			}

			// Role Templates management
			roleTemplatesGroup := authenticatedGroup.Group("/admin/role-templates")
			{
				roleTemplatesGroup.GET("", rbacHandlers.ListRoleTemplates)
				roleTemplatesGroup.GET("/:id", rbacHandlers.GetRoleTemplate)
				roleTemplatesGroup.POST("", middleware.RequireScope(auth.ScopeAdmin), rbacHandlers.CreateRoleTemplate)
				roleTemplatesGroup.PUT("/:id", middleware.RequireScope(auth.ScopeAdmin), rbacHandlers.UpdateRoleTemplate)
				roleTemplatesGroup.DELETE("/:id", middleware.RequireScope(auth.ScopeAdmin), rbacHandlers.DeleteRoleTemplate)
			}

			// Mirror Approval Requests
			approvalsGroup := authenticatedGroup.Group("/admin/approvals")
			{
				approvalsGroup.GET("", middleware.RequireScope(auth.ScopeMirrorsRead), rbacHandlers.ListApprovalRequests)
				approvalsGroup.GET("/:id", middleware.RequireScope(auth.ScopeMirrorsRead), rbacHandlers.GetApprovalRequest)
				approvalsGroup.POST("", middleware.RequireScope(auth.ScopeMirrorsManage), rbacHandlers.CreateApprovalRequest)
				approvalsGroup.PUT("/:id/review", middleware.RequireScope(auth.ScopeAdmin), rbacHandlers.ReviewApproval)
			}

			// Mirror Policies
			policiesGroup := authenticatedGroup.Group("/admin/policies")
			{
				policiesGroup.GET("", middleware.RequireScope(auth.ScopeMirrorsRead), rbacHandlers.ListMirrorPolicies)
				policiesGroup.GET("/:id", middleware.RequireScope(auth.ScopeMirrorsRead), rbacHandlers.GetMirrorPolicy)
				policiesGroup.POST("", middleware.RequireScope(auth.ScopeAdmin), rbacHandlers.CreateMirrorPolicy)
				policiesGroup.PUT("/:id", middleware.RequireScope(auth.ScopeAdmin), rbacHandlers.UpdateMirrorPolicy)
				policiesGroup.DELETE("/:id", middleware.RequireScope(auth.ScopeAdmin), rbacHandlers.DeleteMirrorPolicy)
				policiesGroup.POST("/evaluate", middleware.RequireScope(auth.ScopeMirrorsRead), rbacHandlers.EvaluatePolicy)
			}
		}

		// Development-only endpoints (guarded by DevModeMiddleware)
		devGroup := apiV1.Group("/dev")
		devGroup.Use(admin.DevModeMiddleware())
		{
			devHandlers := admin.NewDevHandlers(cfg, db)
			// Status endpoint (no auth required, but must be in dev mode)
			devGroup.GET("/status", devHandlers.DevStatusHandler())

			// Impersonation endpoints (require auth + admin scope)
			devGroup.Use(middleware.AuthMiddleware(cfg, userRepo, apiKeyRepo, orgRepo))
			devGroup.GET("/users", devHandlers.ListUsersForImpersonationHandler())
			devGroup.POST("/impersonate/:user_id", devHandlers.ImpersonateUserHandler())
		}
	}

	// Webhook endpoints (public, authentication via signature validation)
	router.POST("/webhooks/scm/:module_source_repo_id/:secret", scmWebhookHandler.HandleWebhook)

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
			"version":     "0.1.0",
			"api_version": "v1",
			"protocols": gin.H{
				"modules":   "v1",
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
