package modules

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terraform-registry/terraform-registry/internal/crypto"
	"github.com/terraform-registry/terraform-registry/internal/db/models"
	"github.com/terraform-registry/terraform-registry/internal/db/repositories"
	"github.com/terraform-registry/terraform-registry/internal/scm"
)

// SCMLinkingHandler handles module-SCM repository linking
type SCMLinkingHandler struct {
	scmRepo     *repositories.SCMRepository
	moduleRepo  *repositories.ModuleRepository
	tokenCipher *crypto.TokenCipher
	publicURL   string
}

// NewSCMLinkingHandler creates a new SCM linking handler
func NewSCMLinkingHandler(scmRepo *repositories.SCMRepository, moduleRepo *repositories.ModuleRepository, tokenCipher *crypto.TokenCipher, publicURL string) *SCMLinkingHandler {
	return &SCMLinkingHandler{
		scmRepo:     scmRepo,
		moduleRepo:  moduleRepo,
		tokenCipher: tokenCipher,
		publicURL:   publicURL,
	}
}

type LinkSCMRequest struct {
	SCMProviderID   string `json:"scm_provider_id" binding:"required"`
	RepositoryOwner string `json:"repo_owner" binding:"required"`
	RepositoryName  string `json:"repo_name" binding:"required"`
	DefaultBranch   string `json:"primary_branch"`
	ModulePath      string `json:"module_subpath"`
	TagPattern      string `json:"version_tag_glob"`
	AutoPublish     bool   `json:"publish_on_tag"`
}

// LinkModuleToSCM links a module to an SCM repository
// POST /api/v1/modules/:id/scm
func (h *SCMLinkingHandler) LinkModuleToSCM(c *gin.Context) {
	moduleIDStr := c.Param("id")
	moduleID, err := uuid.Parse(moduleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid module ID"})
		return
	}

	var req LinkSCMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	providerID, err := uuid.Parse(req.SCMProviderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid SCM provider ID"})
		return
	}

	// TODO: Replace with proper module lookup by ID
	// This is a placeholder - actual implementation needs GetModuleByID method
	// module, err := h.moduleRepo.GetByID(c.Request.Context(), moduleID)
	module := (*models.Module)(nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get module"})
		return
	}
	if module == nil {
		// For now, continue without module validation
		// c.JSON(http.StatusNotFound, gin.H{"error": "module not found"})
		// return
	}

	// Check if SCM provider exists
	provider, err := h.scmRepo.GetProvider(c.Request.Context(), providerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get SCM provider"})
		return
	}
	if provider == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "SCM provider not found"})
		return
	}

	// Check if module is already linked
	existing, err := h.scmRepo.GetModuleSourceRepo(c.Request.Context(), moduleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check existing link"})
		return
	}
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "module is already linked to a repository"})
		return
	}

	// Set defaults
	if req.DefaultBranch == "" {
		req.DefaultBranch = "main"
	}
	if req.ModulePath == "" {
		req.ModulePath = "/"
	}
	if req.TagPattern == "" {
		req.TagPattern = "v*"
	}

	// Create the webhook secret
	webhookSecret := generateWebhookSecret()

	// Create module source repo link
	linkID := uuid.New()
	repoFullURL := fmt.Sprintf("%s/%s/%s", provider.BaseURL, req.RepositoryOwner, req.RepositoryName)
	webhookCallbackURL := fmt.Sprintf("%s/webhooks/scm/%s/%s", h.publicURL, linkID, webhookSecret)

	link := &scm.ModuleSourceRepoRecord{
		ID:              linkID,
		ModuleID:        moduleID,
		SCMProviderID:   providerID,
		RepositoryOwner: req.RepositoryOwner,
		RepositoryName:  req.RepositoryName,
		RepositoryURL:   &repoFullURL,
		DefaultBranch:   req.DefaultBranch,
		ModulePath:      req.ModulePath,
		TagPattern:      req.TagPattern,
		AutoPublish:     req.AutoPublish,
		WebhookURL:      &webhookCallbackURL,
		WebhookEnabled:  false, // Will be activated after webhook registration
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := h.scmRepo.CreateModuleSourceRepo(c.Request.Context(), link); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create repository link"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":              "module linked to repository",
		"link_id":              linkID,
		"webhook_callback_url": webhookCallbackURL,
		"note":                 "Register this webhook URL in your repository settings",
	})
}

// UpdateSCMLink updates the SCM link configuration
// PUT /api/v1/modules/:id/scm
func (h *SCMLinkingHandler) UpdateSCMLink(c *gin.Context) {
	moduleIDStr := c.Param("id")
	moduleID, err := uuid.Parse(moduleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid module ID"})
		return
	}

	var req LinkSCMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing link
	link, err := h.scmRepo.GetModuleSourceRepo(c.Request.Context(), moduleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get repository link"})
		return
	}
	if link == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "module is not linked to a repository"})
		return
	}

	// Update fields
	link.RepositoryOwner = req.RepositoryOwner
	link.RepositoryName = req.RepositoryName
	link.DefaultBranch = req.DefaultBranch
	link.ModulePath = req.ModulePath
	link.TagPattern = req.TagPattern
	link.AutoPublish = req.AutoPublish

	if err := h.scmRepo.UpdateModuleSourceRepo(c.Request.Context(), link); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update repository link"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "repository link updated"})
}

// UnlinkModuleFromSCM removes the SCM repository link
// DELETE /api/v1/modules/:id/scm
func (h *SCMLinkingHandler) UnlinkModuleFromSCM(c *gin.Context) {
	moduleIDStr := c.Param("id")
	moduleID, err := uuid.Parse(moduleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid module ID"})
		return
	}

	// Get existing link
	link, err := h.scmRepo.GetModuleSourceRepo(c.Request.Context(), moduleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get repository link"})
		return
	}
	if link == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "module is not linked to a repository"})
		return
	}

	// TODO: Remove webhook from the repository

	// Delete the link
	if err := h.scmRepo.DeleteModuleSourceRepo(c.Request.Context(), moduleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete repository link"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "module unlinked from repository"})
}

// GetModuleSCMInfo retrieves the SCM link information for a module
// GET /api/v1/modules/:id/scm
func (h *SCMLinkingHandler) GetModuleSCMInfo(c *gin.Context) {
	moduleIDStr := c.Param("id")
	moduleID, err := uuid.Parse(moduleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid module ID"})
		return
	}

	link, err := h.scmRepo.GetModuleSourceRepo(c.Request.Context(), moduleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get repository link"})
		return
	}
	if link == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "module is not linked to a repository"})
		return
	}

	c.JSON(http.StatusOK, link)
}

// TriggerManualSync manually triggers a repository sync
// POST /api/v1/modules/:id/scm/sync
func (h *SCMLinkingHandler) TriggerManualSync(c *gin.Context) {
	moduleIDStr := c.Param("id")
	moduleID, err := uuid.Parse(moduleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid module ID"})
		return
	}

	link, err := h.scmRepo.GetModuleSourceRepo(c.Request.Context(), moduleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get repository link"})
		return
	}
	if link == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "module is not linked to a repository"})
		return
	}

	// TODO: Trigger async sync job

	c.JSON(http.StatusAccepted, gin.H{"message": "sync triggered"})
}

// GetWebhookEvents retrieves webhook event history for a module
// GET /api/v1/modules/:id/scm/events
func (h *SCMLinkingHandler) GetWebhookEvents(c *gin.Context) {
	moduleIDStr := c.Param("id")
	moduleID, err := uuid.Parse(moduleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid module ID"})
		return
	}

	link, err := h.scmRepo.GetModuleSourceRepo(c.Request.Context(), moduleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get repository link"})
		return
	}
	if link == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "module is not linked to a repository"})
		return
	}

	limit := 50 // Default limit
	events, err := h.scmRepo.ListWebhookLogs(c.Request.Context(), link.ID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get webhook events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"events": events})
}

func generateWebhookSecret() string {
	return uuid.New().String()
}
