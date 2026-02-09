package admin

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terraform-registry/terraform-registry/internal/config"
	"github.com/terraform-registry/terraform-registry/internal/crypto"
	"github.com/terraform-registry/terraform-registry/internal/db/repositories"
	"github.com/terraform-registry/terraform-registry/internal/scm"
)

// SCMProviderHandlers handles SCM provider CRUD operations
type SCMProviderHandlers struct {
	cfg         *config.Config
	scmRepo     *repositories.SCMRepository
	tokenCipher *crypto.TokenCipher
}

// NewSCMProviderHandlers creates a new SCM provider handlers instance
func NewSCMProviderHandlers(cfg *config.Config, scmRepo *repositories.SCMRepository, tokenCipher *crypto.TokenCipher) *SCMProviderHandlers {
	return &SCMProviderHandlers{
		cfg:         cfg,
		scmRepo:     scmRepo,
		tokenCipher: tokenCipher,
	}
}

type CreateSCMProviderRequest struct {
	OrganizationID uuid.UUID        `json:"organization_id" binding:"required"`
	ProviderType   scm.ProviderType `json:"provider_type" binding:"required"`
	Name           string           `json:"name" binding:"required"`
	BaseURL        *string          `json:"base_url,omitempty"`
	ClientID       string           `json:"client_id"`
	ClientSecret   string           `json:"client_secret"`
	WebhookSecret  string           `json:"webhook_secret" binding:"required"`
}

type UpdateSCMProviderRequest struct {
	Name          *string `json:"name,omitempty"`
	BaseURL       *string `json:"base_url,omitempty"`
	ClientID      *string `json:"client_id,omitempty"`
	ClientSecret  *string `json:"client_secret,omitempty"`
	WebhookSecret *string `json:"webhook_secret,omitempty"`
	IsActive      *bool   `json:"is_active,omitempty"`
}

// CreateProvider creates a new SCM provider configuration
// POST /api/v1/scm-providers
func (h *SCMProviderHandlers) CreateProvider(c *gin.Context) {
	var req CreateSCMProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate provider type
	if !req.ProviderType.Valid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider type"})
		return
	}

	// PAT-based providers don't require OAuth credentials
	if req.ProviderType.IsPATBased() {
		if req.BaseURL == nil || *req.BaseURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "base_url is required for Bitbucket Data Center"})
			return
		}
		if req.ClientID == "" {
			req.ClientID = "pat-auth"
		}
		if req.ClientSecret == "" {
			req.ClientSecret = "not-applicable"
		}
	} else {
		if req.ClientID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "client_id is required for OAuth providers"})
			return
		}
		if req.ClientSecret == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "client_secret is required for OAuth providers"})
			return
		}
	}

	// Encrypt client secret
	clientSecretEncrypted, err := h.tokenCipher.Seal(req.ClientSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encrypt secret"})
		return
	}

	provider := &scm.SCMProviderRecord{
		ID:                    uuid.New(),
		OrganizationID:        req.OrganizationID,
		ProviderType:          req.ProviderType,
		Name:                  req.Name,
		BaseURL:               req.BaseURL,
		ClientID:              req.ClientID,
		ClientSecretEncrypted: clientSecretEncrypted,
		WebhookSecret:         req.WebhookSecret,
		IsActive:              true,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	if err := h.scmRepo.CreateProvider(c.Request.Context(), provider); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create provider"})
		return
	}

	c.JSON(http.StatusCreated, provider)
}

// ListProviders lists all SCM provider configurations
// GET /api/v1/scm-providers
func (h *SCMProviderHandlers) ListProviders(c *gin.Context) {
	orgIDStr := c.Query("organization_id")

	var providers []*scm.SCMProviderRecord
	var err error

	if orgIDStr != "" {
		orgID, parseErr := uuid.Parse(orgIDStr)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization_id"})
			return
		}
		providers, err = h.scmRepo.ListProviders(c.Request.Context(), orgID)
	} else {
		// Pass uuid.Nil to list all providers
		providers, err = h.scmRepo.ListProviders(c.Request.Context(), uuid.Nil)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list providers"})
		return
	}

	c.JSON(http.StatusOK, providers)
}

// GetProvider retrieves a single SCM provider by ID
// GET /api/v1/scm-providers/:id
func (h *SCMProviderHandlers) GetProvider(c *gin.Context) {
	providerIDStr := c.Param("id")
	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider ID"})
		return
	}

	provider, err := h.scmRepo.GetProvider(c.Request.Context(), providerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get provider"})
		return
	}

	if provider == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		return
	}

	c.JSON(http.StatusOK, provider)
}

// UpdateProvider updates an SCM provider configuration
// PUT /api/v1/scm-providers/:id
func (h *SCMProviderHandlers) UpdateProvider(c *gin.Context) {
	providerIDStr := c.Param("id")
	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider ID"})
		return
	}

	var req UpdateSCMProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	provider, err := h.scmRepo.GetProvider(c.Request.Context(), providerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get provider"})
		return
	}

	if provider == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		return
	}

	// Update fields
	if req.Name != nil {
		provider.Name = *req.Name
	}
	if req.BaseURL != nil {
		provider.BaseURL = req.BaseURL
	}
	if req.ClientID != nil {
		provider.ClientID = *req.ClientID
	}
	if req.ClientSecret != nil {
		encryptedSecret, err := h.tokenCipher.Seal(*req.ClientSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encrypt secret"})
			return
		}
		provider.ClientSecretEncrypted = encryptedSecret
	}
	if req.WebhookSecret != nil {
		provider.WebhookSecret = *req.WebhookSecret
	}
	if req.IsActive != nil {
		provider.IsActive = *req.IsActive
	}

	provider.UpdatedAt = time.Now()

	if err := h.scmRepo.UpdateProvider(c.Request.Context(), provider); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update provider"})
		return
	}

	c.JSON(http.StatusOK, provider)
}

// DeleteProvider deletes an SCM provider configuration
// DELETE /api/v1/scm-providers/:id
func (h *SCMProviderHandlers) DeleteProvider(c *gin.Context) {
	providerIDStr := c.Param("id")
	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider ID"})
		return
	}

	if err := h.scmRepo.DeleteProvider(c.Request.Context(), providerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete provider"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "provider deleted"})
}
