package admin

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terraform-registry/terraform-registry/internal/config"
	"github.com/terraform-registry/terraform-registry/internal/crypto"
	"github.com/terraform-registry/terraform-registry/internal/db/repositories"
	"github.com/terraform-registry/terraform-registry/internal/scm"
)

// SCMOAuthHandlers handles SCM OAuth flows
type SCMOAuthHandlers struct {
	cfg         *config.Config
	scmRepo     *repositories.SCMRepository
	userRepo    *repositories.UserRepository
	tokenCipher *crypto.TokenCipher
}

// NewSCMOAuthHandlers creates a new SCM OAuth handlers instance
func NewSCMOAuthHandlers(cfg *config.Config, scmRepo *repositories.SCMRepository, userRepo *repositories.UserRepository, tokenCipher *crypto.TokenCipher) *SCMOAuthHandlers {
	return &SCMOAuthHandlers{
		cfg:         cfg,
		scmRepo:     scmRepo,
		userRepo:    userRepo,
		tokenCipher: tokenCipher,
	}
}

// InitiateOAuth starts the OAuth flow for an SCM provider
// GET /api/v1/scm-providers/:id/oauth/authorize
func (h *SCMOAuthHandlers) InitiateOAuth(c *gin.Context) {
	providerIDStr := c.Param("id")
	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider ID"})
		return
	}

	// Get user ID from context (set by auth middleware)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID"})
		return
	}

	// Get provider configuration
	provider, err := h.scmRepo.GetProvider(c.Request.Context(), providerID)
	if err != nil || provider == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		return
	}

	// Build connector
	connector, err := scm.BuildConnector(&scm.ConnectorSettings{
		Kind:         provider.ProviderType,
		ClientID:     provider.ClientID,
		ClientSecret: provider.ClientSecretEncrypted,
		CallbackURL:  fmt.Sprintf("%s/api/v1/scm-providers/%s/oauth/callback", h.cfg.Server.BaseURL, providerID),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create connector"})
		return
	}

	// Generate authorization URL
	state := fmt.Sprintf("%s:%s", userID, providerID)
	authURL := connector.AuthorizationEndpoint(state, []string{})

	c.JSON(http.StatusOK, gin.H{
		"authorization_url": authURL,
		"state":             state,
	})
}

// HandleOAuthCallback processes the OAuth callback
// GET /api/v1/scm-providers/:id/oauth/callback
func (h *SCMOAuthHandlers) HandleOAuthCallback(c *gin.Context) {
	providerIDStr := c.Param("id")
	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider ID"})
		return
	}

	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing authorization code"})
		return
	}

	// Parse state to get user ID
	var userID uuid.UUID
	fmt.Sscanf(state, "%s:", &userID)

	// Get provider configuration
	provider, err := h.scmRepo.GetProvider(c.Request.Context(), providerID)
	if err != nil || provider == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		return
	}

	// Decrypt client secret for token exchange
	clientSecret, err := h.tokenCipher.Open(provider.ClientSecretEncrypted)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decrypt client secret"})
		return
	}

	// Build connector
	connector, err := scm.BuildConnector(&scm.ConnectorSettings{
		Kind:         provider.ProviderType,
		ClientID:     provider.ClientID,
		ClientSecret: clientSecret,
		CallbackURL:  fmt.Sprintf("%s/api/v1/scm-providers/%s/oauth/callback", h.cfg.Server.BaseURL, providerID),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create connector"})
		return
	}

	// Complete OAuth flow
	oauthToken, err := connector.CompleteAuthorization(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("OAuth flow failed: %v", err)})
		return
	}

	// Encrypt access token
	encryptedAccessToken, err := h.tokenCipher.Seal(oauthToken.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encrypt access token"})
		return
	}

	// Encrypt refresh token if present
	var encryptedRefreshToken *string
	if oauthToken.RefreshToken != "" {
		encrypted, err := h.tokenCipher.Seal(oauthToken.RefreshToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encrypt refresh token"})
			return
		}
		encryptedRefreshToken = &encrypted
	}

	// Format scopes as comma-separated string
	scopesStr := ""
	if len(oauthToken.Scopes) > 0 {
		for i, scope := range oauthToken.Scopes {
			if i > 0 {
				scopesStr += ","
			}
			scopesStr += scope
		}
	}

	// Store or update token
	tokenRecord := &scm.SCMUserTokenRecord{
		ID:                    uuid.New(),
		UserID:                userID,
		SCMProviderID:         providerID,
		AccessTokenEncrypted:  encryptedAccessToken,
		RefreshTokenEncrypted: encryptedRefreshToken,
		TokenType:             oauthToken.TokenType,
		ExpiresAt:             oauthToken.ExpiresAt,
		Scopes:                &scopesStr,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	// Check if token already exists
	existingToken, err := h.scmRepo.GetUserToken(c.Request.Context(), userID, providerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check existing token"})
		return
	}

	if existingToken != nil {
		// Update existing token
		tokenRecord.ID = existingToken.ID
		tokenRecord.CreatedAt = existingToken.CreatedAt
		if err := h.scmRepo.SaveUserToken(c.Request.Context(), tokenRecord); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update token"})
			return
		}
	} else {
		// Create new token
		if err := h.scmRepo.SaveUserToken(c.Request.Context(), tokenRecord); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store token"})
			return
		}
	}

	// Redirect to frontend success page
	redirectURL := fmt.Sprintf("%s/admin/scm-providers/%s/connected", h.cfg.Server.BaseURL, providerID)
	c.Redirect(http.StatusFound, redirectURL)
}

// RevokeOAuth revokes a user's OAuth token for a provider
// DELETE /api/v1/scm-providers/:id/oauth/token
func (h *SCMOAuthHandlers) RevokeOAuth(c *gin.Context) {
	providerIDStr := c.Param("id")
	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider ID"})
		return
	}

	// Get user ID from context
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID"})
		return
	}

	if err := h.scmRepo.DeleteUserToken(c.Request.Context(), userID, providerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OAuth token revoked"})
}

// RefreshToken manually refreshes an OAuth token
// POST /api/v1/scm-providers/:id/oauth/refresh
func (h *SCMOAuthHandlers) RefreshToken(c *gin.Context) {
	providerIDStr := c.Param("id")
	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider ID"})
		return
	}

	// Get user ID from context
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID"})
		return
	}

	// Get existing token
	tokenRecord, err := h.scmRepo.GetUserToken(c.Request.Context(), userID, providerID)
	if err != nil || tokenRecord == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "OAuth token not found"})
		return
	}

	// Get provider configuration
	provider, err := h.scmRepo.GetProvider(c.Request.Context(), providerID)
	if err != nil || provider == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		return
	}

	// Decrypt refresh token
	var refreshToken string
	if tokenRecord.RefreshTokenEncrypted != nil {
		refreshToken, err = h.tokenCipher.Open(*tokenRecord.RefreshTokenEncrypted)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decrypt refresh token"})
			return
		}
	}

	// Decrypt client secret
	clientSecret, err := h.tokenCipher.Open(provider.ClientSecretEncrypted)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decrypt client secret"})
		return
	}

	// Build connector
	connector, err := scm.BuildConnector(&scm.ConnectorSettings{
		Kind:         provider.ProviderType,
		ClientID:     provider.ClientID,
		ClientSecret: clientSecret,
		CallbackURL:  fmt.Sprintf("%s/api/v1/scm-providers/%s/oauth/callback", h.cfg.Server.BaseURL, providerID),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create connector"})
		return
	}

	// Convert scopes string back to array
	var scopes []string
	if tokenRecord.Scopes != nil && *tokenRecord.Scopes != "" {
		for _, scope := range splitString(*tokenRecord.Scopes, ",") {
			scopes = append(scopes, scope)
		}
	}

	// Refresh token using the refresh token string
	newToken, err := connector.RenewToken(context.Background(), refreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("token refresh failed: %v", err)})
		return
	}

	// Encrypt new tokens
	encryptedAccessToken, err := h.tokenCipher.Seal(newToken.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encrypt new access token"})
		return
	}

	var encryptedRefreshToken *string
	if newToken.RefreshToken != "" {
		encrypted, err := h.tokenCipher.Seal(newToken.RefreshToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encrypt new refresh token"})
			return
		}
		encryptedRefreshToken = &encrypted
	}

	// Update token record
	tokenRecord.AccessTokenEncrypted = encryptedAccessToken
	tokenRecord.RefreshTokenEncrypted = encryptedRefreshToken
	tokenRecord.ExpiresAt = newToken.ExpiresAt
	tokenRecord.UpdatedAt = time.Now()

	if err := h.scmRepo.SaveUserToken(c.Request.Context(), tokenRecord); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "token refreshed",
		"expires_at": newToken.ExpiresAt,
	})
}

// Helper function to split string
func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	result := []string{}
	current := ""
	for _, char := range s {
		if string(char) == sep {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
