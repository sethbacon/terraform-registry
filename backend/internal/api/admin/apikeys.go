package admin

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/terraform-registry/terraform-registry/internal/auth"
	"github.com/terraform-registry/terraform-registry/internal/config"
	"github.com/terraform-registry/terraform-registry/internal/db/models"
	"github.com/terraform-registry/terraform-registry/internal/db/repositories"
)

// APIKeyHandlers handles API key management endpoints
type APIKeyHandlers struct {
	cfg        *config.Config
	db         *sql.DB
	apiKeyRepo *repositories.APIKeyRepository
}

// NewAPIKeyHandlers creates a new APIKeyHandlers instance
func NewAPIKeyHandlers(cfg *config.Config, db *sql.DB) *APIKeyHandlers {
	return &APIKeyHandlers{
		cfg:        cfg,
		db:         db,
		apiKeyRepo: repositories.NewAPIKeyRepository(db),
	}
}

// CreateAPIKeyRequest represents the request to create a new API key
type CreateAPIKeyRequest struct {
	Name           string   `json:"name" binding:"required"`
	OrganizationID string   `json:"organization_id" binding:"required"`
	Scopes         []string `json:"scopes" binding:"required"`
	ExpiresAt      *string  `json:"expires_at"` // RFC3339 format
}

// CreateAPIKeyResponse represents the response when creating an API key
type CreateAPIKeyResponse struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Key       string     `json:"key"` // Only returned once during creation
	KeyPrefix string     `json:"key_prefix"`
	Scopes    []string   `json:"scopes"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
}

// ListAPIKeysHandler lists all API keys for the authenticated user
// GET /api/v1/apikeys
func (h *APIKeyHandlers) ListAPIKeysHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		userIDVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not authenticated",
			})
			return
		}

		userID, ok := userIDVal.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid user ID format",
			})
			return
		}

		// Get organization filter if provided
		orgID := c.Query("organization_id")

		var keys []*models.APIKey
		var err error

		if orgID != "" {
			// List keys for specific organization
			keys, err = h.apiKeyRepo.ListByOrganization(c.Request.Context(), orgID)
		} else {
			// List keys for user across all organizations
			keys, err = h.apiKeyRepo.ListByUser(c.Request.Context(), userID)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to list API keys",
			})
			return
		}

		// Return keys (without sensitive data)
		c.JSON(http.StatusOK, gin.H{
			"keys": keys,
		})
	}
}

// CreateAPIKeyHandler creates a new API key
// POST /api/v1/apikeys
func (h *APIKeyHandlers) CreateAPIKeyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateAPIKeyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request",
			})
			return
		}

		// Get user ID from context
		userIDVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not authenticated",
			})
			return
		}

		userID, ok := userIDVal.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid user ID format",
			})
			return
		}

		// Validate scopes
		if err := auth.ValidateScopes(req.Scopes); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid scopes: " + err.Error(),
			})
			return
		}

		// Parse expiration if provided
		var expiresAt *time.Time
		if req.ExpiresAt != nil {
			parsed, err := time.Parse(time.RFC3339, *req.ExpiresAt)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid expires_at format. Use RFC3339",
				})
				return
			}
			expiresAt = &parsed
		}

		// Generate API key
		keyPrefix := "tfr" // Terraform Registry
		fullKey, keyHash, displayPrefix, err := auth.GenerateAPIKey(keyPrefix)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate API key",
			})
			return
		}

		// Create API key in database
		apiKey := &models.APIKey{
			UserID:         &userID,
			OrganizationID: req.OrganizationID,
			Name:           req.Name,
			KeyHash:        keyHash,
			KeyPrefix:      displayPrefix,
			Scopes:         req.Scopes,
			ExpiresAt:      expiresAt,
			CreatedAt:      time.Now(),
		}

		if err := h.apiKeyRepo.Create(c.Request.Context(), apiKey); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create API key",
			})
			return
		}

		// Return full key (only time it's visible)
		c.JSON(http.StatusCreated, CreateAPIKeyResponse{
			ID:        apiKey.ID,
			Name:      apiKey.Name,
			Key:       fullKey, // IMPORTANT: Only returned once
			KeyPrefix: displayPrefix,
			Scopes:    apiKey.Scopes,
			ExpiresAt: apiKey.ExpiresAt,
			CreatedAt: apiKey.CreatedAt,
		})
	}
}

// GetAPIKeyHandler retrieves a specific API key
// GET /api/v1/apikeys/:id
func (h *APIKeyHandlers) GetAPIKeyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		keyID := c.Param("id")

		// Get API key
		apiKey, err := h.apiKeyRepo.GetByID(c.Request.Context(), keyID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve API key",
			})
			return
		}

		if apiKey == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "API key not found",
			})
			return
		}

		// Check authorization (user can only access their own keys)
		userIDVal, _ := c.Get("user_id")
		userID, _ := userIDVal.(string)

		if apiKey.UserID == nil || *apiKey.UserID != userID {
			// Check if user has admin scope
			scopesVal, _ := c.Get("scopes")
			scopes, _ := scopesVal.([]string)
			if !auth.HasScope(scopes, auth.ScopeAdmin) {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Access denied",
				})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"key": apiKey,
		})
	}
}

// DeleteAPIKeyHandler deletes an API key
// DELETE /api/v1/apikeys/:id
func (h *APIKeyHandlers) DeleteAPIKeyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		keyID := c.Param("id")

		// Get API key first to check authorization
		apiKey, err := h.apiKeyRepo.GetByID(c.Request.Context(), keyID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve API key",
			})
			return
		}

		if apiKey == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "API key not found",
			})
			return
		}

		// Check authorization
		userIDVal, _ := c.Get("user_id")
		userID, _ := userIDVal.(string)

		if apiKey.UserID == nil || *apiKey.UserID != userID {
			// Check if user has admin scope
			scopesVal, _ := c.Get("scopes")
			scopes, _ := scopesVal.([]string)
			if !auth.HasScope(scopes, auth.ScopeAdmin) {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Access denied",
				})
				return
			}
		}

		// Delete API key
		if err := h.apiKeyRepo.Delete(c.Request.Context(), keyID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete API key",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "API key deleted successfully",
		})
	}
}

// UpdateAPIKeyHandler updates an API key (name, scopes, expiration)
// PUT /api/v1/apikeys/:id
func (h *APIKeyHandlers) UpdateAPIKeyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		keyID := c.Param("id")

		var req struct {
			Name      *string  `json:"name"`
			Scopes    []string `json:"scopes"`
			ExpiresAt *string  `json:"expires_at"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request",
			})
			return
		}

		// Get API key
		apiKey, err := h.apiKeyRepo.GetByID(c.Request.Context(), keyID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve API key",
			})
			return
		}

		if apiKey == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "API key not found",
			})
			return
		}

		// Check authorization
		userIDVal, _ := c.Get("user_id")
		userID, _ := userIDVal.(string)

		if apiKey.UserID == nil || *apiKey.UserID != userID {
			scopesVal, _ := c.Get("scopes")
			scopes, _ := scopesVal.([]string)
			if !auth.HasScope(scopes, auth.ScopeAdmin) {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Access denied",
				})
				return
			}
		}

		// Update fields
		if req.Name != nil {
			apiKey.Name = *req.Name
		}

		if req.Scopes != nil {
			if err := auth.ValidateScopes(req.Scopes); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid scopes: " + err.Error(),
				})
				return
			}
			apiKey.Scopes = req.Scopes
		}

		if req.ExpiresAt != nil {
			parsed, err := time.Parse(time.RFC3339, *req.ExpiresAt)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid expires_at format. Use RFC3339",
				})
				return
			}
			apiKey.ExpiresAt = &parsed
		}

		// Update in database
		if err := h.apiKeyRepo.Update(c.Request.Context(), apiKey); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update API key",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"key": apiKey,
		})
	}
}
