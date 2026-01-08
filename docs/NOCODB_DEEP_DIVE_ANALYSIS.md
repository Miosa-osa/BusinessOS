# NocoDB Deep Dive Analysis
## Complete Reverse Engineering Documentation for BusinessOS Tables Module

**Analysis Date:** January 2026
**NocoDB Version:** 0.265.1
**Purpose:** Reverse engineer NocoDB architecture to inspire BusinessOS Tables module implementation

---

## Table of Contents
1. [Project Structure](#1-project-structure)
2. [Data Modeling & Database Schema](#2-data-modeling--database-schema)
3. [Core Features List](#3-core-features-list)
4. [Column Types](#4-column-types)
5. [Frontend Architecture](#5-frontend-architecture)
6. [Backend Architecture](#6-backend-architecture)
7. [View Types](#7-view-types)
8. [Key Files Reference](#8-key-files-reference)
9. [Implementation Insights](#9-implementation-insights)

---

## 1. PROJECT STRUCTURE

### Monorepo Organization
```
nocodb-develop/
├── packages/
│   ├── nocodb/              # Backend (NestJS + TypeScript)
│   ├── nocodb-sdk/          # Shared SDK/Types
│   ├── nocodb-sdk-v2/       # SDK v2
│   ├── nc-gui/              # Frontend (Nuxt 3 + Vue 3)
│   ├── nc-lib-gui/          # GUI library
│   ├── nc-mail-assets/      # Email templates
│   ├── noco-integrations/   # Integration system
│   ├── nc-secret-mgr/       # Secret management
│   └── nc-integration-scaffolder/
├── docker-compose/
├── scripts/
└── tests/
```

### Backend Structure (`packages/nocodb/`)
```
src/
├── controllers/          # NestJS API controllers
├── services/            # Business logic layer
├── models/              # Data models (ORM-like)
├── db/                  # Database layer
│   └── BaseModelSqlv2.ts   # Core CRUD abstraction
├── meta/                # Metadata management
│   └── migrations/      # Schema migrations
├── cache/               # Redis caching layer
├── utils/               # Utilities
├── helpers/             # Helper functions
├── schema/              # Schema definitions
└── interface/           # TypeScript interfaces
```

### Frontend Structure (`packages/nc-gui/`)
```
nc-gui/
├── components/
│   ├── smartsheet/      # Core table components
│   │   ├── grid/        # Grid view
│   │   ├── kanban/      # Kanban view
│   │   ├── calendar/    # Calendar view
│   │   ├── form/        # Form view
│   │   ├── gallery/     # Gallery view
│   │   ├── column/      # Column editors
│   │   ├── toolbar/     # Toolbar (filters, sort, etc.)
│   │   └── expanded-form/  # Row detail modal
│   ├── general/         # Reusable components
│   └── workspace/       # Workspace UI
├── composables/         # Vue composables (state + logic)
│   ├── useViewStore.ts
│   ├── useCalendarViewStore.ts
│   ├── useKanbanViewStore.ts
│   ├── useData.ts       # Data fetching/mutation
│   └── useColumnCreateStore.ts
├── pages/               # Nuxt pages
├── layouts/             # Layouts
└── lib/                 # Utilities
```

---

## 2. DATA MODELING & DATABASE SCHEMA

### Core Entity Hierarchy

```
Workspace (Enterprise feature)
  └── Base (Project)
      └── Source (Database connection)
          └── Model (Table)
              ├── Column
              │   ├── Column Options (type-specific)
              │   └── Validations
              └── View
                  ├── View Columns (visibility, width)
                  ├── Filters
                  ├── Sorts
                  └── Row Color Conditions
```

### Meta Tables (Metadata Storage)

NocoDB stores ALL configuration in meta tables (not in the actual data database):

```typescript
// From packages/nocodb/src/utils/globals.ts
export enum MetaTable {
  PROJECT = 'nc_bases_v2',              // Bases/Workspaces
  SOURCES = 'nc_sources_v2',            // Database connections
  MODELS = 'nc_models_v2',              // Tables
  COLUMNS = 'nc_columns_v2',            // Columns

  // Column Type Options
  COL_RELATIONS = 'nc_col_relations_v2',
  COL_SELECT_OPTIONS = 'nc_col_select_options_v2',
  COL_LOOKUP = 'nc_col_lookup_v2',
  COL_ROLLUP = 'nc_col_rollup_v2',
  COL_FORMULA = 'nc_col_formula_v2',
  COL_QRCODE = 'nc_col_qrcode_v2',
  COL_BARCODE = 'nc_col_barcode_v2',
  COL_LONG_TEXT = 'nc_col_long_text_v2',
  COL_BUTTON = 'nc_col_button_v2',

  // Views
  VIEWS = 'nc_views_v2',
  GRID_VIEW = 'nc_grid_view_v2',
  GRID_VIEW_COLUMNS = 'nc_grid_view_columns_v2',
  FORM_VIEW = 'nc_form_view_v2',
  FORM_VIEW_COLUMNS = 'nc_form_view_columns_v2',
  GALLERY_VIEW = 'nc_gallery_view_v2',
  GALLERY_VIEW_COLUMNS = 'nc_gallery_view_columns_v2',
  KANBAN_VIEW = 'nc_kanban_view_v2',
  KANBAN_VIEW_COLUMNS = 'nc_kanban_view_columns_v2',
  CALENDAR_VIEW = 'nc_calendar_view_v2',
  CALENDAR_VIEW_COLUMNS = 'nc_calendar_view_columns_v2',
  CALENDAR_VIEW_RANGE = 'nc_calendar_view_range_v2',
  MAP_VIEW = 'nc_map_view_v2',
  MAP_VIEW_COLUMNS = 'nc_map_view_columns_v2',

  // View Features
  FILTER_EXP = 'nc_filter_exp_v2',
  SORT = 'nc_sort_v2',
  SHARED_VIEWS = 'nc_shared_views_v2',
  ROW_COLOR_CONDITIONS = 'nc_row_color_conditions',

  // Hooks & Automation
  HOOKS = 'nc_hooks_v2',
  HOOK_LOGS = 'nc_hook_logs_v2',

  // Other
  COMMENTS = 'nc_comments',
  FILE_REFERENCES = 'nc_file_references',
  INTEGRATIONS = 'nc_integrations_v2',
  // ... more
}
```

### Model (Table) Schema

```typescript
// From packages/nocodb/src/models/Model.ts
export default class Model implements TableType {
  // Identifiers
  id: string;
  base_id: string;
  source_id: string;
  fk_workspace_id?: string;

  // Table info
  table_name: string;        // Physical table name in DB
  title: string;             // Display name
  description?: string;
  type: ModelTypes;          // TABLE | VIEW

  // Flags
  mm: boolean;               // Is many-to-many junction table
  enabled: boolean;
  deleted: boolean;
  pin: boolean;
  show_all_fields: boolean;

  // Configuration
  order: number;
  meta: Record<string, any>;  // JSON metadata
  schema: any;                // Schema info

  // Relations (loaded separately)
  columns?: Column[];
  columnsById?: { [id: string]: Column };
  views?: View[];

  // Computed
  primaryKey: Column;         // Auto-detected PK
  displayValue: Column;       // Display value column (pv flag)
}
```

### Column Schema

```typescript
// From packages/nocodb/src/models/Column.ts
export default class Column<T = any> implements ColumnType {
  // Identifiers
  id: string;
  fk_model_id: string;
  base_id: string;
  source_id: string;

  // Column definition
  column_name: string;        // Physical column name
  title: string;              // Display name
  description: string;
  uidt: UITypes;              // UI Type (see UITypes enum)

  // Database properties
  dt: string;                 // Data type (e.g., 'varchar', 'int')
  np: string;                 // Numeric precision
  ns: string;                 // Numeric scale
  clen: string;               // Column length
  cop: string;                // Column options

  // Flags
  pk: boolean;                // Is primary key
  pv: boolean;                // Is display value
  rqd: boolean;               // Required
  un: boolean;                // Unsigned
  ai: boolean;                // Auto increment
  unique: boolean;
  system: boolean;            // System column (hidden)
  readonly?: boolean;

  // Default value
  cdf: string;                // Column default

  // Metadata
  order: number;
  meta: any;                  // JSON metadata
  validate: any;              // Validation rules

  // Type-specific options (loaded separately)
  colOptions: T;              // LinkToAnotherRecordColumn | FormulaColumn | etc.
}
```

### Column Types Storage Pattern

Each column type that needs extra config has its own meta table:

```typescript
// LinkToAnotherRecord (Relations)
interface LinkToAnotherRecordColumn {
  fk_column_id: string;
  type: 'hm' | 'mm' | 'bt';        // has-many, many-many, belongs-to
  fk_child_column_id: string;
  fk_parent_column_id: string;
  fk_related_model_id: string;
  virtual: boolean;
}

// Formula
interface FormulaColumn {
  fk_column_id: string;
  formula: string;                   // Formula expression
  parsed_tree: any;                  // Parsed AST
}

// Lookup (pull data from related table)
interface LookupColumn {
  fk_column_id: string;
  fk_relation_column_id: string;     // Which relation to follow
  fk_lookup_column_id: string;       // Which column to pull
}

// Rollup (aggregate from related records)
interface RollupColumn {
  fk_column_id: string;
  fk_relation_column_id: string;
  fk_rollup_column_id: string;
  rollup_function: 'count' | 'sum' | 'avg' | 'min' | 'max' | ...;
}

// Select Options
interface SelectOption {
  id: string;
  fk_column_id: string;
  title: string;
  color: string;
  order: number;
}
```

### View Schema

```typescript
// From packages/nocodb/src/models/View.ts
interface View {
  id: string;
  fk_model_id: string;
  base_id: string;
  source_id: string;

  title: string;
  type: ViewTypes;            // GRID | KANBAN | GALLERY | FORM | CALENDAR
  is_default: boolean;
  show_system_fields: boolean;

  // View-specific properties (loaded from type-specific table)
  view?: GridView | KanbanView | GalleryView | FormView | CalendarView;

  // View columns (visibility, width, order)
  columns?: GridViewColumn[] | KanbanViewColumn[] | ...;

  // Filters & Sorts
  filter?: Filter;
  sorts?: Sort[];

  // Row coloring
  row_color?: RowColorCondition[];

  // Metadata
  order: number;
  password?: string;          // For shared views
  uuid?: string;              // For shared views
  meta?: any;
}
```

---

## 3. CORE FEATURES LIST

### Table Features
- ✅ **CRUD Operations** - Full Create, Read, Update, Delete
- ✅ **Multiple Database Support** - MySQL, PostgreSQL, SQLite, SQL Server
- ✅ **External Database Connection** - Connect to existing databases
- ✅ **Import/Export** - CSV, Excel, JSON
- ✅ **Batch Operations** - Bulk update, delete
- ✅ **Record Linking** - Foreign key relationships
- ✅ **Attachment Management** - File uploads with S3/local storage

### View Types
1. **Grid View** - Traditional spreadsheet
2. **Form View** - Data entry forms
3. **Gallery View** - Card-based view with images
4. **Kanban View** - Kanban boards (group by single-select)
5. **Calendar View** - Calendar with date fields
6. **Map View** - Geographic data visualization

### Filtering & Sorting
- ✅ **Advanced Filters**
  - Multiple conditions (AND/OR logic)
  - Nested filter groups
  - 20+ operators per field type
  - Filter by related records (lookup filters)
- ✅ **Multi-level Sorting**
  - Sort by multiple columns
  - Ascending/Descending
  - Sort by formula/rollup results

### Formulas & Computed Fields
- ✅ **Formula Columns** - Excel-like formulas (100+ functions)
- ✅ **Lookup Columns** - Pull data from related tables
- ✅ **Rollup Columns** - Aggregate data from related records (COUNT, SUM, AVG, etc.)
- ✅ **QR Code / Barcode** - Generate from other field values

### Grouping & Aggregation
- ✅ **Group By** - Group records by any field
- ✅ **Aggregations** - Count, Sum, Average per group
- ✅ **Collapsed/Expanded Groups**

### Automation & Webhooks
- ✅ **Webhooks** - Trigger on After Insert/Update/Delete
- ✅ **Email Notifications**
- ✅ **Slack Notifications**
- ✅ **Custom Scripts** (via extensions)

### Collaboration
- ✅ **Comments** - Row-level comments with @mentions
- ✅ **Activity Log** - Audit trail for all changes
- ✅ **User/Collaborator Fields** - Assign users to records
- ✅ **Real-time Updates** - WebSocket for live collaboration

### API & Integrations
- ✅ **REST API** - Auto-generated for every table
- ✅ **GraphQL API** (experimental)
- ✅ **API Tokens** - Authentication
- ✅ **Swagger/OpenAPI Docs** - Auto-generated
- ✅ **Zapier Integration**
- ✅ **Third-party OAuth** - Google, GitHub, etc.

### Sharing & Permissions
- ✅ **Public Shared Views** - Share read-only or editable views
- ✅ **Password Protection** - For shared views
- ✅ **Role-based Access** - Owner, Creator, Editor, Commenter, Viewer
- ✅ **Column-level Permissions**
- ✅ **Row-level Permissions** (Enterprise)

### UI/UX Features
- ✅ **Row Expand** - Modal for detailed record view
- ✅ **Row Height** - Short, Medium, Tall, Extra Tall
- ✅ **Column Width** - Resizable
- ✅ **Column Reordering** - Drag & drop
- ✅ **Column Hide/Show**
- ✅ **Row Coloring** - Conditional formatting
- ✅ **Dark Mode**
- ✅ **Keyboard Shortcuts**
- ✅ **Undo/Redo** (limited)

### Data Validation
- ✅ **Required Fields**
- ✅ **Unique Constraints**
- ✅ **Custom Validation Rules** (via formulas)
- ✅ **Type Validation** (email, URL, phone, etc.)

### Import/Sync
- ✅ **CSV Import**
- ✅ **Excel Import**
- ✅ **Airtable Import**
- ✅ **JSON Import**
- ✅ **Sync from Airtable** (periodic sync)

---

## 4. COLUMN TYPES

### Complete UITypes Enum

```typescript
// From packages/nocodb-sdk/src/lib/UITypes.ts
enum UITypes {
  // Basic Types
  ID = 'ID',                              // Auto-increment ID
  SingleLineText = 'SingleLineText',      // VARCHAR
  LongText = 'LongText',                  // TEXT (supports Rich Text mode)
  Attachment = 'Attachment',              // File uploads (JSON array)
  Checkbox = 'Checkbox',                  // BOOLEAN

  // Select Types
  MultiSelect = 'MultiSelect',            // Multiple options
  SingleSelect = 'SingleSelect',          // Dropdown

  // Number Types
  Number = 'Number',                      // INTEGER
  Decimal = 'Decimal',                    // DECIMAL
  Currency = 'Currency',                  // DECIMAL with currency symbol
  Percent = 'Percent',                    // DECIMAL as percentage
  Duration = 'Duration',                  // Time duration (h:mm:ss)
  Rating = 'Rating',                      // Star rating (1-10)

  // Date/Time Types
  Date = 'Date',                          // DATE
  DateTime = 'DateTime',                  // DATETIME
  Time = 'Time',                          // TIME
  Year = 'Year',                          // YEAR
  CreatedTime = 'CreatedTime',            // Auto timestamp (create)
  LastModifiedTime = 'LastModifiedTime',  // Auto timestamp (update)

  // String Validation Types
  Email = 'Email',                        // Validated email
  URL = 'URL',                            // Validated URL
  PhoneNumber = 'PhoneNumber',            // Formatted phone

  // Relationship Types
  LinkToAnotherRecord = 'LinkToAnotherRecord',  // Foreign key (old)
  Links = 'Links',                        // Foreign key (new)
  ForeignKey = 'ForeignKey',              // System FK column

  // Computed Types
  Formula = 'Formula',                    // Calculated field
  Lookup = 'Lookup',                      // Pull from related table
  Rollup = 'Rollup',                      // Aggregate from related
  Count = 'Count',                        // Count of related records

  // Special Types
  QrCode = 'QrCode',                      // QR code from another field
  Barcode = 'Barcode',                    // Barcode from another field
  Button = 'Button',                      // Action button (webhook/script)
  Geometry = 'Geometry',                  // Geospatial data
  GeoData = 'GeoData',                    // Lat/Long
  JSON = 'JSON',                          // JSON data
  SpecificDBType = 'SpecificDBType',      // Raw DB type

  // User/Collaborator Types
  User = 'User',                          // User reference
  Collaborator = 'Collaborator',          // User (legacy)
  CreatedBy = 'CreatedBy',                // Auto user (create)
  LastModifiedBy = 'LastModifiedBy',      // Auto user (update)

  // System Types
  AutoNumber = 'AutoNumber',              // Auto-incrementing number
  Order = 'Order',                        // Row order (for drag-drop)
}
```

### Column Type Categories

#### **1. Simple Data Types** (stored directly in column)
- SingleLineText, LongText, Email, URL, PhoneNumber
- Number, Decimal, Currency, Percent, Duration, Rating
- Date, DateTime, Time, Year
- Checkbox
- Attachment (JSON array)
- JSON
- Geometry, GeoData

#### **2. Select Types** (with separate options table)
- SingleSelect → `nc_col_select_options_v2`
- MultiSelect → `nc_col_select_options_v2`
- User → Similar to select with user list

#### **3. Relationship Types** (with separate relations table)
- LinkToAnotherRecord → `nc_col_relations_v2`
- Links → `nc_col_relations_v2`
- ForeignKey (system-generated, read-only)

**Relation Types:**
```typescript
type: 'hm' | 'mm' | 'bt'
// hm = has-many (one-to-many)
// mm = many-to-many (junction table)
// bt = belongs-to (many-to-one)
```

#### **4. Computed/Virtual Types** (not stored, calculated on read)
- **Formula** → `nc_col_formula_v2`
  - JavaScript-like expressions
  - 100+ built-in functions
  - Can reference other columns
  - Parsed AST stored in `parsed_tree`

- **Lookup** → `nc_col_lookup_v2`
  - Follow a relation
  - Pull value from related record

- **Rollup** → `nc_col_rollup_v2`
  - Follow a relation
  - Aggregate multiple related records
  - Functions: count, sum, avg, min, max, concatenate, etc.

- **Count** - Count of linked records (simplified rollup)

#### **5. Generated Types** (computed from other columns)
- QrCode → `nc_col_qrcode_v2` (references another column)
- Barcode → `nc_col_barcode_v2` (references another column)
- Button → `nc_col_button_v2` (webhook/script actions)

#### **6. System/Auto Types** (auto-populated)
- ID (auto-increment primary key)
- AutoNumber (custom auto-increment)
- CreatedTime (timestamp on insert)
- LastModifiedTime (timestamp on update)
- CreatedBy (user on insert)
- LastModifiedBy (user on update)
- Order (for manual row ordering)

### Column Storage Examples

**Example: SingleSelect Column**
```sql
-- In nc_columns_v2
INSERT INTO nc_columns_v2 (
  id, fk_model_id, column_name, title, uidt, dt, ...
) VALUES (
  'col_xyz', 'model_abc', 'status', 'Status', 'SingleSelect', 'varchar', ...
);

-- In nc_col_select_options_v2
INSERT INTO nc_col_select_options_v2 (fk_column_id, title, color, order) VALUES
  ('col_xyz', 'Todo', 'blue', 1),
  ('col_xyz', 'In Progress', 'yellow', 2),
  ('col_xyz', 'Done', 'green', 3);
```

**Example: LinkToAnotherRecord (One-to-Many)**
```sql
-- In nc_columns_v2 (in "Orders" table)
INSERT INTO nc_columns_v2 (
  id, fk_model_id, column_name, title, uidt, ...
) VALUES (
  'col_order_customer', 'model_orders', 'customer_id', 'Customer',
  'LinkToAnotherRecord', ...
);

-- In nc_col_relations_v2
INSERT INTO nc_col_relations_v2 (
  fk_column_id, type, fk_child_column_id, fk_parent_column_id, fk_related_model_id
) VALUES (
  'col_order_customer', 'bt', 'customer_id', 'id', 'model_customers'
);
```

**Example: Formula Column**
```sql
-- In nc_columns_v2
INSERT INTO nc_columns_v2 (
  id, fk_model_id, column_name, title, uidt, virtual, ...
) VALUES (
  'col_total', 'model_orders', NULL, 'Total', 'Formula', true, ...
);

-- In nc_col_formula_v2
INSERT INTO nc_col_formula_v2 (fk_column_id, formula, parsed_tree) VALUES (
  'col_total',
  '{quantity} * {unit_price}',
  '{"type":"BinaryExpression","operator":"*",...}'  -- AST
);
```

---

## 5. FRONTEND ARCHITECTURE

### Tech Stack
- **Framework:** Nuxt 3 (Vue 3 + SSR)
- **State Management:** Pinia stores + Vue composables
- **Styling:** Windi CSS (Tailwind variant) + Ant Design Vue
- **Icons:** Iconify
- **Rich Text:** TipTap
- **Data Grid:** Custom virtual scrolling grid
- **Charts:** Chart.js
- **Drag & Drop:** Sortable.js
- **WebSocket:** Socket.io-client

### State Management Pattern

**Composables = Store + Logic**

Each major feature has a composable that manages:
1. Reactive state
2. API calls
3. Business logic
4. UI state

Example composables:
```typescript
// useViewStore.ts - Manages current view state
const viewStore = useViewStore()
viewStore.openedViewsTab  // Current view
viewStore.activeView      // Active view data
viewStore.views           // All views

// useData.ts - Manages table data CRUD
const data = useData()
data.loadData()           // Fetch rows
data.insertRow()
data.updateRow()
data.deleteRow()

// useCalendarViewStore.ts - Calendar-specific logic
const calendar = useCalendarViewStore()
calendar.calendarRange
calendar.loadEvents()

// useKanbanViewStore.ts - Kanban-specific logic
const kanban = useKanbanViewStore()
kanban.groupingField
kanban.moveRecord()
```

### Component Architecture

**Smartsheet = Core Table Component**

```
SmartsheetGrid.vue
├── SmartsheetHeader.vue        # Column headers
├── SmartsheetToolbar.vue       # Filter, sort, group toolbar
└── SmartsheetCanvas.vue        # Virtualized grid
    └── Cell components (by type)
        ├── CellText.vue
        ├── CellNumber.vue
        ├── CellSelect.vue
        ├── CellDate.vue
        ├── CellAttachment.vue
        ├── CellLTAR.vue       # LinkToAnotherRecord
        └── ... (one per UIType)
```

**View Type Components**

Each view type is a separate component:
```
smartsheet/
├── Grid.vue              # Grid view (default)
├── Kanban.vue            # Kanban board
├── Calendar.vue          # Calendar
├── Gallery.vue           # Gallery/cards
└── Form.vue              # Form view
```

### Cell Editing Pattern

Each cell type has:
1. **Display Component** (read mode)
2. **Edit Component** (edit mode)
3. **Validation**
4. **Formatting**

Example:
```vue
<!-- CellSelect.vue -->
<template>
  <div @click="edit = true">
    <!-- Display mode -->
    <div v-if="!edit" class="cell-display">
      {{ displayValue }}
    </div>

    <!-- Edit mode -->
    <a-select v-else v-model="vModel" @blur="edit = false">
      <a-select-option
        v-for="option in column.colOptions.options"
        :value="option.title"
      >
        {{ option.title }}
      </a-select-option>
    </a-select>
  </div>
</template>
```

### Virtual Scrolling

NocoDB uses custom virtual scrolling for performance:
- Only renders visible rows + buffer
- Dynamically calculates row heights
- Handles 100k+ rows smoothly
- Uses IntersectionObserver for optimization

### Real-time Updates

```typescript
// WebSocket connection per base
const socket = io('/base/{baseId}')

socket.on('rowUpdate', (data) => {
  // Update local state
  data.updateRow(data.rowId, data.changes)
})

socket.on('rowInsert', (data) => {
  data.insertRow(data.row)
})
```

---

## 6. BACKEND ARCHITECTURE

### Tech Stack
- **Framework:** NestJS (TypeScript + Decorators)
- **ORM:** Custom (not TypeORM/Prisma)
- **Query Builder:** Knex.js
- **Database:** Supports MySQL, PostgreSQL, SQLite, SQL Server
- **Cache:** Redis (optional, in-memory fallback)
- **File Storage:** Local, S3, Minio, GCS
- **Auth:** Passport.js (JWT, OAuth)
- **Validation:** Ajv (JSON Schema)

### Architecture Layers

```
Controller Layer (NestJS)
    ↓
Service Layer (Business Logic)
    ↓
Model Layer (Data Access)
    ↓
BaseModelSqlv2 (CRUD Abstraction)
    ↓
Knex (Query Builder)
    ↓
Database
```

### Model Pattern (Active Record-ish)

Unlike TypeORM, NocoDB uses a custom model pattern:

```typescript
// Model class = Static methods + Instance methods
class Model {
  // Static: Fetch from meta tables
  static async get(id): Promise<Model> {
    // 1. Check cache
    // 2. Query nc_models_v2
    // 3. Parse metadata
    // 4. Return Model instance
  }

  static async list(baseId): Promise<Model[]> { }
  static async insert(data): Promise<Model> { }

  // Instance: Work with columns/views
  async getColumns(): Promise<Column[]> { }
  async getViews(): Promise<View[]> { }
  async delete(): Promise<void> { }
}
```

**All metadata queries go through:**
```typescript
ncMeta.metaGet2(workspace_id, base_id, MetaTable.MODELS, id)
ncMeta.metaList2(workspace_id, base_id, MetaTable.COLUMNS, { fk_model_id })
ncMeta.metaInsert2(workspace_id, base_id, MetaTable.VIEWS, viewData)
ncMeta.metaUpdate(workspace_id, base_id, MetaTable.COLUMNS, data, id)
ncMeta.metaDelete(workspace_id, base_id, MetaTable.MODELS, id)
```

### BaseModelSqlv2 (Core CRUD Engine)

**The most important class** - handles ALL data operations:

```typescript
class BaseModelSqlv2 {
  constructor({
    dbDriver,      // Knex instance
    model,         // Model metadata
    viewId,        // Optional view for filtering
  }) { }

  // CRUD
  async list(args: {
    where?: string;
    filterArr?: Filter[];
    sortArr?: Sort[];
    offset?: number;
    limit?: number;
    fields?: string[];
  }): Promise<any[]> {
    // 1. Build query with Knex
    // 2. Apply filters
    // 3. Apply sorts
    // 4. Apply pagination
    // 5. Load related data (lookups, rollups)
    // 6. Format response
  }

  async findOne(args): Promise<any> { }
  async insert(data): Promise<any> { }
  async updateByPk(id, data): Promise<any> { }
  async delByPk(id): Promise<any> { }

  // Bulk operations
  async bulkInsert(data[]): Promise<void> { }
  async bulkUpdate(data[]): Promise<void> { }
  async bulkDelete(ids[]): Promise<void> { }

  // Aggregations
  async count(args): Promise<number> { }
  async groupBy(args): Promise<any[]> { }

  // Relations
  async readByPk(id, includeLTAR = true): Promise<any> { }
  async nestedRead(id, relationColumnId): Promise<any[]> { }
  async nestedInsert(id, relationColumnId, data): Promise<void> { }
  async nestedDelete(id, relationColumnId, childId): Promise<void> { }
}
```

### Query Building Pattern

```typescript
// Example: Building a filtered query
async list(args) {
  const qb = this.dbDriver(this.tnPath)  // SELECT * FROM table

  // Apply view filters
  if (this.viewId) {
    const view = await View.get(this.viewId)
    const filters = await view.getFilters()
    await this.applyFilters(qb, filters)
  }

  // Apply user filters
  if (args.filterArr) {
    await this.applyFilters(qb, args.filterArr)
  }

  // Apply sorts
  if (args.sortArr) {
    for (const sort of args.sortArr) {
      qb.orderBy(sort.fk_column_id, sort.direction)
    }
  }

  // Pagination
  qb.offset(args.offset).limit(args.limit)

  // Execute
  const data = await qb

  // Post-process (load lookups, format, etc.)
  return this.formatResponse(data)
}
```

### Filter System

Filters support complex nested conditions:

```typescript
interface Filter {
  id: string;
  fk_view_id: string;
  fk_parent_id?: string;  // For nested filters
  logical_op?: 'and' | 'or';
  comparison_op?: 'eq' | 'neq' | 'gt' | 'lt' | 'like' | 'in' | ...;
  value?: string;
  fk_column_id?: string;
  is_group?: boolean;     // Is this a group filter?
  children?: Filter[];    // Child filters
}
```

**Filter to SQL:**
```typescript
async applyFilters(qb, filters) {
  for (const filter of filters) {
    if (filter.is_group) {
      // Nested group
      qb[filter.logical_op === 'or' ? 'orWhere' : 'andWhere'](subQb => {
        this.applyFilters(subQb, filter.children)
      })
    } else {
      // Individual filter
      const column = await Column.get(filter.fk_column_id)
      qb.where(
        column.column_name,
        this.opMap[filter.comparison_op],
        this.formatValue(filter.value, column)
      )
    }
  }
}
```

### Caching Strategy

**3-Level Cache:**

1. **In-Memory (L1)** - LRU cache in Node process
2. **Redis (L2)** - Shared across instances
3. **Database (L3)** - Source of truth

```typescript
// Cache scopes
CacheScope.MODEL       // Model metadata
CacheScope.COLUMN      // Column metadata
CacheScope.VIEW        // View metadata
CacheScope.SINGLE_QUERY // Query results (temp cache)

// Cache keys
`${CacheScope.MODEL}:${modelId}`
`${CacheScope.COLUMN}:${modelId}:list`
`${CacheScope.VIEW}:${viewId}`
```

**Cache invalidation:**
```typescript
// When column updated
await NocoCache.del(`${CacheScope.COLUMN}:${columnId}`)
await NocoCache.del(`${CacheScope.COLUMN}:${modelId}:list`)

// When model deleted (cascade)
await NocoCache.deepDel(
  `${CacheScope.MODEL}:${modelId}`,
  CacheDelDirection.CHILD_TO_PARENT  // Delete children first
)
```

### API Structure

**RESTful API Pattern:**
```
/api/v2/tables/{tableId}/
├── GET    /                    # Get table metadata
├── POST   /rows                # Create row
├── GET    /rows                # List rows (with filters/sorts)
├── GET    /rows/{rowId}        # Get single row
├── PATCH  /rows/{rowId}        # Update row
├── DELETE /rows/{rowId}        # Delete row
├── GET    /rows/count          # Count rows
├── POST   /rows/bulk           # Bulk insert
├── PATCH  /rows/bulk           # Bulk update
├── DELETE /rows/bulk           # Bulk delete
├── GET    /rows/export/{type}  # Export (csv/excel)
├── POST   /rows/import         # Import
└── ...
```

**View-specific endpoints:**
```
/api/v2/tables/{tableId}/views/{viewId}/
├── GET    /rows                # List with view filters
├── POST   /rows                # Create (view may be a form)
└── ...
```

**Nested relations:**
```
/api/v2/tables/{tableId}/rows/{rowId}/
├── GET    /{relationColumnId}/rows    # Get related rows
├── POST   /{relationColumnId}/rows    # Link/create related row
└── DELETE /{relationColumnId}/rows/{childId}  # Unlink
```

---

## 7. VIEW TYPES

### Grid View (Default)

**Features:**
- Virtualized scrolling
- Resizable columns
- Reorderable columns
- Row height options (short/medium/tall/extra-tall)
- Inline editing
- Batch operations (select multiple rows)
- Frozen columns (lock left columns)
- Group by field

**Storage:**
```typescript
interface GridView {
  fk_view_id: string;
  meta?: {
    expanded_form_width?: 'small' | 'medium' | 'large';
  };
}

interface GridViewColumn {
  id: string;
  fk_view_id: string;
  fk_column_id: string;
  show: boolean;           // Visible?
  order: number;           // Column order
  width?: string;          // Column width (px)
  group_by?: boolean;      // Group by this column?
  group_by_order?: number;
  group_by_sort?: 'asc' | 'desc';
}
```

### Kanban View

**Features:**
- Cards grouped by SingleSelect field
- Drag & drop between groups
- Collapse/expand groups
- Custom card template
- Limit cards per group

**Storage:**
```typescript
interface KanbanView {
  fk_view_id: string;
  fk_grp_col_id: string;   // Which SingleSelect column to group by
  meta?: {
    fk_cover_image_col_id?: string;  // Cover image column
    groupingFieldColOptions?: any;    // Cached options
  };
}

interface KanbanViewColumn {
  id: string;
  fk_view_id: string;
  fk_column_id: string;
  show: boolean;
  order: number;
}
```

**Kanban Logic:**
```typescript
// Group records by SingleSelect value
const groups = {}
for (const option of groupingColumn.colOptions.options) {
  groups[option.title] = []
}

for (const row of rows) {
  const groupValue = row[groupingColumn.title] || 'Uncategorized'
  groups[groupValue].push(row)
}

// Render columns
for (const [groupName, records] of Object.entries(groups)) {
  renderKanbanColumn(groupName, records)
}

// Drag & drop = Update record's grouping field value
```

### Gallery View

**Features:**
- Card-based layout
- Cover image from Attachment column
- Custom card fields
- Grid or list layout

**Storage:**
```typescript
interface GalleryView {
  fk_view_id: string;
  fk_cover_image_col_id?: string;  // Cover image column
  meta?: {
    card_cover_ratio?: number;      // Image aspect ratio
  };
}

interface GalleryViewColumn {
  id: string;
  fk_view_id: string;
  fk_column_id: string;
  show: boolean;
  order: number;
}
```

### Form View

**Features:**
- Public form for data entry
- Field configuration (label, help text, required)
- Multi-step forms
- Success message/redirect
- Email notifications on submission
- CAPTCHA support
- Limit submissions

**Storage:**
```typescript
interface FormView {
  fk_view_id: string;
  heading?: string;
  subheading?: string;
  success_msg?: string;
  redirect_url?: string;
  redirect_after_secs?: number;
  email?: string;              // Send notification to
  banner_image_url?: string;
  logo_url?: string;
  submit_another_form?: boolean;
  show_blank_form?: boolean;
  meta?: {
    // ... theme, colors, etc.
  };
}

interface FormViewColumn {
  id: string;
  fk_view_id: string;
  fk_column_id: string;
  show: boolean;
  order: number;
  label?: string;              // Override label
  help?: string;               // Help text
  description?: string;
  required?: boolean;          // Override required
  meta?: {
    // ... field-specific config
  };
}
```

### Calendar View

**Features:**
- Month, Week, Day views
- Events from Date/DateTime columns
- Drag & drop to reschedule
- Color by field
- Multiple date fields (ranges)

**Storage:**
```typescript
interface CalendarView {
  fk_view_id: string;
  meta?: {
    selectedDate?: string;
    selectedDateRange?: {
      start: string;
      end: string;
    };
  };
}

interface CalendarViewColumn {
  id: string;
  fk_view_id: string;
  fk_column_id: string;
  show: boolean;
  order: number;
  bold?: boolean;
  italic?: boolean;
  underline?: boolean;
}

// NEW: Calendar ranges (support multiple date fields)
interface CalendarViewRange {
  id: string;
  fk_view_id: string;
  fk_from_column_id: string;   // Start date
  fk_to_column_id?: string;    // End date (optional, for ranges)
}
```

**Calendar Logic:**
```typescript
// Load calendar events
const ranges = await CalendarViewRange.list(viewId)

const events = []
for (const range of ranges) {
  const fromCol = await Column.get(range.fk_from_column_id)
  const toCol = range.fk_to_column_id
    ? await Column.get(range.fk_to_column_id)
    : null

  for (const row of rows) {
    events.push({
      id: row.id,
      title: row[displayColumn.title],
      start: row[fromCol.title],
      end: toCol ? row[toCol.title] : row[fromCol.title],
      allDay: fromCol.uidt === 'Date',
    })
  }
}
```

### Map View

**Features:**
- Marker for each record with GeoData
- Cluster markers
- Popup on click
- Filter by map bounds

**Storage:**
```typescript
interface MapView {
  fk_view_id: string;
  fk_geo_data_col_id: string;  // GeoData column
  meta?: {
    fk_cover_image_col_id?: string;
    center?: { lat: number; lng: number };
    zoom?: number;
  };
}

interface MapViewColumn {
  id: string;
  fk_view_id: string;
  fk_column_id: string;
  show: boolean;
  order: number;
}
```

---

## 8. KEY FILES REFERENCE

### Backend Core Files

**Models:**
```
packages/nocodb/src/models/
├── Model.ts                    # Table model (1295 lines)
├── Column.ts                   # Column model (2000+ lines)
├── View.ts                     # View base model
├── GridView.ts, KanbanView.ts, etc.
├── Filter.ts                   # Filter logic
├── Sort.ts                     # Sort logic
├── LinkToAnotherRecordColumn.ts
├── FormulaColumn.ts
├── LookupColumn.ts
├── RollupColumn.ts
└── SelectOption.ts
```

**Database Layer:**
```
packages/nocodb/src/db/
└── BaseModelSqlv2.ts           # Core CRUD (3000+ lines) ⭐ MOST IMPORTANT
```

**Controllers:**
```
packages/nocodb/src/controllers/
├── tables.controller.ts        # Table CRUD
├── columns.controller.ts       # Column CRUD
├── views.controller.ts         # View CRUD
├── data-alias.controller.ts    # Row CRUD
├── filters.controller.ts
├── sorts.controller.ts
└── ...
```

**Services:**
```
packages/nocodb/src/services/
├── tables.service.ts
├── columns.service.ts
├── views.service.ts
├── data-table.service.ts       # Row CRUD logic
└── ...
```

**Utilities:**
```
packages/nocodb/src/utils/
├── globals.ts                  # MetaTable enum, constants
└── modelUtils.ts               # Model helpers
```

### Frontend Core Files

**Composables (State Management):**
```
packages/nc-gui/composables/
├── useViewStore.ts             # View state
├── useData.ts                  # Data CRUD ⭐ IMPORTANT
├── useKanbanViewStore.ts       # Kanban logic
├── useCalendarViewStore.ts     # Calendar logic
├── useColumnCreateStore.ts     # Column creation
├── useExpandedFormStore.ts     # Expand row modal
└── useMultiSelect.ts           # Batch operations
```

**Grid Components:**
```
packages/nc-gui/components/smartsheet/
├── Grid.vue                    # Main grid view
├── grid/
│   ├── Table.vue               # Grid table wrapper
│   ├── canvas/
│   │   ├── index.vue           # Virtual scroll canvas
│   │   └── cells/              # Cell components
│   │       ├── CellText.vue
│   │       ├── CellSelect.vue
│   │       ├── CellLTAR.vue    # Relations
│   │       └── ... (one per type)
│   └── utils/
│       └── cellTypes.ts        # Cell type registry
├── toolbar/                    # Toolbar (filter, sort, etc.)
│   ├── Filter.vue
│   ├── Sort.vue
│   └── ...
└── expanded-form/              # Row detail modal
    └── index.vue
```

**View Components:**
```
packages/nc-gui/components/smartsheet/
├── Kanban.vue
├── kanban/
│   ├── Card.vue
│   └── Stack.vue
├── Calendar.vue
├── calendar/
│   ├── MonthView/
│   ├── WeekView/
│   └── DayView/
├── Gallery.vue
└── Form.vue
```

### SDK (Shared Types)

```
packages/nocodb-sdk/src/lib/
├── UITypes.ts                  # UITypes enum ⭐ CRITICAL
├── Api.ts                      # Auto-generated API client
├── globals.ts                  # Shared constants
└── formula/                    # Formula parsing
```

---

## 9. IMPLEMENTATION INSIGHTS

### What Makes NocoDB Work

#### 1. **Separation of Metadata vs Data**
- All table/column/view configuration in meta tables
- Actual user data in separate tables
- Allows dynamic schema without migrations
- Can connect to external databases without modifying them

#### 2. **Virtual Columns**
- Formula, Lookup, Rollup never stored
- Computed on-the-fly during queries
- Keeps data normalized
- Allows referencing deleted/changed columns gracefully

#### 3. **View as Filter + Sort + Hide**
- Views don't duplicate data
- View = saved filter/sort configuration
- Multiple views on same table = different perspectives
- View-specific column order/width/visibility

#### 4. **Extensible Column Types**
- Base column in `nc_columns_v2`
- Type-specific options in separate tables
- Easy to add new types without schema changes
- Clean separation of concerns

#### 5. **Smart Caching**
- Metadata heavily cached (changes infrequently)
- Data queries cached temporarily
- Hierarchical invalidation (delete model → delete columns)
- Redis for multi-instance deployments

#### 6. **Formula System**
- JavaScript-like syntax (familiar to users)
- Parsed to AST, stored in DB
- Type inference from AST
- Runtime evaluation during queries
- 100+ built-in functions (SUM, AVERAGE, CONCATENATE, IF, etc.)

#### 7. **Relationship Handling**
- Three types: has-many, many-to-many, belongs-to
- Automatic junction tables for many-to-many
- Nested read/write for related records
- Lookup/Rollup built on top of relations

#### 8. **API Auto-generation**
- Every table gets full CRUD API
- Swagger docs auto-generated
- Handles complex queries (filters, sorts, pagination)
- GraphQL support (experimental)

---

### Architecture Decisions to Learn From

#### ✅ **Good Decisions**

1. **Metadata-driven architecture**
   - Schema changes without downtime
   - Support for external databases
   - Easy to add new features

2. **Virtual scrolling**
   - Handles huge datasets
   - Smooth UX even with 100k rows

3. **View system**
   - Flexible without data duplication
   - Empowers users to create custom views

4. **Type system**
   - Clear separation of UI types vs DB types
   - Extensible for new types

5. **Caching strategy**
   - Smart use of Redis
   - Hierarchical invalidation

#### ⚠️ **Potential Issues**

1. **No real ORM**
   - Custom model layer (reinventing the wheel)
   - Harder to maintain than TypeORM/Prisma

2. **Complex query building**
   - Manual Knex queries everywhere
   - Prone to N+1 query problems

3. **Tight coupling**
   - Frontend tightly coupled to backend structure
   - Hard to swap backends

4. **Limited type safety**
   - Dynamic schema = runtime errors
   - TypeScript types not fully utilized

5. **Performance at scale**
   - Complex queries with many lookups/rollups can be slow
   - No query optimization layer

---

### Key Takeaways for BusinessOS Tables

#### Must-Have Features

1. **Column Types to Support:**
   - Text, Number, Date, Select, Multi-select, Checkbox, Attachment
   - **Relations** (Link to another table) - CRITICAL
   - **Formula** (computed fields) - HIGH VALUE
   - **Rollup** (aggregate from relations) - HIGH VALUE
   - User/Collaborator (assign people)
   - Created/Modified time & by

2. **View Types to Support:**
   - **Grid** (default, must-have)
   - **Kanban** (high user value)
   - **Calendar** (nice-to-have)
   - **Form** (for public data entry)
   - Gallery (lower priority)

3. **Features to Prioritize:**
   - ✅ Filters (with AND/OR logic)
   - ✅ Sorting (multi-level)
   - ✅ Grouping (by field)
   - ✅ Inline editing
   - ✅ Row expand (detail view)
   - ✅ Bulk operations
   - ✅ Import/Export CSV
   - ✅ API access
   - ✅ Webhooks
   - ✅ Real-time updates

4. **Architecture Patterns to Use:**
   - Metadata in separate tables (flexible schema)
   - View = saved filter/sort/hide configuration
   - Virtual columns (computed, not stored)
   - Cache metadata aggressively
   - Use proper ORM (Drizzle for Go, or similar)

5. **What to Avoid:**
   - Don't reinvent ORM
   - Don't over-engineer formula parsing
   - Don't build custom virtual scrolling (use library)
   - Don't support 20+ view types initially (start with 2-3)

---

## Conclusion

NocoDB is a sophisticated Airtable clone with:
- **~55 column types** (including virtual types)
- **6 view types** (Grid, Kanban, Gallery, Form, Calendar, Map)
- **Advanced filtering** (nested AND/OR with 20+ operators)
- **Formula system** (100+ functions)
- **API auto-generation** (REST + GraphQL)
- **External database support** (MySQL, Postgres, etc.)
- **Collaboration features** (comments, real-time, webhooks)

**For BusinessOS Tables**, focus on:
1. Core column types (10-15 types max initially)
2. Grid + Kanban views (80% of value)
3. Basic formulas (SUM, AVERAGE, CONCATENATE)
4. Relations (LinkToAnotherTable)
5. Filters & sorts
6. API access

Start simple, iterate based on user feedback. Don't try to match NocoDB's feature set immediately.

---

**Next Steps:**
1. Review this document with team
2. Create simplified schema design for BusinessOS
3. Decide on initial column types (suggest 10-12)
4. Design API structure
5. Build proof-of-concept Grid view
6. Add filtering/sorting
7. Add relations
8. Add Kanban view
9. Add formulas (v2)
10. Expand from there

---

**Document Author:** Claude (Codebase Analyzer Agent)
**For:** BusinessOS Tables Module Planning
**Date:** January 8, 2026
