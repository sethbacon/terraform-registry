# Terraform Registry API Quick Reference

## Authentication

### Login Flow

1. **Initiate OAuth Login**
```http
GET /api/v1/auth/login?provider=oidc
GET /api/v1/auth/login?provider=azuread
```

Redirects to OAuth provider (OIDC or Azure AD).

2. **OAuth Callback** (handled automatically)
```http
GET /api/v1/auth/callback?code=xxx&state=yyy
```

Returns JWT token:
```json
{
  "token": "eyJhbGciOiJIUzI1...",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe"
  },
  "expires_in": 86400
}
```

3. **Use JWT Token**
```http
Authorization: Bearer eyJhbGciOiJIUzI1...
```

### Token Management

**Refresh JWT Token**
```http
POST /api/v1/auth/refresh
Authorization: Bearer <current-token>

Response:
{
  "token": "new-jwt-token",
  "expires_in": 86400
}
```

**Get Current User**
```http
GET /api/v1/auth/me
Authorization: Bearer <token>

Response:
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe",
    "created_at": "2026-02-02T10:00:00Z"
  }
}
```

## API Key Management

Requires `api_keys:manage` scope.

**Create API Key**
```http
POST /api/v1/apikeys
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "CI/CD Pipeline Key",
  "organization_id": "org-uuid",
  "scopes": ["modules:read", "modules:write"],
  "expires_at": "2027-02-02T00:00:00Z"  // optional
}

Response:
{
  "id": "key-uuid",
  "name": "CI/CD Pipeline Key",
  "key": "tfr_abc123xyz...",  // Only shown once!
  "key_prefix": "tfr_abc123",
  "scopes": ["modules:read", "modules:write"],
  "expires_at": "2027-02-02T00:00:00Z",
  "created_at": "2026-02-02T10:00:00Z"
}
```

**List API Keys**
```http
GET /api/v1/apikeys?organization_id=org-uuid
Authorization: Bearer <token>

Response:
{
  "keys": [
    {
      "id": "key-uuid",
      "name": "CI/CD Pipeline Key",
      "key_prefix": "tfr_abc123",
      "scopes": ["modules:read", "modules:write"],
      "last_used_at": "2026-02-02T12:00:00Z",
      "created_at": "2026-02-02T10:00:00Z"
    }
  ]
}
```

**Get API Key**
```http
GET /api/v1/apikeys/:id
Authorization: Bearer <token>
```

**Update API Key**
```http
PUT /api/v1/apikeys/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Updated Name",
  "scopes": ["modules:read", "modules:write", "providers:read"],
  "expires_at": "2028-02-02T00:00:00Z"
}
```

**Delete API Key**
```http
DELETE /api/v1/apikeys/:id
Authorization: Bearer <token>

Response:
{
  "message": "API key deleted successfully"
}
```

## User Management

### List Users (requires `users:read`)
```http
GET /api/v1/users?page=1&per_page=20
Authorization: Bearer <token>

Response:
{
  "users": [...],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 150
  }
}
```

### Search Users (requires `users:read`)
```http
GET /api/v1/users/search?q=john&page=1&per_page=20
Authorization: Bearer <token>
```

### Get User (requires `users:read`)
```http
GET /api/v1/users/:id
Authorization: Bearer <token>

Response:
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe",
    "created_at": "2026-01-01T00:00:00Z"
  },
  "organizations": [...]
}
```

### Create User (requires `users:write`)
```http
POST /api/v1/users
Authorization: Bearer <token>
Content-Type: application/json

{
  "email": "newuser@example.com",
  "name": "New User",
  "oidc_sub": "oauth2-sub-identifier"  // optional
}

Response:
{
  "user": {
    "id": "uuid",
    "email": "newuser@example.com",
    "name": "New User",
    "created_at": "2026-02-02T10:00:00Z"
  }
}
```

### Update User (requires `users:write`)
```http
PUT /api/v1/users/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Updated Name",
  "email": "newemail@example.com"
}
```

### Delete User (requires `users:write`)
```http
DELETE /api/v1/users/:id
Authorization: Bearer <token>

Response:
{
  "message": "User deleted successfully"
}
```

## Organization Management

### List Organizations
```http
GET /api/v1/organizations?page=1&per_page=20
Authorization: Bearer <token>

Response:
{
  "organizations": [...],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 50
  }
}
```

### Search Organizations
```http
GET /api/v1/organizations/search?q=acme&page=1&per_page=20
Authorization: Bearer <token>
```

### Get Organization
```http
GET /api/v1/organizations/:id
Authorization: Bearer <token>

Response:
{
  "organization": {
    "id": "uuid",
    "name": "acme",
    "display_name": "ACME Corporation",
    "created_at": "2026-01-01T00:00:00Z"
  },
  "members": [
    {
      "organization_id": "uuid",
      "user_id": "user-uuid",
      "role": "owner",
      "created_at": "2026-01-01T00:00:00Z"
    }
  ]
}
```

### Create Organization (requires `admin` scope)
```http
POST /api/v1/organizations
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "acme",
  "display_name": "ACME Corporation"
}

Response:
{
  "organization": {
    "id": "uuid",
    "name": "acme",
    "display_name": "ACME Corporation",
    "created_at": "2026-02-02T10:00:00Z"
  }
}
```

### Update Organization (requires `admin` scope)
```http
PUT /api/v1/organizations/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "display_name": "ACME Inc."
}
```

### Delete Organization (requires `admin` scope)
```http
DELETE /api/v1/organizations/:id
Authorization: Bearer <token>

Response:
{
  "message": "Organization deleted successfully"
}
```

### Add Member (requires `admin` scope)
```http
POST /api/v1/organizations/:id/members
Authorization: Bearer <token>
Content-Type: application/json

{
  "user_id": "user-uuid",
  "role": "member"  // owner, admin, or member
}

Response:
{
  "member": {
    "organization_id": "org-uuid",
    "user_id": "user-uuid",
    "role": "member",
    "created_at": "2026-02-02T10:00:00Z"
  }
}
```

### Update Member Role (requires `admin` scope)
```http
PUT /api/v1/organizations/:id/members/:user_id
Authorization: Bearer <token>
Content-Type: application/json

{
  "role": "admin"
}

Response:
{
  "member": {
    "organization_id": "org-uuid",
    "user_id": "user-uuid",
    "role": "admin",
    "created_at": "2026-01-01T00:00:00Z"
  }
}
```

### Remove Member (requires `admin` scope)
```http
DELETE /api/v1/organizations/:id/members/:user_id
Authorization: Bearer <token>

Response:
{
  "message": "Member removed successfully"
}
```

## Scopes Reference

| Scope | Description |
|-------|-------------|
| `modules:read` | Read access to modules |
| `modules:write` | Write access to modules (includes read) |
| `providers:read` | Read access to providers |
| `providers:write` | Write access to providers (includes read) |
| `users:read` | Read access to user management |
| `users:write` | Write access to user management (includes read) |
| `api_keys:manage` | Manage API keys |
| `audit:read` | Read audit logs |
| `admin` | Full access to all resources (wildcard) |

## Error Responses

**Unauthorized (401)**
```json
{
  "error": "Invalid credentials"
}
```

**Forbidden (403)**
```json
{
  "error": "Missing required scope",
  "details": "Required scope: modules:write"
}
```

**Not Found (404)**
```json
{
  "error": "Resource not found"
}
```

**Bad Request (400)**
```json
{
  "error": "Invalid request: email is required"
}
```

**Conflict (409)**
```json
{
  "error": "User with this email already exists"
}
```

## Example: Complete Authentication Flow

```bash
# 1. Get login URL (open in browser)
curl http://localhost:8080/api/v1/auth/login?provider=oidc

# 2. After OAuth callback, you'll get a JWT token
# Store this token for subsequent requests

# 3. Create an API key for programmatic access
curl -X POST http://localhost:8080/api/v1/apikeys \
  -H "Authorization: Bearer <jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My API Key",
    "organization_id": "org-uuid",
    "scopes": ["modules:read", "modules:write"]
  }'

# Response includes the full key (save it!)
# {
#   "key": "tfr_abc123xyz..."
# }

# 4. Use the API key for subsequent requests
curl http://localhost:8080/api/v1/modules/search \
  -H "Authorization: Bearer tfr_abc123xyz..."
```

## Testing with curl

**Upload a module (with API key)**
```bash
curl -X POST http://localhost:8080/api/v1/modules \
  -H "Authorization: Bearer tfr_abc123xyz..." \
  -F "file=@terraform-module.tar.gz" \
  -F "namespace=acme" \
  -F "name=vpc" \
  -F "system=aws" \
  -F "version=1.0.0"
```

**List modules**
```bash
curl http://localhost:8080/v1/modules/acme/vpc/aws/versions
```

**Download module**
```bash
curl http://localhost:8080/v1/modules/acme/vpc/aws/1.0.0/download
```
