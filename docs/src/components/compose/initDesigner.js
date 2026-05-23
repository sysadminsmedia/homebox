import { buildComposeYaml } from './buildYaml.js';

/**
 * Wires up all form fields inside a single compose-designer root element and
 * keeps the generated YAML output in sync.
 *
 * @param {HTMLElement} root
 */
function initComposeDesigner(root) {
    const fields = {
        imageVariant: root.querySelector('[data-field="imageVariant"]'),
        storageType: root.querySelector('[data-field="storageType"]'),
        databaseType: root.querySelector('[data-field="databaseType"]'),
        postgresUser: root.querySelector('[data-field="postgresUser"]'),
        postgresDatabase: root.querySelector('[data-field="postgresDatabase"]'),
        postgresStorageType: root.querySelector('[data-field="postgresStorageType"]'),
        postgresBindPath: root.querySelector('[data-field="postgresBindPath"]'),
        sqlitePath: root.querySelector('[data-field="sqlitePath"]'),
        proxyType: root.querySelector('[data-field="proxyType"]'),
        hostname: root.querySelector('[data-field="hostname"]'),
        storageBackend: root.querySelector('[data-field="storageBackend"]'),
        s3ConnString: root.querySelector('[data-field="s3ConnString"]'),
        awsAccessKeyId: root.querySelector('[data-field="awsAccessKeyId"]'),
        awsSecretAccessKey: root.querySelector('[data-field="awsSecretAccessKey"]'),
        gcpConnString: root.querySelector('[data-field="gcpConnString"]'),
        gcpCredentialsPath: root.querySelector('[data-field="gcpCredentialsPath"]'),
        azureConnString: root.querySelector('[data-field="azureConnString"]'),
        azureStorageAccount: root.querySelector('[data-field="azureStorageAccount"]'),
        azureStorageKey: root.querySelector('[data-field="azureStorageKey"]'),
        oidcEnabled: root.querySelector('[data-field="oidcEnabled"]'),
        oidcIssuerUrl: root.querySelector('[data-field="oidcIssuerUrl"]'),
        oidcClientId: root.querySelector('[data-field="oidcClientId"]'),
        oidcClientSecret: root.querySelector('[data-field="oidcClientSecret"]'),
        oidcScope: root.querySelector('[data-field="oidcScope"]'),
        oidcAllowedGroups: root.querySelector('[data-field="oidcAllowedGroups"]'),
        oidcAutoRedirect: root.querySelector('[data-field="oidcAutoRedirect"]'),
        oidcVerifyEmail: root.querySelector('[data-field="oidcVerifyEmail"]'),
        allowLocalLogin: root.querySelector('[data-field="allowLocalLogin"]'),
        allowAnalytics: root.querySelector('[data-field="allowAnalytics"]'),
        allowRegistration: root.querySelector('[data-field="allowRegistration"]'),
        githubReleaseCheck: root.querySelector('[data-field="githubReleaseCheck"]'),
        logLevel: root.querySelector('[data-field="logLevel"]'),
        logFormat: root.querySelector('[data-field="logFormat"]'),
        maxUploadSize: root.querySelector('[data-field="maxUploadSize"]'),
        bindPath: root.querySelector('[data-field="bindPath"]'),
        postgresPassword: root.querySelector('[data-field="postgresPassword"]'),
    };

    const bindPathRow = root.querySelector('[data-bind-path-row]');
    const dbPostgresRows = root.querySelectorAll('[data-db-postgres-row]');
    const dbPostgresBindRows = root.querySelectorAll('[data-db-postgres-bind-row]');
    const dbSqliteRows = root.querySelectorAll('[data-db-sqlite-row]');
    const storageLocalSection = root.querySelector('[data-storage-local-section]');
    const storageS3Section = root.querySelector('[data-storage-s3-section]');
    const storageGcpSection = root.querySelector('[data-storage-gcp-section]');
    const storageAzureSection = root.querySelector('[data-storage-azure-section]');
    const oidcRows = root.querySelectorAll('[data-oidc-row]');
    const output = root.querySelector('[data-output-wrapper] code');
    const copyButton = root.querySelector('[data-copy]');
    const rootlessWarning = root.querySelector('[data-warning-rootless]');
    const hardenedWarning = root.querySelector('[data-warning-hardened]');

    const render = () => {
        const state = {
            imageVariant: fields.imageVariant.value,
            storageType: fields.storageType.value,
            databaseType: fields.databaseType.value,
            postgresUser: fields.postgresUser.value || 'homebox',
            postgresDatabase: fields.postgresDatabase.value || 'homebox',
            postgresStorageType: fields.postgresStorageType.value || 'volume',
            postgresBindPath: fields.postgresBindPath.value || '/path/to/postgres/data',
            sqlitePath: fields.sqlitePath.value.trim(),
            proxyType: fields.proxyType.value,
            hostname: fields.hostname.value,
            storageBackend: fields.storageBackend.value,
            s3ConnString: fields.s3ConnString.value || 's3://my-bucket?region=us-east-1',
            awsAccessKeyId: fields.awsAccessKeyId.value || 'your_access_key',
            awsSecretAccessKey: fields.awsSecretAccessKey.value || 'your_secret_key',
            gcpConnString: fields.gcpConnString.value || 'gcs://my-bucket',
            gcpCredentialsPath:
                fields.gcpCredentialsPath.value || '/run/secrets/gcp-service-account.json',
            azureConnString: fields.azureConnString.value || 'azblob://my-container',
            azureStorageAccount: fields.azureStorageAccount.value || 'your_account',
            azureStorageKey: fields.azureStorageKey.value || 'your_storage_key',
            oidcEnabled: fields.oidcEnabled.checked,
            oidcIssuerUrl: fields.oidcIssuerUrl.value,
            oidcClientId: fields.oidcClientId.value,
            oidcClientSecret: fields.oidcClientSecret.value,
            oidcScope: fields.oidcScope.value || 'openid profile email',
            oidcAllowedGroups: fields.oidcAllowedGroups.value,
            oidcAutoRedirect: fields.oidcAutoRedirect.checked,
            oidcVerifyEmail: fields.oidcVerifyEmail.checked,
            allowLocalLogin: fields.allowLocalLogin.checked,
            allowAnalytics: fields.allowAnalytics.checked,
            allowRegistration: fields.allowRegistration.checked,
            githubReleaseCheck: fields.githubReleaseCheck.checked,
            logLevel: fields.logLevel.value || 'info',
            logFormat: fields.logFormat.value || 'text',
            maxUploadSize: fields.maxUploadSize.value || '10',
            bindPath: fields.bindPath.value || '/path/to/data/folder',
            postgresPassword: fields.postgresPassword.value || 'your_secure_password',
        };

        storageLocalSection.style.display = state.storageBackend === 'local' ? '' : 'none';
        bindPathRow.style.display =
            state.storageBackend === 'local' && state.storageType === 'bind' ? '' : 'none';

        dbPostgresRows.forEach((row) => {
            row.style.display = state.databaseType === 'postgres' ? '' : 'none';
        });
        dbPostgresBindRows.forEach((row) => {
            row.style.display =
                state.databaseType === 'postgres' && state.postgresStorageType === 'bind'
                    ? ''
                    : 'none';
        });
        dbSqliteRows.forEach((row) => {
            row.style.display = state.databaseType === 'sqlite' ? '' : 'none';
        });

        storageS3Section.style.display = state.storageBackend === 's3' ? '' : 'none';
        storageGcpSection.style.display = state.storageBackend === 'gcp' ? '' : 'none';
        storageAzureSection.style.display = state.storageBackend === 'azure' ? '' : 'none';

        oidcRows.forEach((row) => {
            row.style.display = state.oidcEnabled ? '' : 'none';
        });

        const isRootlessLike =
            state.imageVariant === 'rootless' || state.imageVariant === 'hardened';
        rootlessWarning.classList.toggle('hidden', !(isRootlessLike && state.storageType === 'bind'));
        hardenedWarning.classList.toggle('hidden', state.imageVariant !== 'hardened');

        output.textContent = buildComposeYaml(state);
    };

    Object.values(fields).forEach((field) => {
        field.addEventListener('input', render);
        field.addEventListener('change', render);
    });

    copyButton.addEventListener('click', async () => {
        const text = output.textContent || '';
        await navigator.clipboard.writeText(text);
        const previous = copyButton.textContent;
        copyButton.textContent = 'Copied';
        setTimeout(() => {
            copyButton.textContent = previous;
        }, 1000);
    });

    render();
}

document.querySelectorAll('[data-compose-designer]').forEach(initComposeDesigner);

