package admin

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terraform-registry/terraform-registry/internal/config"
	"github.com/terraform-registry/terraform-registry/internal/crypto"
	"github.com/terraform-registry/terraform-registry/internal/db/models"
	"github.com/terraform-registry/terraform-registry/internal/db/repositories"
)

// StorageHandlers handles storage configuration CRUD operations
type StorageHandlers struct {
	cfg                   *config.Config
	storageConfigRepo     *repositories.StorageConfigRepository
	tokenCipher           *crypto.TokenCipher
}

// NewStorageHandlers creates a new storage handlers instance
func NewStorageHandlers(cfg *config.Config, storageConfigRepo *repositories.StorageConfigRepository, tokenCipher *crypto.TokenCipher) *StorageHandlers {
	return &StorageHandlers{
		cfg:               cfg,
		storageConfigRepo: storageConfigRepo,
		tokenCipher:       tokenCipher,
	}
}

// GetSetupStatus returns the current setup status
// GET /api/v1/setup/status
func (h *StorageHandlers) GetSetupStatus(c *gin.Context) {
	ctx := c.Request.Context()

	configured, err := h.storageConfigRepo.IsStorageConfigured(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check setup status"})
		return
	}

	settings, err := h.storageConfigRepo.GetSystemSettings(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get system settings"})
		return
	}

	response := gin.H{
		"storage_configured": configured,
		"setup_required":     !configured,
	}

	if settings != nil && settings.StorageConfiguredAt.Valid {
		response["storage_configured_at"] = settings.StorageConfiguredAt.Time
	}

	c.JSON(http.StatusOK, response)
}

// GetActiveStorageConfig returns the currently active storage configuration
// GET /api/v1/storage/config
func (h *StorageHandlers) GetActiveStorageConfig(c *gin.Context) {
	ctx := c.Request.Context()

	storageConfig, err := h.storageConfigRepo.GetActiveStorageConfig(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get storage configuration"})
		return
	}

	if storageConfig == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active storage configuration"})
		return
	}

	c.JSON(http.StatusOK, storageConfig.ToResponse())
}

// ListStorageConfigs lists all storage configurations
// GET /api/v1/storage/configs
func (h *StorageHandlers) ListStorageConfigs(c *gin.Context) {
	ctx := c.Request.Context()

	configs, err := h.storageConfigRepo.ListStorageConfigs(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list storage configurations"})
		return
	}

	responses := make([]models.StorageConfigResponse, len(configs))
	for i, cfg := range configs {
		responses[i] = cfg.ToResponse()
	}

	c.JSON(http.StatusOK, responses)
}

// GetStorageConfig returns a storage configuration by ID
// GET /api/v1/storage/configs/:id
func (h *StorageHandlers) GetStorageConfig(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid configuration ID"})
		return
	}

	storageConfig, err := h.storageConfigRepo.GetStorageConfig(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get storage configuration"})
		return
	}

	if storageConfig == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "storage configuration not found"})
		return
	}

	c.JSON(http.StatusOK, storageConfig.ToResponse())
}

// CreateStorageConfig creates a new storage configuration
// POST /api/v1/storage/configs
func (h *StorageHandlers) CreateStorageConfig(c *gin.Context) {
	ctx := c.Request.Context()

	var input models.StorageConfigInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate backend type
	if input.BackendType != "local" && input.BackendType != "azure" && input.BackendType != "s3" && input.BackendType != "gcs" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid backend_type: must be local, azure, s3, or gcs"})
		return
	}

	// Validate required fields based on backend type
	if err := h.validateStorageConfig(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	var userUUID uuid.NullUUID
	if exists {
		if uid, ok := userID.(uuid.UUID); ok {
			userUUID = uuid.NullUUID{UUID: uid, Valid: true}
		}
	}

	// Check if this is the first storage configuration (initial setup)
	configured, err := h.storageConfigRepo.IsStorageConfigured(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check setup status"})
		return
	}

	// Build the storage config model
	storageConfig, err := h.buildStorageConfig(&input, userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// If storage is already configured, deactivate existing configs first
	if configured {
		if err := h.storageConfigRepo.DeactivateAllStorageConfigs(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update existing configurations"})
			return
		}
	}

	// Create the new configuration
	if err := h.storageConfigRepo.CreateStorageConfig(ctx, storageConfig); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create storage configuration"})
		return
	}

	// Mark storage as configured if this is the first setup
	if !configured && userUUID.Valid {
		if err := h.storageConfigRepo.SetStorageConfigured(ctx, userUUID.UUID); err != nil {
			// Log but don't fail - config was created
		}
	}

	c.JSON(http.StatusCreated, storageConfig.ToResponse())
}

// UpdateStorageConfig updates a storage configuration
// PUT /api/v1/storage/configs/:id
func (h *StorageHandlers) UpdateStorageConfig(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid configuration ID"})
		return
	}

	existing, err := h.storageConfigRepo.GetStorageConfig(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get storage configuration"})
		return
	}

	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "storage configuration not found"})
		return
	}

	// Check guard rails - if storage is configured and this is the active config,
	// only allow updates that don't change the backend type
	configured, _ := h.storageConfigRepo.IsStorageConfigured(ctx)
	if configured && existing.IsActive {
		// Allow updates but log a warning
		// In a real implementation, you might want to add more checks here
	}

	var input models.StorageConfigInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate backend type change - not allowed if active
	if configured && existing.IsActive && input.BackendType != existing.BackendType {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "cannot change backend type of active configuration; create a new configuration instead",
		})
		return
	}

	// Validate required fields
	if err := h.validateStorageConfig(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	var userUUID uuid.NullUUID
	if exists {
		if uid, ok := userID.(uuid.UUID); ok {
			userUUID = uuid.NullUUID{UUID: uid, Valid: true}
		}
	}

	// Update the config
	if err := h.updateStorageConfigFromInput(existing, &input, userUUID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.storageConfigRepo.UpdateStorageConfig(ctx, existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update storage configuration"})
		return
	}

	c.JSON(http.StatusOK, existing.ToResponse())
}

// DeleteStorageConfig deletes a storage configuration
// DELETE /api/v1/storage/configs/:id
func (h *StorageHandlers) DeleteStorageConfig(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid configuration ID"})
		return
	}

	existing, err := h.storageConfigRepo.GetStorageConfig(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get storage configuration"})
		return
	}

	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "storage configuration not found"})
		return
	}

	// Don't allow deleting the active configuration
	if existing.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete the active storage configuration"})
		return
	}

	if err := h.storageConfigRepo.DeleteStorageConfig(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete storage configuration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "storage configuration deleted"})
}

// ActivateStorageConfig activates a storage configuration
// POST /api/v1/storage/configs/:id/activate
func (h *StorageHandlers) ActivateStorageConfig(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid configuration ID"})
		return
	}

	existing, err := h.storageConfigRepo.GetStorageConfig(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get storage configuration"})
		return
	}

	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "storage configuration not found"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	var userUUID uuid.UUID
	if exists {
		if uid, ok := userID.(uuid.UUID); ok {
			userUUID = uid
		}
	}

	if err := h.storageConfigRepo.ActivateStorageConfig(ctx, id, userUUID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to activate storage configuration"})
		return
	}

	// Refresh the config
	existing, _ = h.storageConfigRepo.GetStorageConfig(ctx, id)

	c.JSON(http.StatusOK, gin.H{
		"message": "storage configuration activated",
		"config":  existing.ToResponse(),
	})
}

// TestStorageConfig tests a storage configuration without saving
// POST /api/v1/storage/configs/test
func (h *StorageHandlers) TestStorageConfig(c *gin.Context) {
	var input models.StorageConfigInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate required fields
	if err := h.validateStorageConfig(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement actual connection testing
	// For now, just validate the configuration
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "storage configuration is valid",
	})
}

// Helper functions

func (h *StorageHandlers) validateStorageConfig(input *models.StorageConfigInput) error {
	switch input.BackendType {
	case "local":
		if input.LocalBasePath == "" {
			return &ValidationError{Field: "local_base_path", Message: "required for local storage"}
		}
	case "azure":
		if input.AzureAccountName == "" {
			return &ValidationError{Field: "azure_account_name", Message: "required for Azure storage"}
		}
		if input.AzureContainerName == "" {
			return &ValidationError{Field: "azure_container_name", Message: "required for Azure storage"}
		}
		if input.AzureAccountKey == "" {
			return &ValidationError{Field: "azure_account_key", Message: "required for Azure storage"}
		}
	case "s3":
		if input.S3Bucket == "" {
			return &ValidationError{Field: "s3_bucket", Message: "required for S3 storage"}
		}
		if input.S3Region == "" {
			return &ValidationError{Field: "s3_region", Message: "required for S3 storage"}
		}
		// Validate auth method specific requirements
		if input.S3AuthMethod == "static" {
			if input.S3AccessKeyID == "" || input.S3SecretAccessKey == "" {
				return &ValidationError{Field: "s3_access_key_id", Message: "required for static auth"}
			}
		}
		if input.S3AuthMethod == "assume_role" || input.S3AuthMethod == "oidc" {
			if input.S3RoleARN == "" {
				return &ValidationError{Field: "s3_role_arn", Message: "required for assume_role/oidc auth"}
			}
		}
	case "gcs":
		if input.GCSBucket == "" {
			return &ValidationError{Field: "gcs_bucket", Message: "required for GCS storage"}
		}
		// For service_account auth, require credentials
		if input.GCSAuthMethod == "service_account" {
			if input.GCSCredentialsFile == "" && input.GCSCredentialsJSON == "" {
				return &ValidationError{Field: "gcs_credentials", Message: "credentials_file or credentials_json required for service_account auth"}
			}
		}
	}
	return nil
}

func (h *StorageHandlers) buildStorageConfig(input *models.StorageConfigInput, userID uuid.NullUUID) (*models.StorageConfig, error) {
	now := time.Now()
	config := &models.StorageConfig{
		ID:          uuid.New(),
		BackendType: input.BackendType,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
		CreatedBy:   userID,
		UpdatedBy:   userID,
	}

	// Set backend-specific fields
	switch input.BackendType {
	case "local":
		config.LocalBasePath = sql.NullString{String: input.LocalBasePath, Valid: input.LocalBasePath != ""}
		if input.LocalServeDirectly != nil {
			config.LocalServeDirectly = sql.NullBool{Bool: *input.LocalServeDirectly, Valid: true}
		} else {
			config.LocalServeDirectly = sql.NullBool{Bool: true, Valid: true}
		}

	case "azure":
		config.AzureAccountName = sql.NullString{String: input.AzureAccountName, Valid: input.AzureAccountName != ""}
		config.AzureContainerName = sql.NullString{String: input.AzureContainerName, Valid: input.AzureContainerName != ""}
		config.AzureCDNURL = sql.NullString{String: input.AzureCDNURL, Valid: input.AzureCDNURL != ""}
		if input.AzureAccountKey != "" {
			encrypted, err := h.tokenCipher.Seal(input.AzureAccountKey)
			if err != nil {
				return nil, err
			}
			config.AzureAccountKeyEncrypted = sql.NullString{String: encrypted, Valid: true}
		}

	case "s3":
		config.S3Endpoint = sql.NullString{String: input.S3Endpoint, Valid: input.S3Endpoint != ""}
		config.S3Region = sql.NullString{String: input.S3Region, Valid: input.S3Region != ""}
		config.S3Bucket = sql.NullString{String: input.S3Bucket, Valid: input.S3Bucket != ""}
		config.S3AuthMethod = sql.NullString{String: input.S3AuthMethod, Valid: input.S3AuthMethod != ""}
		config.S3RoleARN = sql.NullString{String: input.S3RoleARN, Valid: input.S3RoleARN != ""}
		config.S3RoleSessionName = sql.NullString{String: input.S3RoleSessionName, Valid: input.S3RoleSessionName != ""}
		config.S3ExternalID = sql.NullString{String: input.S3ExternalID, Valid: input.S3ExternalID != ""}
		config.S3WebIdentityTokenFile = sql.NullString{String: input.S3WebIdentityTokenFile, Valid: input.S3WebIdentityTokenFile != ""}
		if input.S3AccessKeyID != "" {
			encrypted, err := h.tokenCipher.Seal(input.S3AccessKeyID)
			if err != nil {
				return nil, err
			}
			config.S3AccessKeyIDEncrypted = sql.NullString{String: encrypted, Valid: true}
		}
		if input.S3SecretAccessKey != "" {
			encrypted, err := h.tokenCipher.Seal(input.S3SecretAccessKey)
			if err != nil {
				return nil, err
			}
			config.S3SecretAccessKeyEncrypted = sql.NullString{String: encrypted, Valid: true}
		}

	case "gcs":
		config.GCSBucket = sql.NullString{String: input.GCSBucket, Valid: input.GCSBucket != ""}
		config.GCSProjectID = sql.NullString{String: input.GCSProjectID, Valid: input.GCSProjectID != ""}
		config.GCSAuthMethod = sql.NullString{String: input.GCSAuthMethod, Valid: input.GCSAuthMethod != ""}
		config.GCSCredentialsFile = sql.NullString{String: input.GCSCredentialsFile, Valid: input.GCSCredentialsFile != ""}
		config.GCSEndpoint = sql.NullString{String: input.GCSEndpoint, Valid: input.GCSEndpoint != ""}
		if input.GCSCredentialsJSON != "" {
			encrypted, err := h.tokenCipher.Seal(input.GCSCredentialsJSON)
			if err != nil {
				return nil, err
			}
			config.GCSCredentialsJSONEncrypted = sql.NullString{String: encrypted, Valid: true}
		}
	}

	return config, nil
}

func (h *StorageHandlers) updateStorageConfigFromInput(config *models.StorageConfig, input *models.StorageConfigInput, userID uuid.NullUUID) error {
	config.BackendType = input.BackendType
	config.UpdatedBy = userID

	// Update backend-specific fields
	switch input.BackendType {
	case "local":
		config.LocalBasePath = sql.NullString{String: input.LocalBasePath, Valid: input.LocalBasePath != ""}
		if input.LocalServeDirectly != nil {
			config.LocalServeDirectly = sql.NullBool{Bool: *input.LocalServeDirectly, Valid: true}
		}

	case "azure":
		config.AzureAccountName = sql.NullString{String: input.AzureAccountName, Valid: input.AzureAccountName != ""}
		config.AzureContainerName = sql.NullString{String: input.AzureContainerName, Valid: input.AzureContainerName != ""}
		config.AzureCDNURL = sql.NullString{String: input.AzureCDNURL, Valid: input.AzureCDNURL != ""}
		if input.AzureAccountKey != "" {
			encrypted, err := h.tokenCipher.Seal(input.AzureAccountKey)
			if err != nil {
				return err
			}
			config.AzureAccountKeyEncrypted = sql.NullString{String: encrypted, Valid: true}
		}

	case "s3":
		config.S3Endpoint = sql.NullString{String: input.S3Endpoint, Valid: input.S3Endpoint != ""}
		config.S3Region = sql.NullString{String: input.S3Region, Valid: input.S3Region != ""}
		config.S3Bucket = sql.NullString{String: input.S3Bucket, Valid: input.S3Bucket != ""}
		config.S3AuthMethod = sql.NullString{String: input.S3AuthMethod, Valid: input.S3AuthMethod != ""}
		config.S3RoleARN = sql.NullString{String: input.S3RoleARN, Valid: input.S3RoleARN != ""}
		config.S3RoleSessionName = sql.NullString{String: input.S3RoleSessionName, Valid: input.S3RoleSessionName != ""}
		config.S3ExternalID = sql.NullString{String: input.S3ExternalID, Valid: input.S3ExternalID != ""}
		config.S3WebIdentityTokenFile = sql.NullString{String: input.S3WebIdentityTokenFile, Valid: input.S3WebIdentityTokenFile != ""}
		if input.S3AccessKeyID != "" {
			encrypted, err := h.tokenCipher.Seal(input.S3AccessKeyID)
			if err != nil {
				return err
			}
			config.S3AccessKeyIDEncrypted = sql.NullString{String: encrypted, Valid: true}
		}
		if input.S3SecretAccessKey != "" {
			encrypted, err := h.tokenCipher.Seal(input.S3SecretAccessKey)
			if err != nil {
				return err
			}
			config.S3SecretAccessKeyEncrypted = sql.NullString{String: encrypted, Valid: true}
		}

	case "gcs":
		config.GCSBucket = sql.NullString{String: input.GCSBucket, Valid: input.GCSBucket != ""}
		config.GCSProjectID = sql.NullString{String: input.GCSProjectID, Valid: input.GCSProjectID != ""}
		config.GCSAuthMethod = sql.NullString{String: input.GCSAuthMethod, Valid: input.GCSAuthMethod != ""}
		config.GCSCredentialsFile = sql.NullString{String: input.GCSCredentialsFile, Valid: input.GCSCredentialsFile != ""}
		config.GCSEndpoint = sql.NullString{String: input.GCSEndpoint, Valid: input.GCSEndpoint != ""}
		if input.GCSCredentialsJSON != "" {
			encrypted, err := h.tokenCipher.Seal(input.GCSCredentialsJSON)
			if err != nil {
				return err
			}
			config.GCSCredentialsJSONEncrypted = sql.NullString{String: encrypted, Valid: true}
		}
	}

	return nil
}

// ValidationError represents a validation error for a specific field
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
