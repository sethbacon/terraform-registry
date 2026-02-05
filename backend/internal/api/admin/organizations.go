package admin

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/terraform-registry/terraform-registry/internal/config"
	"github.com/terraform-registry/terraform-registry/internal/db/models"
	"github.com/terraform-registry/terraform-registry/internal/db/repositories"
)

// OrganizationHandlers handles organization management endpoints
type OrganizationHandlers struct {
	cfg     *config.Config
	db      *sql.DB
	orgRepo *repositories.OrganizationRepository
}

// NewOrganizationHandlers creates a new OrganizationHandlers instance
func NewOrganizationHandlers(cfg *config.Config, db *sql.DB) *OrganizationHandlers {
	return &OrganizationHandlers{
		cfg:     cfg,
		db:      db,
		orgRepo: repositories.NewOrganizationRepository(db),
	}
}

// ListOrganizationsHandler lists all organizations with pagination
// GET /api/v1/organizations?page=1&per_page=20
func (h *OrganizationHandlers) ListOrganizationsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse pagination parameters
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

		if page < 1 {
			page = 1
		}
		if perPage < 1 || perPage > 100 {
			perPage = 20
		}

		offset := (page - 1) * perPage

		// Get organizations from repository
		orgs, err := h.orgRepo.List(c.Request.Context(), perPage, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to list organizations",
			})
			return
		}

		// Get total count
		total, err := h.orgRepo.Count(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to count organizations",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"organizations": orgs,
			"pagination": gin.H{
				"page":     page,
				"per_page": perPage,
				"total":    total,
			},
		})
	}
}

// GetOrganizationHandler retrieves a specific organization by ID
// GET /api/v1/organizations/:id
func (h *OrganizationHandlers) GetOrganizationHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID := c.Param("id")

		org, err := h.orgRepo.GetByID(c.Request.Context(), orgID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve organization",
			})
			return
		}

		if org == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Organization not found",
			})
			return
		}

		// Get organization members
		members, err := h.orgRepo.ListMembers(c.Request.Context(), orgID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve organization members",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"organization": org,
			"members":      members,
		})
	}
}

// CreateOrganizationRequest represents the request to create a new organization
type CreateOrganizationRequest struct {
	Name        string `json:"name" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
}

// CreateOrganizationHandler creates a new organization
// POST /api/v1/organizations
func (h *OrganizationHandlers) CreateOrganizationHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateOrganizationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request: " + err.Error(),
			})
			return
		}

		// Check if organization already exists
		existingOrg, err := h.orgRepo.GetByName(c.Request.Context(), req.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to check existing organization",
			})
			return
		}

		if existingOrg != nil {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Organization with this name already exists",
			})
			return
		}

		// Create organization
		org := &models.Organization{
			Name:        req.Name,
			DisplayName: req.DisplayName,
		}

		if err := h.orgRepo.Create(c.Request.Context(), org); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create organization",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"organization": org,
		})
	}
}

// UpdateOrganizationRequest represents the request to update an organization
type UpdateOrganizationRequest struct {
	DisplayName *string `json:"display_name"`
}

// UpdateOrganizationHandler updates an organization
// PUT /api/v1/organizations/:id
func (h *OrganizationHandlers) UpdateOrganizationHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID := c.Param("id")

		var req UpdateOrganizationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request: " + err.Error(),
			})
			return
		}

		// Get existing organization
		org, err := h.orgRepo.GetByID(c.Request.Context(), orgID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve organization",
			})
			return
		}

		if org == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Organization not found",
			})
			return
		}

		// Update fields
		if req.DisplayName != nil {
			org.DisplayName = *req.DisplayName
		}

		// Update in database
		if err := h.orgRepo.Update(c.Request.Context(), org); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update organization",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"organization": org,
		})
	}
}

// DeleteOrganizationHandler deletes an organization
// DELETE /api/v1/organizations/:id
func (h *OrganizationHandlers) DeleteOrganizationHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID := c.Param("id")

		// Check if organization exists
		org, err := h.orgRepo.GetByID(c.Request.Context(), orgID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve organization",
			})
			return
		}

		if org == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Organization not found",
			})
			return
		}

		// Delete organization (cascading deletes will handle related records)
		if err := h.orgRepo.Delete(c.Request.Context(), orgID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete organization",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Organization deleted successfully",
		})
	}
}

// AddMemberRequest represents the request to add a member to an organization
type AddMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required"`
}

// AddMemberHandler adds a member to an organization
// POST /api/v1/organizations/:id/members
func (h *OrganizationHandlers) AddMemberHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID := c.Param("id")

		var req AddMemberRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request: " + err.Error(),
			})
			return
		}

		// Validate role
		validRoles := map[string]bool{
			"owner":  true,
			"admin":  true,
			"member": true,
		}

		if !validRoles[req.Role] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid role. Must be 'owner', 'admin', or 'member'",
			})
			return
		}

		// Check if organization exists
		org, err := h.orgRepo.GetByID(c.Request.Context(), orgID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve organization",
			})
			return
		}

		if org == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Organization not found",
			})
			return
		}

		// Add member
		member := &models.OrganizationMember{
			OrganizationID: orgID,
			UserID:         req.UserID,
			Role:           req.Role,
			CreatedAt:      time.Now(),
		}

		if err := h.orgRepo.AddMember(c.Request.Context(), member); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to add member to organization",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"member": member,
		})
	}
}

// UpdateMemberRequest represents the request to update a member's role
type UpdateMemberRequest struct {
	Role string `json:"role" binding:"required"`
}

// UpdateMemberHandler updates a member's role in an organization
// PUT /api/v1/organizations/:id/members/:user_id
func (h *OrganizationHandlers) UpdateMemberHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID := c.Param("id")
		userID := c.Param("user_id")

		var req UpdateMemberRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request: " + err.Error(),
			})
			return
		}

		// Validate role
		validRoles := map[string]bool{
			"owner":  true,
			"admin":  true,
			"member": true,
		}

		if !validRoles[req.Role] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid role. Must be 'owner', 'admin', or 'member'",
			})
			return
		}

		// Get existing member
		member, err := h.orgRepo.GetMember(c.Request.Context(), orgID, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve member",
			})
			return
		}

		if member == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Member not found in organization",
			})
			return
		}

		// Update role
		member.Role = req.Role
		if err := h.orgRepo.UpdateMember(c.Request.Context(), member); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update member role",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"member": member,
		})
	}
}

// RemoveMemberHandler removes a member from an organization
// DELETE /api/v1/organizations/:id/members/:user_id
func (h *OrganizationHandlers) RemoveMemberHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID := c.Param("id")
		userID := c.Param("user_id")

		// Remove member
		if err := h.orgRepo.RemoveMember(c.Request.Context(), orgID, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to remove member from organization",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Member removed successfully",
		})
	}
}

// SearchOrganizationsHandler searches organizations by name
// GET /api/v1/organizations/search?q=query&page=1&per_page=20
func (h *OrganizationHandlers) SearchOrganizationsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Search query is required",
			})
			return
		}

		// Parse pagination
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

		if page < 1 {
			page = 1
		}
		if perPage < 1 || perPage > 100 {
			perPage = 20
		}

		offset := (page - 1) * perPage

		// Search organizations
		orgs, err := h.orgRepo.Search(c.Request.Context(), query, perPage, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to search organizations",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"organizations": orgs,
			"pagination": gin.H{
				"page":     page,
				"per_page": perPage,
			},
		})
	}
}
