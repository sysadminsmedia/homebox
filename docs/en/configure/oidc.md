# OpenID Connect (OIDC) Authentication

Homebox supports OpenID Connect (OIDC) authentication, allowing you to integrate with external identity providers like Keycloak, Auth0, PocketID, and others.

## Overview

When OIDC is enabled, users can authenticate using their existing accounts from your identity provider instead of creating separate Homebox accounts. The system automatically creates user accounts on first login and maps claims from the identity provider to user attributes.

## Configuration

### Environment Variables

To enable OIDC authentication, configure the following environment variables:

```bash
# Enable OIDC
HBOX_OIDC_ENABLED=true

# Identity Provider Configuration
HBOX_OIDC_ISSUER=https://your-identity-provider.com
HBOX_OIDC_CLIENT_ID=your_client_id
HBOX_OIDC_CLIENT_SECRET=your_client_secret
HBOX_OIDC_REDIRECT_URL=https://your-homebox-domain.com/api/v1/auth/oidc/callback

# Optional: Customize claim mappings
HBOX_OIDC_SCOPES=openid profile email
HBOX_OIDC_USERNAME_CLAIM=preferred_username
HBOX_OIDC_EMAIL_CLAIM=email
HBOX_OIDC_NAME_CLAIM=name
```

### Docker Compose Example

```yaml
services:
  homebox:
    image: ghcr.io/sysadminsmedia/homebox:latest
    environment:
      # Basic configuration
      - HBOX_LOG_LEVEL=info
      - HBOX_OPTIONS_ALLOW_REGISTRATION=true
      
      # OIDC Configuration
      - HBOX_OIDC_ENABLED=true
      - HBOX_OIDC_ISSUER=https://auth.example.com
      - HBOX_OIDC_CLIENT_ID=homebox-client
      - HBOX_OIDC_CLIENT_SECRET=your-secret-here
      - HBOX_OIDC_REDIRECT_URL=https://homebox.example.com/api/v1/auth/oidc/callback
    ports:
      - "3100:7745"
    volumes:
      - homebox-data:/data
```

## Provider Setup

### General Requirements

Your OIDC provider must support:
- Authorization Code flow
- The `openid`, `profile`, and `email` scopes
- Standard OIDC discovery (`.well-known/openid_configuration`)

### Redirect URI

Configure your OIDC client with the following redirect URI:
```
https://your-homebox-domain.com/api/v1/auth/oidc/callback
```

### Common Providers

#### Keycloak

1. Create a new client in your Keycloak realm
2. Set **Client Type** to `OpenID Connect`
3. Set **Valid Redirect URIs** to your Homebox callback URL
4. Enable **Standard Flow** (Authorization Code)
5. Configure client credentials if using confidential client

Configuration:
```bash
HBOX_OIDC_ISSUER=https://keycloak.example.com/realms/your-realm
HBOX_OIDC_CLIENT_ID=homebox
HBOX_OIDC_CLIENT_SECRET=your-client-secret
```

#### Auth0

1. Create a new application in Auth0 dashboard
2. Choose **Regular Web Application**
3. Set **Allowed Callback URLs** to your Homebox callback URL
4. Note the Domain, Client ID, and Client Secret

Configuration:
```bash
HBOX_OIDC_ISSUER=https://your-tenant.auth0.com
HBOX_OIDC_CLIENT_ID=your-auth0-client-id
HBOX_OIDC_CLIENT_SECRET=your-auth0-client-secret
```

#### PocketID

1. Access your PocketID admin interface
2. Create a new OAuth2/OIDC client
3. Set the redirect URI to your Homebox callback URL
4. Configure required scopes: `openid`, `profile`, `email`

Configuration:
```bash
HBOX_OIDC_ISSUER=https://your-pocketid-instance.com
HBOX_OIDC_CLIENT_ID=your-pocketid-client-id
HBOX_OIDC_CLIENT_SECRET=your-pocketid-client-secret
```

## Claim Mapping

Homebox maps OIDC claims to user attributes as follows:

| Homebox Field | Default Claim | Environment Variable | Description |
|---------------|---------------|---------------------|-------------|
| Username | `preferred_username` | `HBOX_OIDC_USERNAME_CLAIM` | Used as the primary username |
| Email | `email` | `HBOX_OIDC_EMAIL_CLAIM` | User's email address |
| Display Name | `name` | `HBOX_OIDC_NAME_CLAIM` | Full name for display |

If the username claim is not available, the email will be used as the username.

## User Management

### Automatic User Creation

When a user successfully authenticates via OIDC for the first time:
1. Homebox creates a new user account automatically
2. User information is populated from OIDC claims
3. The user is added to the default group
4. Future logins update user information from current claims

### Local vs OIDC Users

- OIDC users cannot change their password in Homebox (managed by identity provider)
- OIDC users can still be managed by administrators (roles, groups, etc.)
- Local user registration can be disabled when using OIDC exclusively

## Authentication Flow

1. User clicks "Login with OIDC" on Homebox login page
2. User is redirected to the identity provider
3. User authenticates with the identity provider
4. Identity provider redirects back to Homebox with authorization code
5. Homebox exchanges code for tokens and verifies ID token
6. User is automatically redirected to Homebox home page
7. Session is established with authentication cookies

## API Endpoints

When OIDC is enabled, the following endpoints are available:

### Get OIDC Configuration
```http
GET /api/v1/auth/oidc/config
```

Returns OIDC configuration status:
```json
{
  "enabled": true,
  "authUrl": "/api/v1/auth/oidc/login"
}
```

### Initiate OIDC Login
```http
GET /api/v1/auth/oidc/login
```

Redirects to the identity provider for authentication.

### OIDC Callback
```http
GET /api/v1/auth/oidc/callback?code=...&state=...
```

Handles the OAuth2 callback and redirects to the home page on success.

## Troubleshooting

### Common Issues

1. **"Invalid redirect URI"**
   - Ensure the redirect URI in your identity provider exactly matches your Homebox callback URL
   - Check for trailing slashes or protocol mismatches

2. **"Failed to verify ID token"**
   - Verify the issuer URL is correct and accessible
   - Check that the client ID matches your identity provider configuration

3. **"No username or email found in claims"**
   - Verify your identity provider includes the required claims in ID tokens
   - Check claim mapping configuration
   - Ensure scopes include `profile` and `email`

4. **Users redirected to JSON instead of home page**
   - This was a bug in older versions - ensure you're using the latest version
   - The callback should automatically redirect to `/home`

### Debug Mode

Enable debug logging to troubleshoot OIDC issues:
```bash
HBOX_LOG_LEVEL=debug
```

This will provide detailed logs of the OIDC authentication flow.

## Security Considerations

- Always use HTTPS in production for both Homebox and your identity provider
- Keep client secrets secure and rotate them regularly
- Validate that your identity provider's SSL certificates are properly configured
- Consider using PKCE (Proof Key for Code Exchange) if supported by your provider
- Regularly review and audit user access through your identity provider

## Migration from Local Authentication

To migrate from local to OIDC authentication:

1. Configure OIDC as described above
2. Ensure existing users' email addresses match their identity provider accounts
3. Users can continue using local authentication alongside OIDC
4. Optionally disable registration: `HBOX_OPTIONS_ALLOW_REGISTRATION=false`
5. Gradually migrate users to OIDC authentication