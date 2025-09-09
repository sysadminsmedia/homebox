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

export enum UserRole {
  DefaultRole = "user",
  RoleUser = "user",
  RoleOwner = "owner",
}

export enum MaintenanceFilterStatus {
  MaintenanceFilterStatusScheduled = "scheduled",
  MaintenanceFilterStatusCompleted = "completed",
  MaintenanceFilterStatusBoth = "both",
}

export enum ItemType {
  ItemTypeLocation = "location",
  ItemTypeItem = "item",
}

export enum ItemfieldType {
  TypeText = "text",
  TypeNumber = "number",
  TypeBoolean = "boolean",
  TypeTime = "time",
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
  local: string;
  name: string;
  symbol: string;
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
  /** Item holds the value of the item edge. */
  item: EntItem;
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
  /** InvitationTokens holds the value of the invitation_tokens edge. */
  invitation_tokens: EntGroupInvitationToken[];
  /** Items holds the value of the items edge. */
  items: EntItem[];
  /** Labels holds the value of the labels edge. */
  labels: EntLabel[];
  /** Locations holds the value of the locations edge. */
  locations: EntLocation[];
  /** Notifiers holds the value of the notifiers edge. */
  notifiers: EntNotifier[];
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

export interface EntItem {
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
   * The values are being populated by the ItemQuery when eager-loading is set.
   */
  edges: EntItemEdges;
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
  /** PurchaseFrom holds the value of the "purchase_from" field. */
  purchase_from: string;
  /** PurchasePrice holds the value of the "purchase_price" field. */
  purchase_price: number;
  /** PurchaseTime holds the value of the "purchase_time" field. */
  purchase_time: string;
  /** Quantity holds the value of the "quantity" field. */
  quantity: number;
  /** SerialNumber holds the value of the "serial_number" field. */
  serial_number: string;
  /** SoldNotes holds the value of the "sold_notes" field. */
  sold_notes: string;
  /** SoldPrice holds the value of the "sold_price" field. */
  sold_price: number;
  /** SoldTime holds the value of the "sold_time" field. */
  sold_time: string;
  /** SoldTo holds the value of the "sold_to" field. */
  sold_to: string;
  /** SyncChildItemsLocations holds the value of the "sync_child_items_locations" field. */
  sync_child_items_locations: boolean;
  /** UpdatedAt holds the value of the "updated_at" field. */
  updated_at: string;
  /** WarrantyDetails holds the value of the "warranty_details" field. */
  warranty_details: string;
  /** WarrantyExpires holds the value of the "warranty_expires" field. */
  warranty_expires: string;
}

export interface EntItemEdges {
  /** Attachments holds the value of the attachments edge. */
  attachments: EntAttachment[];
  /** Children holds the value of the children edge. */
  children: EntItem[];
  /** Fields holds the value of the fields edge. */
  fields: EntItemField[];
  /** Group holds the value of the group edge. */
  group: EntGroup;
  /** Label holds the value of the label edge. */
  label: EntLabel[];
  /** Location holds the value of the location edge. */
  location: EntLocation;
  /** MaintenanceEntries holds the value of the maintenance_entries edge. */
  maintenance_entries: EntMaintenanceEntry[];
  /** Parent holds the value of the parent edge. */
  parent: EntItem;
}

export interface EntItemField {
  /** BooleanValue holds the value of the "boolean_value" field. */
  boolean_value: boolean;
  /** CreatedAt holds the value of the "created_at" field. */
  created_at: string;
  /** Description holds the value of the "description" field. */
  description: string;
  /**
   * Edges holds the relations/edges for other nodes in the graph.
   * The values are being populated by the ItemFieldQuery when eager-loading is set.
   */
  edges: EntItemFieldEdges;
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
  type: ItemfieldType;
  /** UpdatedAt holds the value of the "updated_at" field. */
  updated_at: string;
}

export interface EntItemFieldEdges {
  /** Item holds the value of the item edge. */
  item: EntItem;
}

export interface EntLabel {
  /** Color holds the value of the "color" field. */
  color: string;
  /** CreatedAt holds the value of the "created_at" field. */
  created_at: string;
  /** Description holds the value of the "description" field. */
  description: string;
  /**
   * Edges holds the relations/edges for other nodes in the graph.
   * The values are being populated by the LabelQuery when eager-loading is set.
   */
  edges: EntLabelEdges;
  /** ID of the ent. */
  id: string;
  /** Name holds the value of the "name" field. */
  name: string;
  /** UpdatedAt holds the value of the "updated_at" field. */
  updated_at: string;
}

export interface EntLabelEdges {
  /** Group holds the value of the group edge. */
  group: EntGroup;
  /** Items holds the value of the items edge. */
  items: EntItem[];
}

export interface EntLocation {
  /** CreatedAt holds the value of the "created_at" field. */
  created_at: string;
  /** Description holds the value of the "description" field. */
  description: string;
  /**
   * Edges holds the relations/edges for other nodes in the graph.
   * The values are being populated by the LocationQuery when eager-loading is set.
   */
  edges: EntLocationEdges;
  /** ID of the ent. */
  id: string;
  /** Name holds the value of the "name" field. */
  name: string;
  /** UpdatedAt holds the value of the "updated_at" field. */
  updated_at: string;
}

export interface EntLocationEdges {
  /** Children holds the value of the children edge. */
  children: EntLocation[];
  /** Group holds the value of the group edge. */
  group: EntGroup;
  /** Items holds the value of the items edge. */
  items: EntItem[];
  /** Parent holds the value of the parent edge. */
  parent: EntLocation;
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
  /** ID of the ent. */
  id: string;
  /** ItemID holds the value of the "item_id" field. */
  item_id: string;
  /** Name holds the value of the "name" field. */
  name: string;
  /** ScheduledDate holds the value of the "scheduled_date" field. */
  scheduled_date: Date | string;
  /** UpdatedAt holds the value of the "updated_at" field. */
  updated_at: string;
}

export interface EntMaintenanceEntryEdges {
  /** Item holds the value of the item edge. */
  item: EntItem;
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

export interface EntUser {
  /** ActivatedOn holds the value of the "activated_on" field. */
  activated_on: string;
  /** CreatedAt holds the value of the "created_at" field. */
  created_at: string;
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
  /** Role holds the value of the "role" field. */
  role: UserRole;
  /** Settings holds the value of the "settings" field. */
  settings: SchemaUserSettings;
  /** Superuser holds the value of the "superuser" field. */
  superuser: boolean;
  /** UpdatedAt holds the value of the "updated_at" field. */
  updated_at: string;
}

export interface EntUserEdges {
  /** AuthTokens holds the value of the auth_tokens edge. */
  auth_tokens: EntAuthTokens[];
  /** Group holds the value of the group edge. */
  group: EntGroup;
  /** Notifiers holds the value of the notifiers edge. */
  notifiers: EntNotifier[];
}

export interface BarcodeProduct {
  barcode: string;
  imageBase64: string;
  imageURL: string;
  item: ItemCreate;
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

export interface Group {
  createdAt: Date | string;
  currency: string;
  id: string;
  name: string;
  updatedAt: Date | string;
}

export interface GroupStatistics {
  totalItemPrice: number;
  totalItems: number;
  totalLabels: number;
  totalLocations: number;
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

export interface ItemCreate {
  /** @maxLength 1000 */
  description: string;
  labelIds: string[];
  /** Edges */
  locationId: string;
  /**
   * @minLength 1
   * @maxLength 255
   */
  name: string;
  parentId?: string | null;
  quantity: number;
}

export interface ItemField {
  booleanValue: boolean;
  id: string;
  name: string;
  numberValue: number;
  textValue: string;
  type: string;
}

export interface ItemOut {
  archived: boolean;
  /** @example "0" */
  assetId: string;
  attachments: ItemAttachment[];
  createdAt: Date | string;
  description: string;
  fields: ItemField[];
  id: string;
  imageId?: string | null;
  insured: boolean;
  labels: LabelSummary[];
  /** Warranty */
  lifetimeWarranty: boolean;
  /** Edges */
  location?: LocationSummary | null;
  manufacturer: string;
  modelNumber: string;
  name: string;
  /** Extras */
  notes: string;
  parent?: ItemSummary | null;
  purchaseFrom: string;
  purchasePrice: number;
  /** Purchase */
  purchaseTime: Date | string;
  quantity: number;
  serialNumber: string;
  soldNotes: string;
  soldPrice: number;
  /** Sold */
  soldTime: Date | string;
  soldTo: string;
  syncChildItemsLocations: boolean;
  thumbnailId?: string | null;
  updatedAt: Date | string;
  warrantyDetails: string;
  warrantyExpires: Date | string;
}

export interface ItemPatch {
  id: string;
  quantity?: number | null;
}

export interface ItemPath {
  id: string;
  name: string;
  type: ItemType;
}

export interface ItemSummary {
  archived: boolean;
  /** @example "0" */
  assetId: string;
  createdAt: Date | string;
  description: string;
  id: string;
  imageId?: string | null;
  insured: boolean;
  labels: LabelSummary[];
  /** Edges */
  location?: LocationSummary | null;
  name: string;
  purchasePrice: number;
  quantity: number;
  /** Sale details */
  soldTime: Date | string;
  thumbnailId?: string | null;
  updatedAt: Date | string;
}

export interface ItemUpdate {
  archived: boolean;
  assetId: string;
  /** @maxLength 1000 */
  description: string;
  fields: ItemField[];
  id: string;
  insured: boolean;
  labelIds: string[];
  /** Warranty */
  lifetimeWarranty: boolean;
  /** Edges */
  locationId: string;
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
  /** @maxLength 255 */
  purchaseFrom: string;
  purchasePrice?: number | null;
  /** Purchase */
  purchaseTime: Date | string;
  quantity: number;
  /** Identifications */
  serialNumber: string;
  soldNotes: string;
  soldPrice?: number | null;
  /** Sold */
  soldTime: Date | string;
  /** @maxLength 255 */
  soldTo: string;
  syncChildItemsLocations: boolean;
  warrantyDetails: string;
  warrantyExpires: Date | string;
}

export interface LabelCreate {
  color: string;
  /** @maxLength 1000 */
  description: string;
  /**
   * @minLength 1
   * @maxLength 255
   */
  name: string;
}

export interface LabelOut {
  color: string;
  createdAt: Date | string;
  description: string;
  id: string;
  name: string;
  updatedAt: Date | string;
}

export interface LabelSummary {
  color: string;
  createdAt: Date | string;
  description: string;
  id: string;
  name: string;
  updatedAt: Date | string;
}

export interface LocationCreate {
  description: string;
  name: string;
  parentId?: string | null;
}

export interface LocationOut {
  children: LocationSummary[];
  createdAt: Date | string;
  description: string;
  id: string;
  name: string;
  parent: LocationSummary;
  totalPrice: number;
  updatedAt: Date | string;
}

export interface LocationOutCount {
  createdAt: Date | string;
  description: string;
  id: string;
  itemCount: number;
  name: string;
  updatedAt: Date | string;
}

export interface LocationSummary {
  createdAt: Date | string;
  description: string;
  id: string;
  name: string;
  updatedAt: Date | string;
}

export interface LocationUpdate {
  description: string;
  id: string;
  name: string;
  parentId?: string | null;
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

export interface PaginationResultItemSummary {
  items: ItemSummary[];
  page: number;
  pageSize: number;
  total: number;
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
  email: string;
  groupId: string;
  groupName: string;
  id: string;
  isOwner: boolean;
  isSuperuser: boolean;
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

export interface SchemaDuplicateSettings {
  copyAttachments: boolean;
  copyCustomFields: boolean;
  copyMaintenance: boolean;
  copyPrefixOverride: string;
}

export interface SchemaUserSettings {
  displayLegacyHeader: boolean;
  duplicateSettings: SchemaDuplicateSettings;
  editorAdvancedView: boolean;
  itemDisplayView: string;
  itemsPerPage: number;
  itemsPerTablePage: number;
  language: string;
  locale: string;
  overrideFormatLocale: string;
  showDetails: boolean;
  showEmpty: boolean;
  theme: string;
}

export interface Latest {
  date: Date | string;
  version: string;
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
  title: string;
  versions: string[];
}

export interface ActionAmountResult {
  completed: number;
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

export interface GroupInvitation {
  expiresAt: Date | string;
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

export interface ItemAttachmentToken {
  token: string;
}

export interface LoginForm {
  /** @example "admin" */
  password: string;
  stayLoggedIn: boolean;
  /** @example "admin@admin.com" */
  username: string;
}

export interface TokenResponse {
  attachmentToken: string;
  expiresAt: Date | string;
  token: string;
}

export interface Wrapped {
  item: any;
}

export interface ValidateErrorResponse {
  error: string;
  fields: string;
}
