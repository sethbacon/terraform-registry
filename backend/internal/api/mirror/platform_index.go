package mirror

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/terraform-registry/terraform-registry/internal/config"
	"github.com/terraform-registry/terraform-registry/internal/db/repositories"
	"github.com/terraform-registry/terraform-registry/internal/storage"
	"github.com/terraform-registry/terraform-registry/internal/validation"
)

// PlatformIndexHandler handles network mirror platform index requests
// Implements: GET /terraform/providers/:hostname/:namespace/:type/:version.json
// Returns download URLs and hashes for all platforms of a specific version
func PlatformIndexHandler(db *sql.DB, cfg *config.Config) gin.HandlerFunc {
	providerRepo := repositories.NewProviderRepository(db)
	orgRepo := repositories.NewOrganizationRepository(db)

	return func(c *gin.Context) {
		// Note: hostname is in the path for compatibility with Network Mirror Protocol
		hostname := c.Param("hostname")
		namespace := c.Param("namespace")
		providerType := c.Param("type")

		// Extract version from versionfile parameter (format: version.json)
		versionfile := c.Param("versionfile")

		// Strip .json suffix if present
		version := versionfile
		if len(version) > 5 && version[len(version)-5:] == ".json" {
			version = version[:len(version)-5]
		}

		// Log hostname for debugging (not used in single-tenant mode)
		_ = hostname

		// Validate semantic versioning
		if err := validation.ValidateSemver(version); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": []string{"Invalid version format - must be valid semantic versioning"},
			})
			return
		}

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

		// Get provider version
		providerVersion, err := providerRepo.GetVersion(c.Request.Context(), provider.ID, version)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to query provider version",
			})
			return
		}

		if providerVersion == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"errors": []string{"Provider version not found"},
			})
			return
		}

		// Get all platforms for this version
		platforms, err := providerRepo.ListPlatforms(c.Request.Context(), providerVersion.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to list provider platforms",
			})
			return
		}

		// Get storage backend
		storageBackend, err := storage.NewStorage(cfg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to initialize storage backend",
			})
			return
		}

		// Format response per Network Mirror Protocol spec
		// https://www.terraform.io/docs/internals/provider-network-mirror-protocol.html
		//
		// Response format:
		// {
		//   "archives": {
		//     "darwin_amd64": {
		//       "url": "providers/...",
		//       "hashes": ["h1:abcd...", "zh:abcd..."]
		//     },
		//     "linux_amd64": {
		//       "url": "providers/...",
		//       "hashes": ["h1:abcd...", "zh:abcd..."]
		//     }
		//   }
		// }
		archives := make(map[string]gin.H)

		for _, platform := range platforms {
			// Generate platform key (os_arch)
			platformKey := fmt.Sprintf("%s_%s", platform.OS, platform.Arch)

			// Get download URL from storage
			// For Network Mirror, we use a longer TTL (1 hour)
			downloadURL, err := storageBackend.GetURL(c.Request.Context(), platform.StoragePath, 1*time.Hour)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("Failed to generate download URL for %s", platformKey),
				})
				return
			}

			// Format hashes according to Network Mirror Protocol
			// Network Mirror expects:
			// - "h1:" prefix = SHA256 hash in base64
			// - "zh:" prefix = ZIP hash (also SHA256, but may include file headers)
			//
			// For now, we'll provide the h1 hash from our SHA256 checksum
			hashes := []string{
				formatH1Hash(platform.Shasum),
			}

			// Add to archives
			archives[platformKey] = gin.H{
				"url":    downloadURL,
				"hashes": hashes,
			}
		}

		response := gin.H{
			"archives": archives,
		}

		c.JSON(http.StatusOK, response)
	}
}

// formatH1Hash converts a hex SHA256 checksum to the "h1:" format used by Terraform
// "h1:" format is SHA256 hash in base64 encoding
func formatH1Hash(hexChecksum string) string {
	// Convert hex string to bytes
	// Since we have a hex string, we need to decode it first
	hashBytes := make([]byte, sha256.Size)
	for i := 0; i < len(hexChecksum) && i/2 < sha256.Size; i += 2 {
		var b byte
		fmt.Sscanf(hexChecksum[i:i+2], "%02x", &b)
		hashBytes[i/2] = b
	}

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(hashBytes)

	return "h1:" + encoded
}
