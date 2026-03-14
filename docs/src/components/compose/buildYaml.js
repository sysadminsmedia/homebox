import { toBool, pushEnv } from './utils.js';
import { proxyServiceBlock } from './proxyService.js';

/**
 * Assembles a full compose.yml string from the current designer state.
 *
 * @param {Record<string, any>} state
 * @returns {string}
 */
export function buildComposeYaml(state) {
    const imageTag =
        state.imageVariant === 'regular' ? 'latest' : `latest-${state.imageVariant}`;

    const homeboxVolumes =
        state.storageType === 'bind'
            ? `      - ${state.bindPath}:/data/`
            : '      - homebox-data:/data/';

    const postgresVolume =
        state.postgresStorageType === 'bind'
            ? `      - ${state.postgresBindPath}:/var/lib/postgresql/data`
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
                  `      POSTGRES_PASSWORD: ${state.postgresPassword}`,
                  `      POSTGRES_USER: ${state.postgresUser}`,
                  `      POSTGRES_DB: ${state.postgresDatabase}`,
                  '',
              ].join('\n')
            : '';

    const envLines = [
        '      - HBOX_LOG_LEVEL=' + state.logLevel,
        '      - HBOX_LOG_FORMAT=' + state.logFormat,
        '      - HBOX_WEB_MAX_UPLOAD_SIZE=' + state.maxUploadSize,
        '      - HBOX_OPTIONS_ALLOW_ANALYTICS=' + toBool(state.allowAnalytics),
        '      - HBOX_OPTIONS_ALLOW_REGISTRATION=' + toBool(state.allowRegistration),
        '      - HBOX_OPTIONS_GITHUB_RELEASE_CHECK=' + toBool(state.githubReleaseCheck),
    ];

    if (state.databaseType === 'postgres') {
        envLines.push('      - HBOX_DATABASE_DRIVER=postgres');
        envLines.push('      - HBOX_DATABASE_HOST=postgres');
        envLines.push('      - HBOX_DATABASE_PORT=5432');
        envLines.push(`      - HBOX_DATABASE_USERNAME=${state.postgresUser}`);
        envLines.push(`      - HBOX_DATABASE_PASSWORD=${state.postgresPassword}`);
        envLines.push(`      - HBOX_DATABASE_DATABASE=${state.postgresDatabase}`);
    } else if (state.sqlitePath) {
        envLines.push(`      - HBOX_DATABASE_SQLITE_PATH=${state.sqlitePath}`);
    }

    if (state.proxyType !== 'none') {
        envLines.push('      - HBOX_OPTIONS_TRUST_PROXY=true');
        pushEnv(envLines, 'HBOX_OPTIONS_HOSTNAME', state.hostname);
    }

    if (state.storageBackend === 's3') {
        envLines.push(`      - HBOX_STORAGE_CONN_STRING=${state.s3ConnString}`);
        envLines.push(`      - AWS_ACCESS_KEY_ID=${state.awsAccessKeyId}`);
        envLines.push(`      - AWS_SECRET_ACCESS_KEY=${state.awsSecretAccessKey}`);
    }

    if (state.storageBackend === 'gcp') {
        envLines.push(`      - HBOX_STORAGE_CONN_STRING=${state.gcpConnString}`);
        envLines.push(`      - GOOGLE_APPLICATION_CREDENTIALS=${state.gcpCredentialsPath}`);
    }

    if (state.storageBackend === 'azure') {
        envLines.push(`      - HBOX_STORAGE_CONN_STRING=${state.azureConnString}`);
        envLines.push(`      - AZURE_STORAGE_ACCOUNT=${state.azureStorageAccount}`);
        envLines.push(`      - AZURE_STORAGE_KEY=${state.azureStorageKey}`);
    }

    if (state.oidcEnabled) {
        envLines.push('      - HBOX_OIDC_ENABLED=true');
        envLines.push(`      - HBOX_OIDC_ISSUER_URL=${state.oidcIssuerUrl}`);
        envLines.push(`      - HBOX_OIDC_CLIENT_ID=${state.oidcClientId}`);
        envLines.push(`      - HBOX_OIDC_CLIENT_SECRET=${state.oidcClientSecret}`);
        envLines.push(`      - HBOX_OIDC_SCOPE=${state.oidcScope}`);
        pushEnv(envLines, 'HBOX_OIDC_ALLOWED_GROUPS', state.oidcAllowedGroups);
        envLines.push(`      - HBOX_OIDC_AUTO_REDIRECT=${toBool(state.oidcAutoRedirect)}`);
        envLines.push(`      - HBOX_OIDC_VERIFY_EMAIL=${toBool(state.oidcVerifyEmail)}`);
        envLines.push(`      - HBOX_OPTIONS_ALLOW_LOCAL_LOGIN=${toBool(state.allowLocalLogin)}`);
    }

    const dependsOn =
        state.databaseType === 'postgres' ? '    depends_on:\n      - postgres\n' : '';
    const sidecarService = proxyServiceBlock(state);

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

    return [
        'services:',
        '  homebox:',
        `    image: ghcr.io/sysadminsmedia/homebox:${imageTag}`,
        '    restart: always',
        dependsOn.trimEnd(),
        '    environment:',
        envLines.join('\n'),
        '    volumes:',
        homeboxVolumes,
        '    ports:',
        '      - 3100:7745',
        '',
        postgresService.trimEnd(),
        sidecarService.trimEnd(),
        volumesSection.trimEnd(),
        '',
    ]
        .filter((line) => line !== '')
        .join('\n');
}

