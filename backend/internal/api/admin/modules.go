package admin

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/terraform-registry/terraform-registry/internal/config"
	"github.com/terraform-registry/terraform-registry/internal/db/models"
	"github.com/terraform-registry/terraform-registry/internal/db/repositories"
	"github.com/terraform-registry/terraform-registry/internal/storage"
)

// ModuleAdminHandlers handles administrative module operations
type ModuleAdminHandlers struct {
	moduleRepo     *repositories.ModuleRepository
	orgRepo        *repositories.OrganizationRepository
	storageBackend storage.Storage
	cfg            *config.Config
}

// NewModuleAdminHandlers creates a new module admin handlers instance
func NewModuleAdminHandlers(db *sql.DB, storageBackend storage.Storage, cfg *config.Config) *ModuleAdminHandlers {
	return &ModuleAdminHandlers{
		moduleRepo:     repositories.NewModuleRepository(db),
		orgRepo:        repositories.NewOrganizationRepository(db),
		storageBackend: storageBackend,
		cfg:            cfg,
	}
}

// CreateModuleRecord creates a module record without a version file.
// This is used by the SCM publishing flow to register a module before linking it to a repository.
// POST /api/v1/admin/modules/create
func (h *ModuleAdminHandlers) CreateModuleRecord(c *gin.Context) {
	var req struct {
		Namespace   string `json:"namespace" binding:"required"`
		Name        string `json:"name" binding:"required"`
		System      string `json:"system" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	org, err := h.orgRepo.GetDefaultOrganization(c.Request.Context())
	if err != nil || org == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get organization context"})
		return
	}

	// Return existing module if it already exists
	existing, err := h.moduleRepo.GetModule(c.Request.Context(), org.ID, req.Namespace, req.Name, req.System)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query module"})
		return
	}
	if existing != nil {
		c.JSON(http.StatusOK, existing)
		return
	}

	module := &models.Module{
		OrganizationID: org.ID,
		Namespace:      req.Namespace,
		Name:           req.Name,
		System:         req.System,
	}
	if req.Description != "" {
		module.Description = &req.Description
	}
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			module.CreatedBy = &uid
		}
	}

	if err := h.moduleRepo.CreateModule(c.Request.Context(), module); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create module"})
		return
	}

	c.JSON(http.StatusCreated, module)
}

// GetModule retrieves a specific module by namespace, name, and system
// GET /api/v1/modules/:namespace/:name/:system
func (h *ModuleAdminHandlers) GetModule(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	system := c.Param("system")

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

	// Get module
	module, err := h.moduleRepo.GetModule(c.Request.Context(), orgID, namespace, name, system)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get module"})
		return
	}

	if module == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
		return
	}

	// Get versions for the module
	versions, err := h.moduleRepo.ListVersions(c.Request.Context(), module.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list module versions"})
		return
	}

	// Format versions and calculate total downloads
	versionsList := make([]gin.H, 0, len(versions))
	var totalDownloads int64
	for _, v := range versions {
		totalDownloads += v.DownloadCount
		versionData := gin.H{
			"id":             v.ID,
			"version":        v.Version,
			"size_bytes":     v.SizeBytes,
			"checksum":       v.Checksum,
			"download_count": v.DownloadCount,
			"deprecated":     v.Deprecated,
			"created_at":     v.CreatedAt,
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
		"id":             module.ID,
		"namespace":      module.Namespace,
		"name":           module.Name,
		"system":         module.System,
		"description":    module.Description,
		"source":         module.Source,
		"download_count": totalDownloads,
		"versions":       versionsList,
		"created_at":     module.CreatedAt,
		"updated_at":     module.UpdatedAt,
	})
}

// DeleteModule deletes a module and all its versions
// DELETE /api/v1/modules/:namespace/:name/:system
func (h *ModuleAdminHandlers) DeleteModule(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	system := c.Param("system")

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

	// Get module
	module, err := h.moduleRepo.GetModule(c.Request.Context(), orgID, namespace, name, system)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get module"})
		return
	}

	if module == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
		return
	}

	// Get all versions to delete their files from storage
	versions, err := h.moduleRepo.ListVersions(c.Request.Context(), module.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list module versions"})
		return
	}

	// Delete files from storage for each version
	for _, v := range versions {
		if v.StoragePath != "" {
			// Try to delete from storage (ignore errors - file might not exist)
			_ = h.storageBackend.Delete(c.Request.Context(), v.StoragePath)
		}
	}

	// Delete module from database (cascades to versions)
	if err := h.moduleRepo.DeleteModule(c.Request.Context(), module.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete module: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Module deleted successfully",
		"namespace": namespace,
		"name":      name,
		"system":    system,
	})
}

// DeleteVersion deletes a specific version of a module
// DELETE /api/v1/modules/:namespace/:name/:system/versions/:version
func (h *ModuleAdminHandlers) DeleteVersion(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	system := c.Param("system")
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

	// Get module
	module, err := h.moduleRepo.GetModule(c.Request.Context(), orgID, namespace, name, system)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get module"})
		return
	}

	if module == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
		return
	}

	// Get the specific version
	versionRecord, err := h.moduleRepo.GetVersion(c.Request.Context(), module.ID, version)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get version"})
		return
	}

	if versionRecord == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Version not found"})
		return
	}

	// Delete file from storage
	if versionRecord.StoragePath != "" {
		_ = h.storageBackend.Delete(c.Request.Context(), versionRecord.StoragePath)
	}

	// Delete version from database
	if err := h.moduleRepo.DeleteVersion(c.Request.Context(), versionRecord.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete version: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Version deleted successfully",
		"namespace": namespace,
		"name":      name,
		"system":    system,
		"version":   version,
	})
}

// DeprecateVersionRequest represents a request to deprecate a version
type DeprecateModuleVersionRequest struct {
	Message string `json:"message,omitempty"`
}

// DeprecateVersion marks a specific version as deprecated
// POST /api/v1/modules/:namespace/:name/:system/versions/:version/deprecate
func (h *ModuleAdminHandlers) DeprecateVersion(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	system := c.Param("system")
	version := c.Param("version")

	var req DeprecateModuleVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Empty body is OK - message is optional
		req = DeprecateModuleVersionRequest{}
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

	// Get module
	module, err := h.moduleRepo.GetModule(c.Request.Context(), orgID, namespace, name, system)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get module"})
		return
	}

	if module == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
		return
	}

	// Get the specific version
	versionRecord, err := h.moduleRepo.GetVersion(c.Request.Context(), module.ID, version)
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

	if err := h.moduleRepo.DeprecateVersion(c.Request.Context(), versionRecord.ID, message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deprecate version: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Version deprecated successfully",
		"namespace": namespace,
		"name":      name,
		"system":    system,
		"version":   version,
	})
}

// UndeprecateVersion removes the deprecated status from a version
// DELETE /api/v1/modules/:namespace/:name/:system/versions/:version/deprecate
func (h *ModuleAdminHandlers) UndeprecateVersion(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	system := c.Param("system")
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

	// Get module
	module, err := h.moduleRepo.GetModule(c.Request.Context(), orgID, namespace, name, system)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get module"})
		return
	}

	if module == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
		return
	}

	// Get the specific version
	versionRecord, err := h.moduleRepo.GetVersion(c.Request.Context(), module.ID, version)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get version"})
		return
	}

	if versionRecord == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Version not found"})
		return
	}

	// Undeprecate the version
	if err := h.moduleRepo.UndeprecateVersion(c.Request.Context(), versionRecord.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to undeprecate version: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Version deprecation removed successfully",
		"namespace": namespace,
		"name":      name,
		"system":    system,
		"version":   version,
	})
}
