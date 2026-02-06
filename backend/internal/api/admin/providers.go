package admin

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/terraform-registry/terraform-registry/internal/config"
	"github.com/terraform-registry/terraform-registry/internal/db/repositories"
	"github.com/terraform-registry/terraform-registry/internal/storage"
)

// ProviderAdminHandlers handles administrative provider operations
type ProviderAdminHandlers struct {
	providerRepo   *repositories.ProviderRepository
	orgRepo        *repositories.OrganizationRepository
	storageBackend storage.Storage
	cfg            *config.Config
}

// NewProviderAdminHandlers creates a new provider admin handlers instance
func NewProviderAdminHandlers(db *sql.DB, storageBackend storage.Storage, cfg *config.Config) *ProviderAdminHandlers {
	return &ProviderAdminHandlers{
		providerRepo:   repositories.NewProviderRepository(db),
		orgRepo:        repositories.NewOrganizationRepository(db),
		storageBackend: storageBackend,
		cfg:            cfg,
	}
}

// GetProvider retrieves a specific provider by namespace and type
// GET /api/v1/providers/:namespace/:type
func (h *ProviderAdminHandlers) GetProvider(c *gin.Context) {
	namespace := c.Param("namespace")
	providerType := c.Param("type")

	// Get organization context (default org for single-tenant mode)
	org, err := h.orgRepo.GetDefaultOrganization(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get organization context"})
		return
	}

	var orgID string
	if org != nil {
		orgID = org.ID
	}

	// Get provider
	provider, err := h.providerRepo.GetProvider(c.Request.Context(), orgID, namespace, providerType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get provider"})
		return
	}

	if provider == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
		return
	}

	// Get versions for the provider
	versions, err := h.providerRepo.ListVersions(c.Request.Context(), provider.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list provider versions"})
		return
	}

	// Format versions
	versionsList := make([]gin.H, 0, len(versions))
	for _, v := range versions {
		platforms, _ := h.providerRepo.ListPlatforms(c.Request.Context(), v.ID)
		platformsList := make([]gin.H, 0, len(platforms))
		for _, p := range platforms {
			platformsList = append(platformsList, gin.H{
				"os":             p.OS,
				"arch":           p.Arch,
				"filename":       p.Filename,
				"shasum":         p.Shasum,
				"download_count": p.DownloadCount,
			})
		}

		versionData := gin.H{
			"id":         v.ID,
			"version":    v.Version,
			"protocols":  v.Protocols,
			"platforms":  platformsList,
			"deprecated": v.Deprecated,
			"created_at": v.CreatedAt,
		}
		if v.DeprecatedAt != nil {
			versionData["deprecated_at"] = v.DeprecatedAt
		}
		if v.DeprecationMessage != nil {
			versionData["deprecation_message"] = v.DeprecationMessage
		}
		versionsList = append(versionsList, versionData)
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          provider.ID,
		"namespace":   provider.Namespace,
		"type":        provider.Type,
		"description": provider.Description,
		"source":      provider.Source,
		"versions":    versionsList,
		"created_at":  provider.CreatedAt,
		"updated_at":  provider.UpdatedAt,
	})
}

// DeleteProvider deletes a provider and all its versions/platforms
// DELETE /api/v1/providers/:namespace/:type
func (h *ProviderAdminHandlers) DeleteProvider(c *gin.Context) {
	namespace := c.Param("namespace")
	providerType := c.Param("type")

	// Get organization context
	org, err := h.orgRepo.GetDefaultOrganization(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get organization context"})
		return
	}

	var orgID string
	if org != nil {
		orgID = org.ID
	}

	// Get provider
	provider, err := h.providerRepo.GetProvider(c.Request.Context(), orgID, namespace, providerType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get provider"})
		return
	}

	if provider == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
		return
	}

	// Get all versions to delete their files from storage
	versions, err := h.providerRepo.ListVersions(c.Request.Context(), provider.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list provider versions"})
		return
	}

	// Delete files from storage for each version
	for _, v := range versions {
		platforms, _ := h.providerRepo.ListPlatforms(c.Request.Context(), v.ID)
		for _, p := range platforms {
			if p.StoragePath != "" {
				// Try to delete from storage (ignore errors - file might not exist)
				_ = h.storageBackend.Delete(c.Request.Context(), p.StoragePath)
			}
		}
	}

	// Delete provider from database (cascades to versions and platforms)
	if err := h.providerRepo.DeleteProvider(c.Request.Context(), provider.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete provider: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Provider deleted successfully",
		"namespace": namespace,
		"type":      providerType,
	})
}

// DeleteVersion deletes a specific version of a provider
// DELETE /api/v1/providers/:namespace/:type/versions/:version
func (h *ProviderAdminHandlers) DeleteVersion(c *gin.Context) {
	namespace := c.Param("namespace")
	providerType := c.Param("type")
	version := c.Param("version")

	// Get organization context
	org, err := h.orgRepo.GetDefaultOrganization(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get organization context"})
		return
	}

	var orgID string
	if org != nil {
		orgID = org.ID
	}

	// Get provider
	provider, err := h.providerRepo.GetProvider(c.Request.Context(), orgID, namespace, providerType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get provider"})
		return
	}

	if provider == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
		return
	}

	// Get the specific version
	versionRecord, err := h.providerRepo.GetVersion(c.Request.Context(), provider.ID, version)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get version"})
		return
	}

	if versionRecord == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Version not found"})
		return
	}

	// Delete files from storage
	platforms, _ := h.providerRepo.ListPlatforms(c.Request.Context(), versionRecord.ID)
	for _, p := range platforms {
		if p.StoragePath != "" {
			_ = h.storageBackend.Delete(c.Request.Context(), p.StoragePath)
		}
	}

	// Delete version from database (cascades to platforms)
	if err := h.providerRepo.DeleteVersion(c.Request.Context(), versionRecord.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete version: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Version deleted successfully",
		"namespace": namespace,
		"type":      providerType,
		"version":   version,
	})
}

// DeprecateVersionRequest represents a request to deprecate a version
type DeprecateVersionRequest struct {
	Message string `json:"message,omitempty"`
}

// DeprecateVersion marks a specific version as deprecated
// POST /api/v1/providers/:namespace/:type/versions/:version/deprecate
func (h *ProviderAdminHandlers) DeprecateVersion(c *gin.Context) {
	namespace := c.Param("namespace")
	providerType := c.Param("type")
	version := c.Param("version")

	var req DeprecateVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Empty body is OK - message is optional
		req = DeprecateVersionRequest{}
	}

	// Get organization context
	org, err := h.orgRepo.GetDefaultOrganization(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get organization context"})
		return
	}

	var orgID string
	if org != nil {
		orgID = org.ID
	}

	// Get provider
	provider, err := h.providerRepo.GetProvider(c.Request.Context(), orgID, namespace, providerType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get provider"})
		return
	}

	if provider == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
		return
	}

	// Get the specific version
	versionRecord, err := h.providerRepo.GetVersion(c.Request.Context(), provider.ID, version)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get version"})
		return
	}

	if versionRecord == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Version not found"})
		return
	}

	// Deprecate the version
	var message *string
	if req.Message != "" {
		message = &req.Message
	}

	if err := h.providerRepo.DeprecateVersion(c.Request.Context(), versionRecord.ID, message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deprecate version: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Version deprecated successfully",
		"namespace": namespace,
		"type":      providerType,
		"version":   version,
	})
}

// UndeprecateVersion removes the deprecated status from a version
// DELETE /api/v1/providers/:namespace/:type/versions/:version/deprecate
func (h *ProviderAdminHandlers) UndeprecateVersion(c *gin.Context) {
	namespace := c.Param("namespace")
	providerType := c.Param("type")
	version := c.Param("version")

	// Get organization context
	org, err := h.orgRepo.GetDefaultOrganization(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get organization context"})
		return
	}

	var orgID string
	if org != nil {
		orgID = org.ID
	}

	// Get provider
	provider, err := h.providerRepo.GetProvider(c.Request.Context(), orgID, namespace, providerType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get provider"})
		return
	}

	if provider == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
		return
	}

	// Get the specific version
	versionRecord, err := h.providerRepo.GetVersion(c.Request.Context(), provider.ID, version)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get version"})
		return
	}

	if versionRecord == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Version not found"})
		return
	}

	// Undeprecate the version
	if err := h.providerRepo.UndeprecateVersion(c.Request.Context(), versionRecord.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to undeprecate version: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Version deprecation removed successfully",
		"namespace": namespace,
		"type":      providerType,
		"version":   version,
	})
}
