package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/terraform-registry/terraform-registry/internal/db/models"
	"github.com/terraform-registry/terraform-registry/internal/db/repositories"
)

// AuditMiddleware logs authenticated actions
func AuditMiddleware(auditRepo *repositories.AuditRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request first
		c.Next()

		// Only log successful write operations
		if c.Request.Method == "GET" || c.Request.Method == "OPTIONS" || c.Writer.Status() >= 400 {
			return
		}

		// Extract context
		userID, _ := c.Get("user_id")
		orgID, _ := c.Get("organization_id")

		// Create audit log entry
		action := fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path)
		ipAddress := c.ClientIP()

		auditLog := &models.AuditLog{
			Action:    action,
			IPAddress: &ipAddress,
			CreatedAt: time.Now(),
		}

		// Set user ID if present
		if userID != nil {
			if userIDStr, ok := userID.(string); ok {
				auditLog.UserID = &userIDStr
			}
		}

		// Set organization ID if present
		if orgID != nil {
			if orgIDStr, ok := orgID.(string); ok {
				auditLog.OrganizationID = &orgIDStr
			}
		}

		// Set resource type based on URL path
		// This is a simple heuristic based on the path
		if contains(c.Request.URL.Path, "/modules") {
			resourceType := "module"
			auditLog.ResourceType = &resourceType
		} else if contains(c.Request.URL.Path, "/mirrors") {
			resourceType := "mirror"
			auditLog.ResourceType = &resourceType
			// Add specific mirror action details
			if contains(c.Request.URL.Path, "/sync") {
				action = "mirror.sync_triggered"
			} else if c.Request.Method == "POST" {
				action = "mirror.created"
			} else if c.Request.Method == "PUT" {
				action = "mirror.updated"
			} else if c.Request.Method == "DELETE" {
				action = "mirror.deleted"
			}
			auditLog.Action = action
		} else if contains(c.Request.URL.Path, "/providers") {
			resourceType := "provider"
			auditLog.ResourceType = &resourceType
		} else if contains(c.Request.URL.Path, "/users") {
			resourceType := "user"
			auditLog.ResourceType = &resourceType
		} else if contains(c.Request.URL.Path, "/api-keys") {
			resourceType := "api_key"
			auditLog.ResourceType = &resourceType
		}

		// Extract metadata from context if available
		metadata := make(map[string]interface{})

		if authMethod, exists := c.Get("auth_method"); exists {
			metadata["auth_method"] = authMethod
		}

		if len(metadata) > 0 {
			auditLog.Metadata = metadata
		}

		// Async log creation (non-blocking)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := auditRepo.CreateAuditLog(ctx, auditLog)
			if err != nil {
				// Log error but don't fail the request
				// In production, you might want to log this to a separate error tracking system
				fmt.Printf("Failed to create audit log: %v\n", err)
			}
		}()
	}
}

// contains is a simple helper to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		indexOf(s, substr) >= 0))
}

// indexOf returns the index of the first instance of substr in s, or -1 if substr is not present
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
