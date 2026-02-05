package providers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/terraform-registry/terraform-registry/internal/config"
	"github.com/terraform-registry/terraform-registry/internal/db/repositories"
)

// SearchHandler handles provider search requests
// Implements: GET /api/v1/providers/search?q=<query>&namespace=<namespace>&limit=<limit>&offset=<offset>
func SearchHandler(db *sql.DB, cfg *config.Config) gin.HandlerFunc {
	providerRepo := repositories.NewProviderRepository(db)
	orgRepo := repositories.NewOrganizationRepository(db)

	return func(c *gin.Context) {
		// Get query parameters
		query := c.Query("q")
		namespace := c.Query("namespace")

		// Pagination parameters
		limitStr := c.DefaultQuery("limit", "20")
		offsetStr := c.DefaultQuery("offset", "0")

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			limit = 20 // Default to 20, max 100
		}

		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			offset = 0
		}

		// Get organization context
		var orgID string
		if cfg.MultiTenancy.Enabled {
			org, err := orgRepo.GetDefaultOrganization(c.Request.Context())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to get organization context",
				})
				return
			}
			if org == nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Default organization not found",
				})
				return
			}
			orgID = org.ID
		}
		// In single-tenant mode, orgID will be empty string which the repository will handle

		// Search providers
		providers, total, err := providerRepo.SearchProviders(
			c.Request.Context(),
			orgID,
			query,
			namespace,
			limit,
			offset,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to search providers",
			})
			return
		}

		// Format results
		results := make([]gin.H, len(providers))
		for i, p := range providers {
			// Get latest version for each provider
			versions, _ := providerRepo.ListVersions(c.Request.Context(), p.ID)
			var latestVersion string
			var totalDownloads int64
			if len(versions) > 0 {
				latestVersion = versions[0].Version
				// For providers, downloads are tracked at the platform level
				// We would need to sum across all platforms for all versions
				// For now, set to 0 as we don't have a direct query for this
				totalDownloads = 0
			}

			results[i] = gin.H{
				"id":             p.ID,
				"namespace":      p.Namespace,
				"type":           p.Type,
				"description":    p.Description,
				"source":         p.Source,
				"latest_version": latestVersion,
				"download_count": totalDownloads,
				"created_at":     p.CreatedAt,
				"updated_at":     p.UpdatedAt,
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"providers": results,
			"meta": gin.H{
				"limit":  limit,
				"offset": offset,
				"total":  total,
			},
		})
	}
}
