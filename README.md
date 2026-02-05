# Terraform Registry

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![React](https://img.shields.io/badge/React-18+-61DAFB?logo=react)](https://reactjs.org/)

An enterprise-grade, self-hosted Terraform registry implementing all HashiCorp protocols for modules, providers, and network mirroring.

## Features

### Core Protocols

- **✅ Module Registry Protocol** - Publish and discover Terraform modules (Phase 2 - Complete)
- **✅ Provider Registry Protocol** - Distribute custom Terraform providers (Phase 3 - Complete)
- **✅ Provider Network Mirror Protocol** - Mirror providers for air-gapped environments (Phase 3 - Complete)

### Backend

- **✅ Go Backend** with Gin framework for high performance
- **✅ PostgreSQL** database for robust metadata storage
- **Storage backends**:
  - **✅ Local filesystem** (Phase 2 - Complete)
  - Azure Blob Storage (Planned)
  - S3-compatible storage (AWS S3, MinIO, etc.) (Planned)
- **✅ GPG signature verification** framework for provider security (Phase 3 - Complete)

### Authentication & Authorization

- **API Key authentication** for CLI and automation
- **Azure AD / Entra ID** integration
- **Generic OIDC** provider support (Okta, Auth0, Google, etc.)
- **Role-based access control (RBAC)**
- **Multi-tenancy** support (configurable)

### Frontend

- **React + TypeScript SPA** with Material-UI
- Module and provider browsing with search
- Web-based upload interface
- User and permission management
- Usage analytics dashboard
- Dark mode support

### DevOps Integration

- **Azure DevOps extension** with OIDC authentication
- Custom pipeline task for publishing
- Service connection integration

### Deployment

- Docker Compose for single-server deployments
- Kubernetes with Helm charts
- Azure Container Apps
- Standalone binary

### Enterprise Features

- OpenTelemetry instrumentation
- Prometheus metrics
- Structured logging
- Rate limiting
- Audit logging
- Health checks

## Quick Start

### Prerequisites

- Go 1.22 or later
- PostgreSQL 14 or later
- Node.js 18 or later (for frontend development)
- Docker and Docker Compose (for containerized deployment)

### Development Setup

1. **Clone the repository**

   ```bash
   git clone https://github.com/your-org/terraform-registry.git
   cd terraform-registry
   ```

2. **Set up PostgreSQL**

   ```bash
   # Using Docker
   docker run -d \
     --name postgres \
     -e POSTGRES_DB=terraform_registry \
     -e POSTGRES_USER=registry \
     -e POSTGRES_PASSWORD=registry \
     -p 5432:5432 \
     postgres:16
   ```

3. **Configure the application**

   ```bash
   cp config.example.yaml config.yaml
   # Edit config.yaml with your settings
   ```

4. **Run database migrations**

   ```bash
   cd backend
   go run cmd/server/main.go migrate up
   ```

5. **Start the backend server**

   ```bash
   cd backend
   go run cmd/server/main.go serve
   ```

6. **Start the frontend (development)**

   ```bash
   cd frontend
   npm install
   npm run dev
   ```

The application will be available at:

- Backend API: <http://localhost:8080>
- Frontend UI: <http://localhost:5173>

### Docker Compose Deployment

```bash
cd deployments
docker-compose up -d
```

Access the registry at <http://localhost:8080>

## Usage with Terraform

### Publishing Modules (✅ Available Now)

Upload a module to the registry:

```bash
# Package your module
tar -czf module.tar.gz -C /path/to/module .

# Upload to registry
curl -X POST http://localhost:8080/api/v1/modules \
  -F "namespace=myorg" \
  -F "name=vpc" \
  -F "system=aws" \
  -F "version=1.0.0" \
  -F "description=My VPC module" \
  -F "file=@module.tar.gz"
```

### Using Modules (✅ Available Now)

Configure Terraform to use modules from your registry:

```hcl
module "vpc" {
  source  = "localhost:8080/myorg/vpc/aws"
  version = "1.0.0"

  vpc_name = "production"
  vpc_cidr = "10.0.0.0/16"
}
```

### Searching Modules (✅ Available Now)

```bash
# Search for modules
curl "http://localhost:8080/api/v1/modules/search?q=vpc"

# List versions
curl "http://localhost:8080/v1/modules/myorg/vpc/aws/versions"
```

### Providers (✅ Available Now)

Upload a provider to the registry:

```bash
# Upload provider binary for a specific platform
curl -X POST http://localhost:8080/api/v1/providers \
  -F "namespace=myorg" \
  -F "type=custom" \
  -F "version=1.0.0" \
  -F "os=linux" \
  -F "arch=amd64" \
  -F "protocols=[\"5.0\",\"6.0\"]" \
  -F "file=@terraform-provider-custom_1.0.0_linux_amd64.zip"
```

Configure Terraform to use a provider from your registry:

```hcl
terraform {
  required_providers {
    custom = {
      source  = "localhost:8080/myorg/custom"
      version = "1.0.0"
    }
  }
}

provider "custom" {
  # Provider configuration
}
```

List available versions:

```bash
curl "http://localhost:8080/v1/providers/myorg/custom/versions"
```

### Network Mirror Protocol (✅ Available Now)

Configure Terraform CLI to use your registry as a provider mirror for air-gapped environments:

```hcl
# In ~/.terraformrc or terraform.rc
provider_installation {
  network_mirror {
    url = "http://localhost:8080/terraform/providers/"
  }
}
```

Test the Network Mirror endpoints:

```bash
# Get version index
curl "http://localhost:8080/terraform/providers/registry.terraform.io/hashicorp/azurerm/index.json"

# Get platform index for a specific version
curl "http://localhost:8080/terraform/providers/registry.terraform.io/hashicorp/azurerm/3.85.0.json"
```

## Documentation

Comprehensive documentation is available in the [docs](docs/) directory:

- [Architecture Overview](docs/architecture.md)
- [API Reference](docs/api-reference.md)
- [Deployment Guide](docs/deployment.md)
- [Configuration Reference](docs/configuration.md)
- [Development Guide](docs/development.md)
- [Azure DevOps Integration](docs/azure-devops-integration.md)
- [Terraform Usage Guide](docs/terraform-usage.md)
- [Troubleshooting](docs/troubleshooting.md)

## Configuration

The application is configured via `config.yaml` or environment variables. Key configuration areas:

- **Server settings** - Host, port, base URL, timeouts
- **Database** - PostgreSQL connection settings
- **Storage backends** - Azure Blob, S3, or local filesystem
- **Authentication** - API keys, OIDC, Azure AD
- **Multi-tenancy** - Enable/disable organization isolation
- **Security** - CORS, rate limiting, TLS
- **Telemetry** - Metrics, tracing, logging

See [Configuration Reference](docs/configuration.md) for complete details.

## API Endpoints

### Service Discovery

```cmd
GET /.well-known/terraform.json
```

### Module Registry

```cmd
GET  /v1/modules/:namespace/:name/:system/versions
GET  /v1/modules/:namespace/:name/:system/:version/download
POST /api/v1/modules (upload)
```

### Provider Registry

```cmd
GET  /v1/providers/:namespace/:type/versions
GET  /v1/providers/:namespace/:type/:version/download/:os/:arch
POST /api/v1/providers (upload)
```

### Network Mirror

```cmd
GET /v1/providers/:hostname/:namespace/:type/index.json
GET /v1/providers/:hostname/:namespace/:type/:version.json
```

See [API Reference](docs/api-reference.md) for complete API documentation.

## Development

### Project Structure

```cmd
terraform-registry/
├── backend/          # Go backend application
├── frontend/         # React TypeScript SPA
├── azure-devops-extension/  # Azure DevOps extension
├── deployments/      # Deployment configurations
├── docs/             # Documentation
└── scripts/          # Build and utility scripts
```

### Running Tests

**Backend:**

```bash
cd backend
go test ./...
```

**Frontend:**

```bash
cd frontend
npm test
```

### Building

**Backend:**

```bash
cd backend
go build -o terraform-registry cmd/server/main.go
```

**Frontend:**

```bash
cd frontend
npm run build
```

**Docker Image:**

```bash
docker build -t terraform-registry:latest .
```

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## Security

Security is a top priority. We implement:

- GPG signature verification for all provider binaries
- API key hashing with bcrypt
- HTTPS/TLS enforcement in production
- Rate limiting to prevent abuse
- Input validation and SQL injection prevention
- Audit logging for administrative actions
- CORS policy enforcement

Please report security vulnerabilities to <security@example.com>.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [HashiCorp](https://www.hashicorp.com/) for Terraform and protocol specifications
- [Gin Web Framework](https://gin-gonic.com/) for the Go web framework
- [React](https://reactjs.org/) and [Material-UI](https://mui.com/) for the frontend
- All contributors who help improve this project

## Support

- Documentation: [docs/](docs/)
- Issues: [GitHub Issues](https://github.com/your-org/terraform-registry/issues)
- Discussions: [GitHub Discussions](https://github.com/your-org/terraform-registry/discussions)

## Roadmap

- [x] Phase 1: Project Foundation & Backend Core (Complete)
- [x] Phase 2: Module Registry Protocol (Complete)
- [x] Phase 3: Provider Registry and Network Mirror (Complete)
- [ ] Phase 4: Authentication and Authorization
- [ ] Phase 5: Frontend Web UI
- [ ] Phase 6: Azure DevOps Extension
- [ ] Phase 7: Kubernetes Deployment
- [ ] Phase 8: Production Hardening
- [ ] Phase 9: Performance Optimization

## Status

This project is under active development. Current phase: **Phase 4 - Authentication and Authorization**

**Completed Phases:**

- ✅ Phase 1: Core backend infrastructure, database, configuration, Docker deployment
- ✅ Phase 2: Complete Module Registry Protocol with storage abstraction layer
- ✅ Phase 3: Provider Registry Protocol & Network Mirror Protocol with multi-platform support

---

Built with ❤️ for the Terraform community
