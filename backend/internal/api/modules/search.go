package modules

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/terraform-registry/terraform-registry/internal/config"
	"github.com/terraform-registry/terraform-registry/internal/db/repositories"
)

// SearchHandler handles module search requests
// Implements: GET /api/v1/modules/search?q=<query>&namespace=<namespace>&system=<system>&limit=<limit>&offset=<offset>
func SearchHandler(db *sql.DB, cfg *config.Config) gin.HandlerFunc {
	moduleRepo := repositories.NewModuleRepository(db)
	orgRepo := repositories.NewOrganizationRepository(db)

	return func(c *gin.Context) {
		// Get query parameters
		query := c.Query("q")
		namespace := c.Query("namespace")
		system := c.Query("system")

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

		// Search modules
		modules, total, err := moduleRepo.SearchModules(
			c.Request.Context(),
			org.ID,
			query,
			namespace,
			system,
			limit,
			offset,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to search modules",
			})
			return
		}

		// Format results
		results := make([]gin.H, len(modules))
		for i, m := range modules {
			results[i] = gin.H{
				"id":          m.ID,
				"namespace":   m.Namespace,
				"name":        m.Name,
				"system":      m.System,
				"description": m.Description,
				"source":      m.Source,
				"created_at":  m.CreatedAt,
				"updated_at":  m.UpdatedAt,
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"modules": results,
			"meta": gin.H{
				"limit":  limit,
				"offset": offset,
				"total":  total,
			},
		})
	}
}
