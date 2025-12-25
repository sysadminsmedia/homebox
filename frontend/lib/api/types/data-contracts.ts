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

export enum TemplatefieldType {
  TypeText = "text",
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

export enum PrinterStatus {
  DefaultStatus = "unknown",
  StatusOnline = "online",
  StatusOffline = "offline",
  StatusUnknown = "unknown",
}

export enum LabelmakerContentType {
  /** Can encode any text (URLs, names, etc.) */
  ContentTypeAny = "any",
  /** Letters, numbers, limited symbols */
  ContentTypeAlphanumeric = "alphanumeric",
  /** Digits only */
  ContentTypeNumeric = "numeric",
}

export enum LabelmakerBarcodeFormat {
  BarcodeQR = "qr",
  BarcodeCode128 = "code128",
  BarcodeCode39 = "code39",
  BarcodeDataMatrix = "datamatrix",
  BarcodeEAN13 = "ean13",
  BarcodeEAN8 = "ean8",
  BarcodeUPCA = "upca",
  BarcodeUPCE = "upce",
}

export enum ItemfieldType {
  TypeText = "text",
  TypeNumber = "number",
  TypeBoolean = "boolean",
  TypeTime = "time",
}

export enum GithubComSysadminsmediaHomeboxBackendInternalDataEntPrinterPrinterType {
  DefaultPrinterType = "ipp",
  PrinterTypeIpp = "ipp",
  PrinterTypeCups = "cups",
  PrinterTypeBrotherRaster = "brother_raster",
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
  /** ItemTemplates holds the value of the item_templates edge. */
  item_templates: EntItemTemplate[];
  /** Items holds the value of the items edge. */
  items: EntItem[];
  /** LabelTemplates holds the value of the label_templates edge. */
  label_templates: EntLabelTemplate[];
  /** Labels holds the value of the labels edge. */
  labels: EntLabel[];
  /** Locations holds the value of the locations edge. */
  locations: EntLocation[];
  /** Notifiers holds the value of the notifiers edge. */
  notifiers: EntNotifier[];
  /** Printers holds the value of the printers edge. */
  printers: EntPrinter[];
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

export interface EntItemTemplate {
  /** CreatedAt holds the value of the "created_at" field. */
  created_at: string;
  /** Default description for items created from this template */
  default_description: string;
  /** DefaultInsured holds the value of the "default_insured" field. */
  default_insured: boolean;
  /** Default label IDs for items created from this template */
  default_label_ids: string[];
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
  /** DefaultWarrantyDetails holds the value of the "default_warranty_details" field. */
  default_warranty_details: string;
  /** Description holds the value of the "description" field. */
  description: string;
  /**
   * Edges holds the relations/edges for other nodes in the graph.
   * The values are being populated by the ItemTemplateQuery when eager-loading is set.
   */
  edges: EntItemTemplateEdges;
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

export interface EntItemTemplateEdges {
  /** Fields holds the value of the fields edge. */
  fields: EntTemplateField[];
  /** Group holds the value of the group edge. */
  group: EntGroup;
  /** Location holds the value of the location edge. */
  location: EntLocation;
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

export interface EntLabelTemplate {
  /** Fabric.js compatible canvas JSON */
  canvas_data: Record<string, any>;
  /** CreatedAt holds the value of the "created_at" field. */
  created_at: string;
  /** Description holds the value of the "description" field. */
  description: string;
  /** Output DPI for rendering */
  dpi: number;
  /**
   * Edges holds the relations/edges for other nodes in the graph.
   * The values are being populated by the LabelTemplateQuery when eager-loading is set.
   */
  edges: EntLabelTemplateEdges;
  /** Label height in mm */
  height: number;
  /** ID of the ent. */
  id: string;
  /** Whether template is shared with group */
  is_shared: boolean;
  /** Brother media type like 'DK-22251' for direct printing */
  media_type: string;
  /** Name holds the value of the "name" field. */
  name: string;
  /** Output format: png, pdf */
  output_format: string;
  /** User who created this template */
  owner_id: string;
  /** Preset size key like 'brother_dk2205' */
  preset: string;
  /** UpdatedAt holds the value of the "updated_at" field. */
  updated_at: string;
  /** Label width in mm */
  width: number;
}

export interface EntLabelTemplateEdges {
  /** Group holds the value of the group edge. */
  group: EntGroup;
  /** Owner holds the value of the owner edge. */
  owner: EntUser;
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

export interface EntPrinter {
  /** IPP URI (ipp://host:port/path) or CUPS printer name */
  address: string;
  /** CreatedAt holds the value of the "created_at" field. */
  created_at: string;
  /** Description holds the value of the "description" field. */
  description: string;
  /** Printer DPI for optimal rendering */
  dpi: number;
  /**
   * Edges holds the relations/edges for other nodes in the graph.
   * The values are being populated by the PrinterQuery when eager-loading is set.
   */
  edges: EntPrinterEdges;
  /** ID of the ent. */
  id: string;
  /** Whether this is the default label printer */
  is_default: boolean;
  /** Expected label height in mm for validation */
  label_height_mm: number;
  /** Expected label width in mm for validation */
  label_width_mm: number;
  /** When status was last verified */
  last_status_check: string;
  /** Media type identifier for IPP (e.g., 'labels') */
  media_type: string;
  /** Name holds the value of the "name" field. */
  name: string;
  /** Type of printer connection */
  printer_type: GithubComSysadminsmediaHomeboxBackendInternalDataEntPrinterPrinterType;
  /** Cached printer status */
  status: PrinterStatus;
  /** UpdatedAt holds the value of the "updated_at" field. */
  updated_at: string;
}

export interface EntPrinterEdges {
  /** Group holds the value of the group edge. */
  group: EntGroup;
}

export interface EntTemplateField {
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
  /** TextValue holds the value of the "text_value" field. */
  text_value: string;
  /** Type holds the value of the "type" field. */
  type: TemplatefieldType;
  /** UpdatedAt holds the value of the "updated_at" field. */
  updated_at: string;
}

export interface EntTemplateFieldEdges {
  /** ItemTemplate holds the value of the item_template edge. */
  item_template: EntItemTemplate;
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
  /** OidcIssuer holds the value of the "oidc_issuer" field. */
  oidc_issuer: string;
  /** OidcSubject holds the value of the "oidc_subject" field. */
  oidc_subject: string;
  /** Role holds the value of the "role" field. */
  role: UserRole;
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
  /** LabelTemplates holds the value of the label_templates edge. */
  label_templates: EntLabelTemplate[];
  /** Notifiers holds the value of the notifiers edge. */
  notifiers: EntNotifier[];
}

export interface LabelmakerBarcodeFormatInfo {
  /** What kind of content this format supports */
  contentType: LabelmakerContentType;
  description: string;
  format: LabelmakerBarcodeFormat;
  is2D: boolean;
  /** 0 means variable/unlimited */
  maxLength: number;
  name: string;
}

export interface LabelmakerLabelPreset {
  brand: string;
  /** Whether this is continuous tape */
  continuous: boolean;
  description: string;
  /** Height in mm */
  height: number;
  key: string;
  name: string;
  /**
   * Sheet layout information (for Avery-style sheet labels)
   * If SheetLayout is set, labels are arranged on a printable sheet
   */
  sheetLayout: LabelmakerSheetLayout;
  /** Whether this supports two-color printing (black/red) */
  twoColor: boolean;
  /** Width in mm */
  width: number;
}

export interface LabelmakerSheetLayout {
  /** Number of labels across */
  columns: number;
  /** Horizontal gap between labels in mm */
  gutterH: number;
  /** Vertical gap between labels in mm */
  gutterV: number;
  /** Left margin in mm */
  marginLeft: number;
  /** Top margin in mm */
  marginTop: number;
  /** Sheet height in mm (e.g., 279.4 for Letter, 297 for A4) */
  pageHeight: number;
  /** Sheet width in mm (e.g., 215.9 for Letter, 210 for A4) */
  pageWidth: number;
  /** Number of labels down */
  rows: number;
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
  labelIds?: string[] | null;
  locationId?: string | null;
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

export interface ItemTemplateCreate {
  /** @maxLength 1000 */
  defaultDescription?: string | null;
  defaultInsured: boolean;
  defaultLabelIds?: string[] | null;
  defaultLifetimeWarranty: boolean;
  /** Default location and labels */
  defaultLocationId?: string | null;
  /** @maxLength 255 */
  defaultManufacturer?: string | null;
  /** @maxLength 255 */
  defaultModelNumber?: string | null;
  /** @maxLength 255 */
  defaultName?: string | null;
  /** Default values for items */
  defaultQuantity?: number | null;
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

export interface ItemTemplateOut {
  createdAt: Date | string;
  defaultDescription: string;
  defaultInsured: boolean;
  defaultLabels: TemplateLabelSummary[];
  defaultLifetimeWarranty: boolean;
  /** Default location and labels */
  defaultLocation: TemplateLocationSummary;
  defaultManufacturer: string;
  defaultModelNumber: string;
  defaultName: string;
  /** Default values for items */
  defaultQuantity: number;
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

export interface ItemTemplateSummary {
  createdAt: Date | string;
  description: string;
  id: string;
  name: string;
  updatedAt: Date | string;
}

export interface ItemTemplateUpdate {
  /** @maxLength 1000 */
  defaultDescription?: string | null;
  defaultInsured: boolean;
  defaultLabelIds?: string[] | null;
  defaultLifetimeWarranty: boolean;
  /** Default location and labels */
  defaultLocationId?: string | null;
  /** @maxLength 255 */
  defaultManufacturer?: string | null;
  /** @maxLength 255 */
  defaultModelNumber?: string | null;
  /** @maxLength 255 */
  defaultName?: string | null;
  /** Default values for items */
  defaultQuantity?: number | null;
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

export interface LabelTemplateCreate {
  canvasData: Record<string, any>;
  /** @maxLength 1000 */
  description: string;
  /**
   * @min 72
   * @max 600
   */
  dpi: number;
  height: number;
  isShared: boolean;
  mediaType?: string | null;
  /**
   * @minLength 1
   * @maxLength 255
   */
  name: string;
  outputFormat: "png" | "pdf";
  preset?: string | null;
  width: number;
}

export interface LabelTemplateOut {
  canvasData: Record<string, any>;
  createdAt: Date | string;
  description: string;
  dpi: number;
  height: number;
  id: string;
  isOwner: boolean;
  isShared: boolean;
  mediaType: string;
  name: string;
  outputFormat: string;
  ownerId: string;
  preset: string;
  updatedAt: Date | string;
  width: number;
}

export interface LabelTemplateSummary {
  createdAt: Date | string;
  description: string;
  height: number;
  id: string;
  isOwner: boolean;
  isShared: boolean;
  name: string;
  preset: string;
  updatedAt: Date | string;
  width: number;
}

export interface LabelTemplateUpdate {
  canvasData: Record<string, any>;
  /** @maxLength 1000 */
  description: string;
  /**
   * @min 72
   * @max 600
   */
  dpi: number;
  height: number;
  id: string;
  isShared: boolean;
  mediaType?: string | null;
  /**
   * @minLength 1
   * @maxLength 255
   */
  name: string;
  outputFormat: "png" | "pdf";
  preset?: string | null;
  width: number;
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

export interface PrinterCreate {
  /**
   * @minLength 1
   * @maxLength 512
   */
  address: string;
  /** @maxLength 1000 */
  description: string;
  /**
   * @min 72
   * @max 1200
   */
  dpi: number;
  isDefault: boolean;
  labelHeightMm?: number | null;
  labelWidthMm?: number | null;
  mediaType?: string | null;
  /**
   * @minLength 1
   * @maxLength 255
   */
  name: string;
  printerType: "ipp" | "cups" | "brother_raster";
}

export interface PrinterOut {
  address: string;
  createdAt: Date | string;
  description: string;
  dpi: number;
  id: string;
  isDefault: boolean;
  labelHeightMm: number;
  labelWidthMm: number;
  lastStatusCheck: string;
  mediaType: string;
  name: string;
  printerType: string;
  status: string;
  updatedAt: Date | string;
}

export interface PrinterSummary {
  address: string;
  createdAt: Date | string;
  description: string;
  dpi: number;
  id: string;
  isDefault: boolean;
  labelHeightMm: number;
  labelWidthMm: number;
  name: string;
  printerType: string;
  status: string;
  updatedAt: Date | string;
}

export interface PrinterUpdate {
  /**
   * @minLength 1
   * @maxLength 512
   */
  address: string;
  /** @maxLength 1000 */
  description: string;
  /**
   * @min 72
   * @max 1200
   */
  dpi: number;
  id: string;
  isDefault: boolean;
  labelHeightMm?: number | null;
  labelWidthMm?: number | null;
  mediaType?: string | null;
  /**
   * @minLength 1
   * @maxLength 255
   */
  name: string;
  printerType: "ipp" | "cups" | "brother_raster";
}

export interface TemplateField {
  id: string;
  name: string;
  textValue: string;
  type: string;
}

export interface TemplateLabelSummary {
  id: string;
  name: string;
}

export interface TemplateLocationSummary {
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
  email: string;
  groupId: string;
  groupName: string;
  id: string;
  isOwner: boolean;
  isSuperuser: boolean;
  name: string;
  oidcIssuer: string;
  oidcSubject: string;
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
  title: string;
  versions: string[];
}

export interface ActionAmountResult {
  completed: number;
}

export interface BrotherMediaInfo {
  id: string;
  isContinuous: boolean;
  /** 0 for continuous */
  lengthMm: number;
  name: string;
  twoColor: boolean;
  widthMm: number;
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

export interface ItemTemplateCreateItemRequest {
  /** @maxLength 1000 */
  description: string;
  labelIds: string[];
  locationId: string;
  /**
   * @minLength 1
   * @maxLength 255
   */
  name: string;
  quantity: number;
}

export interface LabelPrintItem {
  id: string;
  /** Number of copies for this item */
  quantity: number;
}

export interface LabelPrintLocation {
  id: string;
  /** Number of copies for this location */
  quantity: number;
}

export interface LabelTemplatePrintLocationsRequest {
  /** Default copies per label */
  copies: number;
  /** Simple list (1 copy each) */
  locationIds: string[];
  /** Locations with individual quantities */
  locations: LabelPrintLocation[];
  /** If nil, uses default printer */
  printerId: string;
}

export interface LabelTemplatePrintRequest {
  /** Default copies per label (used if item.quantity is 0) */
  copies: number;
  /** Simple list (1 copy each) - for backward compatibility */
  itemIds: string[];
  /** Items with individual quantities */
  items: LabelPrintItem[];
  /** If nil, uses default printer */
  printerId: string;
}

export interface LabelTemplatePrintResponse {
  jobId: number;
  labelCount: number;
  message: string;
  printerName: string;
  success: boolean;
}

export interface LabelTemplateRenderLocationsRequest {
  /** "png" or "pdf", defaults to "png" */
  format: string;
  /** @minItems 1 */
  locationIds: string[];
  /** "Letter", "A4", or "Custom" for PDF */
  pageSize: string;
  /** Draw light borders around labels for cutting */
  showCutGuides: boolean;
}

export interface LabelTemplateRenderRequest {
  /** Optional: canvas data for live preview (overrides saved template) */
  canvasData: string;
  /** "png" or "pdf", defaults to "png" */
  format: string;
  /** @minItems 1 */
  itemIds: string[];
  /** "Letter", "A4", or "Custom" for PDF */
  pageSize: string;
  /** Draw light borders around labels for cutting */
  showCutGuides: boolean;
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

export interface PrinterStatusResponse {
  mediaReady: string[];
  message: string;
  status: string;
  supportsIpp: boolean;
}

export interface PrinterTestRequest {
  message: string;
}

export interface PrinterTestResponse {
  jobId: number;
  message: string;
  success: boolean;
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
