# NocoDB Code Patterns & Examples
## Technical Implementation Details for BusinessOS Tables

---

## Table of Contents
1. [Database Schema Examples](#database-schema-examples)
2. [Backend Code Patterns](#backend-code-patterns)
3. [Frontend Code Patterns](#frontend-code-patterns)
4. [API Endpoint Examples](#api-endpoint-examples)
5. [Formula System Details](#formula-system-details)
6. [Filter System Implementation](#filter-system-implementation)

---

## Database Schema Examples

### Meta Tables Structure

```sql
-- Core table metadata
CREATE TABLE nc_models_v2 (
  id VARCHAR(20) PRIMARY KEY,
  base_id VARCHAR(20) NOT NULL,
  source_id VARCHAR(20),
  table_name VARCHAR(255),    -- Physical table name
  title VARCHAR(255),          -- Display name
  type VARCHAR(10),            -- 'table' or 'view'
  meta JSON,                   -- Flexible metadata
  `schema` JSON,
  enabled BOOLEAN DEFAULT TRUE,
  mm BOOLEAN DEFAULT FALSE,    -- Is junction table
  tags VARCHAR(255),
  `order` FLOAT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Column metadata
CREATE TABLE nc_columns_v2 (
  id VARCHAR(20) PRIMARY KEY,
  base_id VARCHAR(20) NOT NULL,
  fk_model_id VARCHAR(20) NOT NULL,
  column_name VARCHAR(255),    -- Physical column name (NULL for virtual)
  title VARCHAR(255) NOT NULL, -- Display name
  uidt VARCHAR(50) NOT NULL,   -- UIType (e.g., 'SingleLineText', 'Number')

  -- Database properties
  dt VARCHAR(50),              -- Database type (e.g., 'varchar', 'int')
  np VARCHAR(20),              -- Numeric precision
  ns VARCHAR(20),              -- Numeric scale
  clen VARCHAR(20),            -- Column length
  cop VARCHAR(255),            -- Column options

  -- Flags
  pk BOOLEAN DEFAULT FALSE,    -- Is primary key
  pv BOOLEAN DEFAULT FALSE,    -- Is display value
  rqd BOOLEAN DEFAULT FALSE,   -- Required
  un BOOLEAN DEFAULT FALSE,    -- Unsigned
  ai BOOLEAN DEFAULT FALSE,    -- Auto increment
  unique BOOLEAN DEFAULT FALSE,
  `system` BOOLEAN DEFAULT FALSE,  -- System column (hidden by default)
  virtual BOOLEAN DEFAULT FALSE,   -- Virtual column (not in DB)

  -- Default value
  cdf TEXT,                    -- Column default

  -- Metadata
  `order` FLOAT,
  meta JSON,
  description TEXT,

  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  FOREIGN KEY (fk_model_id) REFERENCES nc_models_v2(id) ON DELETE CASCADE
);

-- Select options (for SingleSelect and MultiSelect)
CREATE TABLE nc_col_select_options_v2 (
  id VARCHAR(20) PRIMARY KEY,
  fk_column_id VARCHAR(20) NOT NULL,
  title VARCHAR(255) NOT NULL,
  color VARCHAR(20),
  `order` FLOAT,
  meta JSON,
  FOREIGN KEY (fk_column_id) REFERENCES nc_columns_v2(id) ON DELETE CASCADE
);

-- Relations (for LinkToAnotherRecord)
CREATE TABLE nc_col_relations_v2 (
  id VARCHAR(20) PRIMARY KEY,
  fk_column_id VARCHAR(20) NOT NULL,
  type VARCHAR(10) NOT NULL,   -- 'hm', 'mm', 'bt'
  fk_child_column_id VARCHAR(20),    -- Foreign key column (child)
  fk_parent_column_id VARCHAR(20),   -- Primary key column (parent)
  fk_related_model_id VARCHAR(20),   -- Related table
  fk_mm_model_id VARCHAR(20),        -- Junction table (for mm)
  fk_mm_child_column_id VARCHAR(20), -- FK in junction (child side)
  fk_mm_parent_column_id VARCHAR(20), -- FK in junction (parent side)
  ur VARCHAR(255),             -- Update rule (e.g., 'CASCADE')
  dr VARCHAR(255),             -- Delete rule (e.g., 'CASCADE')
  virtual BOOLEAN DEFAULT FALSE,
  meta JSON,
  FOREIGN KEY (fk_column_id) REFERENCES nc_columns_v2(id) ON DELETE CASCADE,
  FOREIGN KEY (fk_related_model_id) REFERENCES nc_models_v2(id)
);

-- Formula columns
CREATE TABLE nc_col_formula_v2 (
  id VARCHAR(20) PRIMARY KEY,
  fk_column_id VARCHAR(20) NOT NULL,
  formula TEXT NOT NULL,
  parsed_tree JSON,            -- AST of parsed formula
  error TEXT,                  -- Parse error if any
  meta JSON,
  FOREIGN KEY (fk_column_id) REFERENCES nc_columns_v2(id) ON DELETE CASCADE
);

-- Lookup columns
CREATE TABLE nc_col_lookup_v2 (
  id VARCHAR(20) PRIMARY KEY,
  fk_column_id VARCHAR(20) NOT NULL,
  fk_relation_column_id VARCHAR(20) NOT NULL,  -- Which relation to follow
  fk_lookup_column_id VARCHAR(20) NOT NULL,    -- Which column to pull
  meta JSON,
  FOREIGN KEY (fk_column_id) REFERENCES nc_columns_v2(id) ON DELETE CASCADE,
  FOREIGN KEY (fk_relation_column_id) REFERENCES nc_columns_v2(id),
  FOREIGN KEY (fk_lookup_column_id) REFERENCES nc_columns_v2(id)
);

-- Rollup columns
CREATE TABLE nc_col_rollup_v2 (
  id VARCHAR(20) PRIMARY KEY,
  fk_column_id VARCHAR(20) NOT NULL,
  fk_relation_column_id VARCHAR(20) NOT NULL,
  fk_rollup_column_id VARCHAR(20) NOT NULL,
  rollup_function VARCHAR(50) NOT NULL,  -- 'count', 'sum', 'avg', etc.
  meta JSON,
  FOREIGN KEY (fk_column_id) REFERENCES nc_columns_v2(id) ON DELETE CASCADE
);

-- Views
CREATE TABLE nc_views_v2 (
  id VARCHAR(20) PRIMARY KEY,
  base_id VARCHAR(20) NOT NULL,
  fk_model_id VARCHAR(20) NOT NULL,
  title VARCHAR(255) NOT NULL,
  type VARCHAR(50) NOT NULL,   -- 'grid', 'kanban', 'gallery', 'form', 'calendar'
  is_default BOOLEAN DEFAULT FALSE,
  show_system_fields BOOLEAN DEFAULT FALSE,
  lock_type VARCHAR(20),
  uuid VARCHAR(255),           -- For shared views
  password VARCHAR(255),       -- For protected shared views
  `order` FLOAT,
  meta JSON,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (fk_model_id) REFERENCES nc_models_v2(id) ON DELETE CASCADE
);

-- Grid view (no extra fields, just inherits from views)
CREATE TABLE nc_grid_view_v2 (
  fk_view_id VARCHAR(20) PRIMARY KEY,
  row_height ENUM('short', 'medium', 'tall', 'extra-tall') DEFAULT 'short',
  meta JSON,
  FOREIGN KEY (fk_view_id) REFERENCES nc_views_v2(id) ON DELETE CASCADE
);

-- Grid view columns (visibility, width, order)
CREATE TABLE nc_grid_view_columns_v2 (
  id VARCHAR(20) PRIMARY KEY,
  fk_view_id VARCHAR(20) NOT NULL,
  fk_column_id VARCHAR(20) NOT NULL,
  `show` BOOLEAN DEFAULT TRUE,
  `order` FLOAT,
  width VARCHAR(20),
  group_by BOOLEAN DEFAULT FALSE,
  group_by_order FLOAT,
  group_by_sort VARCHAR(10),   -- 'asc' or 'desc'
  meta JSON,
  FOREIGN KEY (fk_view_id) REFERENCES nc_views_v2(id) ON DELETE CASCADE,
  FOREIGN KEY (fk_column_id) REFERENCES nc_columns_v2(id) ON DELETE CASCADE
);

-- Kanban view
CREATE TABLE nc_kanban_view_v2 (
  fk_view_id VARCHAR(20) PRIMARY KEY,
  fk_grp_col_id VARCHAR(20),   -- SingleSelect column to group by
  meta JSON,
  FOREIGN KEY (fk_view_id) REFERENCES nc_views_v2(id) ON DELETE CASCADE,
  FOREIGN KEY (fk_grp_col_id) REFERENCES nc_columns_v2(id)
);

-- Filters
CREATE TABLE nc_filter_exp_v2 (
  id VARCHAR(20) PRIMARY KEY,
  fk_view_id VARCHAR(20),
  fk_parent_id VARCHAR(20),    -- Parent filter (for nesting)
  is_group BOOLEAN DEFAULT FALSE,
  logical_op VARCHAR(10),      -- 'and' or 'or'
  comparison_op VARCHAR(50),   -- 'eq', 'neq', 'gt', 'lt', 'like', etc.
  comparison_sub_op VARCHAR(50),
  value TEXT,
  fk_column_id VARCHAR(20),
  FOREIGN KEY (fk_view_id) REFERENCES nc_views_v2(id) ON DELETE CASCADE,
  FOREIGN KEY (fk_parent_id) REFERENCES nc_filter_exp_v2(id) ON DELETE CASCADE,
  FOREIGN KEY (fk_column_id) REFERENCES nc_columns_v2(id)
);

-- Sorts
CREATE TABLE nc_sort_v2 (
  id VARCHAR(20) PRIMARY KEY,
  fk_view_id VARCHAR(20) NOT NULL,
  fk_column_id VARCHAR(20),
  direction VARCHAR(10),       -- 'asc' or 'desc'
  `order` FLOAT,
  FOREIGN KEY (fk_view_id) REFERENCES nc_views_v2(id) ON DELETE CASCADE,
  FOREIGN KEY (fk_column_id) REFERENCES nc_columns_v2(id)
);
```

---

## Backend Code Patterns

### 1. Model Loading Pattern

```typescript
// packages/nocodb/src/models/Model.ts

export default class Model {
  // Cached instance
  columns?: Column[];
  views?: View[];

  // Get model with all metadata
  public static async getWithInfo(
    context: NcContext,
    { id }: { id: string },
    ncMeta = Noco.ncMeta,
  ): Promise<Model> {
    // 1. Check cache
    let modelData = await NocoCache.get(
      `${CacheScope.MODEL}:${id}`,
      CacheGetType.TYPE_OBJECT,
    );

    // 2. Load from database if not cached
    if (!modelData) {
      modelData = await ncMeta.metaGet2(
        context.workspace_id,
        context.base_id,
        MetaTable.MODELS,
        id,
      );

      if (modelData) {
        // Parse JSON metadata
        modelData.meta = parseMetaProp(modelData);
        // Cache it
        await NocoCache.set(`${CacheScope.MODEL}:${id}`, modelData);
      }
    }

    if (modelData) {
      const m = this.castType(modelData);

      // 3. Load related data
      await m.getViews(context, false, ncMeta);

      // 4. Get default view ID
      const defaultViewId = m.views.find((view) => view.is_default).id;

      // 5. Load columns with default view context
      await m.getColumns(context, ncMeta, defaultViewId);

      // 6. Compute columns hash (for cache invalidation)
      await m.getColumnsHash(context, ncMeta);

      return m;
    }

    return null;
  }

  // Load columns with caching
  public async getColumns(
    context: NcContext,
    ncMeta = Noco.ncMeta,
    defaultViewId = undefined,
    updateColumns = true,
  ): Promise<Column[]> {
    // Load from Column model
    const columns = await Column.list(
      context,
      {
        fk_model_id: this.id,
        fk_default_view_id: defaultViewId,
      },
      ncMeta,
    );

    if (!updateColumns) return columns;

    // Store on instance
    this.columns = columns;

    // Create columnsById lookup
    this.columnsById = this.columns.reduce((agg, c) => {
      agg[c.id] = c;
      return agg;
    }, {});

    return this.columns;
  }

  // Getter: Primary key column
  public get primaryKey(): Column {
    if (!this.columns) return null;

    // Prefer auto-increment/generated PK
    return (
      this.columns.find((c) => c.pk && (c.ai || c.meta?.ag)) ||
      this.columns?.find((c) => c.pk)
    );
  }

  // Getter: Display value column
  public get displayValue(): Column {
    if (!this.columns) return null;

    // Find column marked as display value (pv flag)
    const pCol = this.columns?.find((c) => c.pv);
    if (pCol) return pCol;

    // For many-to-many junction tables, use first column
    if (this.mm) {
      return this.columns[0];
    }

    // Otherwise, use column next to PK (or first if PK is last)
    const pkIndex = this.columns.indexOf(this.primaryKey);
    if (pkIndex < this.columns.length - 1) {
      return this.columns[pkIndex + 1];
    }
    return this.columns[0];
  }
}
```

### 2. Column Creation Pattern

```typescript
// packages/nocodb/src/models/Column.ts

export default class Column<T = any> {
  public static async insert<T>(
    context: NcContext,
    column: Partial<ColumnReqType>,
    ncMeta = Noco.ncMeta,
  ) {
    // 1. Extract base column properties
    const insertObj = extractProps(column, [
      'id',
      'fk_model_id',
      'column_name',
      'title',
      'uidt',
      'dt',
      'np',
      'ns',
      'clen',
      'cop',
      'pk',
      'rqd',
      'un',
      'ct',
      'ai',
      'unique',
      'cdf',
      'order',
      'base_id',
      'source_id',
      'system',
      'meta',
      'virtual',
      'description',
    ]);

    // 2. Convert meta to JSON if object
    if (insertObj.meta && typeof insertObj.meta === 'object') {
      insertObj.meta = JSON.stringify(insertObj.meta);
    }

    // 3. Generate order if not provided
    if (!insertObj.order) {
      insertObj.order = await ncMeta.metaGetNextOrder(
        MetaTable.COLUMNS,
        { fk_model_id: insertObj.fk_model_id },
      );
    }

    // 4. Insert base column
    const { id } = await ncMeta.metaInsert2(
      context.workspace_id,
      context.base_id,
      MetaTable.COLUMNS,
      insertObj,
    );

    // 5. Insert type-specific options
    if (column.uidt === UITypes.SingleSelect || column.uidt === UITypes.MultiSelect) {
      // Insert select options
      await SelectOption.bulkInsert(
        context,
        {
          fk_column_id: id,
          options: column.colOptions?.options || [],
        },
        ncMeta,
      );
    } else if (column.uidt === UITypes.LinkToAnotherRecord) {
      // Insert relation metadata
      await LinkToAnotherRecordColumn.insert(
        context,
        {
          ...column.colOptions,
          fk_column_id: id,
        },
        ncMeta,
      );
    } else if (column.uidt === UITypes.Formula) {
      // Parse and insert formula
      await FormulaColumn.insert(
        context,
        {
          fk_column_id: id,
          formula: column.colOptions?.formula,
        },
        ncMeta,
      );
    }
    // ... handle other types

    // 6. Clear cache
    await NocoCache.appendToList(
      CacheScope.COLUMN,
      [column.fk_model_id],
      `${CacheScope.COLUMN}:${id}`,
    );

    // 7. Return created column
    return this.get(context, { colId: id }, ncMeta);
  }
}
```

### 3. BaseModelSqlv2 Query Pattern

```typescript
// packages/nocodb/src/db/BaseModelSqlv2.ts

export class BaseModelSqlv2 {
  // List records with filters, sorts, pagination
  async list(args: {
    where?: string;
    filterArr?: Filter[];
    sortArr?: Sort[];
    offset?: number;
    limit?: number;
    fieldsSet?: Set<string>;
  }) {
    const { where, filterArr = [], sortArr = [], offset, limit } = args;

    // 1. Build base query
    const qb = this.dbDriver(this.tnPath);

    // 2. Apply view filters (if viewId provided)
    if (this.viewId) {
      const view = await View.get(this.context, this.viewId);
      const viewFilters = await view.getFilters(this.context);
      await this.applyFilterArr(qb, viewFilters);
    }

    // 3. Apply user-provided filters
    if (filterArr?.length) {
      await this.applyFilterArr(qb, filterArr);
    }

    // 4. Apply where clause
    if (where) {
      qb.whereRaw(where);
    }

    // 5. Apply sorts
    if (sortArr?.length) {
      await this.applySortArr(qb, sortArr);
    }

    // 6. Apply pagination
    if (offset) qb.offset(offset);
    if (limit) qb.limit(limit);

    // 7. Execute query
    const data = await qb;

    // 8. Post-process: Load virtual columns (lookups, rollups, formulas)
    return await this.extractVirtualAndLoadRelatedData(data, {
      fieldsSet: args.fieldsSet,
    });
  }

  // Apply filter array to query builder
  async applyFilterArr(qb: Knex.QueryBuilder, filterArr: Filter[]) {
    if (!filterArr?.length) return;

    for (const filter of filterArr) {
      if (filter.is_group) {
        // Nested filter group
        const method = filter.logical_op === 'or' ? 'orWhere' : 'andWhere';
        qb[method]((nestedQb) => {
          this.applyFilterArr(nestedQb, filter.children || []);
        });
      } else {
        // Individual filter
        await this.applyFilter(qb, filter);
      }
    }
  }

  // Apply single filter
  async applyFilter(qb: Knex.QueryBuilder, filter: Filter) {
    const column = await Column.get(this.context, {
      colId: filter.fk_column_id,
    });

    const field = this.aliasToColumn[column.title];

    switch (filter.comparison_op) {
      case 'eq':
        qb.where(field, '=', this.sanitizeValue(filter.value, column));
        break;
      case 'neq':
        qb.where(field, '!=', this.sanitizeValue(filter.value, column));
        break;
      case 'gt':
        qb.where(field, '>', this.sanitizeValue(filter.value, column));
        break;
      case 'lt':
        qb.where(field, '<', this.sanitizeValue(filter.value, column));
        break;
      case 'like':
        qb.where(field, 'LIKE', `%${filter.value}%`);
        break;
      case 'in':
        const values = filter.value.split(',');
        qb.whereIn(field, values);
        break;
      // ... more operators
    }
  }

  // Apply sorts
  async applySortArr(qb: Knex.QueryBuilder, sortArr: Sort[]) {
    for (const sort of sortArr) {
      const column = await Column.get(this.context, {
        colId: sort.fk_column_id,
      });

      const field = this.aliasToColumn[column.title];
      qb.orderBy(field, sort.direction || 'asc');
    }
  }

  // Load virtual columns (formula, lookup, rollup)
  async extractVirtualAndLoadRelatedData(data: any[], args: any) {
    if (!data?.length) return data;

    const proto = await this.getProto();

    // Process each row
    for (const row of data) {
      // Apply prototype (adds getters for virtual columns)
      Object.setPrototypeOf(row, proto);

      // Load lookups
      for (const col of this.model.columns.filter(
        (c) => c.uidt === UITypes.Lookup,
      )) {
        const lookupVal = await this.getLookupValue(row, col);
        row[col.title] = lookupVal;
      }

      // Load rollups
      for (const col of this.model.columns.filter(
        (c) => c.uidt === UITypes.Rollup,
      )) {
        const rollupVal = await this.getRollupValue(row, col);
        row[col.title] = rollupVal;
      }

      // Evaluate formulas
      for (const col of this.model.columns.filter(
        (c) => c.uidt === UITypes.Formula,
      )) {
        const formulaVal = await this.getFormulaValue(row, col);
        row[col.title] = formulaVal;
      }
    }

    return data;
  }
}
```

### 4. Controller Pattern

```typescript
// packages/nocodb/src/controllers/data-alias.controller.ts

@Controller()
export class DataAliasController {
  @Get('/api/v2/tables/:modelId/rows')
  @Acl('dataList')
  async dataList(
    @TenantContext() context: NcContext,
    @Req() req: NcRequest,
    @Param('modelId') modelId: string,
    @Query('offset') offset?: number,
    @Query('limit') limit?: number,
    @Query('where') where?: string,
    @Query('filterArrJson') filterArrJson?: string,
    @Query('sortArrJson') sortArrJson?: string,
  ) {
    // 1. Parse filters and sorts from JSON
    const filterArr = filterArrJson ? JSON.parse(filterArrJson) : [];
    const sortArr = sortArrJson ? JSON.parse(sortArrJson) : [];

    // 2. Call service
    return await this.dataAliasService.dataList(context, {
      modelId,
      query: {
        offset,
        limit,
        where,
        filterArr,
        sortArr,
      },
    });
  }

  @Post('/api/v2/tables/:modelId/rows')
  @Acl('dataInsert')
  async dataInsert(
    @TenantContext() context: NcContext,
    @Param('modelId') modelId: string,
    @Body() body: any,
  ) {
    return await this.dataAliasService.dataInsert(context, {
      modelId,
      body,
    });
  }

  @Patch('/api/v2/tables/:modelId/rows/:rowId')
  @Acl('dataUpdate')
  async dataUpdate(
    @TenantContext() context: NcContext,
    @Param('modelId') modelId: string,
    @Param('rowId') rowId: string,
    @Body() body: any,
  ) {
    return await this.dataAliasService.dataUpdate(context, {
      modelId,
      rowId,
      body,
    });
  }
}
```

---

## Frontend Code Patterns

### 1. Data Management Composable

```typescript
// packages/nc-gui/composables/useData.ts

export function useData(args: {
  modelValue?: Row;
  syncVisibleColumns?: boolean;
}) {
  const { modelValue, syncVisibleColumns = true } = args;

  // Reactive state
  const formattedData = ref<Row[]>([]);
  const paginationData = ref({ page: 1, pageSize: 25 });
  const selectedRows = ref<Set<number>>(new Set());

  // Load data from API
  async function loadData(params?: {
    offset?: number;
    where?: string;
    limit?: number;
  }) {
    const { offset, where, limit = paginationData.value.pageSize } = params || {};

    try {
      // Build filter array from view
      const filterArr = await buildFilterArr();

      // Build sort array from view
      const sortArr = await buildSortArr();

      // Call API
      const response = await api.dbTableRow.list(
        'noco',
        base.value.id,
        meta.value.id,
        {
          offset,
          limit,
          where,
          filterArrJson: JSON.stringify(filterArr),
          sortArrJson: JSON.stringify(sortArr),
        },
      );

      // Update reactive data
      formattedData.value = formatData(response.list);
      paginationData.value.totalRows = response.pageInfo.totalRows;

      return response;
    } catch (e) {
      message.error('Failed to load data');
      console.error(e);
    }
  }

  // Insert row
  async function insertRow(
    currentRow: Row,
    ltarState?: Record<string, any>,
    args = { metaValue?: TableType; viewMetaValue?: ViewType },
  ) {
    const { metaValue = meta.value, viewMetaValue = viewMeta.value } = args;

    // Prepare data
    const data = { ...currentRow };
    delete data.ncRowId;

    try {
      // Call API
      const insertedRow = await api.dbTableRow.create(
        'noco',
        base.value.id,
        metaValue.id,
        data,
      );

      // Add to local state
      formattedData.value.push(formatRow(insertedRow));

      // Emit event
      emit('row:inserted', insertedRow);

      message.success('Row created successfully');
      return insertedRow;
    } catch (e) {
      message.error('Failed to create row');
      console.error(e);
    }
  }

  // Update row
  async function updateRowProperty(
    row: Row,
    property: string,
    value: any,
  ) {
    try {
      // Optimistic update
      row[property] = value;

      // Call API
      await api.dbTableRow.update(
        'noco',
        base.value.id,
        meta.value.id,
        row.ncRowId,
        { [property]: value },
      );

      // Emit event
      emit('row:updated', row);
    } catch (e) {
      message.error('Failed to update row');
      console.error(e);
      // Revert on error
      await loadData();
    }
  }

  // Delete row
  async function deleteRow(rowIndex: number) {
    try {
      const row = formattedData.value[rowIndex];

      // Call API
      await api.dbTableRow.delete(
        'noco',
        base.value.id,
        meta.value.id,
        row.ncRowId,
      );

      // Remove from local state
      formattedData.value.splice(rowIndex, 1);

      // Emit event
      emit('row:deleted', row);

      message.success('Row deleted successfully');
    } catch (e) {
      message.error('Failed to delete row');
      console.error(e);
    }
  }

  return {
    formattedData,
    loadData,
    insertRow,
    updateRowProperty,
    deleteRow,
    selectedRows,
    paginationData,
  };
}
```

### 2. Kanban View Store

```typescript
// packages/nc-gui/composables/useKanbanViewStore.ts

export function useKanbanViewStore() {
  const { $api } = useNuxtApp();

  // Reactive state
  const kanbanMetaData = ref<KanbanType>();
  const groupingField = ref<ColumnType>();
  const groupingFieldColOptions = ref<SelectOptionsType[]>();

  const countByStack = ref<Map<string, number>>(new Map());

  // Load kanban view data
  async function loadKanbanMeta() {
    if (!viewMeta?.value?.id) return;

    const kanbanData = await $api.dbView.kanbanRead(viewMeta.value.id);

    kanbanMetaData.value = kanbanData;

    // Get grouping column
    groupingField.value = meta.value.columns.find(
      (col) => col.id === kanbanData.fk_grp_col_id,
    );

    // Get select options
    if (groupingField.value?.colOptions) {
      groupingFieldColOptions.value =
        groupingField.value.colOptions.options;
    }
  }

  // Load records grouped by stack
  async function loadKanbanData() {
    if (!groupingFieldColOptions.value) return;

    const records = await $api.dbViewRow.list('noco', base.value.id, meta.value.id, viewMeta.value.id, {
      limit: 1000,  // Load all for kanban
    });

    // Group by stack value
    const grouped = new Map<string, Row[]>();

    for (const option of groupingFieldColOptions.value) {
      grouped.set(option.title, []);
    }
    grouped.set('Uncategorized', []);

    for (const record of records.list) {
      const stackValue = record[groupingField.value.title] || 'Uncategorized';
      if (!grouped.has(stackValue)) {
        grouped.set(stackValue, []);
      }
      grouped.get(stackValue).push(record);
    }

    formattedData.value = grouped;

    // Update counts
    countByStack.value.clear();
    for (const [stack, records] of grouped.entries()) {
      countByStack.value.set(stack, records.length);
    }
  }

  // Move card to different stack
  async function moveCard(
    record: Row,
    fromStack: string,
    toStack: string,
  ) {
    try {
      // Update grouping field value
      await $api.dbTableRow.update(
        'noco',
        base.value.id,
        meta.value.id,
        record.ncRowId,
        {
          [groupingField.value.title]: toStack === 'Uncategorized' ? null : toStack,
        },
      );

      // Update local state
      const fromRecords = formattedData.value.get(fromStack);
      const toRecords = formattedData.value.get(toStack);

      const recordIndex = fromRecords.findIndex((r) => r.ncRowId === record.ncRowId);
      if (recordIndex > -1) {
        fromRecords.splice(recordIndex, 1);
        toRecords.push({ ...record, [groupingField.value.title]: toStack });
      }

      // Update counts
      countByStack.value.set(fromStack, fromRecords.length);
      countByStack.value.set(toStack, toRecords.length);
    } catch (e) {
      message.error('Failed to move card');
      console.error(e);
    }
  }

  return {
    kanbanMetaData,
    groupingField,
    groupingFieldColOptions,
    formattedData,
    countByStack,
    loadKanbanMeta,
    loadKanbanData,
    moveCard,
  };
}
```

### 3. Cell Component Pattern

```vue
<!-- packages/nc-gui/components/smartsheet/grid/canvas/cells/CellText.vue -->
<script setup lang="ts">
import type { ColumnType } from 'nocodb-sdk'

const props = defineProps<{
  column: ColumnType
  modelValue: any
  active?: boolean
  readOnly?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: any): void
}>()

// Edit state
const isEditing = ref(false)
const editValue = ref(props.modelValue)

// Watchers
watch(() => props.active, (active) => {
  if (active) {
    isEditing.value = true
    editValue.value = props.modelValue
  }
})

watch(() => props.modelValue, (newVal) => {
  if (!isEditing.value) {
    editValue.value = newVal
  }
})

// Methods
function onBlur() {
  isEditing.value = false
  if (editValue.value !== props.modelValue) {
    emit('update:modelValue', editValue.value)
  }
}

function onKeyDown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    onBlur()
  } else if (e.key === 'Escape') {
    editValue.value = props.modelValue
    isEditing.value = false
  }
}
</script>

<template>
  <div class="nc-cell-text">
    <!-- Display mode -->
    <div
      v-if="!isEditing"
      class="cell-display"
      @dblclick="!readOnly && (isEditing = true)"
    >
      {{ modelValue }}
    </div>

    <!-- Edit mode -->
    <input
      v-else
      ref="inputRef"
      v-model="editValue"
      class="cell-input"
      type="text"
      @blur="onBlur"
      @keydown="onKeyDown"
    />
  </div>
</template>

<style scoped>
.nc-cell-text {
  width: 100%;
  height: 100%;
}

.cell-display {
  padding: 2px 5px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cell-input {
  width: 100%;
  height: 100%;
  border: none;
  padding: 2px 5px;
  outline: none;
}
</style>
```

---

## API Endpoint Examples

### List Rows (with filters/sorts)

```http
GET /api/v2/tables/{tableId}/rows
  ?offset=0
  &limit=25
  &where=(age,gt,18)
  &filterArrJson=[{"fk_column_id":"col_123","comparison_op":"eq","value":"Active"}]
  &sortArrJson=[{"fk_column_id":"col_456","direction":"asc"}]
```

Response:
```json
{
  "list": [
    {
      "id": 1,
      "Title": "Task 1",
      "Status": "Todo",
      "Priority": "High",
      "Created Time": "2024-01-01T10:00:00Z"
    }
  ],
  "pageInfo": {
    "totalRows": 100,
    "page": 1,
    "pageSize": 25,
    "isFirstPage": true,
    "isLastPage": false
  }
}
```

### Create Row

```http
POST /api/v2/tables/{tableId}/rows
Content-Type: application/json

{
  "Title": "New Task",
  "Status": "Todo",
  "Priority": "Medium"
}
```

### Update Row

```http
PATCH /api/v2/tables/{tableId}/rows/{rowId}
Content-Type: application/json

{
  "Status": "In Progress"
}
```

### Nested Relations

```http
# Get related records
GET /api/v2/tables/{tableId}/rows/{rowId}/{relationColumnId}/rows

# Link existing record
POST /api/v2/tables/{tableId}/rows/{rowId}/{relationColumnId}/rows/{childRowId}

# Create and link new record
POST /api/v2/tables/{tableId}/rows/{rowId}/{relationColumnId}/rows
{
  "Title": "Related Item"
}

# Unlink record
DELETE /api/v2/tables/{tableId}/rows/{rowId}/{relationColumnId}/rows/{childRowId}
```

---

## Formula System Details

### Formula Parsing

```typescript
// Simplified formula parsing
interface FormulaNode {
  type: 'BinaryExpression' | 'UnaryExpression' | 'CallExpression' | 'Literal' | 'Identifier';
  operator?: string;
  left?: FormulaNode;
  right?: FormulaNode;
  callee?: string;
  arguments?: FormulaNode[];
  value?: any;
  name?: string;
}

// Example: Parse "{quantity} * {unit_price}"
const ast: FormulaNode = {
  type: 'BinaryExpression',
  operator: '*',
  left: {
    type: 'Identifier',
    name: 'quantity'
  },
  right: {
    type: 'Identifier',
    name: 'unit_price'
  }
};

// Evaluate formula
function evaluateFormula(ast: FormulaNode, row: Record<string, any>): any {
  switch (ast.type) {
    case 'Literal':
      return ast.value;

    case 'Identifier':
      return row[ast.name];

    case 'BinaryExpression':
      const left = evaluateFormula(ast.left, row);
      const right = evaluateFormula(ast.right, row);

      switch (ast.operator) {
        case '+': return left + right;
        case '-': return left - right;
        case '*': return left * right;
        case '/': return left / right;
        case '==': return left === right;
        case '!=': return left !== right;
        case '>': return left > right;
        case '<': return left < right;
        // ... more operators
      }
      break;

    case 'CallExpression':
      const args = ast.arguments.map(arg => evaluateFormula(arg, row));
      return formulaFunctions[ast.callee](...args);
  }
}

// Built-in functions
const formulaFunctions = {
  SUM: (...args) => args.reduce((a, b) => a + b, 0),
  AVERAGE: (...args) => args.reduce((a, b) => a + b, 0) / args.length,
  CONCATENATE: (...args) => args.join(''),
  IF: (condition, trueVal, falseVal) => condition ? trueVal : falseVal,
  // ... 100+ more functions
};
```

---

## Filter System Implementation

### Nested Filter Structure

```typescript
// Filter tree structure
interface Filter {
  id: string;
  fk_view_id: string;
  fk_parent_id?: string;
  is_group: boolean;
  logical_op?: 'and' | 'or';       // For groups
  comparison_op?: string;           // For leaf filters
  value?: string;
  fk_column_id?: string;
  children?: Filter[];              // Loaded separately
}

// Example: Complex nested filter
const filterTree: Filter = {
  id: 'f1',
  fk_view_id: 'view_123',
  is_group: true,
  logical_op: 'and',
  children: [
    {
      id: 'f2',
      fk_view_id: 'view_123',
      fk_parent_id: 'f1',
      is_group: false,
      comparison_op: 'eq',
      fk_column_id: 'col_status',
      value: 'Active',
    },
    {
      id: 'f3',
      fk_view_id: 'view_123',
      fk_parent_id: 'f1',
      is_group: true,
      logical_op: 'or',
      children: [
        {
          id: 'f4',
          fk_view_id: 'view_123',
          fk_parent_id: 'f3',
          is_group: false,
          comparison_op: 'gt',
          fk_column_id: 'col_priority',
          value: '5',
        },
        {
          id: 'f5',
          fk_view_id: 'view_123',
          fk_parent_id: 'f3',
          is_group: false,
          comparison_op: 'like',
          fk_column_id: 'col_title',
          value: 'urgent',
        },
      ],
    },
  ],
};

// SQL: WHERE status = 'Active' AND (priority > 5 OR title LIKE '%urgent%')
```

### Filter Operators

```typescript
const comparisonOperators = {
  // Equality
  eq: '=',
  neq: '!=',

  // Numeric
  gt: '>',
  gte: '>=',
  lt: '<',
  lte: '<=',

  // String
  like: 'LIKE',
  nlike: 'NOT LIKE',

  // Array
  in: 'IN',
  nin: 'NOT IN',

  // Null checks
  empty: 'IS NULL',
  notempty: 'IS NOT NULL',
  null: 'IS NULL',
  notnull: 'IS NOT NULL',

  // Date-specific
  within: 'WITHIN',  // e.g., within last 7 days
  pastWeek: 'PAST_WEEK',
  pastMonth: 'PAST_MONTH',

  // Boolean
  checked: 'IS TRUE',
  notchecked: 'IS FALSE',

  // Multi-select
  anyof: 'ANY OF',    // Match any of the values
  nanyof: 'NOT ANY OF',

  // Relation
  btw: 'BETWEEN',
  nbtw: 'NOT BETWEEN',
};
```

---

This document provides the technical implementation patterns you need to reverse engineer NocoDB for your BusinessOS Tables module. Focus on the core patterns (metadata storage, BaseModelSqlv2 CRUD, view filtering) and adapt them to your Go/Svelte stack.
