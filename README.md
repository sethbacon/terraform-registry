<!-- markdownlint-disable MD024 -->

# Enterprise Terraform Registry

A fully-featured, enterprise-grade Terraform registry implementing all three HashiCorp protocols with modern web UI and multi-tenancy support.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://go.dev/)
[![React](https://img.shields.io/badge/React-18+-61DAFB?logo=react)](https://react.dev/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5+-3178C6?logo=typescript)](https://www.typescriptlang.org/)

## 🚀 Features

### ✅ Fully Implemented

#### Terraform Protocol Support

- **Module Registry Protocol** - Complete implementation for hosting and discovering Terraform modules
- **Provider Registry Protocol** - Full provider hosting with platform-specific binaries
- **Provider Network Mirror Protocol** - Efficient provider mirroring and caching

#### Authentication & Authorization

- **API Key Authentication** - Secure token-based access with scoped permissions
- **OIDC Integration** - Generic OIDC provider support for SSO
- **Azure AD / Entra ID** - Native Azure Active Directory integration
- **Role-Based Access Control (RBAC)** - Fine-grained permissions with organization roles

#### Multi-Tenancy

- **Organization Management** - Isolated namespaces for teams and projects
- **User Management** - Comprehensive user administration
- **Organization Membership** - Role-based team collaboration (owner, admin, member, viewer)
- **Configurable Modes** - Single-tenant or multi-tenant deployment

#### SCM Integration

- **GitHub Integration** - Connect modules to GitHub repositories with OAuth
- **Azure DevOps Integration** - Native support for Azure Repos
- **GitLab Integration** - Full GitLab repository support
- **Bitbucket Data Center** - Support for self-hosted Bitbucket instances with PAT authentication
- **Webhook Support** - Automatic publishing on repository events
- **Immutable Publishing** - Version control integration with commit SHA pinning
- **Module README Support** - Automatic README extraction and rendering from module sources

#### Provider Mirroring

- **Upstream Registry Client** - Mirror providers from registry.terraform.io
- **Automated Sync** - Scheduled synchronization of provider versions (10-minute intervals)
- **Configurable Filters** - Namespace and provider-specific mirroring
- **Sync History** - Track mirror operations and status
- **Manual Triggers** - On-demand synchronization via UI or API
- **GPG Signature Verification** - Cryptographic validation of provider binaries
- **SHA256 Checksums** - File integrity verification for all downloads
- **Enhanced RBAC** - Mirror-specific scopes (mirrors:read, mirrors:manage)
- **Audit Logging** - Full tracking of all mirror operations
- **Mirror Management UI** - Complete admin interface for mirror configuration

#### Storage Backends

- **Local Filesystem** - Direct file serving for development and simple deployments
- **Azure Blob Storage** - Cloud storage with SAS tokens, CDN URLs, and flexible access tiers
- **AWS S3 / S3-Compatible** - Support for AWS S3, MinIO, DigitalOcean Spaces with presigned URLs and multipart uploads
- **Google Cloud Storage** - Native GCS integration with signed URLs and resumable uploads
- **Pluggable Architecture** - Extensible storage interface for adding new backends
- **Storage Configuration UI** - Admin dashboard for configuring and switching between storage backends
- **Multiple Auth Methods** - Support for IAM roles, service accounts, workload identity, and explicit credentials

#### API Key Management

- **API Key Lifecycle** - Complete key management with creation, rotation, and deletion
- **Expiration Dates** - Set optional expiration dates for automatic key invalidation
- **Key Rotation** - Rotate keys with grace periods (1-72 hours) for seamless transitions
- **Scope Management** - Fine-grained permission control for each key
- **Scope-Based Authorization** - Keys can be restricted to specific operations (modules:read, modules:write, providers:read, providers:write, mirrors:manage, admin:*)

#### Deployment Options

- **Docker Compose** - Complete development and production setups with all services
- **Kubernetes + Kustomize** - Production-ready manifests with base and environment-specific overlays (dev, production)
- **Helm Chart** - Fully parameterized Helm chart with support for all storage backends and configurations
- **Azure Container Apps** - Bicep templates for Azure Container Apps deployment
- **AWS ECS Fargate** - CloudFormation stack with VPC, RDS PostgreSQL, ALB, and auto-scaling
- **Google Cloud Run** - Knative services with Cloud SQL, Cloud Storage, and Secret Manager integration
- **Standalone Binary** - Systemd service with Nginx reverse proxy for production deployments
- **Terraform IaC** - Complete Infrastructure-as-Code for AWS, Azure, and GCP with parameterized storage backend configuration

#### Modern Web Interface

- **React 18+ SPA** - Fast, responsive single-page application
- **TypeScript** - Full type safety across the frontend
- **Material-UI** - Professional, accessible component library
- **Module Browser** - Search, filter, and explore modules with pagination and README rendering
- **Provider Browser** - Discover and manage provider versions with platform information
- **Admin Dashboard** - System statistics and management tools
- **Upload Interface** - Easy module and provider publishing with SCM linking
- **SCM Management** - Connect and configure GitHub, Azure DevOps, GitLab, and Bitbucket repositories
- **Mirror Management** - Configure and trigger provider synchronization from upstream registries
- **API Key Management** - Create, edit, rotate, and expire API keys with scope controls
- **Responsive Design** - Works on desktop, tablet, and mobile

### 📋 Planned

#### Testing & Documentation (Phase 7)

- Unit and integration tests for backend (target 80%+ coverage)
- E2E tests for critical user flows
- Comprehensive API documentation
- Deployment guides for all platforms
- Troubleshooting and FAQ documentation
- Contributing guidelines and development setup

#### Production Polish (Phase 8)

- OpenTelemetry instrumentation and Prometheus metrics
- Grafana dashboards for monitoring
- Email alerts for API key expiration
- Performance optimization and benchmarks
- Security hardening review
- Open source license attribution audit

#### Azure DevOps Extension (Phase 5B - Deferred)

- Azure DevOps pipeline task for publishing
- OIDC authentication with workload identity
- Visual Studio Marketplace distribution

## 🏗️ Architecture

```txt
┌──────────────────────────────────────────────────────────────────┐
│                    React TypeScript SPA                          │
│  Module Browser │ Provider Browser │ Admin Dashboard │ Auth UI  │
└────────────────────────────┬─────────────────────────────────────┘
                             │
                             │ REST API / Protocol Endpoints
                             │
┌────────────────────────────▼─────────────────────────────────────┐
│                      Go 1.24 Backend (Gin)                       │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐         │
│  │ Modules  │  │Providers │  │  Mirror  │  │  Admin   │         │
│  │   API    │  │   API    │  │   API    │  │   API    │         │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘         │
│  ┌─────────────────────────────────────────────────────────┐     │
│  │      Authentication & Authorization                     │     │
│  │  JWT │ API Keys │ OIDC │ Azure AD │ RBAC  │ Audit Log   │     │
│  └─────────────────────────────────────────────────────────┘     │
│  ┌─────────────────────────────────────────────────────────┐     │
│  │         SCM Integration & Processing                    │     │
│  │  GitHub │ Azure DevOps │ GitLab │ Bitbucket DC          │     │
│  └─────────────────────────────────────────────────────────┘     │
│  ┌─────────────────────────────────────────────────────────┐     │
│  │           Storage Abstraction Layer                     │     │
│  │  Local │ Azure Blob │ S3-Compatible │ GCS              │     │
│  └─────────────────────────────────────────────────────────┘     │
└────────────────────────────┬─────────────────────────────────────┘
                             │
                    ┌────────┴────────┐
                    │   PostgreSQL    │
                    │   Database      │
                    └─────────────────┘
```

## 📦 Installation

### Prerequisites

- Go 1.24 or later
- Node.js 18+ and npm
- PostgreSQL 14+
- Docker & Docker Compose (for containerized deployment)

### Quick Start with Docker Compose

```bash
# Clone the repository
git clone https://github.com/yourusername/terraform-registry.git
cd terraform-registry

# Start all services
cd deployments
docker-compose up -d

# Backend: http://localhost:8080
# PostgreSQL: localhost:5432
```

### Manual Setup

#### Backend

```bash
cd backend

# Install dependencies
go mod download

# Set up configuration
cp config.example.yaml config.yaml
# Edit config.yaml with your settings

# Run database migrations
go run cmd/server/main.go migrate up

# Start the server
go run cmd/server/main.go serve
```

#### Frontend

```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build
```

## 🔧 Configuration

### Environment Variables

All configuration can be set via environment variables (prefix: `TFR_`) or YAML config file.

```bash
# Database
export TFR_DATABASE_HOST=localhost
export TFR_DATABASE_PORT=5432
export TFR_DATABASE_USER=registry
export TFR_DATABASE_PASSWORD=your_password
export TFR_DATABASE_NAME=terraform_registry
export TFR_DATABASE_SSL_MODE=disable

# Server
export TFR_SERVER_PORT=8080
export TFR_SERVER_HOST=0.0.0.0
export TFR_SERVER_BASE_URL=http://localhost:8080

# Storage (choose one: local | azure | s3 | gcs)
export TFR_STORAGE_DEFAULT_BACKEND=local
export TFR_STORAGE_LOCAL_BASE_PATH=/var/lib/terraform-registry
export TFR_STORAGE_LOCAL_SERVE_DIRECTLY=true

# Azure Blob Storage (if using azure backend)
export TFR_STORAGE_AZURE_ACCOUNT_NAME=myaccount
export TFR_STORAGE_AZURE_CONTAINER_NAME=terraform-registry
export TFR_STORAGE_AZURE_ACCOUNT_KEY=your_key

# AWS S3 (if using s3 backend)
export TFR_STORAGE_S3_BUCKET=terraform-registry
export TFR_STORAGE_S3_REGION=us-east-1
export TFR_STORAGE_S3_AUTH_METHOD=default  # default, static, oidc, assume_role

# Google Cloud Storage (if using gcs backend)
export TFR_STORAGE_GCS_BUCKET=terraform-registry
export TFR_STORAGE_GCS_PROJECT_ID=my-project
export TFR_STORAGE_GCS_AUTH_METHOD=default  # default, service_account, workload_identity

# Authentication
export TFR_AUTH_API_KEYS_ENABLED=true
export TFR_AUTH_OIDC_ENABLED=false
export TFR_AUTH_AZURE_AD_ENABLED=false

# Multi-tenancy
export TFR_MULTI_TENANCY_ENABLED=false
export TFR_MULTI_TENANCY_DEFAULT_ORGANIZATION=default

# Security
export TFR_JWT_SECRET=your_jwt_secret
export ENCRYPTION_KEY=your_32_byte_encryption_key
```

### Configuration File

See `backend/config.example.yaml` for a complete configuration reference.

## 📚 Usage

### Using with Terraform

```hcl
# Configure the registry
terraform {
  required_providers {
    mycloud = {
      source  = "registry.example.com/myorg/mycloud"
      version = "~> 1.0"
    }
  }
}

# Use modules from the registry
module "vpc" {
  source  = "registry.example.com/myorg/vpc/aws"
  version = "2.1.0"
  
  cidr_block = "10.0.0.0/16"
}
```

### Publishing Modules

#### Via UI

1. Navigate to Admin → Upload
2. Select module archive (tar.gz or zip)
3. Fill in metadata (namespace, name, system, version)
4. Click Upload

#### Via API

```bash
curl -X POST https://registry.example.com/api/v1/modules \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -F "file=@module.tar.gz" \
  -F "namespace=myorg" \
  -F "name=vpc" \
  -F "system=aws" \
  -F "version=1.0.0"
```

### Publishing Providers

```bash
curl -X POST https://registry.example.com/api/v1/providers \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -F "file=@terraform-provider-mycloud_1.0.0_linux_amd64.zip" \
  -F "namespace=myorg" \
  -F "type=mycloud" \
  -F "version=1.0.0" \
  -F "os=linux" \
  -F "arch=amd64" \
  -F "gpg_public_key=@public_key.asc"
```

## 🧪 Development

### Running Tests

```bash
# Backend tests
cd backend
go test ./...

# Frontend tests
cd frontend
npm test
```

### Building

```bash
# Build backend binary
cd backend
go build -o terraform-registry cmd/server/main.go

# Build frontend for production
cd frontend
npm run build
```

## 📖 Documentation

- [Implementation Plan](IMPLEMENTATION_PLAN.md) - Detailed project roadmap
- [Changelog](CHANGELOG.md) - Version history and changes
- [API Reference](docs/api-reference.md) - Complete API documentation (coming soon)
- [Architecture](docs/architecture.md) - System design details (coming soon)
- [Deployment Guide](docs/deployment.md) - Production deployment (coming soon)

## 🗺️ Roadmap

### Completed: Phase 6 - Storage Backends, Deployments, and SCM Enhancements

- ✅ Azure Blob Storage backend with SAS tokens and CDN support
- ✅ AWS S3 / S3-compatible storage backend with multiple auth methods
- ✅ Google Cloud Storage backend with signed URLs
- ✅ Storage configuration UI for runtime backend switching
- ✅ Docker Compose development and production setups
- ✅ Kubernetes + Kustomize with base and overlays
- ✅ Helm Chart for production deployments
- ✅ Azure Container Apps with Bicep templates
- ✅ AWS ECS Fargate with CloudFormation
- ✅ Google Cloud Run deployment support
- ✅ Standalone binary with systemd and Nginx
- ✅ Terraform IaC for AWS, Azure, and GCP
- ✅ Bitbucket Data Center SCM integration (4th provider)
- ✅ API key rotation and lifecycle management
- ✅ Enhanced storage configuration variables in all Terraform modules

### Current: Phase 7 - Documentation & Testing (Sessions 26-28)

- Unit and integration tests for backend (target 80%+ coverage)
- E2E tests for critical user flows
- Comprehensive API documentation
- Deployment guides for all platforms
- Troubleshooting and FAQ documentation

### Next: Phase 8 - Production Polish (Sessions 29-32)

- OpenTelemetry instrumentation and Prometheus metrics
- Grafana dashboards
- Email alerts for API key expiration
- Performance optimization and benchmarks
- Security hardening review
- Open source license attribution audit

## 🤝 Contributing

Contributions are welcome! Please read our contributing guidelines before submitting PRs.

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [HashiCorp Terraform](https://www.terraform.io/) - Module and Provider protocols
- [Gin Web Framework](https://gin-gonic.com/) - Go HTTP framework
- [React](https://react.dev/) - Frontend framework
- [Material-UI](https://mui.com/) - Component library

## 📊 Project Status

**Current Version:** 1.0.0 (Session 25 Complete - Phase 6 Complete)

**Implementation Progress:**

- ✅ Phase 1: Project Foundation (100%)
- ✅ Phase 2: Module Registry Protocol (100%)
- ✅ Phase 3: Provider Registry & Network Mirror (100%)
- ✅ Phase 4: Authentication & Authorization (100%)
- ✅ Phase 5: Frontend SPA (100%)
- ✅ Phase 5A: SCM Integration (100%) - GitHub, Azure DevOps, GitLab, Bitbucket Data Center
- ✅ Phase 5C: Provider Network Mirroring (100%)
- ✅ Phase 6: Storage Backends & Deployment (100%) - All 4 storage backends, 7 deployment options
- 📋 Phase 7: Documentation & Testing (In Progress)
- 📋 Phase 8: Production Polish (Planned)
- 📋 Phase 5B: Azure DevOps Extension (Deferred)

---

Built with ❤️ for the Terraform community
