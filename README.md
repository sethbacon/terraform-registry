# Terraform Registry

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![React](https://img.shields.io/badge/React-18+-61DAFB?logo=react)](https://reactjs.org/)

An enterprise-grade, self-hosted Terraform registry implementing all HashiCorp protocols for modules, providers, and network mirroring.

## Features

### Core Protocols
- **Module Registry Protocol** - Publish and discover Terraform modules
- **Provider Registry Protocol** - Distribute custom Terraform providers
- **Provider Network Mirror Protocol** - Mirror providers for air-gapped environments

### Backend
- **Go Backend** with Gin framework for high performance
- **PostgreSQL** database for robust metadata storage
- **Multi-storage backends**:
  - Azure Blob Storage
  - S3-compatible storage (AWS S3, MinIO, etc.)
  - Local filesystem
- **GPG signature verification** for provider security

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
- Backend API: http://localhost:8080
- Frontend UI: http://localhost:5173

### Docker Compose Deployment

```bash
cd deployments
docker-compose up -d
```

Access the registry at http://localhost:8080

## Usage with Terraform

### Modules

Configure Terraform to use your registry:

```hcl
module "example" {
  source  = "registry.example.com/namespace/module-name/aws"
  version = "1.0.0"
}
```

### Providers

Configure a provider from your registry:

```hcl
terraform {
  required_providers {
    custom = {
      source  = "registry.example.com/namespace/custom"
      version = "1.0.0"
    }
  }
}
```

### Network Mirror

Configure Terraform CLI to use your registry as a provider mirror:

```hcl
provider_installation {
  network_mirror {
    url = "https://registry.example.com/v1/providers/"
  }
}
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
```
GET /.well-known/terraform.json
```

### Module Registry
```
GET  /v1/modules/:namespace/:name/:system/versions
GET  /v1/modules/:namespace/:name/:system/:version/download
POST /api/v1/modules (upload)
```

### Provider Registry
```
GET  /v1/providers/:namespace/:type/versions
GET  /v1/providers/:namespace/:type/:version/download/:os/:arch
POST /api/v1/providers (upload)
```

### Network Mirror
```
GET /v1/providers/:hostname/:namespace/:type/index.json
GET /v1/providers/:hostname/:namespace/:type/:version.json
```

See [API Reference](docs/api-reference.md) for complete API documentation.

## Development

### Project Structure

```
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

Please report security vulnerabilities to security@example.com.

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

- [ ] Phase 1: Core backend and Module Registry (In Progress)
- [ ] Phase 2: Provider Registry and Network Mirror
- [ ] Phase 3: Authentication and Authorization
- [ ] Phase 4: Frontend Web UI
- [ ] Phase 5: Azure DevOps Extension
- [ ] Phase 6: Kubernetes Deployment
- [ ] Phase 7: Production Hardening
- [ ] Phase 8: Performance Optimization

## Status

This project is under active development. Current phase: **Phase 1 - Project Foundation & Backend Core**

---

Built with ❤️ for the Terraform community
