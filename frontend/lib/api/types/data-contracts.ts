/* post-processed by ./scripts/process-types.go */
/* eslint-disable */
/* tslint:disable */
// @ts-nocheck
/*
 * ---------------------------------------------------------------
 * ## THIS FILE WAS GENERATED VIA SWAGGER-TYPESCRIPT-API        ##
 * ##                                                           ##
 * ## AUTHOR: acacode                                           ##
 * ## SOURCE: https://github.com/acacode/swagger-typescript-api ##
 * ---------------------------------------------------------------
 */

export enum UsergroupRole {
  DefaultRole = "user",
  RoleUser = "user",
  RoleOwner = "owner",
}

export enum TemplatefieldType {
  TypeText = "text",
  TypeNumber = "number",
  TypeBoolean = "boolean",
  TypeTime = "time",
}

export enum MaintenanceFilterStatus {
  MaintenanceFilterStatusScheduled = "scheduled",
  MaintenanceFilterStatusCompleted = "completed",
  MaintenanceFilterStatusBoth = "both",
}

export enum EntityPathType {
  EntityPathTypeLocation = "location",
  EntityPathTypeItem = "item",
}

export enum ExportStatus {
  DefaultStatus = "pending",
  StatusPending = "pending",
  StatusRunning = "running",
  StatusCompleted = "completed",
  StatusFailed = "failed",
}

export enum ExportOrigin {
  DefaultOrigin = "manual",
  OriginManual = "manual",
  OriginScheduled = "scheduled",
}

export enum ExportKind {
  DefaultKind = "export",
  KindExport = "export",
  KindImport = "import",
}

export enum EntityfieldType {
  TypeText = "text",
  TypeNumber = "number",
  TypeBoolean = "boolean",
  TypeTime = "time",
}

export enum BackupdestinationType {
  DefaultType = "primary",
  TypePrimary = "primary",
  TypeLocal = "local",
  TypeS3 = "s3",
  TypeGcs = "gcs",
  TypeAzblob = "azblob",
  TypeSftp = "sftp",
  TypeWebdav = "webdav",
  TypeGdrive = "gdrive",
  TypeOnedrive = "onedrive",
  TypeDropbox = "dropbox",
  TypeSmb = "smb",
}

export enum BackupdestinationHealthStatus {
  DefaultHealthStatus = "unknown",
  HealthStatusUnknown = "unknown",
  HealthStatusHealthy = "healthy",
  HealthStatusUnreachable = "unreachable",
}

export enum BackupdestinationFrequency {
  DefaultFrequency = "daily",
  FrequencyHourly = "hourly",
  FrequencyDaily = "daily",
  FrequencyWeekly = "weekly",
  FrequencyMonthly = "monthly",
  FrequencyCron = "cron",
}

export enum AuthrolesRole {
  DefaultRole = "user",
  RoleAdmin = "admin",
  RoleUser = "user",
  RoleAttachments = "attachments",
}

export enum AttachmentType {
  DefaultType = "attachment",
  TypePhoto = "photo",
  TypeManual = "manual",
  TypeWarranty = "warranty",
  TypeAttachment = "attachment",
  TypeReceipt = "receipt",
  TypeThumbnail = "thumbnail",
}

export interface CurrenciesCurrency {
  code: string;
  decimals: number;
  local: string;
  name: string;
  symbol: string;
}

export interface EntAPIKey {
  /** CreatedAt holds the value of the "created_at" field. */
  created_at: string;
  /**
   * Edges holds the relations/edges for other nodes in the graph.
   * The values are being populated by the APIKeyQuery when eager-loading is set.
   */
  edges: EntAPIKeyEdges;
  /** ExpiresAt holds the value of the "expires_at" field. */
  expires_at: string;
  /** ID of the ent. */
  id: string;
  /** LastUsedAt holds the value of the "last_used_at" field. */
  last_used_at: string;
  /** Name holds the value of the "name" field. */
  name: string;
  /** UpdatedAt holds the value of the "updated_at" field. */
  updated_at: string;
  /** UserID holds the value of the "user_id" field. */
  user_id: string;
}

export interface EntAPIKeyEdges {
  /** User holds the value of the user edge. */
  user: EntUser;
}

export interface EntAttachment {
  /** CreatedAt holds the value of the "created_at" field. */
  created_at: string;
  /**
   * Edges holds the relations/edges for other nodes in the graph.
   * The values are being populated by the AttachmentQuery when eager-loading is set.
   */
  edges: EntAttachmentEdges;
  /** ID of the ent. */
  id: string;
  /** MimeType holds the value of the "mime_type" field. */
  mime_type: string;
  /** Path holds the value of the "path" field. */
  path: string;
  /** Primary holds the value of the "primary" field. */
  primary: boolean;
  /** Title holds the value of the "title" field. */
  title: string;
  /** Type holds the value of the "type" field. */
  type: AttachmentType;
  /** UpdatedAt holds the value of the "updated_at" field. */
  updated_at: string;
}

export interface EntAttachmentEdges {
  /** Entity holds the value of the entity edge. */
  entity: EntEntity;
  /** Thumbnail holds the value of the thumbnail edge. */
  thumbnail: EntAttachment;
}

export interface EntAuthRoles {
  /**
   * Edges holds the relations/edges for other nodes in the graph.
   * The values are being populated by the AuthRolesQuery when eager-loading is set.
   */
  edges: EntAuthRolesEdges;
  /** ID of the ent. */
  id: number;
  /** Role holds the value of the "role" field. */
  role: AuthrolesRole;
}

export interface EntAuthRolesEdges {
  /** Token holds the value of the token edge. */
  token: EntAuthTokens;
}

export interface EntAuthTokens {
  /** CreatedAt holds the value of the "created_at" field. */
  created_at: string;
  /**
   * Edges holds the relations/edges for other nodes in the graph.
   * The values are being populated by the AuthTokensQuery when eager-loading is set.
   */
  edges: EntAuthTokensEdges;
  /** ExpiresAt holds the value of the "expires_at" field. */
  expires_at: string;
  /** ID of the ent. */
  id: string;
  /** Token holds the value of the "token" field. */
  token: number[];
  /** UpdatedAt holds the value of the "updated_at" field. */
  updated_at: string;
}

export interface EntAuthTokensEdges {
  /** Roles holds the value of the roles edge. */
  roles: EntAuthRoles;
  /** User holds the value of the user edge. */
  user: EntUser;
}

export interface EntBackupDestination {
  /** AlertFailureThreshold holds the value of the "alert_failure_threshold" field. */
  alert_failure_threshold: number;
  /** AlertStaleHours holds the value of the "alert_stale_hours" field. */
  alert_stale_hours: number;
  /** AlertedFailure holds the value of the "alerted_failure" field. */
  alerted_failure: boolean;
  /** AlertedStale holds the value of the "alerted_stale" field. */
  alerted_stale: boolean;
  /** AlertedUnreachable holds the value of the "alerted_unreachable" field. */
  alerted_unreachable: boolean;
  /** AlertsEnabled holds the value of the "alerts_enabled" field. */
  alerts_enabled: boolean;
  /** AtHour holds the value of the "at_hour" field. */
  at_hour: number;
  /** AtMinute holds the value of the "at_minute" field. */
  at_minute: number;
  /** ConnString holds the value of the "conn_string" field. */
  conn_string: string;
  /** CreatedAt holds the value of the "created_at" field. */
  created_at: string;
  /** CronExpr holds the value of the "cron_expr" field. */
  cron_expr: string;
  /** DayOfMonth holds the value of the "day_of_month" field. */
  day_of_month: number;
  /** Description holds the value of the "description" field. */
  description: string;
  /**
   * Edges holds the relations/edges for other nodes in the graph.
   * The values are being populated by the BackupDestinationQuery when eager-loading is set.
   */
  edges: EntBackupDestinationEdges;
  /** Enabled holds the value of the "enabled" field. */
  enabled: boolean;
  /** Frequency holds the value of the "frequency" field. */
  frequency: BackupdestinationFrequency;
  /** GroupID holds the value of the "group_id" field. */
  group_id: string;
  /** HealthCheckedAt holds the value of the "health_checked_at" field. */
  health_checked_at: string;
  /** HealthError holds the value of the "health_error" field. */
  health_error: string;
  /** HealthFailures holds the value of the "health_failures" field. */
  health_failures: number;
  /** HealthIntervalMinutes holds the value of the "health_interval_minutes" field. */
  health_interval_minutes: number;
  /** HealthStatus holds the value of the "health_status" field. */
  health_status: BackupdestinationHealthStatus;
  /** HostKey holds the value of the "host_key" field. */
  host_key: string;
  /** ID of the ent. */
  id: string;
  /** IntervalHours holds the value of the "interval_hours" field. */
  interval_hours: number;
  /** KeepDaily holds the value of the "keep_daily" field. */
  keep_daily: number;
  /** KeepMonthly holds the value of the "keep_monthly" field. */
  keep_monthly: number;
  /** KeepWeekly holds the value of the "keep_weekly" field. */
  keep_weekly: number;
  /** LastError holds the value of the "last_error" field. */
  last_error: string;
  /** LastFingerprint holds the value of the "last_fingerprint" field. */
  last_fingerprint: string;
  /** LastRunAt holds the value of the "last_run_at" field. */
  last_run_at: string;
  /** LastSkippedAt holds the value of the "last_skipped_at" field. */
  last_skipped_at: string;
  /** LastSuccessAt holds the value of the "last_success_at" field. */
  last_success_at: string;
  /** Name holds the value of the "name" field. */
  name: string;
  /** NextRunAt holds the value of the "next_run_at" field. */
  next_run_at: string;
  /** Prefix holds the value of the "prefix" field. */
  prefix: string;
  /** ScheduleEnabled holds the value of the "schedule_enabled" field. */
  schedule_enabled: boolean;
  /** SkipIfUnchanged holds the value of the "skip_if_unchanged" field. */
  skip_if_unchanged: boolean;
  /** Type holds the value of the "type" field. */
  type: BackupdestinationType;
  /** UpdatedAt holds the value of the "updated_at" field. */
  updated_at: string;
  /** Username holds the value of the "username" field. */
  username: string;
  /** Weekday holds the value of the "weekday" field. */
  weekday: number;
}

export interface EntBackupDestinationEdges {
  /** Group holds the value of the group edge. */
  group: EntGroup;
}

export interface EntEntity {
  /** Archived holds the value of the "archived" field. */
  archived: boolean;
  /** AssetID holds the value of the "asset_id" field. */
  asset_id: number;
  /** CreatedAt holds the value of the "created_at" field. */
  created_at: string;
  /** Description holds the value of the "description" field. */
  description: string;
  /**
   * Edges holds the relations/edges for other nodes in the graph.
   * The values are being populated by the EntityQuery when eager-loading is set.
   */
  edges: EntEntityEdges;
  /** ID of the ent. */
  id: string;
  /** ImportRef holds the value of the "import_ref" field. */
  import_ref: string;
  /** Insured holds the value of the "insured" field. */
  insured: boolean;
  /** LifetimeWarranty holds the value of the "lifetime_warranty" field. */
  lifetime_warranty: boolean;
  /** Manufacturer holds the value of the "manufacturer" field. */
  manufacturer: string;
  /** ModelNumber holds the value of the "model_number" field. */
  model_number: string;
  /** Name holds the value of the "name" field. */
  name: string;
  /** Notes holds the value of the "notes" field. */
  notes: string;
  /** PurchaseDate holds the value of the "purchase_date" field. */
  purchase_date: Date | string;
  /** PurchaseFrom holds the value of the "purchase_from" field. */
  purchase_from: string;
  /** PurchasePrice holds the value of the "purchase_price" field. */
  purchase_price: number;
  /** Quantity holds the value of the "quantity" field. */
  quantity: number;
  /** SerialNumber holds the value of the "serial_number" field. */
  serial_number: string;
  /** SoldDate holds the value of the "sold_date" field. */
  sold_date: Date | string;
  /** SoldNotes holds the value of the "sold_notes" field. */
  sold_notes: string;
  /** SoldPrice holds the value of the "sold_price" field. */
  sold_price: number;
  /** SoldTo holds the value of the "sold_to" field. */
  sold_to: string;
  /** SyncChildEntityLocations holds the value of the "sync_child_entity_locations" field. */
  sync_child_entity_locations: boolean;
  /** UpdatedAt holds the value of the "updated_at" field. */
  updated_at: string;
  /** WarrantyDetails holds the value of the "warranty_details" field. */
  warranty_details: string;
  /** WarrantyExpires holds the value of the "warranty_expires" field. */
  warranty_expires: string;
}

export interface EntEntityEdges {
  /** Attachments holds the value of the attachments edge. */
  attachments: EntAttachment[];
  /** Children holds the value of the children edge. */
  children: EntEntity[];
  /** EntityType holds the value of the entity_type edge. */
  entity_type: EntEntityType;
  /** Fields holds the value of the fields edge. */
  fields: EntEntityField[];
  /** Group holds the value of the group edge. */
  group: EntGroup;
  /** Location holds the value of the location edge. */
  location: EntEntity;
  /** LocationEntities holds the value of the location_entities edge. */
  location_entities: EntEntity[];
  /** MaintenanceEntries holds the value of the maintenance_entries edge. */
  maintenance_entries: EntMaintenanceEntry[];
  /** Parent holds the value of the parent edge. */
  parent: EntEntity;
  /** Tag holds the value of the tag edge. */
  tag: EntTag[];
}

export interface EntEntityField {
  /** BooleanValue holds the value of the "boolean_value" field. */
  boolean_value: boolean;
  /** CreatedAt holds the value of the "created_at" field. */
  created_at: string;
  /** Description holds the value of the "description" field. */
  description: string;
  /**
   * Edges holds the relations/edges for other nodes in the graph.
   * The values are being populated by the EntityFieldQuery when eager-loading is set.
   */
  edges: EntEntityFieldEdges;
  /** ID of the ent. */
  id: string;
  /** Name holds the value of the "name" field. */
  name: string;
  /** NumberValue holds the value of the "number_value" field. */
  number_value: number;
  /** TextValue holds the value of the "text_value" field. */
  text_value: string;
  /** TimeValue holds the value of the "time_value" field. */
  time_value: string;
  /** Type holds the value of the "type" field. */
  type: EntityfieldType;
  /** UpdatedAt holds the value of the "updated_at" field. */
  updated_at: string;
}

export interface EntEntityFieldEdges {
  /** Entity holds the value of the entity edge. */
  entity: EntEntity;
}

export interface EntEntityTemplate {
  /** CreatedAt holds the value of the "created_at" field. */
  created_at: string;
  /** Default description for items created from this template */
  default_description: string;
  /** DefaultInsured holds the value of the "default_insured" field. */
  default_insured: boolean;
  /** DefaultLifetimeWarranty holds the value of the "default_lifetime_warranty" field. */
  default_lifetime_warranty: boolean;
  /** DefaultManufacturer holds the value of the "default_manufacturer" field. */
  default_manufacturer: string;
  /** Default model number for items created from this template */
  default_model_number: string;
  /** Default name template for items (can use placeholders) */
  default_name: string;
  /** DefaultQuantity holds the value of the "default_quantity" field. */
  default_quantity: number;
  /** Default tag IDs for items created from this template */
  default_tag_ids: string[];
  /** DefaultWarrantyDetails holds the value of the "default_warranty_details" field. */
  default_warranty_details: string;
  /** Description holds the value of the "description" field. */
  description: string;
  /**
   * Edges holds the relations/edges for other nodes in the graph.
   * The values are being populated by the EntityTemplateQuery when eager-loading is set.
   */
  edges: EntEntityTemplateEdges;
  /** ID of the ent. */
  id: string;
  /** Whether to include purchase fields in items created from this template */
  include_purchase_fields: boolean;
  /** Whether to include sold fields in items created from this template */
  include_sold_fields: boolean;
  /** Whether to include warranty fields in items created from this template */
  include_warranty_fields: boolean;
  /** Name holds the value of the "name" field. */
  name: string;
  /** Notes holds the value of the "notes" field. */
  notes: string;
  /** UpdatedAt holds the value of the "updated_at" field. */
  updated_at: string;
}

export interface EntEntityTemplateEdges {
  /** Fields holds the value of the fields edge. */
  fields: EntTemplateField[];
  /** Group holds the value of the group edge. */
  group: EntGroup;
  /** Location holds the value of the location edge. */
  location: EntEntity;
}

export interface EntEntityType {
  /** CreatedAt holds the value of the "created_at" field. */
  created_at: string;
  /** Description holds the value of the "description" field. */
  description: string;
  /**
   * Edges holds the relations/edges for other nodes in the graph.
   * The values are being populated by the EntityTypeQuery when eager-loading is set.
   */
  edges: EntEntityTypeEdges;
  /** Icon holds the value of the "icon" field. */
  icon: string;
  /** ID of the ent. */
  id: string;
  /** IsLocation holds the value of the "is_location" field. */
  is_location: boolean;
  /** Name holds the value of the "name" field. */
  name: string;
  /** UpdatedAt holds the value of the "updated_at" field. */
  updated_at: string;
}

export interface EntEntityTypeEdges {
  /** DefaultTemplate holds the value of the default_template edge. */
  default_template: EntEntityTemplate;
  /** Entities holds the value of the entities edge. */
  entities: EntEntity[];
  /** Group holds the value of the group edge. */
  group: EntGroup;
}

export interface EntExport {
  /** ArtifactPath holds the value of the "artifact_path" field. */
  artifact_path: string;
  /** CreatedAt holds the value of the "created_at" field. */
  created_at: string;
  /** DestinationID holds the value of the "destination_id" field. */
  destination_id: string;
  /**
   * Edges holds the relations/edges for other nodes in the graph.
   * The values are being populated by the ExportQuery when eager-loading is set.
   */
  edges: EntExportEdges;
  /** Error holds the value of the "error" field. */
  error: string;
  /** GroupID holds the value of the "group_id" field. */
  group_id: string;
  /** ID of the ent. */
  id: string;
  /** Kind holds the value of the "kind" field. */
  kind: ExportKind;
  /** Origin holds the value of the "origin" field. */
  origin: ExportOrigin;
  /** Progress holds the value of the "progress" field. */
  progress: number;
  /** SizeBytes holds the value of the "size_bytes" field. */
  size_bytes: number;
  /** Status holds the value of the "status" field. */
  status: ExportStatus;
  /** UpdatedAt holds the value of the "updated_at" field. */
  updated_at: string;
}

export interface EntExportEdges {
  /** Group holds the value of the group edge. */
  group: EntGroup;
}

export interface EntGroup {
  /** CreatedAt holds the value of the "created_at" field. */
  created_at: string;
  /** Currency holds the value of the "currency" field. */
  currency: string;
  /**
   * Edges holds the relations/edges for other nodes in the graph.
   * The values are being populated by the GroupQuery when eager-loading is set.
   */
  edges: EntGroupEdges;
  /** ID of the ent. */
  id: string;
  /** Name holds the value of the "name" field. */
  name: string;
  /** UpdatedAt holds the value of the "updated_at" field. */
  updated_at: string;
}

export interface EntGroupEdges {
  /** BackupDestinations holds the value of the backup_destinations edge. */
  backup_destinations: EntBackupDestination[];
  /** Entities holds the value of the entities edge. */
  entities: EntEntity[];
  /** EntityTemplates holds the value of the entity_templates edge. */
  entity_templates: EntEntityTemplate[];
  /** EntityTypes holds the value of the entity_types edge. */
  entity_types: EntEntityType[];
  /** Exports holds the value of the exports edge. */
  exports: EntExport[];
  /** InvitationTokens holds the value of the invitation_tokens edge. */
  invitation_tokens: EntGroupInvitationToken[];
  /** Notifiers holds the value of the notifiers edge. */
  notifiers: EntNotifier[];
  /** Tags holds the value of the tags edge. */
  tags: EntTag[];
  /** UserGroups holds the value of the user_groups edge. */
  user_groups: EntUserGroup[];
  /** Users holds the value of the users edge. */
  users: EntUser[];
}

export interface EntGroupInvitationToken {
  /** CreatedAt holds the value of the "created_at" field. */
  created_at: string;
  /**
   * Edges holds the relations/edges for other nodes in the graph.
   * The values are being populated by the GroupInvitationTokenQuery when eager-loading is set.
   */
  edges: EntGroupInvitationTokenEdges;
  /** ExpiresAt holds the value of the "expires_at" field. */
  expires_at: string;
  /** ID of the ent. */
  id: string;
  /** Token holds the value of the "token" field. */
  token: number[];
  /** UpdatedAt holds the value of the "updated_at" field. */
  updated_at: string;
  /** Uses holds the value of the "uses" field. */
  uses: number;
}

export interface EntGroupInvitationTokenEdges {
  /** Group holds the value of the group edge. */
  group: EntGroup;
}

export interface EntMaintenanceEntry {
  /** Cost holds the value of the "cost" field. */
  cost: number;
  /** CreatedAt holds the value of the "created_at" field. */
  created_at: string;
  /** Date holds the value of the "date" field. */
  date: Date | string;
  /** Description holds the value of the "description" field. */
  description: string;
  /**
   * Edges holds the relations/edges for other nodes in the graph.
   * The values are being populated by the MaintenanceEntryQuery when eager-loading is set.
   */
  edges: EntMaintenanceEntryEdges;
  /** EntityID holds the value of the "entity_id" field. */
  entity_id: string;
  /** ID of the ent. */
  id: string;
  /** Name holds the value of the "name" field. */
  name: string;
  /** ScheduledDate holds the value of the "scheduled_date" field. */
  scheduled_date: Date | string;
  /** UpdatedAt holds the value of the "updated_at" field. */
  updated_at: string;
}

export interface EntMaintenanceEntryEdges {
  /** Entity holds the value of the entity edge. */
  entity: EntEntity;
}

export interface EntNotifier {
  /** CreatedAt holds the value of the "created_at" field. */
  created_at: string;
  /**
   * Edges holds the relations/edges for other nodes in the graph.
   * The values are being populated by the NotifierQuery when eager-loading is set.
   */
  edges: EntNotifierEdges;
  /** GroupID holds the value of the "group_id" field. */
  group_id: string;
  /** ID of the ent. */
  id: string;
  /** IsActive holds the value of the "is_active" field. */
  is_active: boolean;
  /** Name holds the value of the "name" field. */
  name: string;
  /** UpdatedAt holds the value of the "updated_at" field. */
  updated_at: string;
  /** UserID holds the value of the "user_id" field. */
  user_id: string;
}

export interface EntNotifierEdges {
  /** Group holds the value of the group edge. */
  group: EntGroup;
  /** User holds the value of the user edge. */
  user: EntUser;
}

export interface EntPasswordResetTokens {
  /** CreatedAt holds the value of the "created_at" field. */
  created_at: string;
  /**
   * Edges holds the relations/edges for other nodes in the graph.
   * The values are being populated by the PasswordResetTokensQuery when eager-loading is set.
   */
  edges: EntPasswordResetTokensEdges;
  /** ExpiresAt holds the value of the "expires_at" field. */
  expires_at: string;
  /** ID of the ent. */
  id: string;
  /** Token holds the value of the "token" field. */
  token: number[];
  /** UpdatedAt holds the value of the "updated_at" field. */
  updated_at: string;
  /** UsedAt holds the value of the "used_at" field. */
  used_at: string;
  /** UserID holds the value of the "user_id" field. */
  user_id: string;
}

export interface EntPasswordResetTokensEdges {
  /** User holds the value of the user edge. */
  user: EntUser;
}

export interface EntTag {
  /** Color holds the value of the "color" field. */
  color: string;
  /** CreatedAt holds the value of the "created_at" field. */
  created_at: string;
  /** Description holds the value of the "description" field. */
  description: string;
  /**
   * Edges holds the relations/edges for other nodes in the graph.
   * The values are being populated by the TagQuery when eager-loading is set.
   */
  edges: EntTagEdges;
  /** Icon holds the value of the "icon" field. */
  icon: string;
  /** ID of the ent. */
  id: string;
  /** Name holds the value of the "name" field. */
  name: string;
  /** UpdatedAt holds the value of the "updated_at" field. */
  updated_at: string;
}

export interface EntTagEdges {
  /** Children holds the value of the children edge. */
  children: EntTag[];
  /** Entities holds the value of the entities edge. */
  entities: EntEntity[];
  /** Group holds the value of the group edge. */
  group: EntGroup;
  /** Parent holds the value of the parent edge. */
  parent: EntTag;
}

export interface EntTemplateField {
  /** BooleanValue holds the value of the "boolean_value" field. */
  boolean_value: boolean;
  /** CreatedAt holds the value of the "created_at" field. */
  created_at: string;
  /** Description holds the value of the "description" field. */
  description: string;
  /**
   * Edges holds the relations/edges for other nodes in the graph.
   * The values are being populated by the TemplateFieldQuery when eager-loading is set.
   */
  edges: EntTemplateFieldEdges;
  /** ID of the ent. */
  id: string;
  /** Name holds the value of the "name" field. */
  name: string;
  /** NumberValue holds the value of the "number_value" field. */
  number_value: number;
  /** TextValue holds the value of the "text_value" field. */
  text_value: string;
  /** TimeValue holds the value of the "time_value" field. */
  time_value: string;
  /** Type holds the value of the "type" field. */
  type: TemplatefieldType;
  /** UpdatedAt holds the value of the "updated_at" field. */
  updated_at: string;
}

export interface EntTemplateFieldEdges {
  /** EntityTemplate holds the value of the entity_template edge. */
  entity_template: EntEntityTemplate;
}

export interface EntUser {
  /** ActivatedOn holds the value of the "activated_on" field. */
  activated_on: string;
  /** CreatedAt holds the value of the "created_at" field. */
  created_at: string;
  /** DefaultGroupID holds the value of the "default_group_id" field. */
  default_group_id: string;
  /**
   * Edges holds the relations/edges for other nodes in the graph.
   * The values are being populated by the UserQuery when eager-loading is set.
   */
  edges: EntUserEdges;
  /** Email holds the value of the "email" field. */
  email: string;
  /** ID of the ent. */
  id: string;
  /** IsSuperuser holds the value of the "is_superuser" field. */
  is_superuser: boolean;
  /** Name holds the value of the "name" field. */
  name: string;
  /** OidcIssuer holds the value of the "oidc_issuer" field. */
  oidc_issuer: string;
  /** OidcSubject holds the value of the "oidc_subject" field. */
  oidc_subject: string;
  /** Settings holds the value of the "settings" field. */
  settings: Record<string, any>;
  /** Superuser holds the value of the "superuser" field. */
  superuser: boolean;
  /** UpdatedAt holds the value of the "updated_at" field. */
  updated_at: string;
}

export interface EntUserEdges {
  /** APIKeys holds the value of the api_keys edge. */
  api_keys: EntAPIKey[];
  /** AuthTokens holds the value of the auth_tokens edge. */
  auth_tokens: EntAuthTokens[];
  /** Groups holds the value of the groups edge. */
  groups: EntGroup[];
  /** Notifiers holds the value of the notifiers edge. */
  notifiers: EntNotifier[];
  /** PasswordResetTokens holds the value of the password_reset_tokens edge. */
  password_reset_tokens: EntPasswordResetTokens[];
  /** UserGroups holds the value of the user_groups edge. */
  user_groups: EntUserGroup[];
}

export interface EntUserGroup {
  /**
   * Edges holds the relations/edges for other nodes in the graph.
   * The values are being populated by the UserGroupQuery when eager-loading is set.
   */
  edges: EntUserGroupEdges;
  /** GroupID holds the value of the "group_id" field. */
  group_id: string;
  /** Role holds the value of the "role" field. */
  role: UsergroupRole;
  /** UserID holds the value of the "user_id" field. */
  user_id: string;
}

export interface EntUserGroupEdges {
  /** Group holds the value of the group edge. */
  group: EntGroup;
  /** User holds the value of the user edge. */
  user: EntUser;
}

export interface APIKeyCreate {
  expiresAt?: string | null;
  /**
   * @minLength 1
   * @maxLength 255
   */
  name: string;
}

export interface APIKeyCreatedOut {
  createdAt: Date | string;
  expiresAt?: string | null;
  id: string;
  lastUsedAt?: string | null;
  name: string;
  token: string;
  userId: string;
}

export interface APIKeyOut {
  createdAt: Date | string;
  expiresAt?: string | null;
  id: string;
  lastUsedAt?: string | null;
  name: string;
  userId: string;
}

export interface BackupDestinationOut {
  /**
   * @min 1
   * @max 100
   */
  alertFailureThreshold: number;
  /**
   * AlertStaleHours alerts when no backup succeeded for this long. 0 disables.
   * @min 0
   * @max 8760
   */
  alertStaleHours: number;
  alertsEnabled: boolean;
  /**
   * @min 0
   * @max 23
   */
  atHour: number;
  /**
   * @min 0
   * @max 59
   */
  atMinute: number;
  /** @maxLength 2048 */
  connString: string;
  createdAt: Date | string;
  /**
   * CronExpr is a 5-field cron expression or descriptor, used when Frequency
   * is "cron". It may start with CRON_TZ=Zone to schedule in another zone.
   * @maxLength 255
   */
  cronExpr: string;
  /**
   * @min 1
   * @max 28
   */
  dayOfMonth: number;
  /** @maxLength 1000 */
  description: string;
  enabled: boolean;
  /** Frequency is one of hourly, daily, weekly, monthly, cron. */
  frequency: "hourly" | "daily" | "weekly" | "monthly" | "cron";
  groupId: string;
  /** HasSecret reports whether credentials are stored for the destination. */
  hasSecret: boolean;
  healthCheckedAt?: string | null;
  healthError: string;
  healthFailures: number;
  /**
   * @min 1
   * @max 1440
   */
  healthIntervalMinutes: number;
  healthStatus: string;
  /** @maxLength 255 */
  hostKey: string;
  id: string;
  /**
   * @min 1
   * @max 168
   */
  intervalHours: number;
  /**
   * @min 0
   * @max 3650
   */
  keepDaily: number;
  /**
   * @min 0
   * @max 120
   */
  keepMonthly: number;
  /**
   * @min 0
   * @max 520
   */
  keepWeekly: number;
  lastError: string;
  lastRunAt?: string | null;
  lastSkippedAt?: string | null;
  lastSuccessAt?: string | null;
  /**
   * @minLength 1
   * @maxLength 255
   */
  name: string;
  nextRunAt?: string | null;
  /** @maxLength 255 */
  prefix: string;
  scheduleEnabled: boolean;
  skipIfUnchanged: boolean;
  /**
   * Type is one of primary, local, s3, gcs, azblob, sftp, webdav, gdrive,
   * onedrive, dropbox, smb.
   */
  type:
    | "primary"
    | "local"
    | "s3"
    | "gcs"
    | "azblob"
    | "sftp"
    | "webdav"
    | "gdrive"
    | "onedrive"
    | "dropbox"
    | "smb";
  updatedAt: Date | string;
  /**
   * Username and HostKey serve the sftp and webdav types. HostKey is the
   * SSH host key fingerprint (SHA256:...) an sftp server must present.
   * @maxLength 255
   */
  username: string;
  /**
   * @min 0
   * @max 6
   */
  weekday: number;
}

export interface BackupInput {
  /**
   * @min 1
   * @max 100
   */
  alertFailureThreshold: number;
  /**
   * AlertStaleHours alerts when no backup succeeded for this long. 0 disables.
   * @min 0
   * @max 8760
   */
  alertStaleHours: number;
  alertsEnabled: boolean;
  /**
   * @min 0
   * @max 23
   */
  atHour: number;
  /**
   * @min 0
   * @max 59
   */
  atMinute: number;
  /** @maxLength 2048 */
  connString: string;
  /**
   * CronExpr is a 5-field cron expression or descriptor, used when Frequency
   * is "cron". It may start with CRON_TZ=Zone to schedule in another zone.
   * @maxLength 255
   */
  cronExpr: string;
  /**
   * @min 1
   * @max 28
   */
  dayOfMonth: number;
  /** @maxLength 1000 */
  description: string;
  /**
   * DestinationID is only used when testing unsaved settings: it lets the
   * test reuse the stored credentials of that destination.
   */
  destinationId: string;
  enabled: boolean;
  /** Frequency is one of hourly, daily, weekly, monthly, cron. */
  frequency: "hourly" | "daily" | "weekly" | "monthly" | "cron";
  /**
   * @min 1
   * @max 1440
   */
  healthIntervalMinutes: number;
  /** @maxLength 255 */
  hostKey: string;
  /**
   * @min 1
   * @max 168
   */
  intervalHours: number;
  /**
   * @min 0
   * @max 3650
   */
  keepDaily: number;
  /**
   * @min 0
   * @max 120
   */
  keepMonthly: number;
  /**
   * @min 0
   * @max 520
   */
  keepWeekly: number;
  /**
   * @minLength 1
   * @maxLength 255
   */
  name: string;
  /**
   * OAuthTicket is the one-time ticket returned by the OAuth callback for a
   * cloud-drive destination. It carries the connected account, and is
   * consumed when the destination is saved.
   */
  oauthTicket: string;
  /** Passphrase unlocks a passphrase-protected PrivateKey. */
  passphrase: string;
  /**
   * Password is the sftp or webdav password. Leave empty to keep the stored
   * credentials of an existing destination.
   */
  password: string;
  /** @maxLength 255 */
  prefix: string;
  /** PrivateKey is a PEM private key for sftp, optionally passphrase-protected. */
  privateKey: string;
  scheduleEnabled: boolean;
  skipIfUnchanged: boolean;
  /**
   * Type is one of primary, local, s3, gcs, azblob, sftp, webdav, gdrive,
   * onedrive, dropbox, smb.
   */
  type:
    | "primary"
    | "local"
    | "s3"
    | "gcs"
    | "azblob"
    | "sftp"
    | "webdav"
    | "gdrive"
    | "onedrive"
    | "dropbox"
    | "smb";
  /**
   * Username and HostKey serve the sftp and webdav types. HostKey is the
   * SSH host key fingerprint (SHA256:...) an sftp server must present.
   * @maxLength 255
   */
  username: string;
  /**
   * @min 0
   * @max 6
   */
  weekday: number;
}

export interface BarcodeProduct {
  barcode: string;
  imageBase64: string;
  imageURL: string;
  item: EntityCreate;
  manufacturer: string;
  /** Identifications */
  modelNumber: string;
  /** Extras */
  notes: string;
  search_engine_name: string;
}

export interface DuplicateOptions {
  copyAttachments: boolean;
  copyCustomFields: boolean;
  copyMaintenance: boolean;
  copyPrefix: string;
}

export interface EntityCreate {
  /** @maxLength 1000 */
  description: string;
  entityTypeId: string;
  /** Only needed when ParentID is another item and this lives elsewhere (#1688). */
  locationId?: string | null;
  /** @maxLength 255 */
  manufacturer?: string | null;
  /**
   * Identifications — optional at create time; populated e.g. by the
   * barcode product-search import flow (#1578).
   * @maxLength 255
   */
  modelNumber?: string | null;
  /**
   * @minLength 1
   * @maxLength 255
   */
  name: string;
  parentId?: string | null;
  quantity: number;
  /** Edges */
  tagIds: string[];
}

export interface EntityFieldData {
  booleanValue: boolean;
  id: string;
  name: string;
  numberValue: number;
  textValue: string;
  type: string;
}

export interface EntityListResult {
  items: EntitySummary[];
  page: number;
  pageSize: number;
  total: number;
  totalPrice: number;
}

export interface EntityOut {
  archived: boolean;
  /** @example "0" */
  assetId: string;
  attachments: ItemAttachment[];
  /** Container-specific fields (for entities whose entity_type.is_location = true) */
  children: EntitySummary[];
  createdAt: Date | string;
  description: string;
  entityType?: EntityTypeSummary | null;
  fields: EntityFieldData[];
  id: string;
  imageId?: string | null;
  insured: boolean;
  /** Container-specific (populated when querying locations) */
  itemCount: number;
  /** Warranty */
  lifetimeWarranty: boolean;
  /**
   * Location is resolved: this entity's own location, else the nearest
   * location ancestor. Read-only — write via LocationID.
   */
  location?: EntitySummary | null;
  /**
   * LocationID is set only when this entity has its own location, nil when
   * inherited. Same name as the EntityUpdate field so a GET/PUT round trip
   * doesn't pin an inherited location or drop an explicit one.
   */
  locationId?: string | null;
  manufacturer: string;
  modelNumber: string;
  name: string;
  /** Extras */
  notes: string;
  /** Edges */
  parent?: EntitySummary | null;
  /** Purchase */
  purchaseDate: Date | string;
  purchaseFrom: string;
  purchasePrice: number;
  quantity: number;
  serialNumber: string;
  /** Sold */
  soldDate: Date | string;
  soldNotes: string;
  soldPrice: number;
  soldTo: string;
  syncChildEntityLocations: boolean;
  tags: TagSummary[];
  thumbnailId?: string | null;
  totalPrice: number;
  updatedAt: Date | string;
  warrantyDetails: string;
  warrantyExpires: Date | string;
}

export interface EntityPatch {
  entityTypeId?: string | null;
  id: string;
  locationId?: string | null;
  parentId?: string | null;
  quantity?: number | null;
  tagIds?: string[] | null;
}

export interface EntityPath {
  id: string;
  name: string;
  type: EntityPathType;
}

export interface EntitySummary {
  archived: boolean;
  /** @example "0" */
  assetId: string;
  createdAt: Date | string;
  description: string;
  entityType?: EntityTypeSummary | null;
  id: string;
  imageId?: string | null;
  insured: boolean;
  /** Container-specific (populated when querying locations) */
  itemCount: number;
  name: string;
  /** Edges */
  parent?: EntitySummary | null;
  purchasePrice: number;
  quantity: number;
  /** Sale details */
  soldDate: Date | string;
  tags: TagSummary[];
  thumbnailId?: string | null;
  updatedAt: Date | string;
}

export interface EntityTemplateCreate {
  /** @maxLength 1000 */
  defaultDescription?: string | null;
  defaultInsured: boolean;
  defaultLifetimeWarranty: boolean;
  /** Default location and tags */
  defaultLocationId?: string | null;
  /** @maxLength 255 */
  defaultManufacturer?: string | null;
  /** @maxLength 255 */
  defaultModelNumber?: string | null;
  /** @maxLength 255 */
  defaultName?: string | null;
  /** Default values for entities */
  defaultQuantity?: number | null;
  defaultTagIds?: string[] | null;
  /** @maxLength 1000 */
  defaultWarrantyDetails?: string | null;
  /** @maxLength 1000 */
  description: string;
  /** Custom fields */
  fields: TemplateField[];
  includePurchaseFields: boolean;
  includeSoldFields: boolean;
  /** Metadata flags */
  includeWarrantyFields: boolean;
  /**
   * @minLength 1
   * @maxLength 255
   */
  name: string;
  /** @maxLength 1000 */
  notes: string;
}

export interface EntityTemplateOut {
  createdAt: Date | string;
  defaultDescription: string;
  defaultInsured: boolean;
  defaultLifetimeWarranty: boolean;
  /** Default location and tags */
  defaultLocation: TemplateLocationSummary;
  defaultManufacturer: string;
  defaultModelNumber: string;
  defaultName: string;
  /** Default values for entities */
  defaultQuantity: number;
  defaultTags: TemplateTagSummary[];
  defaultWarrantyDetails: string;
  description: string;
  /** Custom fields */
  fields: TemplateField[];
  id: string;
  includePurchaseFields: boolean;
  includeSoldFields: boolean;
  /** Metadata flags */
  includeWarrantyFields: boolean;
  name: string;
  notes: string;
  updatedAt: Date | string;
}

export interface EntityTemplateSummary {
  createdAt: Date | string;
  description: string;
  id: string;
  name: string;
  updatedAt: Date | string;
}

export interface EntityTemplateUpdate {
  /** @maxLength 1000 */
  defaultDescription?: string | null;
  defaultInsured: boolean;
  defaultLifetimeWarranty: boolean;
  /** Default location and tags */
  defaultLocationId?: string | null;
  /** @maxLength 255 */
  defaultManufacturer?: string | null;
  /** @maxLength 255 */
  defaultModelNumber?: string | null;
  /** @maxLength 255 */
  defaultName?: string | null;
  /** Default values for entities */
  defaultQuantity?: number | null;
  defaultTagIds?: string[] | null;
  /** @maxLength 1000 */
  defaultWarrantyDetails?: string | null;
  /** @maxLength 1000 */
  description: string;
  /** Custom fields */
  fields: TemplateField[];
  id: string;
  includePurchaseFields: boolean;
  includeSoldFields: boolean;
  /** Metadata flags */
  includeWarrantyFields: boolean;
  /**
   * @minLength 1
   * @maxLength 255
   */
  name: string;
  /** @maxLength 1000 */
  notes: string;
}

export interface EntityTypeCreate {
  defaultTemplateId: string;
  icon: string;
  isLocation: boolean;
  name: string;
}

export interface EntityTypeSummary {
  createdAt: Date | string;
  defaultTemplate: EntityTemplateSummary;
  defaultTemplateId: string;
  description: string;
  icon: string;
  id: string;
  isLocation: boolean;
  name: string;
  updatedAt: Date | string;
}

export interface EntityTypeUpdate {
  defaultTemplateId: string;
  icon: string;
  id: string;
  isLocation: boolean;
  name: string;
}

export interface EntityUpdate {
  archived: boolean;
  assetId: string;
  /** @maxLength 1000 */
  description: string;
  entityTypeId: string;
  fields: EntityFieldData[];
  id: string;
  insured: boolean;
  /** Warranty */
  lifetimeWarranty: boolean;
  /**
   * Only needed when ParentID is another item and this lives elsewhere (#1688).
   * Otherwise ParentID carries the location. See resolveLocationOverride.
   */
  locationId?: string | null;
  manufacturer: string;
  modelNumber: string;
  /**
   * @minLength 1
   * @maxLength 255
   */
  name: string;
  /** Extras */
  notes: string;
  parentId?: string | null;
  /** Purchase */
  purchaseDate: Date | string;
  /** @maxLength 255 */
  purchaseFrom: string;
  purchasePrice?: number | null;
  quantity: number;
  /** Identifications */
  serialNumber: string;
  /** Sold */
  soldDate: Date | string;
  soldNotes: string;
  soldPrice?: number | null;
  /** @maxLength 255 */
  soldTo: string;
  syncChildEntityLocations: boolean;
  /** Edges */
  tagIds: string[];
  warrantyDetails: string;
  warrantyExpires: Date | string;
}

export interface ExportOut {
  artifactPath: string;
  createdAt: Date | string;
  /**
   * DestinationID is the backup destination holding the artifact; absent
   * for plain manual exports, which live in the primary storage.
   */
  destinationId?: string | null;
  error: string;
  groupId: string;
  id: string;
  /**
   * Kind is "export" for server-produced backup artifacts, "import" for
   * user-uploaded restore zips. The lifecycle fields below behave the
   * same for both.
   */
  kind: string;
  /**
   * Origin is "manual" for user-triggered exports and "scheduled" for
   * backups started by the scheduler.
   */
  origin: string;
  progress: number;
  sizeBytes: number;
  status: string;
  updatedAt: Date | string;
}

export interface Group {
  createdAt: Date | string;
  currency: string;
  id: string;
  name: string;
  updatedAt: Date | string;
}

export interface GroupInvitation {
  expiresAt: Date | string;
  group: Group;
  id: string;
  uses: number;
}

export interface GroupStatistics {
  totalItemPrice: number;
  totalItems: number;
  totalLocations: number;
  totalTags: number;
  totalUsers: number;
  totalWithWarranty: number;
}

export interface GroupUpdate {
  currency: string;
  name: string;
}

export interface ItemAttachment {
  createdAt: Date | string;
  id: string;
  mimeType: string;
  path: string;
  primary: boolean;
  thumbnail: EntAttachment;
  title: string;
  type: string;
  updatedAt: Date | string;
}

export interface ItemAttachmentUpdate {
  primary: boolean;
  title: string;
  type: string;
}

export interface MaintenanceEntry {
  completedDate: Date | string;
  /** @example "0" */
  cost: string;
  description: string;
  id: string;
  name: string;
  scheduledDate: Date | string;
}

export interface MaintenanceEntryCreate {
  completedDate: Date | string;
  /** @example "0" */
  cost: string;
  description: string;
  name: string;
  scheduledDate: Date | string;
}

export interface MaintenanceEntryUpdate {
  completedDate: Date | string;
  /** @example "0" */
  cost: string;
  description: string;
  name: string;
  scheduledDate: Date | string;
}

export interface MaintenanceEntryWithDetails {
  completedDate: Date | string;
  /** @example "0" */
  cost: string;
  description: string;
  id: string;
  itemID: string;
  itemName: string;
  name: string;
  scheduledDate: Date | string;
}

export interface NotifierCreate {
  isActive: boolean;
  /**
   * @minLength 1
   * @maxLength 255
   */
  name: string;
  url: string;
}

export interface NotifierOut {
  createdAt: Date | string;
  groupId: string;
  id: string;
  isActive: boolean;
  name: string;
  updatedAt: Date | string;
  url: string;
  userId: string;
}

export interface NotifierUpdate {
  isActive: boolean;
  /**
   * @minLength 1
   * @maxLength 255
   */
  name: string;
  url?: string | null;
}

export interface PaginationResultEntitySummary {
  items: EntitySummary[];
  page: number;
  pageSize: number;
  total: number;
}

export interface TagCreate {
  color: string;
  /** @maxLength 1000 */
  description: string;
  /** @maxLength 255 */
  icon: string;
  /**
   * @minLength 1
   * @maxLength 255
   */
  name: string;
  parentId?: string | null;
}

export interface TagOut {
  children: TagSummary[];
  color: string;
  createdAt: Date | string;
  description: string;
  icon: string;
  id: string;
  name: string;
  parent?: TagSummary | null;
  parentId?: string | null;
  updatedAt: Date | string;
}

export interface TagSummary {
  color: string;
  createdAt: Date | string;
  description: string;
  icon: string;
  id: string;
  name: string;
  parentId?: string | null;
  updatedAt: Date | string;
}

export interface TagUpdate {
  color: string;
  /** @maxLength 1000 */
  description: string;
  /** @maxLength 255 */
  icon: string;
  id: string;
  /**
   * @minLength 1
   * @maxLength 255
   */
  name: string;
  parentId?: string | null;
}

export interface TemplateField {
  booleanValue: boolean;
  id: string;
  name: string;
  numberValue: number;
  textValue: string;
  timeValue: string;
  type: string;
}

export interface TemplateLocationSummary {
  id: string;
  name: string;
}

export interface TemplateTagSummary {
  id: string;
  name: string;
}

export interface TotalsByOrganizer {
  id: string;
  name: string;
  total: number;
}

export interface TreeItem {
  children: TreeItem[];
  id: string;
  name: string;
  type: string;
}

export interface UserOut {
  defaultGroupId: string;
  email: string;
  groupIds: string[];
  id: string;
  isSuperuser: boolean;
  name: string;
  oidcIssuer: string;
  oidcSubject: string;
}

export interface UserSummary {
  email: string;
  id: string;
  name: string;
}

export interface UserUpdate {
  email: string;
  name: string;
}

export interface ValueOverTime {
  end: string;
  entries: ValueOverTimeEntry[];
  start: string;
  valueAtEnd: number;
  valueAtStart: number;
}

export interface ValueOverTimeEntry {
  date: Date | string;
  name: string;
  value: number;
}

export interface Latest {
  date: Date | string;
  version: string;
}

export interface OIDCSuggestion {
  destType: string;
  email: string;
  provider: string;
}

export interface TestResult {
  /**
   * HostKey is the SSH host key fingerprint the server presented when it
   * was missing or did not match, so the UI can offer to trust it.
   */
  hostKey: string;
  latencyMs: number;
  message: string;
  ok: boolean;
}

export interface UserRegistration {
  email: string;
  name: string;
  password: string;
  token: string;
}

export interface APISummary {
  allowRegistration: boolean;
  build: Build;
  demo: boolean;
  health: boolean;
  labelPrinting: boolean;
  latest: Latest;
  message: string;
  oidc: OIDCStatus;
  telemetry: TelemetryStatus;
  title: string;
  versions: string[];
}

export interface ActionAmountResult {
  completed: number;
}

export interface BackupOAuthStartIn {
  provider: "google" | "microsoft" | "dropbox";
  /**
   * UseLoginAccount asks the provider to preselect the account the user
   * signed in to Homebox with. Honoured only when that login came from the
   * same provider.
   */
  useLoginAccount: boolean;
}

export interface BackupOAuthStartOut {
  authUrl: string;
}

export interface BackupOptions {
  allowCustomEndpoints: boolean;
  enabled: boolean;
  localEnabled: boolean;
  /** OAuthProviders lists the configured cloud drives: google, microsoft, dropbox. */
  oauthProviders: string[];
  /**
   * OIDCSuggestion offers the cloud drive matching the identity provider the
   * current user signed in with, when there is one.
   */
  oidcSuggestion?: OIDCSuggestion | null;
  remoteEnabled: boolean;
}

export interface Build {
  buildTime: string;
  commit: string;
  version: string;
}

export interface ChangePassword {
  current: string;
  new: string;
}

export interface CreateRequest {
  name: string;
}

export interface EntityTemplateCreateItemRequest {
  /** @maxLength 1000 */
  description: string;
  /**
   * EntityTypeID is the entity type selected by the user. When set it takes
   * precedence; when empty the repository falls back to the group's default.
   */
  entityTypeId: string;
  locationId?: string | null;
  /**
   * @minLength 1
   * @maxLength 255
   */
  name: string;
  parentId: string;
  quantity: number;
  tagIds: string[];
}

export interface ForgotPasswordRequest {
  /** @example "user@example.com" */
  email: string;
}

export interface GroupAcceptInvitationResponse {
  id: string;
  name: string;
}

export interface GroupInvitation {
  expiresAt: Date | string;
  id: string;
  token: string;
  uses: number;
}

export interface GroupInvitationCreate {
  expiresAt: Date | string;
  /**
   * @min 1
   * @max 100
   */
  uses: number;
}

export interface LoginForm {
  /** @example "admin" */
  password: string;
  stayLoggedIn: boolean;
  /** @example "admin@admin.com" */
  username: string;
}

export interface OIDCStatus {
  allowLocal: boolean;
  autoRedirect: boolean;
  buttonText: string;
  enabled: boolean;
}

export interface ResetPasswordRequest {
  /** @minLength 6 */
  password: string;
  /** @minLength 20 */
  token: string;
}

export interface ResultsRepoBackupDestinationOut {
  items: BackupDestinationOut[];
}

export interface ResultsRepoExportOut {
  items: ExportOut[];
}

export interface TelemetryStatus {
  enabled: boolean;
}

export interface TokenResponse {
  attachmentToken: string;
  expiresAt: Date | string;
  token: string;
}

export interface WipeInventoryOptions {
  wipeLocations: boolean;
  wipeMaintenance: boolean;
  wipeTags: boolean;
}

export interface Wrapped {
  item: any;
}

export interface ExternalAttachmentRequest {
  attachment_type: string;
  external_id: string;
  source_type: string;
  title: string;
}

export interface ValidateErrorResponse {
  error: string;
  fields: string;
}
