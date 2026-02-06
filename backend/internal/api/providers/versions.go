package providers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/terraform-registry/terraform-registry/internal/config"
	"github.com/terraform-registry/terraform-registry/internal/db/repositories"
)

// ListVersionsHandler handles listing all versions of a provider
// Implements: GET /v1/providers/:namespace/:type/versions
func ListVersionsHandler(db *sql.DB, cfg *config.Config) gin.HandlerFunc {
	providerRepo := repositories.NewProviderRepository(db)
	orgRepo := repositories.NewOrganizationRepository(db)

	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		providerType := c.Param("type")

		// Get organization context (default org for single-tenant mode)
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

		// Get provider
		provider, err := providerRepo.GetProvider(c.Request.Context(), org.ID, namespace, providerType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to query provider",
			})
			return
		}

		if provider == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"errors": []string{"Provider not found"},
			})
			return
		}

		// Get all versions for the provider
		versions, err := providerRepo.ListVersions(c.Request.Context(), provider.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to list provider versions",
			})
			return
		}

		// Format response per Terraform Provider Registry Protocol spec
		// https://www.terraform.io/docs/internals/provider-registry-protocol.html
		versionsList := make([]gin.H, 0, len(versions))
		for _, v := range versions {
			// Get platforms for this version
			platforms, err := providerRepo.ListPlatforms(c.Request.Context(), v.ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to list provider platforms",
				})
				return
			}

			// Format platforms
			platformsList := make([]gin.H, 0, len(platforms))
			for _, p := range platforms {
				platformsList = append(platformsList, gin.H{
					"id":             p.ID,
					"os":             p.OS,
					"arch":           p.Arch,
					"filename":       p.Filename,
					"shasum":         p.Shasum,
					"download_count": p.DownloadCount,
				})
			}

			versionData := gin.H{
				"id":           v.ID,
				"version":      v.Version,
				"protocols":    v.Protocols,
				"platforms":    platformsList,
				"published_at": v.CreatedAt.Format(time.RFC3339),
				"deprecated":   v.Deprecated,
			}
			if v.DeprecatedAt != nil {
				versionData["deprecated_at"] = v.DeprecatedAt.Format(time.RFC3339)
			}
			if v.DeprecationMessage != nil {
				versionData["deprecation_message"] = *v.DeprecationMessage
			}
			versionsList = append(versionsList, versionData)
		}

		response := gin.H{
			"versions": versionsList,
		}

		c.JSON(http.StatusOK, response)
	}
}
