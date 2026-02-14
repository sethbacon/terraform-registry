<!-- markdownlint-disable MD024 -->

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned

- Phase 7: Comprehensive documentation and testing suite
- Phase 8: Production polish (monitoring, observability, security hardening)

## [1.0.0] - 2026-02-10 - Sessions 22-25 (Phase 6 Complete)

### Added

- **Phase 6 Complete: All Storage Backends, Deployments, and SCM Integrations**

- **Session 22: Bitbucket Data Center SCM Integration**
  - Bitbucket Data Center connector with full API integration (636 LOC)
  - Repository browsing, search, and tag enumeration with commit SHA resolution
  - Webhook creation and management for automated publishing
  - Personal Access Token (PAT) authentication for Bitbucket (no OAuth required)
  - Database migration 000027: Added `bitbucket_host` column to scm_providers
  - Frontend support for Bitbucket DC with dynamic form fields
  - Extended SCMProvider type with bitbucket_host field
  - SCM provider type constant and error handling for Bitbucket

- **Session 23: Storage Configuration in Terraform**
  - Added storage backend variable support to all 3 Terraform configs (AWS, Azure, GCP)
  - Each cloud defaults to native storage with zero additional configuration:
    - AWS: S3 with IAM role authentication
    - Azure: Azure Blob Storage with Direct Account Key
    - GCP: Google Cloud Storage with Workload Identity
  - All 4 storage backends (S3, Azure Blob, GCS, Local) available in every cloud
  - Conditional secrets management via Secrets Manager/Secret Manager/Key Vault
  - Storage configuration merged into task definitions/container apps via locals
  - Enhanced RBAC with storage-specific scope support (storage:read, storage:write, storage:manage)

- **Session 24: API Key Frontend Lifecycle Management**
  - Optional expiration date field in API key creation dialog
  - API key edit dialog for updating name, scopes, and expiration
  - API key rotation with grace period options (1-72h slider)
  - Expiration indicators: Red "Expired", orange "Expires soon" (within 7 days)
  - Scopes column with chip display and overflow tooltip
  - Helper functions for expiration status and datetime conversion
  - `rotateAPIKey()` method in API service
  - Copy-to-clipboard for newly rotated API key values

- **Session 25: Storage Configuration Azure Bug Fix & Phase 6 Completion**
  - Fixed critical bug in Azure Terraform deployment:
    - Corrected `azurerm_storage_account` resource reference to `azurerm_storage_account.main.primary_access_key`
    - Ensures proper Container App secret management
    - Resolves failures related to storage account key initialization
  - Cleaned up database migration state for Bitbucket DC
  - Removed duplicate/empty migration files

### Summary of Phase 6 Deliverables

**Storage Backends (Sessions 16-18):**

- ✅ Azure Blob Storage backend with SAS token support and CDN URLs
- ✅ AWS S3-compatible backend with presigned URLs and multipart uploads
- ✅ Google Cloud Storage backend with signed URLs and resumable uploads
- ✅ All backends support multiple authentication methods
- ✅ SHA256 checksum calculation and verification for all uploads

**Deployment Configurations (Sessions 20-21):**

- ✅ Docker Compose (dev and production)
- ✅ Kubernetes + Kustomize with base and environment overlays (dev, prod)
- ✅ Helm Chart with configurable values and all storage backends
- ✅ Azure Container Apps with Bicep templates
- ✅ AWS ECS Fargate with CloudFormation stack
- ✅ Google Cloud Run with VPC connectors
- ✅ Standalone binary deployment with systemd and nginx
- ✅ Terraform IaC for AWS, Azure, and GCP with full storage configuration

**SCM Enhancements & API Keys (Sessions 22-24):**

- ✅ Bitbucket Data Center as 4th SCM provider alongside GitHub, Azure DevOps, GitLab
- ✅ Complete API key lifecycle management with expiration and rotation
- ✅ Frontend UI for API key creation, editing, and rotation
- ✅ Scope management with checkboxes and overflow handling

### Milestone

- **✅ Phase 6 Complete**: Enterprise-grade Terraform Registry fully implemented
  - All 3 Terraform protocols (Module, Provider, Mirror)
  - 4 storage backends with multi-cloud support
  - 4 SCM providers with automated publishing
  - Complete deployment options (Docker, K8s, PaaS, binary)
  - Full authentication and RBAC system
  - React SPA with comprehensive admin UI
  - Ready for Phase 7 (Testing & Documentation) and Phase 8 (Production Polish)

## [0.9.0] - 2026-02-06 - Session 15

### Added

- **Provider Network Mirroring - Complete Implementation (Phase 5C)**
  - Full `syncProvider()` implementation with actual provider binary downloads
  - Downloads provider binaries from upstream registries (registry.terraform.io)
  - Stores binaries in local storage backend
  - Creates provider, version, and platform records in database
  - SHA256 checksum verification for all downloaded files
  - GPG signature verification using ProtonMail/go-crypto library
  - Mirrored provider tracking tables (migration 011):
    - `mirrored_providers`: tracks which providers came from which mirror
    - `mirrored_provider_versions`: tracks version sync status and verification
  - Organization support for mirror configurations
  - Connected TriggerSync API to background sync job
  - Enhanced RBAC with mirror-specific scopes:
    - `mirrors:read`: View mirror configurations and sync status
    - `mirrors:manage`: Create, update, delete mirrors and trigger syncs
  - Audit logging for all mirror operations via middleware
  - Mirror Management UI page (frontend):
    - List all mirror configurations with status
    - Create/edit/delete mirror configurations
    - Trigger manual sync
    - View sync status and history
    - Namespace and provider filters
    - Navigation in admin sidebar

### Milestone

- **Phase 5C Complete**: Provider network mirroring fully implemented with GPG verification, RBAC, audit logging, and UI

## [0.8.0] - 2026-02-04 - Session 14

### Added

- **Provider Network Mirroring Infrastructure (Phase 5C Session 14)**
  - Database migration 010: `mirror_configurations` and `mirror_sync_history` tables
  - Upstream registry client with Terraform Provider Registry Protocol support
  - Service discovery for upstream registries
  - Provider version enumeration from upstream
  - Package download URL resolution
  - Mirror configuration models and repository layer
  - Full CRUD API endpoints for mirror management (`/api/v1/admin/mirrors/*`)
  - Background sync job infrastructure with 10-minute interval checks
  - Sync history tracking and status monitoring
  - Framework ready for actual provider downloads

### Fixed

- Fixed migration system: renamed migrations to `.up.sql`/`.down.sql` convention
- Created `fix-migration` utility for cleaning dirty migration states

## [0.7.0] - 2026-02-04 - Session 13

### Added

- **SCM Frontend UI & Comprehensive Debugging (Phase 5A Session 13)**
  - Complete SCM provider management interface
  - Repository browser with search and filtering
  - Publishing wizard with commit pinning
  - Description field for module uploads
  - Helper text and tab-specific guidelines for all upload forms
  - Authentication-gated upload buttons on modules/providers pages
  - Network mirrored provider badges for visual differentiation
  - ISO 8601 date formatting for international compatibility

### Fixed

- **Single-Tenant Mode Issues**:
  - Organization filtering now correctly skips when multi-tenancy is disabled
  - Search handlers conditionally check MultiTenancy.Enabled configuration
  - Repository layer handles empty organization ID with proper SQL WHERE clauses
  
- **Frontend Data Visibility**:
  - Module and provider search results now include computed latest_version and download_count
  - Backend aggregates version data and download statistics for search results
  - Fixed undefined values display with proper fallbacks (N/A, 0)
  - Provider download counts correctly handle platform-level aggregation

- **Navigation & Routing**:
  - Fixed route parameters in ModuleDetailPage (provider→system)
  - Fixed route parameters in ProviderDetailPage (name→type)
  - Dashboard cards now navigate correctly to respective pages
  - Quick action cards navigate with state to select correct upload tab

- **Date Display**:
  - Changed from localized dates to ISO 8601 format (YYYY-MM-DD)
  - Applied consistently across all version displays
  - Backend now includes published_at in version responses

- **Provider Pages**:
  - Fixed versions response structure handling (direct array vs. nested)
  - Fixed TypeScript linting errors (unused imports, type mismatches)
  - Provider cards now use provider.type instead of non-existent provider.name
  - Added organization_name and published_at fields to Provider/ProviderVersion types

- **Upload Interface**:
  - Added description field to module upload form
  - FormData creation fixed for proper API compatibility
  - Tab-specific upload guidelines implemented
  - Removed duplicate generic guidelines section

### Technical Details

- Backend search endpoints now query versions to compute latest_version
- Module versions: Sum download_count across all versions
- Provider versions: Platform-level downloads (set to 0 pending aggregation implementation)
- Frontend uses computed values from search results instead of missing model fields
- All dates use RFC3339 format in API responses
- Network mirror differentiation uses provider.source field presence

### Phase Completion

- ✅ **Phase 5A Complete**: SCM integration fully implemented with production-ready UI

## [0.6.0] - 2024-01-XX - Session 11

### Added

- **SCM OAuth Flows & Repository Operations (Phase 5A Session 11)**
  - GitHub connector with complete OAuth 2.0 authorization flow
  - GitHub repository listing, searching, and browsing
  - GitHub branch and tag operations with commit resolution
  - GitHub archive download (tarball/zipball)
  - Azure DevOps connector with OAuth 2.0 flow
  - Azure DevOps project and repository browsing
  - Azure DevOps branch, tag, and commit operations
  - Azure DevOps archive download functionality
  - GitLab connector with OAuth 2.0 flow and token refresh
  - Token encryption/decryption using AES-256-GCM
  - SCM repository data access layer
  - Support for self-hosted SCM instances
  - Connector registry with factory pattern
  - Pagination support for all list operations
  - Repository search functionality

### Technical Details

- **GitHub Integration**:
  - OAuth app flow with code exchange
  - REST API v3 with proper versioning headers
  - Repository filtering and sorting
  - Tag-to-commit SHA resolution
  - Archive download with format selection
  
- **Azure DevOps Integration**:
  - Azure DevOps Services OAuth with JWT assertions
  - Project-based repository organization
  - Git refs API for branches and tags
  - Token refresh support with expiry tracking
  
- **GitLab Integration**:
  - Standard OAuth 2.0 flow
  - Token refresh capability
  - Self-hosted GitLab support
  - Stub implementations ready for completion

### Infrastructure

- Connector interface with consistent API across providers
- Error handling with wrapped remote API errors
- Token expiry checking and validation
- Secure credential management

## [0.5.1] - 2024-01-XX - Session 10

### Added

- **SCM Integration Foundation (Phase 5A Session 10)**
  - Database migration for SCM integration (008_scm_integration.sql)
  - SCM provider configurations table
  - User OAuth tokens table with encryption
  - Module-to-repository linking table
  - Webhook event logging table
  - Version immutability violations tracking
  - SCM provider interface and types
  - Connector abstraction layer
  - Token encryption utilities (AES-256-GCM)
  - Connector registry/factory pattern
  - Error definitions for SCM operations

### Changed

- Extended module_versions table with SCM metadata (commit SHA, source URL, tag)

## [0.5.0] - 2024-01-XX - Session 9

### Added

- **Frontend SPA (Phase 5 Complete)**
  - Complete React 18+ TypeScript application with Vite
  - Material-UI component library integration
  - Module browsing and search pages with pagination
  - Provider browsing and search pages with pagination
  - Module and provider detail pages with version history
  - Admin dashboard with system statistics
  - User management UI (list, create, edit, delete)
  - Organization management UI (list, create, edit, delete)
  - API key management UI with scope configuration
  - Upload interface for modules and providers
  - Authentication context with JWT support
  - Protected routes for admin functionality
  - Responsive design with light theme (dark mode ready)
  - Comprehensive error handling and loading states
  - Optimistic UI updates for better UX
  - Vite dev server with backend proxy on port 3000

### Changed

- Updated implementation plan to reflect frontend completion
- Renamed VCS (Version Control System) to SCM (Source Code Management) throughout project

## [0.4.0] - 2024-01-XX - Session 8

### Added

- **User & Organization Management (Phase 4 Complete)**
  - User management REST endpoints (list, search, create, update, delete)
  - Organization management REST endpoints (list, search, create, update, delete)
  - Organization membership management (add, update, remove members)
  - Role-based organization membership (owner, admin, member, viewer)
  - User search by email and name
  - Organization search by name
  - Pagination support for user and organization listings
  - Audit logging for all administrative actions
  - RBAC middleware integration for endpoint protection

### Changed

- Enhanced authentication middleware with organization context
- Improved API key scoping for multi-tenant operations
- Updated database schema with organization member roles

## [0.3.0] - 2024-01-XX - Session 7

### Added

- **Authentication & Authorization (Phase 4)**
  - JWT-based authentication system
  - API key authentication with bcrypt hashing
  - OIDC provider support (generic)
  - Azure AD / Entra ID integration
  - Role-based access control (RBAC) middleware
  - Scope-based authorization for fine-grained permissions
  - API key management endpoints (create, list, delete)
  - Authentication endpoints (login, logout, refresh)
  - Token encryption for OAuth tokens
  - Configurable single-tenant vs multi-tenant mode
  - User model with OIDC subject support
  - Organization model for multi-tenancy
  - API key model with expiration and scopes

### Changed

- Updated router with authentication middleware
- Protected admin endpoints with proper authorization
- Enhanced database schema with auth tables

## [0.2.0] - 2024-01-XX - Sessions 4-6

### Added

- **Provider Registry Protocol (Phase 3 Complete)**
  - Provider version listing endpoint
  - Provider binary download endpoint with platform support
  - Provider upload endpoint with validation
  - GPG signature verification framework
  - Provider platform matrix support (OS/Architecture)
  - SHA256 checksum validation for provider binaries
  - Provider data models and repositories
  - Provider search functionality

- **Network Mirror Protocol (Phase 3 Complete)**
  - Version index endpoint for provider mirroring
  - Platform index endpoint for specific versions
  - JSON response formatting per Terraform mirror spec
  - Hostname-based provider routing
  - Integration with existing provider storage

### Changed

- Enhanced storage abstraction to support provider binaries
- Updated database schema with provider tables
- Improved validation for provider uploads

## [0.1.0] - 2024-01-XX - Sessions 1-3

### Added

- **Module Registry Protocol (Phase 2 Complete)**
  - Module version listing endpoint
  - Module download endpoint with redirect support
  - Module upload endpoint with validation
  - Module search with pagination
  - Direct file serving for local storage
  - SHA256 checksum generation and verification
  - Semantic version validation
  - Archive format validation (tar.gz, zip)
  - Security checks for path traversal
  - Download tracking and analytics
  - Module data models and repositories

- **Project Foundation (Phase 1 Complete)**
  - Go backend with Gin framework
  - PostgreSQL database with migrations
  - Configuration management (YAML + environment variables)
  - Service discovery endpoint (/.well-known/terraform.json)
  - Health check endpoint
  - Docker Compose setup for local development
  - Dockerfile for backend service
  - Storage abstraction layer
  - Local filesystem storage backend
  - Organization-based multi-tenancy support

### Infrastructure

- PostgreSQL database schema with migrations
- Database repositories for data access layer
- HTTP middleware (logging, CORS, error handling)
- Request validation utilities
- Checksum utilities for file integrity

[Unreleased]: https://github.com/yourusername/terraform-registry/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/yourusername/terraform-registry/compare/v0.9.0...v1.0.0
[0.9.0]: https://github.com/yourusername/terraform-registry/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/yourusername/terraform-registry/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/yourusername/terraform-registry/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/yourusername/terraform-registry/compare/v0.5.1...v0.6.0
[0.5.1]: https://github.com/yourusername/terraform-registry/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/yourusername/terraform-registry/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/yourusername/terraform-registry/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/yourusername/terraform-registry/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/yourusername/terraform-registry/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/yourusername/terraform-registry/releases/tag/v0.1.0
