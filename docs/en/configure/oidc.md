# OIDC Authentication

Homebox supports OpenID Connect (OIDC) authentication, allowing users to log in using external identity providers like Keycloak, Auth0, Google, Microsoft Azure AD, and others.

## Configuration

To enable OIDC authentication, configure the following environment variables or YAML settings:

### Environment Variables

```bash
# Enable OIDC authentication
HBOX_OIDC_ENABLED=true

# OIDC Provider Configuration
HBOX_OIDC_ISSUER_URL=https://your-oidc-provider.com/auth/realms/your-realm
HBOX_OIDC_CLIENT_ID=homebox-client
HBOX_OIDC_CLIENT_SECRET=your-client-secret
HBOX_OIDC_REDIRECT_URL=https://your-homebox-instance.com/api/v1/auth/oidc/callback

# Optional: Customize scopes (default: "openid email profile")
HBOX_OIDC_SCOPES="openid email profile groups"

# Optional: Configure role mapping (defaults shown)
HBOX_OIDC_ROLES_CLAIM=groups
HBOX_OIDC_ADMIN_ROLE=admin
HBOX_OIDC_USER_ROLE=user
```

### YAML Configuration

```yaml
oidc:
  enabled: true
  issuer_url: "https://your-oidc-provider.com/auth/realms/your-realm"
  client_id: "homebox-client"
  client_secret: "your-client-secret"
  redirect_url: "https://your-homebox-instance.com/api/v1/auth/oidc/callback"
  scopes: "openid email profile groups"
  roles_claim: "groups"
  admin_role: "admin"
  user_role: "user"
```

## OIDC Provider Setup

### Keycloak

1. **Create a new client in Keycloak:**
   - Client ID: `homebox-client`
   - Client Protocol: `openid-connect`
   - Access Type: `confidential`

2. **Configure client settings:**
   - Standard Flow Enabled: `ON`
   - Direct Access Grants Enabled: `OFF`
   - Valid Redirect URIs: `https://your-homebox-instance.com/api/v1/auth/oidc/callback`
   - Web Origins: `https://your-homebox-instance.com`

3. **Configure client scopes:**
   - Ensure the client has access to `email`, `profile`, and `groups` scopes
   - Add group membership to the ID token if using role-based access

4. **Get configuration details:**
   - Issuer URL: `https://your-keycloak.com/auth/realms/your-realm`
   - Client Secret: Available in the "Credentials" tab

### Auth0

1. **Create a new application:**
   - Application Type: `Regular Web Application`
   - Technology: `Generic`

2. **Configure application settings:**
   - Allowed Callback URLs: `https://your-homebox-instance.com/api/v1/auth/oidc/callback`
   - Allowed Web Origins: `https://your-homebox-instance.com`
   - Allowed Logout URLs: `https://your-homebox-instance.com`

3. **Get configuration details:**
   - Issuer URL: `https://your-domain.auth0.com/`
   - Client ID and Client Secret: Available in application settings

### Google

1. **Create OAuth 2.0 credentials:**
   - Go to [Google Cloud Console](https://console.cloud.google.com/)
   - Create OAuth 2.0 Client IDs
   - Application type: `Web application`
   - Authorized redirect URIs: `https://your-homebox-instance.com/api/v1/auth/oidc/callback`

2. **Configuration:**
   - Issuer URL: `https://accounts.google.com`
   - Client ID and Client Secret: From Google Cloud Console

### Microsoft Azure AD

1. **Register an application:**
   - Go to Azure Portal > Azure Active Directory > App registrations
   - New registration with redirect URI: `https://your-homebox-instance.com/api/v1/auth/oidc/callback`

2. **Configure API permissions:**
   - Add `openid`, `email`, `profile` permissions
   - Grant admin consent if required

3. **Configuration:**
   - Issuer URL: `https://login.microsoftonline.com/{tenant-id}/v2.0`
   - Client ID: Application (client) ID
   - Client Secret: From "Certificates & secrets"

## Role Mapping

Homebox supports role-based access control through OIDC group claims:

- **Admin Role**: Users in this group become owners with full access
- **User Role**: Standard users with normal permissions
- **No Group**: Users not in any recognized group get standard user access

Configure the `roles_claim` to specify which claim contains group information (default: `groups`).

## User Management

### First-Time Login

When a user logs in via OIDC for the first time:
1. Homebox creates a new user account automatically
2. The user's email and name are populated from OIDC claims
3. No password is set (password field remains empty)
4. The user's role is determined by group membership

### Existing Users

- Local users can continue using username/password authentication
- OIDC and local authentication can coexist
- Users created via OIDC cannot use password authentication

## Security Considerations

1. **HTTPS Required**: Always use HTTPS for production deployments
2. **Client Secret**: Keep the client secret secure and rotate it regularly
3. **Redirect URL**: Ensure redirect URLs are exact matches to prevent attacks
4. **Email Verification**: Only users with verified emails can access Homebox
5. **State Parameter**: Homebox uses state parameters to prevent CSRF attacks

## Troubleshooting

### Common Issues

**"Invalid state parameter"**
- Check that your system time is synchronized
- Verify the redirect URL matches exactly

**"Email not verified"**
- Ensure the OIDC provider returns `email_verified: true`
- Check that email verification is enabled in your OIDC provider

**"Failed to get OIDC provider"**
- Verify the issuer URL is correct and accessible
- Check that the OIDC provider's discovery endpoint is available at `{issuer_url}/.well-known/openid_configuration`

**"No id_token field in oauth2 token"**
- Ensure the client is configured for OpenID Connect (not just OAuth2)
- Verify that the `openid` scope is included

### Debug Logging

Enable debug logging to troubleshoot OIDC issues:

```bash
HBOX_LOG_LEVEL=debug
```

Look for log entries containing "OIDC" or "auth" for detailed authentication flow information.

## Migration from Local Auth

To migrate existing users to OIDC:

1. Ensure OIDC users use the same email addresses as local users
2. Local user accounts will be matched by email automatically
3. Consider disabling new user registration once OIDC is enabled
4. Local authentication remains available for existing users

## Example Docker Compose

```yaml
version: '3.7'
services:
  homebox:
    image: ghcr.io/sysadminsmedia/homebox:latest
    environment:
      HBOX_OIDC_ENABLED: "true"
      HBOX_OIDC_ISSUER_URL: "https://auth.example.com/auth/realms/homebox"
      HBOX_OIDC_CLIENT_ID: "homebox-client"
      HBOX_OIDC_CLIENT_SECRET: "your-secret-here"
      HBOX_OIDC_REDIRECT_URL: "https://homebox.example.com/api/v1/auth/oidc/callback"
      HBOX_OIDC_SCOPES: "openid email profile groups"
    ports:
      - "7745:7745"
```