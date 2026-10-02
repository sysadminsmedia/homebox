import { toBool, pushEnv, envLine, yamlScalar, CONTAINER_PORT } from './utils.js';
import { getLabels, proxyServiceBlock } from './proxyService.js';

/**
 * Assembles a full compose.yml string from the current designer state.
 *
 * @param {Record<string, any>} state
 * @returns {string}
 */
export function buildComposeYaml(state) {
    const imageTag =
        state.imageVariant === 'regular' ? 'latest' : `latest-${state.imageVariant}`;

    const homeboxVolumes = [
        state.storageType === 'bind'
            ? `      - ${yamlScalar(`${state.bindPath}:/data/`)}`
            : '      - homebox-data:/data/',
    ];

    // The credentials file has to exist inside the container for
    // GOOGLE_APPLICATION_CREDENTIALS to resolve, so mount it alongside the data.
    if (state.storageBackend === 'gcp') {
        homeboxVolumes.push(
            `      - ${yamlScalar(`./gcp-service-account.json:${state.gcpCredentialsPath}:ro`)}`
        );
    }

    const postgresVolume =
        state.postgresStorageType === 'bind'
            ? `      - ${yamlScalar(`${state.postgresBindPath}:/var/lib/postgresql/data`)}`
            : '      - postgres:/var/lib/postgresql/data';

    const postgresService =
        state.databaseType === 'postgres'
            ? [
                  '  postgres:',
                  '    image: postgres:17-alpine',
                  '    restart: unless-stopped',
                  '    volumes:',
                  postgresVolume,
                  '    environment:',
                  `      POSTGRES_PASSWORD: ${yamlScalar(state.postgresPassword)}`,
                  `      POSTGRES_USER: ${yamlScalar(state.postgresUser)}`,
                  `      POSTGRES_DB: ${yamlScalar(state.postgresDatabase)}`,
                  // Without this, depends_on only waits for the container to start,
                  // not for the database to accept connections. Exec form keeps the
                  // names as single arguments instead of handing them to a shell.
                  '    healthcheck:',
                  `      test: ["CMD", "pg_isready", "-U", ${JSON.stringify(state.postgresUser)}, "-d", ${JSON.stringify(state.postgresDatabase)}]`,
                  '      interval: 10s',
                  '      timeout: 5s',
                  '      retries: 5',
                  '      start_period: 30s',
              ].join('\n')
            : '';

    const envLines = [
        envLine('HBOX_LOG_LEVEL', state.logLevel),
        envLine('HBOX_LOG_FORMAT', state.logFormat),
        envLine('HBOX_WEB_MAX_UPLOAD_SIZE', state.maxUploadSize),
        envLine('HBOX_WEB_MAX_IMPORT_SIZE', state.maxImportSize),
        envLine('HBOX_OPTIONS_ALLOW_ANALYTICS', toBool(state.allowAnalytics)),
        envLine('HBOX_OPTIONS_ALLOW_REGISTRATION', toBool(state.allowRegistration)),
        envLine('HBOX_OPTIONS_AUTO_INCREMENT_ASSET_ID', toBool(state.autoIncrementAssetId)),
        envLine('HBOX_OPTIONS_GITHUB_RELEASE_CHECK', toBool(state.githubReleaseCheck)),
        // Required: Homebox refuses to start when this is under 32 bytes.
        '      # Required, min 32 bytes. Keep it stable: changing it invalidates all API keys',
        envLine('HBOX_AUTH_API_KEY_PEPPER', state.apiKeyPepper),
    ];

    pushEnv(envLines, 'HBOX_OPTIONS_CURRENCY_CONFIG', state.currencyConfig);

    if (!state.thumbnailEnabled) {
        envLines.push(envLine('HBOX_THUMBNAIL_ENABLED', 'false'));
    } else {
        envLines.push(envLine('HBOX_THUMBNAIL_WIDTH', state.thumbnailWidth));
        envLines.push(envLine('HBOX_THUMBNAIL_HEIGHT', state.thumbnailHeight));
    }

    if (state.databaseType === 'postgres') {
        envLines.push(envLine('HBOX_DATABASE_DRIVER', 'postgres'));
        envLines.push(envLine('HBOX_DATABASE_HOST', 'postgres'));
        envLines.push(envLine('HBOX_DATABASE_PORT', '5432'));
        envLines.push(envLine('HBOX_DATABASE_USERNAME', state.postgresUser));
        envLines.push(envLine('HBOX_DATABASE_PASSWORD', state.postgresPassword));
        envLines.push(envLine('HBOX_DATABASE_DATABASE', state.postgresDatabase));
        // Defaults to "require", which the postgres:17-alpine image above cannot
        // satisfy - it ships with ssl off - so the mode has to be explicit.
        envLines.push(envLine('HBOX_DATABASE_SSL_MODE', state.databaseSslMode));
    } else if (state.sqlitePath) {
        envLines.push(envLine('HBOX_DATABASE_SQLITE_PATH', state.sqlitePath));
    }

    if (state.proxyType !== 'none') {
        envLines.push(envLine('HBOX_OPTIONS_TRUST_PROXY', 'true'));
        pushEnv(envLines, 'HBOX_OPTIONS_HOSTNAME', state.hostname);
    }

    if (state.storageBackend === 's3') {
        envLines.push(envLine('HBOX_STORAGE_CONN_STRING', state.s3ConnString));
        envLines.push(envLine('AWS_ACCESS_KEY_ID', state.awsAccessKeyId));
        envLines.push(envLine('AWS_SECRET_ACCESS_KEY', state.awsSecretAccessKey));
    }

    if (state.storageBackend === 'gcp') {
        envLines.push(envLine('HBOX_STORAGE_CONN_STRING', state.gcpConnString));
        envLines.push(envLine('GOOGLE_APPLICATION_CREDENTIALS', state.gcpCredentialsPath));
    }

    if (state.storageBackend === 'azure') {
        envLines.push(envLine('HBOX_STORAGE_CONN_STRING', state.azureConnString));
        envLines.push(envLine('AZURE_STORAGE_ACCOUNT', state.azureStorageAccount));
        envLines.push(envLine('AZURE_STORAGE_KEY', state.azureStorageKey));
    }

    if (state.oidcEnabled) {
        envLines.push(envLine('HBOX_OIDC_ENABLED', 'true'));
        envLines.push(envLine('HBOX_OIDC_ISSUER_URL', state.oidcIssuerUrl));
        envLines.push(envLine('HBOX_OIDC_CLIENT_ID', state.oidcClientId));
        envLines.push(envLine('HBOX_OIDC_CLIENT_SECRET', state.oidcClientSecret));
        envLines.push(envLine('HBOX_OIDC_SCOPE', state.oidcScope));
        pushEnv(envLines, 'HBOX_OIDC_ALLOWED_GROUPS', state.oidcAllowedGroups);
        envLines.push(envLine('HBOX_OIDC_AUTO_REDIRECT', toBool(state.oidcAutoRedirect)));
        envLines.push(envLine('HBOX_OIDC_VERIFY_EMAIL', toBool(state.oidcVerifyEmail)));
        envLines.push(envLine('HBOX_OPTIONS_ALLOW_LOCAL_LOGIN', toBool(state.allowLocalLogin)));
    }

    const labels = getLabels(state);

    // Waiting on service_healthy avoids the start-up crash loop that a plain
    // list-form depends_on leaves in place while postgres is still initialising.
    const dependsOn =
        state.databaseType === 'postgres'
            ? '    depends_on:\n      postgres:\n        condition: service_healthy\n'
            : '';
    const sidecarService = proxyServiceBlock(state);

    // Behind a reverse proxy the app must not also be reachable on the host port,
    // or requests can bypass whatever TLS and auth the proxy terminates.
    const portsSection =
        state.proxyType === 'none'
            ? ['    ports:', `      - "3100:${CONTAINER_PORT}"`]
            : [
                  '    expose:',
                  `      - "${CONTAINER_PORT}"`,
                  '    # Add a ports: mapping here only if you also need to reach Homebox',
                  '    # directly, bypassing the proxy.',
              ];

    const volumeLines = [];
    if (state.storageType === 'volume') {
        volumeLines.push('  homebox-data:');
    }
    if (state.databaseType === 'postgres' && state.postgresStorageType === 'volume') {
        volumeLines.push('  postgres:');
    }
    if (state.proxyType === 'caddy') {
        volumeLines.push('  caddy-data:');
        volumeLines.push('  caddy-config:');
    }

    const volumesSection = volumeLines.length
        ? `\nvolumes:\n${volumeLines.join('\n')}`
        : '';

    const homeboxService = [
        '  homebox:',
        `    image: ghcr.io/sysadminsmedia/homebox:${imageTag}`,
        '    restart: unless-stopped',
        dependsOn.trimEnd(),
        '    environment:',
        envLines.join('\n'),
        '    volumes:',
        homeboxVolumes.join('\n'),
        portsSection.join('\n'),
        labels,
    ]
        .filter((line) => line !== '')
        .join('\n');

    const services = [homeboxService, postgresService.trimEnd(), sidecarService.trimEnd()]
        .filter((block) => block !== '')
        .join('\n\n');

    return `services:\n${services}\n${volumesSection.trimEnd()}\n`;
}
