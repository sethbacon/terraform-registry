package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/terraform-registry/terraform-registry/internal/auth"
	"github.com/terraform-registry/terraform-registry/internal/db/repositories"
)

// RequireScope checks if authenticated user has the required scope
func RequireScope(scope auth.Scope) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get scopes from context (set by AuthMiddleware)
		scopesVal, exists := c.Get("scopes")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
			})
			return
		}

		userScopes, ok := scopesVal.([]string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Invalid scopes format",
			})
			return
		}

		if !auth.HasScope(userScopes, scope) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "Missing required scope",
				"details": "Required scope: " + string(scope),
			})
			return
		}

		c.Next()
	}
}

// RequireAnyScope checks if authenticated user has at least one of the required scopes
func RequireAnyScope(scopes ...auth.Scope) gin.HandlerFunc {
	return func(c *gin.Context) {
		scopesVal, exists := c.Get("scopes")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
			})
			return
		}

		userScopes, ok := scopesVal.([]string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Invalid scopes format",
			})
			return
		}

		if !auth.HasAnyScope(userScopes, scopes) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Missing required scope",
			})
			return
		}

		c.Next()
	}
}

// RequireAllScopes checks if authenticated user has all of the required scopes
func RequireAllScopes(scopes ...auth.Scope) gin.HandlerFunc {
	return func(c *gin.Context) {
		scopesVal, exists := c.Get("scopes")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
			})
			return
		}

		userScopes, ok := scopesVal.([]string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Invalid scopes format",
			})
			return
		}

		if !auth.HasAllScopes(userScopes, scopes) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Missing one or more required scopes",
			})
			return
		}

		c.Next()
	}
}

// RequireRole checks if user has the required organization role
func RequireRole(minRole string, orgRepo *repositories.OrganizationRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context
		userVal, userExists := c.Get("user_id")
		if !userExists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "User not authenticated",
			})
			return
		}

		userID, ok := userVal.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Invalid user ID format",
			})
			return
		}

		// Get organization from context
		orgVal, orgExists := c.Get("organization_id")
		if !orgExists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Organization context not found",
			})
			return
		}

		orgID, ok := orgVal.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Invalid organization ID format",
			})
			return
		}

		// Check membership and role
		member, err := orgRepo.GetMember(c.Request.Context(), orgID, userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to check organization membership",
			})
			return
		}

		if member == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Not a member of organization",
			})
			return
		}

		// Role hierarchy: viewer < member < admin < owner
		if !hasRequiredRole(member.Role, minRole) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "Insufficient role",
				"details": "Required role: " + minRole + ", current role: " + member.Role,
			})
			return
		}

		c.Next()
	}
}

// hasRequiredRole checks if userRole meets the minimum required role
func hasRequiredRole(userRole, minRole string) bool {
	roleHierarchy := map[string]int{
		"viewer": 1,
		"member": 2,
		"admin":  3,
		"owner":  4,
	}

	userLevel, userExists := roleHierarchy[userRole]
	minLevel, minExists := roleHierarchy[minRole]

	if !userExists || !minExists {
		return false
	}

	return userLevel >= minLevel
}
