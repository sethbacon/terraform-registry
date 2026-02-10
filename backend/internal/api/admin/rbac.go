package admin

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terraform-registry/terraform-registry/internal/db/models"
	"github.com/terraform-registry/terraform-registry/internal/db/repositories"
)

// RBACHandlers handles RBAC-related API endpoints
type RBACHandlers struct {
	rbacRepo *repositories.RBACRepository
}

// NewRBACHandlers creates a new RBAC handlers instance
func NewRBACHandlers(rbacRepo *repositories.RBACRepository) *RBACHandlers {
	return &RBACHandlers{rbacRepo: rbacRepo}
}

// ============================================================================
// Role Templates
// ============================================================================

// ListRoleTemplates returns all available role templates
// GET /api/v1/admin/role-templates
func (h *RBACHandlers) ListRoleTemplates(c *gin.Context) {
	templates, err := h.rbacRepo.ListRoleTemplates(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list role templates"})
		return
	}

	c.JSON(http.StatusOK, templates)
}

// GetRoleTemplate returns a single role template
// GET /api/v1/admin/role-templates/:id
func (h *RBACHandlers) GetRoleTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role template ID"})
		return
	}

	template, err := h.rbacRepo.GetRoleTemplate(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get role template"})
		return
	}

	if template == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role template not found"})
		return
	}

	c.JSON(http.StatusOK, template)
}

// CreateRoleTemplateRequest represents the request to create a role template
type CreateRoleTemplateRequest struct {
	Name        string   `json:"name" binding:"required"`
	DisplayName string   `json:"display_name" binding:"required"`
	Description string   `json:"description"`
	Scopes      []string `json:"scopes" binding:"required"`
}

// CreateRoleTemplate creates a new role template
// POST /api/v1/admin/role-templates
func (h *RBACHandlers) CreateRoleTemplate(c *gin.Context) {
	var req CreateRoleTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if name already exists
	existing, err := h.rbacRepo.GetRoleTemplateByName(c.Request.Context(), req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing template"})
		return
	}
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Role template with this name already exists"})
		return
	}

	template := &models.RoleTemplate{
		ID:          uuid.New(),
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: &req.Description,
		Scopes:      req.Scopes,
		IsSystem:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.rbacRepo.CreateRoleTemplate(c.Request.Context(), template); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role template"})
		return
	}

	c.JSON(http.StatusCreated, template)
}

// UpdateRoleTemplate updates an existing role template
// PUT /api/v1/admin/role-templates/:id
func (h *RBACHandlers) UpdateRoleTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role template ID"})
		return
	}

	existing, err := h.rbacRepo.GetRoleTemplate(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get role template"})
		return
	}
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role template not found"})
		return
	}
	if existing.IsSystem {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot modify system role templates"})
		return
	}

	var req CreateRoleTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing.DisplayName = req.DisplayName
	existing.Description = &req.Description
	existing.Scopes = req.Scopes
	existing.UpdatedAt = time.Now()

	if err := h.rbacRepo.UpdateRoleTemplate(c.Request.Context(), existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role template"})
		return
	}

	c.JSON(http.StatusOK, existing)
}

// DeleteRoleTemplate deletes a role template
// DELETE /api/v1/admin/role-templates/:id
func (h *RBACHandlers) DeleteRoleTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role template ID"})
		return
	}

	existing, err := h.rbacRepo.GetRoleTemplate(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get role template"})
		return
	}
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role template not found"})
		return
	}
	if existing.IsSystem {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete system role templates"})
		return
	}

	if err := h.rbacRepo.DeleteRoleTemplate(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete role template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role template deleted"})
}

// ============================================================================
// Mirror Approval Requests
// ============================================================================

// ListApprovalRequests lists all approval requests
// GET /api/v1/admin/approvals
func (h *RBACHandlers) ListApprovalRequests(c *gin.Context) {
	var orgID *uuid.UUID
	if orgIDStr := c.Query("organization_id"); orgIDStr != "" {
		id, err := uuid.Parse(orgIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
			return
		}
		orgID = &id
	}

	var status *models.ApprovalStatus
	if statusStr := c.Query("status"); statusStr != "" {
		s := models.ApprovalStatus(statusStr)
		status = &s
	}

	requests, err := h.rbacRepo.ListApprovalRequests(c.Request.Context(), orgID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list approval requests"})
		return
	}

	c.JSON(http.StatusOK, requests)
}

// GetApprovalRequest returns a single approval request
// GET /api/v1/admin/approvals/:id
func (h *RBACHandlers) GetApprovalRequest(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid approval request ID"})
		return
	}

	req, err := h.rbacRepo.GetApprovalRequest(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get approval request"})
		return
	}

	if req == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Approval request not found"})
		return
	}

	c.JSON(http.StatusOK, req)
}

// CreateApprovalRequestRequest represents the request to create an approval request
type CreateApprovalRequestRequest struct {
	MirrorConfigID    string  `json:"mirror_config_id" binding:"required"`
	ProviderNamespace string  `json:"provider_namespace" binding:"required"`
	ProviderName      *string `json:"provider_name"`
	Reason            string  `json:"reason"`
}

// CreateApprovalRequest creates a new approval request
// POST /api/v1/admin/approvals
func (h *RBACHandlers) CreateApprovalRequest(c *gin.Context) {
	var req CreateApprovalRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mirrorConfigID, err := uuid.Parse(req.MirrorConfigID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid mirror config ID"})
		return
	}

	// Get user ID from context
	var requestedBy *uuid.UUID
	if userIDStr, exists := c.Get("user_id"); exists {
		if idStr, ok := userIDStr.(string); ok {
			if id, err := uuid.Parse(idStr); err == nil {
				requestedBy = &id
			}
		}
	}

	// Get organization ID from context
	var orgID *uuid.UUID
	if orgIDStr, exists := c.Get("organization_id"); exists {
		if idStr, ok := orgIDStr.(string); ok {
			if id, err := uuid.Parse(idStr); err == nil {
				orgID = &id
			}
		}
	}

	approval := &models.MirrorApprovalRequest{
		ID:                uuid.New(),
		MirrorConfigID:    mirrorConfigID,
		OrganizationID:    orgID,
		RequestedBy:       requestedBy,
		ProviderNamespace: req.ProviderNamespace,
		ProviderName:      req.ProviderName,
		Reason:            req.Reason,
		Status:            models.ApprovalStatusPending,
		AutoApproved:      false,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := h.rbacRepo.CreateApprovalRequest(c.Request.Context(), approval); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create approval request"})
		return
	}

	c.JSON(http.StatusCreated, approval)
}

// ReviewApprovalRequest represents the request to review an approval
type ReviewApprovalRequest struct {
	Status string `json:"status" binding:"required"` // "approved" or "rejected"
	Notes  string `json:"notes"`
}

// ReviewApproval approves or rejects an approval request
// PUT /api/v1/admin/approvals/:id/review
func (h *RBACHandlers) ReviewApproval(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid approval request ID"})
		return
	}

	var req ReviewApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := models.ApprovalStatus(req.Status)
	if status != models.ApprovalStatusApproved && status != models.ApprovalStatusRejected {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status must be 'approved' or 'rejected'"})
		return
	}

	// Get reviewer ID from context
	var reviewerID uuid.UUID
	if userIDStr, exists := c.Get("user_id"); exists {
		if idStr, ok := userIDStr.(string); ok {
			if id, err := uuid.Parse(idStr); err == nil {
				reviewerID = id
			}
		}
	}

	if err := h.rbacRepo.UpdateApprovalStatus(c.Request.Context(), id, status, reviewerID, req.Notes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update approval status"})
		return
	}

	// Fetch updated approval
	approval, err := h.rbacRepo.GetApprovalRequest(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated approval"})
		return
	}

	c.JSON(http.StatusOK, approval)
}

// ============================================================================
// Mirror Policies
// ============================================================================

// ListMirrorPolicies lists all mirror policies
// GET /api/v1/admin/policies
func (h *RBACHandlers) ListMirrorPolicies(c *gin.Context) {
	var orgID *uuid.UUID
	if orgIDStr := c.Query("organization_id"); orgIDStr != "" {
		id, err := uuid.Parse(orgIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
			return
		}
		orgID = &id
	}

	policies, err := h.rbacRepo.ListMirrorPolicies(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list mirror policies"})
		return
	}

	c.JSON(http.StatusOK, policies)
}

// GetMirrorPolicy returns a single mirror policy
// GET /api/v1/admin/policies/:id
func (h *RBACHandlers) GetMirrorPolicy(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid policy ID"})
		return
	}

	policy, err := h.rbacRepo.GetMirrorPolicy(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get policy"})
		return
	}

	if policy == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Policy not found"})
		return
	}

	c.JSON(http.StatusOK, policy)
}

// CreateMirrorPolicyRequest represents the request to create a mirror policy
type CreateMirrorPolicyRequest struct {
	OrganizationID   *string `json:"organization_id"`
	Name             string  `json:"name" binding:"required"`
	Description      string  `json:"description"`
	PolicyType       string  `json:"policy_type" binding:"required"` // "allow" or "deny"
	UpstreamRegistry *string `json:"upstream_registry"`
	NamespacePattern *string `json:"namespace_pattern"`
	ProviderPattern  *string `json:"provider_pattern"`
	Priority         int     `json:"priority"`
	IsActive         bool    `json:"is_active"`
	RequiresApproval bool    `json:"requires_approval"`
}

// CreateMirrorPolicy creates a new mirror policy
// POST /api/v1/admin/policies
func (h *RBACHandlers) CreateMirrorPolicy(c *gin.Context) {
	var req CreateMirrorPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	policyType := models.PolicyType(req.PolicyType)
	if policyType != models.PolicyTypeAllow && policyType != models.PolicyTypeDeny {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Policy type must be 'allow' or 'deny'"})
		return
	}

	var orgID *uuid.UUID
	if req.OrganizationID != nil {
		id, err := uuid.Parse(*req.OrganizationID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
			return
		}
		orgID = &id
	}

	// Get creator ID from context
	var createdBy *uuid.UUID
	if userIDStr, exists := c.Get("user_id"); exists {
		if idStr, ok := userIDStr.(string); ok {
			if id, err := uuid.Parse(idStr); err == nil {
				createdBy = &id
			}
		}
	}

	policy := &models.MirrorPolicy{
		ID:               uuid.New(),
		OrganizationID:   orgID,
		Name:             req.Name,
		Description:      &req.Description,
		PolicyType:       policyType,
		UpstreamRegistry: req.UpstreamRegistry,
		NamespacePattern: req.NamespacePattern,
		ProviderPattern:  req.ProviderPattern,
		Priority:         req.Priority,
		IsActive:         req.IsActive,
		RequiresApproval: req.RequiresApproval,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		CreatedBy:        createdBy,
	}

	if err := h.rbacRepo.CreateMirrorPolicy(c.Request.Context(), policy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create policy"})
		return
	}

	c.JSON(http.StatusCreated, policy)
}

// UpdateMirrorPolicy updates an existing mirror policy
// PUT /api/v1/admin/policies/:id
func (h *RBACHandlers) UpdateMirrorPolicy(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid policy ID"})
		return
	}

	existing, err := h.rbacRepo.GetMirrorPolicy(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get policy"})
		return
	}
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Policy not found"})
		return
	}

	var req CreateMirrorPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	policyType := models.PolicyType(req.PolicyType)
	if policyType != models.PolicyTypeAllow && policyType != models.PolicyTypeDeny {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Policy type must be 'allow' or 'deny'"})
		return
	}

	existing.Name = req.Name
	existing.Description = &req.Description
	existing.PolicyType = policyType
	existing.UpstreamRegistry = req.UpstreamRegistry
	existing.NamespacePattern = req.NamespacePattern
	existing.ProviderPattern = req.ProviderPattern
	existing.Priority = req.Priority
	existing.IsActive = req.IsActive
	existing.RequiresApproval = req.RequiresApproval
	existing.UpdatedAt = time.Now()

	if err := h.rbacRepo.UpdateMirrorPolicy(c.Request.Context(), existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update policy"})
		return
	}

	c.JSON(http.StatusOK, existing)
}

// DeleteMirrorPolicy deletes a mirror policy
// DELETE /api/v1/admin/policies/:id
func (h *RBACHandlers) DeleteMirrorPolicy(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid policy ID"})
		return
	}

	if err := h.rbacRepo.DeleteMirrorPolicy(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete policy"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Policy deleted"})
}

// EvaluatePolicyRequest represents a request to evaluate policies for a provider
type EvaluatePolicyRequest struct {
	Registry  string `json:"registry" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
	Provider  string `json:"provider" binding:"required"`
}

// EvaluatePolicy evaluates policies for a given provider
// POST /api/v1/admin/policies/evaluate
func (h *RBACHandlers) EvaluatePolicy(c *gin.Context) {
	var req EvaluatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var orgID *uuid.UUID
	if orgIDStr := c.Query("organization_id"); orgIDStr != "" {
		id, err := uuid.Parse(orgIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
			return
		}
		orgID = &id
	}

	result, err := h.rbacRepo.EvaluatePolicies(c.Request.Context(), orgID, req.Registry, req.Namespace, req.Provider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to evaluate policies"})
		return
	}

	c.JSON(http.StatusOK, result)
}
