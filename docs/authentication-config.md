# Authentication Configuration Guide

This guide explains how to configure authentication for the Terraform Registry.

## Configuration File Format

The registry uses YAML configuration files. Here's the complete authentication section:

```yaml
# config.yaml

server:
  base_url: "http://localhost:8080"  # Used for OIDC redirect URLs

auth:
  # Generic OIDC Provider Configuration
  oidc:
    enabled: true
    issuer_url: "https://your-oidc-provider.com"
    client_id: "your-client-id"
    client_secret: "your-client-secret"
    redirect_url: "http://localhost:8080/api/v1/auth/callback"
    scopes:
      - openid
      - email
      - profile
  
  # Azure AD / Entra ID Configuration
  azuread:
    enabled: true
    tenant_id: "your-tenant-id-or-domain.onmicrosoft.com"
    client_id: "your-azure-app-client-id"
    client_secret: "your-azure-app-client-secret"
    redirect_url: "http://localhost:8080/api/v1/auth/callback"

security:
  cors:
    allowed_origins:
      - "http://localhost:3000"      # React frontend
      - "http://localhost:8080"      # API
      - "https://registry.example.com"
```

## Environment Variables

Set these environment variables for sensitive data:

```bash
# JWT Secret (required in production)
export TFR_JWT_SECRET="your-super-secret-jwt-signing-key-min-32-chars"

# OIDC Configuration (alternative to config file)
export TFR_OIDC_ENABLED=true
export TFR_OIDC_ISSUER_URL="https://your-oidc-provider.com"
export TFR_OIDC_CLIENT_ID="your-client-id"
export TFR_OIDC_CLIENT_SECRET="your-client-secret"
export TFR_OIDC_REDIRECT_URL="http://localhost:8080/api/v1/auth/callback"

# Azure AD Configuration (alternative to config file)
export TFR_AZUREAD_ENABLED=true
export TFR_AZUREAD_TENANT_ID="your-tenant-id"
export TFR_AZUREAD_CLIENT_ID="your-client-id"
export TFR_AZUREAD_CLIENT_SECRET="your-client-secret"
export TFR_AZUREAD_REDIRECT_URL="http://localhost:8080/api/v1/auth/callback"
```

## Provider-Specific Setup

### 1. Generic OIDC Provider (Keycloak, Auth0, Okta, etc.)

**Example: Keycloak**

```yaml
auth:
  oidc:
    enabled: true
    issuer_url: "https://keycloak.example.com/realms/terraform-registry"
    client_id: "terraform-registry-client"
    client_secret: "your-keycloak-client-secret"
    redirect_url: "http://localhost:8080/api/v1/auth/callback"
    scopes:
      - openid
      - email
      - profile
```

**Keycloak Setup:**
1. Create a new realm (e.g., "terraform-registry")
2. Create a new client with "openid-connect" protocol
3. Set Valid Redirect URIs: `http://localhost:8080/api/v1/auth/callback`
4. Enable "Standard Flow"
5. Copy Client ID and Secret

**Example: Auth0**

```yaml
auth:
  oidc:
    enabled: true
    issuer_url: "https://your-tenant.auth0.com"
    client_id: "your-auth0-client-id"
    client_secret: "your-auth0-client-secret"
    redirect_url: "http://localhost:8080/api/v1/auth/callback"
    scopes:
      - openid
      - email
      - profile
```

**Auth0 Setup:**
1. Create a new Application (Regular Web Application)
2. Set Allowed Callback URLs: `http://localhost:8080/api/v1/auth/callback`
3. Set Allowed Logout URLs: `http://localhost:8080`
4. Copy Domain, Client ID, and Client Secret

**Example: Okta**

```yaml
auth:
  oidc:
    enabled: true
    issuer_url: "https://your-org.okta.com/oauth2/default"
    client_id: "your-okta-client-id"
    client_secret: "your-okta-client-secret"
    redirect_url: "http://localhost:8080/api/v1/auth/callback"
    scopes:
      - openid
      - email
      - profile
```

**Okta Setup:**
1. Create a new App Integration (OIDC - Web Application)
2. Set Sign-in redirect URIs: `http://localhost:8080/api/v1/auth/callback`
3. Copy Client ID and Client Secret

### 2. Azure AD / Entra ID

```yaml
auth:
  azuread:
    enabled: true
    tenant_id: "common"  # or specific tenant ID
    client_id: "your-app-registration-client-id"
    client_secret: "your-app-registration-client-secret"
    redirect_url: "http://localhost:8080/api/v1/auth/callback"
```

**Azure AD Setup:**

1. **Register Application in Azure Portal:**
   - Go to Azure Active Directory → App registrations → New registration
   - Name: "Terraform Registry"
   - Supported account types: Choose based on your needs
   - Redirect URI: Web → `http://localhost:8080/api/v1/auth/callback`
   - Click Register

2. **Configure Authentication:**
   - Go to Authentication section
   - Add additional redirect URI if needed (for production)
   - Enable "ID tokens" under Implicit grant
   - Save changes

3. **Create Client Secret:**
   - Go to Certificates & secrets
   - New client secret
   - Copy the secret value immediately (shown only once)

4. **Configure API Permissions:**
   - Go to API permissions
   - Add permission → Microsoft Graph → Delegated permissions
   - Add: `openid`, `email`, `profile`, `User.Read`
   - Grant admin consent (if required)

5. **Copy Configuration:**
   - Application (client) ID → `client_id`
   - Directory (tenant) ID → `tenant_id`
   - Client secret value → `client_secret`

**Multi-Tenant Azure AD:**

For multi-tenant apps (users from any Azure AD):

```yaml
auth:
  azuread:
    enabled: true
    tenant_id: "common"  # Allows any Azure AD tenant
    client_id: "your-app-registration-client-id"
    client_secret: "your-app-registration-client-secret"
    redirect_url: "http://localhost:8080/api/v1/auth/callback"
```

**Single-Tenant Azure AD:**

For single-tenant apps (users from specific Azure AD):

```yaml
auth:
  azuread:
    enabled: true
    tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"  # Your specific tenant ID
    client_id: "your-app-registration-client-id"
    client_secret: "your-app-registration-client-secret"
    redirect_url: "http://localhost:8080/api/v1/auth/callback"
```

## Production Deployment

### 1. Environment Variables for Production

```bash
# Required
export TFR_JWT_SECRET="production-secret-at-least-32-chars-long-random-string"

# For Azure AD
export TFR_AZUREAD_ENABLED=true
export TFR_AZUREAD_TENANT_ID="your-tenant-id"
export TFR_AZUREAD_CLIENT_ID="your-client-id"
export TFR_AZUREAD_CLIENT_SECRET="your-client-secret"
export TFR_AZUREAD_REDIRECT_URL="https://registry.example.com/api/v1/auth/callback"
```

### 2. Update Redirect URLs

Update all redirect URLs to use HTTPS:

```yaml
server:
  base_url: "https://registry.example.com"

auth:
  oidc:
    redirect_url: "https://registry.example.com/api/v1/auth/callback"
  
  azuread:
    redirect_url: "https://registry.example.com/api/v1/auth/callback"

security:
  cors:
    allowed_origins:
      - "https://registry.example.com"
      - "https://app.registry.example.com"
```

### 3. CORS Configuration

For production, restrict CORS to specific origins:

```yaml
security:
  cors:
    allowed_origins:
      - "https://registry.example.com"
      - "https://app.registry.example.com"
    # Remove "*" wildcard in production!
```

### 4. Azure AD Production Settings

In Azure Portal:
- Update Redirect URIs to production URLs
- Add production domains to Redirect URIs
- Consider using certificates instead of secrets for higher security
- Enable logging and monitoring
- Set up Conditional Access policies if needed

## Testing Authentication

### 1. Test OIDC/Azure AD Login

```bash
# Start the server
./server

# Open in browser (replace with your setup)
open http://localhost:8080/api/v1/auth/login?provider=oidc
# or
open http://localhost:8080/api/v1/auth/login?provider=azuread
```

You should be redirected to your OIDC provider, and after successful login, redirected back with a JWT token.

### 2. Test JWT Token

```bash
# Save the token from login response
TOKEN="eyJhbGciOiJIUzI1..."

# Test authenticated endpoint
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

### 3. Create and Test API Key

```bash
# Create API key
curl -X POST http://localhost:8080/api/v1/apikeys \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Key",
    "organization_id": "org-uuid",
    "scopes": ["modules:read", "modules:write"]
  }'

# Save the returned key
API_KEY="tfr_abc123..."

# Test API key
curl http://localhost:8080/api/v1/modules/search \
  -H "Authorization: Bearer $API_KEY"
```

## Troubleshooting

### Common Issues

**"Invalid redirect URI"**
- Check that redirect_url in config matches exactly what's configured in your OIDC provider
- Include protocol (http/https), port, and path
- No trailing slashes

**"Invalid client credentials"**
- Verify client_id and client_secret are correct
- Check if secret has expired (Azure AD secrets expire after 1-2 years)
- Ensure app registration is enabled

**"Invalid issuer"**
- For OIDC: Check issuer_url points to correct discovery document
- For Azure AD: Verify tenant_id is correct
- Test discovery endpoint: `curl https://your-oidc-provider.com/.well-known/openid-configuration`

**"Invalid token signature"**
- Ensure TFR_JWT_SECRET is set consistently across all instances
- Check that JWT is not expired (24-hour default)
- Verify token is being passed correctly in Authorization header

**CORS errors in browser**
- Add frontend URL to `security.cors.allowed_origins`
- Ensure OPTIONS requests are allowed
- Check browser console for specific CORS error messages

### Debug Mode

Enable debug logging:

```yaml
logging:
  level: "debug"
  format: "json"
```

This will log authentication attempts, token validation, and scope checks.

## Security Best Practices

1. **JWT Secret**: Use a strong, random secret (min 32 characters)
2. **HTTPS**: Always use HTTPS in production
3. **Secrets**: Never commit secrets to version control
4. **Rotation**: Rotate client secrets and JWT secrets regularly
5. **Expiration**: Set reasonable expiration times for API keys
6. **Scopes**: Grant minimum required scopes (principle of least privilege)
7. **Monitoring**: Monitor authentication failures and API key usage
8. **Audit**: Enable audit logging for sensitive operations
9. **Multi-factor**: Consider enabling MFA in your OIDC provider
10. **Session Storage**: Use Redis or similar for production (not in-memory)

## Example: Complete config.yaml

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  base_url: "https://registry.example.com"

database:
  host: "postgres"
  port: 5432
  name: "terraform_registry"
  user: "tfregistry"
  password: "${DB_PASSWORD}"
  ssl_mode: "require"

storage:
  default_backend: "azure"
  azure:
    account_name: "${AZURE_STORAGE_ACCOUNT}"
    account_key: "${AZURE_STORAGE_KEY}"
    container_name: "terraform-modules"

auth:
  oidc:
    enabled: false
  
  azuread:
    enabled: true
    tenant_id: "${AZURE_TENANT_ID}"
    client_id: "${AZURE_CLIENT_ID}"
    client_secret: "${AZURE_CLIENT_SECRET}"
    redirect_url: "https://registry.example.com/api/v1/auth/callback"

security:
  cors:
    allowed_origins:
      - "https://registry.example.com"

logging:
  level: "info"
  format: "json"
```
