export function generateDockerCompose(config: any): string {
    const services: any = {}
    const volumes: any = {}
    const networks: any = {
        homebox: {
            driver: 'bridge'
        }
    }

    // Generate Homebox service
    services.homebox = generateHomeboxService(config)

    // Add database service if PostgreSQL is selected
    if (config.databaseType === 'postgres') {
        services.postgres = generatePostgresService(config)
        if (config.storageConfig.containerStorage.postgresStorage.type === 'volume') {
            volumes[config.storageConfig.containerStorage.postgresStorage.volumeName] = null
        }
    }

    // Ensure homebox-data volume exists if SQLite is selected
    if (config.databaseType === 'sqlite') {
        volumes['homebox-data'] = null
    }

    // Add reverse proxy services based on HTTPS option
    switch (config.httpsOption) {
        case 'traefik':
            services.traefik = generateTraefikService(config)
            if (config.storageConfig.containerStorage.traefikStorage.type === 'volume') {
                volumes[config.storageConfig.containerStorage.traefikStorage.volumeName] = null
            }
            break
        case 'nginx':
            services.nginx = generateNginxService(config)
            if (config.storageConfig.containerStorage.nginxStorage.type === 'volume') {
                volumes[config.storageConfig.containerStorage.nginxStorage.volumeName] = null
            }
            break
        case 'caddy':
            services.caddy = generateCaddyService(config)
            if (config.storageConfig.containerStorage.caddyStorage.type === 'volume') {
                volumes[config.storageConfig.containerStorage.caddyStorage.volumeName] = null
            }
            break
        case 'cloudflared':
            services.cloudflared = generateCloudflaredService(config)
            if (config.storageConfig.containerStorage.cloudflaredStorage.type === 'volume') {
                volumes[config.storageConfig.containerStorage.cloudflaredStorage.volumeName] = null
            }
            break
    }

    // Add Homebox storage volume only for local storage
    if (config.storageType === 'local' && config.storageConfig.local.type === 'volume') {
        volumes[config.storageConfig.local.volumeName] = null
    }

    const compose = {
        version: '3.8',
        services,
        ...(Object.keys(volumes).length > 0 && {volumes}),
        networks
    }

    return `# Generated Homebox Docker Compose Config Generator 1.0 Beta
# Storage Type: ${config.storageType.toUpperCase()}
# Generated on: ${new Date().toISOString()}
${yaml.stringify(compose)}`
}

function generateHomeboxService(config: any): any {
    const service: any = {
        image: config.rootless ? config.image.replace(':latest', ':latest-rootless') : config.image,
        container_name: 'homebox',
        restart: 'unless-stopped',
        environment: generateEnvironmentVariables(config),
        networks: ['homebox']
    }

    // Add ports for direct access (when no reverse proxy is used)
    if (config.httpsOption === 'none') {
        service.ports = [`${config.port}:7745`]
    }

    // Configure storage based on storage type
    if (config.storageType === 'local') {
        service.volumes = generateLocalStorageVolumes(config)
    } else {
        // For cloud storage, we might still need some local volumes for certain files
        service.volumes = generateCloudStorageVolumes(config)
    }

    // Always mount homebox-data at /data if SQLite is used
    if (config.databaseType === 'sqlite') {
        if (!service.volumes) service.volumes = []
        // Only add if not already present
        if (!service.volumes.some(v => v.startsWith('homebox-data:'))) {
            service.volumes.push('homebox-data:/data')
        }
    }

    return service
}

function generateEnvironmentVariables(config: any): string[] {
    const env: string[] = [
        `HBOX_LOG_LEVEL=${config.logLevel}`,
        `HBOX_LOG_FORMAT=${config.logFormat}`,
        `HBOX_MAX_UPLOAD_SIZE=${config.maxFileUpload}`,
        `HBOX_AUTO_INCREMENT_ASSET_ID=${config.autoIncrementAssetId}`,
        `HBOX_WEB_PORT=7745`
    ]

    // Database configuration
    if (config.databaseType === 'postgres') {
        env.push(
            `HBOX_DATABASE_DRIVER=postgres`,
            `HBOX_DATABASE_HOST=${config.postgresConfig.host}`,
            `HBOX_DATABASE_PORT=${config.postgresConfig.port}`,
            `HBOX_DATABASE_NAME=${config.postgresConfig.database}`,
            `HBOX_DATABASE_USER=${config.postgresConfig.username}`,
            `HBOX_DATABASE_PASS=${config.postgresConfig.password}`
        )
    }

    // Registration settings
    if (!config.allowRegistration) {
        env.push('HBOX_OPTIONS_ALLOW_REGISTRATION=false')
    }

    // Analytics settings
    if (!config.allowAnalytics) {
        env.push('HBOX_OPTIONS_ALLOW_ANALYTICS=false')
    }

    // GitHub release check
    if (!config.checkGithubRelease) {
        env.push('HBOX_OPTIONS_CHECK_GITHUB_RELEASE=false')
    }

    // Storage configuration
    env.push(...generateStorageEnvironmentVariables(config))

    return env
}

function generateStorageEnvironmentVariables(config: any): string[] {
    const env: string[] = []

    switch (config.storageType) {
        case 'local':
            const storagePath = config.storageConfig.local.path || '/data'
            env.push(`HBOX_STORAGE_CONN_STRING=file://${storagePath}`)
            if (config.storageConfig.local.prefixPath) {
                env.push(`HBOX_STORAGE_PREFIX_PATH=${config.storageConfig.local.prefixPath}`)
            }
            break

        case 's3':
            const s3Config = config.storageConfig.s3
            let connectionString = `s3://${s3Config.bucket}?awssdk=${s3Config.awsSdk}`

            if (s3Config.region && !s3Config.isCompatible) {
                connectionString += `&region=${s3Config.region}`
            }

            if (s3Config.endpoint) {
                connectionString += `&endpoint=${s3Config.endpoint}`
            }

            if (s3Config.disableSSL) {
                connectionString += '&disableSSL=true'
            }

            if (s3Config.s3ForcePathStyle) {
                connectionString += '&s3ForcePathStyle=true'
            }

            if (s3Config.sseType) {
                connectionString += `&sseType=${s3Config.sseType}`
            }

            if (s3Config.kmsKeyId) {
                connectionString += `&kmskeyid=${s3Config.kmsKeyId}`
            }

            if (s3Config.fips) {
                connectionString += '&fips=true'
            }

            if (s3Config.dualstack) {
                connectionString += '&dualstack=true'
            }

            if (s3Config.accelerate) {
                connectionString += '&accelerate=true'
            }

            env.push(`HBOX_STORAGE_CONN_STRING=${connectionString}`)

            if (s3Config.prefixPath) {
                env.push(`HBOX_STORAGE_PREFIX_PATH=${s3Config.prefixPath}`)
            }

            // AWS credentials
            env.push(`AWS_ACCESS_KEY_ID=${s3Config.awsAccessKeyId}`)
            env.push(`AWS_SECRET_ACCESS_KEY=${s3Config.awsSecretAccessKey}`)

            if (s3Config.awsSessionToken) {
                env.push(`AWS_SESSION_TOKEN=${s3Config.awsSessionToken}`)
            }
            break

        case 'gcs':
            const gcsConfig = config.storageConfig.gcs
            env.push(`HBOX_STORAGE_CONN_STRING=gcs://${gcsConfig.bucket}`)

            if (gcsConfig.prefixPath) {
                env.push(`HBOX_STORAGE_PREFIX_PATH=${gcsConfig.prefixPath}`)
            }

            env.push(`GOOGLE_APPLICATION_CREDENTIALS=${gcsConfig.credentialsPath}`)
            break

        case 'azure':
            const azureConfig = config.storageConfig.azure
            let azureConnectionString = `azblob://${azureConfig.container}`

            if (azureConfig.useEmulator) {
                azureConnectionString += `?protocol=http&domain=${azureConfig.emulatorEndpoint}`
            }

            env.push(`HBOX_STORAGE_CONN_STRING=${azureConnectionString}`)

            if (azureConfig.prefixPath) {
                env.push(`HBOX_STORAGE_PREFIX_PATH=${azureConfig.prefixPath}`)
            }

            if (!azureConfig.useEmulator) {
                env.push(`AZURE_STORAGE_ACCOUNT=${azureConfig.storageAccount}`)

                if (azureConfig.sasToken) {
                    env.push(`AZURE_STORAGE_SAS_TOKEN=${azureConfig.sasToken}`)
                } else {
                    env.push(`AZURE_STORAGE_KEY=${azureConfig.storageKey}`)
                }
            }
            break
    }

    return env
}

function generateLocalStorageVolumes(config: any): string[] {
    const volumes: string[] = []

    if (config.storageConfig.local.type === 'volume') {
        const mountPath = config.storageConfig.local.path || '/data'
        volumes.push(`${config.storageConfig.local.volumeName}:${mountPath}`)
    } else {
        const mountPath = config.storageConfig.local.path || '/data'
        volumes.push(`${config.storageConfig.local.directory}:${mountPath}`)
    }

    return volumes
}

function generateCloudStorageVolumes(config: any): string[] {
    const volumes: string[] = []

    // For cloud storage, we might still need local volumes for certain files like GCS credentials
    if (config.storageType === 'gcs') {
        volumes.push('/path/to/gcs-credentials.json:/app/gcs-credentials.json:ro')
    }

    return volumes
}

function generatePostgresService(config: any): any {
    const service: any = {
        image: 'postgres:17-alpine',
        container_name: 'homebox_postgres',
        restart: 'unless-stopped',
        environment: [
            `POSTGRES_USER=${config.postgresConfig.username}`,
            `POSTGRES_PASSWORD=${config.postgresConfig.password}`,
            `POSTGRES_DB=${config.postgresConfig.database}`
        ],
        networks: ['homebox']
    }

    if (config.storageConfig.containerStorage.postgresStorage.type === 'volume') {
        service.volumes = [`${config.storageConfig.containerStorage.postgresStorage.volumeName}:/var/lib/postgresql/data`]
    } else {
        service.volumes = [`${config.storageConfig.containerStorage.postgresStorage.directory}:/var/lib/postgresql/data`]
    }

    return service
}

function generateTraefikService(config: any): any {
    const service: any = {
        image: 'traefik:v3.0',
        container_name: 'traefik',
        restart: 'unless-stopped',
        command: [
            '--api.dashboard=true',
            '--providers.docker=true',
            '--providers.docker.exposedbydefault=false',
            '--entrypoints.web.address=:80',
            '--entrypoints.websecure.address=:443',
            '--certificatesresolvers.letsencrypt.acme.tlschallenge=true',
            `--certificatesresolvers.letsencrypt.acme.email=${config.traefikConfig.email}`,
            '--certificatesresolvers.letsencrypt.acme.storage=/letsencrypt/acme.json'
        ],
        ports: ['80:80', '443:443'],
        networks: ['homebox'],
        labels: [
            'traefik.enable=true',
            'traefik.http.routers.traefik.rule=Host(`traefik.${config.traefikConfig.domain}`)',
            'traefik.http.routers.traefik.entrypoints=websecure',
            'traefik.http.routers.traefik.tls.certresolver=letsencrypt',
            'traefik.http.routers.traefik.service=api@internal'
        ]
    }

    if (config.storageConfig.containerStorage.traefikStorage.type === 'volume') {
        service.volumes = [
            '/var/run/docker.sock:/var/run/docker.sock:ro',
            `${config.storageConfig.containerStorage.traefikStorage.volumeName}:/letsencrypt`
        ]
    } else {
        service.volumes = [
            '/var/run/docker.sock:/var/run/docker.sock:ro',
            `${config.storageConfig.containerStorage.traefikStorage.directory}:/letsencrypt`
        ]
    }

    return service
}

function generateNginxService(config: any): any {
    // This would generate an Nginx service with SSL configuration
    // Implementation would depend on specific Nginx configuration needs
    return {
        image: 'nginx:alpine',
        container_name: 'nginx',
        restart: 'unless-stopped',
        ports: [`${config.nginxConfig.port}:443`, '80:80'],
        networks: ['homebox']
    }
}

function generateCaddyService(config: any): any {
    return {
        image: 'caddy:alpine',
        container_name: 'caddy',
        restart: 'unless-stopped',
        ports: ['80:80', '443:443'],
        networks: ['homebox']
    }
}

function generateCloudflaredService(config: any): any {
    return {
        image: 'cloudflare/cloudflared:latest',
        container_name: 'cloudflared',
        restart: 'unless-stopped',
        command: `tunnel --no-autoupdate run --token ${config.cloudflaredConfig.token}`,
        networks: ['homebox']
    }
}

// Simple YAML stringifier (basic implementation

const yaml = {
    stringify(obj: any, indent = 0, parentKey = "", isTopLevel = true): string {
        const spaces = '  '.repeat(indent)
        const nextSpaces = '  '.repeat(indent + 1)
        if (obj === null || obj === undefined) {
            return 'null'
        }
        if (typeof obj === 'string') {
            if (parentKey === 'environment') {
                // Should not be used, handled by stringifyEnv
                return obj
            }
            if (obj.includes(':') || obj.includes('#') || obj.includes('\n') || /^[0-9]/.test(obj) || obj.includes('${')) {
                return `"${obj.replace(/"/g, '\\"')}"`
            }
            return obj
        }
        if (typeof obj === 'number' || typeof obj === 'boolean') {
            return String(obj)
        }
        if (Array.isArray(obj)) {
            if (obj.length === 0) return '[]'
            if (parentKey === 'environment') {
                return yaml.stringifyEnv(obj, indent)
            }
            // For arrays under object keys, indent dashes at the same level as the parent key's value (spaces)
            return '\n' + obj.map(item => `${spaces}- ${this.stringify(item, indent + 1, '', false).replace(/^\s+/, '')}`).join('\n')
        }
        if (typeof obj === 'object') {
            const keys = Object.keys(obj)
            if (keys.length === 0) return '{}'
            return (isTopLevel ? '' : '\n') + keys.map(key => {
                const value = this.stringify(obj[key], indent + 1, key, false)
                // If value is an array, ensure correct indentation
                if (Array.isArray(obj[key])) {
                    // Place key at current indent, then array items at next indent
                    return `${isTopLevel ? '' : spaces}${key}:${value}`
                }
                if (value.startsWith('\n')) {
                    return `${isTopLevel ? '' : spaces}${key}:${value}`
                }
                return `${isTopLevel ? '' : spaces}${key}: ${value}`
            }).join('\n')
        }
        return String(obj)
    },

    stringifyEnv(envArr: string[], indent = 0): string {
        const spaces = '  '.repeat(indent)
        return '\n' + envArr.map(env => {
            const eqIdx = env.indexOf('=')
            if (eqIdx !== -1) {
                const key = env.slice(0, eqIdx + 1)
                let value = env.slice(eqIdx + 1)
                // Only quote the value if it contains special YAML characters
                if (value.match(/[:#\n]|^\d|\${/)) {
                    value = `"${value.replace(/"/g, '\\"')}"`
                }
                return `${spaces}- ${key}${value}`
            }
            return `${spaces}- ${env}`
        }).join('\n')
    }
}
