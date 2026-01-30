# Changelog

All notable changes to the Terraform Registry project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2026-01-30

### Added - Phase 2: Module Registry Protocol

#### Storage Layer

- Storage interface with Upload, Download, Delete, GetURL, Exists, and GetMetadata operations
- Local filesystem storage implementation with SHA256 checksum calculation
- Storage factory with registration pattern for pluggable backends
- File path structure: `modules/{namespace}/{name}/{system}/{version}.tar.gz`
- Support for direct file serving via HTTP

#### Data Layer

- Module and ModuleVersion data models
- ModuleRepository with CRUD operations, version listing, and search
- OrganizationRepository for multi-tenancy support
- Database queries using prepared statements for security
- Download count tracking with async increment

#### Validation & Security

- Semantic versioning validation using HashiCorp go-version library
- Archive validation (gzip/tar format, size limits, path traversal prevention)
- SHA256 checksum calculation and verification
- Maximum archive size limit (100MB)
- Malicious path detection (.git directories, absolute paths, path traversal)

#### HTTP Handlers (Terraform Protocol)

- **List Versions**: `GET /v1/modules/:namespace/:name/:system/versions`
  - Returns Terraform-compliant JSON with version list
- **Download Module**: `GET /v1/modules/:namespace/:name/:system/:version/download`
  - Returns 204 No Content with `X-Terraform-Get` header
  - Async download counter increment
- **Upload Module**: `POST /api/v1/modules`
  - Multipart form upload with validation
  - Automatic module creation/update
  - Duplicate version detection
- **Search Modules**: `GET /api/v1/modules/search`
  - PostgreSQL LIKE queries with pagination
  - Filter by namespace, system, query string
- **Serve Files**: `GET /v1/files/*filepath`
  - Direct file serving for local storage
  - Proper content headers and streaming

#### Dependencies

- Added `github.com/hashicorp/go-version` for semantic versioning

### Changed

- Router updated to initialize storage backend and wire module handlers
- Removed placeholder module endpoints, replaced with functional implementations

### Fixed

- Import cycle issue in storage package resolved using registration pattern

### Implementation Details

- **Files Created**: 15 new files (~1,500 lines of code)
- **Terraform Protocol**: Fully compliant with Module Registry Protocol
- **Testing**: Manual end-to-end testing with curl and actual module uploads
- **Storage**: Local filesystem working, Azure/S3 deferred to future phases

## [0.1.0] - 2026-01-29

### Added - Phase 1: Project Foundation & Backend Core

#### Project Structure

- Complete project directory structure for backend, frontend, Azure DevOps extension, and deployments
- MIT License
- Comprehensive README.md with project overview
- Implementation plan documentation (IMPLEMENTATION_PLAN.md)
- Session notes and tracking (SESSION_NOTES.md, SESSION_1_UPDATE.md)
- Git ignore configuration for Go, Node.js, and IDEs

#### Backend Application (Go)

- Go module initialization with all required dependencies
- Application entry point with command handling (`serve`, `migrate`, `version`)
- Configuration management system using Viper
  - Support for YAML files and environment variables
  - Environment variable expansion for secrets
  - Comprehensive validation
  - Default values for all settings
- Explicit environment variable binding for nested configuration structures
- Debug logging for database configuration (with password masking)

#### Database Layer

- PostgreSQL schema with 11 tables:
  - Core: `organizations`, `users`, `api_keys`, `organization_members`
  - Modules: `modules`, `module_versions`
  - Providers: `providers`, `provider_versions`, `provider_platforms`
  - Analytics: `download_events`, `audit_logs`
- Database migrations using golang-migrate with embedded migration files
- Automatic migration execution on startup
- Support for multi-tenancy (can be enabled/disabled via config)
- Default organization for single-tenant mode
- Indexes for performance optimization

#### HTTP Server (Gin Framework)

- Health check endpoint (`/health`) with database connectivity test
- Readiness check endpoint (`/ready`)
- Terraform service discovery endpoint (`/.well-known/terraform.json`)
- API version endpoint (`/version`)
- Placeholder endpoints for Module Registry Protocol (Phase 2)
- Placeholder endpoints for Provider Registry Protocol (Phase 3)
- Placeholder endpoints for Network Mirror Protocol (Phase 3)
- Placeholder endpoints for Admin API (Phases 4-5)
- CORS middleware with configurable origins
- Logging middleware (JSON and text formats)
- Graceful shutdown on SIGINT/SIGTERM

#### Docker Support

- Multi-stage Dockerfile for minimal production images
- Docker Compose configuration with:
  - PostgreSQL 16 database
  - Backend application
  - Optional Prometheus for metrics
  - Optional Grafana for visualization
- Health checks for all services
- Volume mounts for data persistence
- Network isolation
- Environment variable configuration

#### Configuration

- Comprehensive configuration system supporting:
  - Server settings (host, port, timeouts, TLS)
  - Database connection (host, port, credentials, SSL mode)
  - Storage backends (Azure Blob, S3, local filesystem)
  - Authentication (API keys, OIDC, Azure AD)
  - Multi-tenancy settings
  - Security (CORS, rate limiting, TLS)
  - Logging (level, format, output)
  - Telemetry (metrics, tracing, profiling)
- Example configuration file (config.example.yaml)

### Bug Fixes

- PostgreSQL connection issue: Viper's `AutomaticEnv()` not working with `Unmarshal()` for nested structures
  - Solution: Added explicit environment variable bindings in `bindEnvVars()` function
- Gin routing conflict between Provider Registry and Network Mirror endpoints
  - Solution: Moved Network Mirror endpoints to `/terraform/providers/` path to avoid parameter conflicts

### Technical Details

- **Language**: Go 1.22+
- **Framework**: Gin web framework
- **Database**: PostgreSQL 16
- **Configuration**: Viper
- **Migrations**: golang-migrate
- **Containerization**: Docker with multi-stage builds
- **Lines of Code**: ~2500+ across all files

### Deployment

- Docker Compose deployment fully functional
- All health checks passing
- Database migrations automatic on startup
- Ready for Phase 2 implementation

### Testing

- Manual endpoint testing completed:
  - ✅ `/health` - Returns healthy status with database check
  - ✅ `/.well-known/terraform.json` - Service discovery working
  - ✅ `/version` - API version information
  - ✅ `/ready` - Readiness check
- Docker containers running and healthy:
  - ✅ `terraform-registry-db` (PostgreSQL)
  - ✅ `terraform-registry-backend` (Go application)

### Documentation

- Complete implementation plan with 9 phases
- Session notes with setup instructions
- README with quick start guide
- Configuration examples
- API endpoint documentation in implementation plan

### Next Steps

- **Phase 2**: Module Registry Protocol implementation
  - Storage abstraction layer
  - Module upload and versioning
  - Module download endpoints
  - Comprehensive testing

---
