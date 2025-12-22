# Configure OIDC

HomeBox supports OpenID Connect (OIDC) authentication, allowing users to login using external identity providers like Keycloak, Authentik, Authelia, Google, Microsoft, etc.

::: tip OIDC Provider Documentation
When configuring OIDC, always refer to the documentation provided by your identity provider for specific details and requirements.
:::

## Basic OIDC Setup

1. **Enable OIDC**: Set `HBOX_OIDC_ENABLED=true`.
2. **Provider Configuration**: Set the required provider details:
   - `HBOX_OIDC_ISSUER_URL`: Your OIDC provider's issuer URL.
     - Generally this URL should not have a trailing slash, though it may be required for some providers.
   - `HBOX_OIDC_CLIENT_ID`: Client ID from your OIDC provider.
   - `HBOX_OIDC_CLIENT_SECRET`: Client secret from your OIDC provider.
   - If you are using a reverse proxy, it may be necessary to set `HBOX_OPTIONS_TRUST_PROXY=true` to ensure `https` is correctly detected.
   - If you have set `HBOX_OPTIONS_HOSTNAME` make sure it is just the hostname and does not include `https://` or `http://`.

3. **Configure Redirect URI**: In your OIDC provider, set the redirect URI to:
   `https://your-homebox-domain.example.com/api/v1/users/login/oidc/callback`.

## Advanced OIDC Configuration

- **Group Authorization**: Use `HBOX_OIDC_ALLOWED_GROUPS` to restrict access to specific groups, e.g. `HBOX_OIDC_ALLOWED_GROUPS=admin,homebox`.
  - Some providers require the `groups` scope to return group claims, include it in `HBOX_OIDC_SCOPE` (e.g. `openid profile email groups`) or configure the provider to release the claim.
- **Custom Claims**: Configure `HBOX_OIDC_GROUP_CLAIM`, `HBOX_OIDC_EMAIL_CLAIM`, and `HBOX_OIDC_NAME_CLAIM` if your provider uses different claim names.
  - These default to `HBOX_OIDC_GROUP_CLAIM=groups`, `HBOX_OIDC_EMAIL_CLAIM=email` and `HBOX_OIDC_NAME_CLAIM=name`.
- **Auto Redirect to OIDC**: Set `HBOX_OIDC_AUTO_REDIRECT=true` to automatically redirect users directly to OIDC.
- **Local Login**: Set `HBOX_OPTIONS_ALLOW_LOCAL_LOGIN=false` to completely disable username/password login.
- **Email Verification**: Set `HBOX_OIDC_VERIFY_EMAIL=true` to require email verification from the OIDC provider.

## Security Considerations

::: warning OIDC Security
- Store `HBOX_OIDC_CLIENT_SECRET` securely (use environment variables, not config files).
- Use HTTPS for production deployments.
- Configure proper redirect URIs in your OIDC provider.
- Consider setting `HBOX_OIDC_ALLOWED_GROUPS` for group-based access control.
:::

::: tip CLI Arguments
If you're deploying without docker you can use command line arguments to configure the application. Run `homebox --help` for more information.
:::