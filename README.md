# Enterprise Terraform Registry

A fully-featured, enterprise-grade Terraform registry implementing all three HashiCorp protocols with modern web UI and multi-tenancy support.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
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

#### Storage Backends

- **Local Filesystem** - Direct file serving for development and simple deployments
- **Pluggable Architecture** - Extensible storage interface for cloud providers
- ⏳ Azure Blob Storage (planned - Phase 6)
- ⏳ S3-Compatible Storage (planned - Phase 6)

#### Modern Web Interface

- **React 18+ SPA** - Fast, responsive single-page application
- **TypeScript** - Full type safety across the frontend
- **Material-UI** - Professional, accessible component library
- **Module Browser** - Search, filter, and explore modules with pagination
- **Provider Browser** - Discover and manage provider versions
- **Admin Dashboard** - System statistics and management tools
- **Upload Interface** - Easy module and provider publishing
- **Responsive Design** - Works on desktop, tablet, and mobile

#### Security

- **GPG Verification** - Provider binary signature validation
- **SHA256 Checksums** - File integrity verification
- **Path Traversal Protection** - Secure archive extraction
- **Input Validation** - Semantic versioning and format validation
- **Audit Logging** - Comprehensive activity tracking
- **Rate Limiting** - API protection (coming soon)

### ⏳ In Progress

#### SCM Integration (Phase 5A - Sessions 10-13)

- **Git Provider Support** - GitHub, Azure DevOps, GitLab integration
- **OAuth 2.0 Flows** - Secure repository access
- **Automated Publishing** - Tag-triggered module releases
- **Commit-Pinned Versions** - Immutable version security
- **Webhook Handlers** - Real-time update processing
- **Tag Movement Detection** - Supply chain attack prevention

### 📋 Planned

#### Azure DevOps Extension (Phase 5B)

- Custom pipeline task for publishing
- OIDC authentication with workload identity
- Visual Studio Marketplace distribution

#### Additional Storage & Deployment (Phase 6)

- Azure Blob Storage backend
- S3-compatible storage backend
- Kubernetes Helm charts
- Azure Container Apps configuration

#### Documentation & Testing (Phase 7)

- Comprehensive API documentation
- 80%+ test coverage
- End-to-end test suite
- Performance benchmarks

#### Production Polish (Phase 8)

- OpenTelemetry instrumentation
- Prometheus metrics
- Grafana dashboards
- Performance optimization

## 🏗️ Architecture

```txt
┌─────────────────────────────────────────────────────────────┐
│                     React TypeScript SPA                     │
│  Module Browser │ Provider Browser │ Admin Dashboard │ Auth │
└────────────────────────────┬────────────────────────────────┘
                             │
                             │ REST API
                             │
┌────────────────────────────▼────────────────────────────────┐
│                      Go Backend (Gin)                        │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │ Modules  │  │Providers │  │  Mirror  │  │  Admin   │   │
│  │   API    │  │   API    │  │   API    │  │   API    │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
│  ┌──────────────────────────────────────────────────────┐  │
│  │         Authentication & Authorization               │  │
│  │    API Keys │ OIDC │ Azure AD │ RBAC Middleware     │  │
│  └──────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Storage Abstraction Layer               │  │
│  │    Local FS │ Azure Blob (planned) │ S3 (planned)   │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────────────┬────────────────────────────────┘
                             │
                             │
                    ┌────────▼────────┐
                    │   PostgreSQL    │
                    │   Database      │
                    └─────────────────┘
```

## 📦 Installation

### Prerequisites

- Go 1.21 or later
- Node.js 18+ and npm
- PostgreSQL 14+
- Docker & Docker Compose (for containerized deployment)

### Quick Start with Docker Compose

```bash
# Clone the repository
git clone https://github.com/yourusername/terraform-registry.git
cd terraform-registry

# Start all services
docker-compose up -d

# Backend: http://localhost:8080
# Frontend: http://localhost:3000
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
go run cmd/server/main.go migrate

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

```bash
# Database
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=terraform_registry
export DB_PASSWORD=your_password
export DB_NAME=terraform_registry

# Server
export SERVER_PORT=8080
export SERVER_HOST=0.0.0.0

# Storage
export STORAGE_BACKEND=local
export STORAGE_LOCAL_PATH=/var/lib/terraform-registry

# Authentication
export JWT_SECRET=your_jwt_secret_here
export OIDC_ISSUER_URL=https://your-idp.com
export OIDC_CLIENT_ID=your_client_id
export OIDC_CLIENT_SECRET=your_client_secret

# Azure AD (optional)
export AZURE_AD_TENANT_ID=your_tenant_id
export AZURE_AD_CLIENT_ID=your_client_id
export AZURE_AD_CLIENT_SECRET=your_client_secret
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

### Current: Phase 5A - SCM Integration (Sessions 10-13)

- OAuth integration with GitHub, Azure DevOps, GitLab
- Repository browsing and selection
- Automated tag-based publishing
- Commit-pinned immutable versions

### Next: Phase 5B - Azure DevOps Extension (Sessions 14-16)

- Custom pipeline task
- OIDC authentication
- Marketplace publication

### Future Phases

- Additional storage backends (Azure Blob, S3)
- Comprehensive documentation and testing
- Production hardening and optimization

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

**Current Version:** 0.5.0 (Session 9 Complete)

**Implementation Progress:**

- ✅ Phase 1: Project Foundation (100%)
- ✅ Phase 2: Module Registry Protocol (100%)
- ✅ Phase 3: Provider Registry & Network Mirror (100%)
- ✅ Phase 4: Authentication & Authorization (100%)
- ✅ Phase 5: Frontend SPA (100%)
- ⏳ Phase 5A: SCM Integration (0% - Starting Session 10)
- 📋 Phase 5B: Azure DevOps Extension (Planned)
- 📋 Phase 6: Additional Storage & Deployment (Planned)
- 📋 Phase 7: Documentation & Testing (Planned)
- 📋 Phase 8: Production Polish (Planned)

---

Built with ❤️ for the Terraform community
