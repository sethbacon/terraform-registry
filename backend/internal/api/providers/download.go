package providers

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

// DownloadHandler handles provider download requests
// Implements: GET /v1/providers/:namespace/:type/:version/download/:os/:arch
// Returns JSON with download URL, checksums, and signing keys
func DownloadHandler(db *sql.DB, storageBackend storage.Storage, cfg *config.Config) gin.HandlerFunc {
	providerRepo := repositories.NewProviderRepository(db)
	orgRepo := repositories.NewOrganizationRepository(db)

	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		providerType := c.Param("type")
		version := c.Param("version")
		os := c.Param("os")
		arch := c.Param("arch")

		// Validate semantic versioning
		if err := validation.ValidateSemver(version); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": []string{"Invalid version format - must be valid semantic versioning"},
			})
			return
		}

		// Validate platform
		if err := validation.ValidatePlatform(os, arch); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": []string{err.Error()},
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

		// Get platform binary
		platform, err := providerRepo.GetPlatform(c.Request.Context(), providerVersion.ID, os, arch)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to query provider platform",
			})
			return
		}
		if platform == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"errors": []string{"Provider platform not found"},
			})
			return
		}

		// Get download URL from storage backend
		// TTL of 15 minutes for signed URLs
		downloadURL, err := storageBackend.GetURL(c.Request.Context(), platform.StoragePath, 15*time.Minute)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate download URL",
			})
			return
		}

		// Get SHA256SUMS and signature URLs if available
		shasumsURL := ""
		shasumsSignatureURL := ""
		if providerVersion.ShasumURL != "" {
			shasumsURL = providerVersion.ShasumURL
		}
		if providerVersion.ShasumSignatureURL != "" {
			shasumsSignatureURL = providerVersion.ShasumSignatureURL
		}

		// Increment download counter asynchronously (don't block the response)
		go func() {
			// Use background context to avoid cancellation when request completes
			if err := providerRepo.IncrementDownloadCount(c.Request.Context(), platform.ID); err != nil {
				// Log error but don't fail the request
				// TODO: Add proper logging in Phase 9
			}
		}()

		// Format response per Terraform Provider Registry Protocol spec
		// https://www.terraform.io/docs/internals/provider-registry-protocol.html
		response := gin.H{
			"protocols":             providerVersion.Protocols,
			"os":                    platform.OS,
			"arch":                  platform.Arch,
			"filename":              platform.Filename,
			"download_url":          downloadURL,
			"shasums_url":           shasumsURL,
			"shasums_signature_url": shasumsSignatureURL,
			"shasum":                platform.Shasum,
		}

		// Include signing keys if GPG public key is available
		if providerVersion.GPGPublicKey != "" {
			response["signing_keys"] = gin.H{
				"gpg_public_keys": []gin.H{
					{
						"key_id":      "", // Could be extracted from GPG key if needed
						"ascii_armor": providerVersion.GPGPublicKey,
					},
				},
			}
		}

		c.JSON(http.StatusOK, response)
	}
}
