# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned
- Phase 5B: Azure DevOps pipeline extension for publishing
- Phase 5C: Provider network mirroring with enhanced security roles (see IMPLEMENTATION_PLAN.md Phase 5C)
- Phase 6: Azure Blob Storage and S3-compatible storage backends
- Phase 6: Deployment configurations (Kubernetes, Helm, Azure Container Apps)
- Phase 7: Comprehensive documentation and testing suite
- Phase 8: Production polish (monitoring, observability, security hardening)

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

[Unreleased]: https://github.com/yourusername/terraform-registry/compare/v0.5.0...HEAD
[0.5.0]: https://github.com/yourusername/terraform-registry/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/yourusername/terraform-registry/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/yourusername/terraform-registry/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/yourusername/terraform-registry/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/yourusername/terraform-registry/releases/tag/v0.1.0
