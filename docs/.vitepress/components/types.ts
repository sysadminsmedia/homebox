// types.ts

export type StorageType = "volume" | "directory"
export type HttpsOption = "none" | "traefik" | "nginx" | "caddy" | "cloudflared"
export type DatabaseType = "sqlite" | "postgres"

export interface StorageDetail {
  type: StorageType
  directory: string
  volumeName: string
}

export interface StorageConfig {
  homeboxStorage: StorageDetail
  postgresStorage: StorageDetail
  traefikStorage: StorageDetail
  nginxStorage: StorageDetail
  caddyStorage: StorageDetail
  cloudflaredStorage: StorageDetail
}

export interface PostgresConfig {
  host: string
  port: string
  username: string
  password: string
  database: string
}

export interface TraefikConfig {
  domain: string
  email: string
}

export interface NginxConfig {
  domain: string
  port: string
  sslCertPath: string
  sslKeyPath: string
}

export interface CaddyConfig {
  domain: string
  email: string
}

export interface CloudflaredConfig {
  tunnel: string // Note: This wasn't used in the generator function, but kept for completeness
  domain: string
  token: string
}

export interface AppConfig {
  image: string // Not directly used in generator, but part of the config
  rootless: boolean
  port: string
  logLevel: string
  logFormat: string
  maxFileUpload: string
  allowAnalytics: boolean
  httpsOption: HttpsOption
  traefikConfig: TraefikConfig
  nginxConfig: NginxConfig
  caddyConfig: CaddyConfig
  cloudflaredConfig: CloudflaredConfig
  databaseType: DatabaseType
  postgresConfig: PostgresConfig
  allowRegistration: boolean
  autoIncrementAssetId: boolean
  checkGithubRelease: boolean
  storageConfig: StorageConfig
}

// Types for the generated Docker Compose structure
export interface DockerService {
  image: string
  container_name: string
  restart: string
  environment?: string[]
  volumes: string[]
  ports?: string[]
  expose?: string[]
  labels?: string[]
  command?: string[]
  depends_on?: string[]
}

export interface DockerServices {
  [key: string]: DockerService
}
