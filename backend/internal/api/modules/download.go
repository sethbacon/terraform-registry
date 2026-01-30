package modules

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/terraform-registry/terraform-registry/internal/config"
	"github.com/terraform-registry/terraform-registry/internal/db/repositories"
	"github.com/terraform-registry/terraform-registry/internal/storage"
	"github.com/terraform-registry/terraform-registry/internal/validation"
)

// DownloadHandler handles module download requests
// Implements: GET /v1/modules/:namespace/:name/:system/:version/download
// Returns 204 No Content with X-Terraform-Get header pointing to download URL
func DownloadHandler(db *sql.DB, storageBackend storage.Storage, cfg *config.Config) gin.HandlerFunc {
	moduleRepo := repositories.NewModuleRepository(db)
	orgRepo := repositories.NewOrganizationRepository(db)

	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		name := c.Param("name")
		system := c.Param("system")
		version := c.Param("version")

		// Validate semantic versioning
		if err := validation.ValidateSemver(version); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": []string{"Invalid version format - must be valid semantic versioning"},
			})
			return
		}

		// Get organization context
		org, err := orgRepo.GetDefaultOrganization(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get organization context",
			})
			return
		}
		if org == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Default organization not found - please run migrations",
			})
			return
		}

		// Get module
		module, err := moduleRepo.GetModule(c.Request.Context(), org.ID, namespace, name, system)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to query module",
			})
			return
		}
		if module == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"errors": []string{"Module not found"},
			})
			return
		}

		// Get specific version
		moduleVersion, err := moduleRepo.GetVersion(c.Request.Context(), module.ID, version)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to query module version",
			})
			return
		}
		if moduleVersion == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"errors": []string{"Module version not found"},
			})
			return
		}

		// Get download URL from storage backend
		// TTL of 15 minutes for signed URLs
		downloadURL, err := storageBackend.GetURL(c.Request.Context(), moduleVersion.StoragePath, 15*time.Minute)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate download URL",
			})
			return
		}

		// Increment download counter asynchronously (don't block the response)
		go func() {
			// Use background context to avoid cancellation when request completes
			if err := moduleRepo.IncrementDownloadCount(c.Request.Context(), moduleVersion.ID); err != nil {
				// Log error but don't fail the request
				// TODO: Add proper logging in Phase 9
			}
		}()

		// Return 204 No Content with X-Terraform-Get header
		// This is the Terraform Module Registry Protocol standard response
		c.Header("X-Terraform-Get", downloadURL)
		c.Status(http.StatusNoContent)
	}
}
