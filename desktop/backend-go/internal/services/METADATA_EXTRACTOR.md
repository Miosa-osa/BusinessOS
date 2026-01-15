# App Metadata Extractor

Intelligent service that parses generated apps to extract meaningful metadata including name, category, icon, and description.

## Features

- **Dual Input Support**: Works with both file system paths and in-memory bundle content
- **Intelligent Category Inference**: Analyzes package.json, README, and code to determine app category
- **Scoring Algorithm**: Uses keyword matching with scoring to handle ambiguous cases
- **Graceful Degradation**: Returns sensible defaults when package.json is missing or malformed
- **Icon Mapping**: Automatically maps categories to appropriate Lucide icons
- **Description Generation**: Auto-generates descriptions when none provided
- **200 Character Limit**: Enforces description length with smart word-boundary truncation

## Usage

### Basic Usage

```go
import "github.com/rhl/businessos-backend/internal/services"

// From bundle content (preferred for new deployments)
bundleContent := map[string]string{
    "package.json": `{"name": "invoice-app", "keywords": ["billing"]}`,
    "README.md": "Invoice generation tool...",
}

metadata, err := services.ExtractAppMetadata("", bundleContent)
if err != nil {
    // Handle error (though errors are rare - defaults are returned)
}

fmt.Println(metadata.Name)        // "Invoice App"
fmt.Println(metadata.Category)    // "finance"
fmt.Println(metadata.Icon)        // "DollarSign"
fmt.Println(metadata.Description) // Description from package.json or generated
```

### File System Mode

```go
// From deployed app directory (legacy mode)
appPath := "/path/to/deployed/app"
metadata, err := services.ExtractAppMetadata(appPath, nil)
```

### Integration with Deployment Service

```go
// Inside DeployApp function
files := ParseFileBundle(codeBundle)

// Convert to bundleContent for metadata extraction
bundleContent := make(map[string]string)
for _, file := range files {
    bundleContent[file.Path] = file.Content
}

// Extract metadata with intelligent analysis
metadata, _ := ExtractAppMetadata(appDir, bundleContent)
deployedApp.Metadata = metadata
```

## Supported Categories

| Category | Keywords | Icon |
|----------|----------|------|
| **finance** | invoice, billing, payment, accounting, expense, stripe, paypal | DollarSign |
| **communication** | chat, messaging, email, slack, discord, notification | MessageSquare |
| **productivity** | todo, task, calendar, notes, reminder, planner | Calendar |
| **analytics** | dashboard, analytics, metrics, reporting, chart, visualization | BarChart |
| **ecommerce** | shop, store, cart, product, checkout, marketplace | ShoppingCart |
| **crm** | crm, customer, contact, lead, sales, pipeline | Users |
| **hr** | employee, hr, payroll, recruitment, hiring | UserCheck |
| **inventory** | inventory, stock, warehouse, asset, tracking | Package |
| **marketing** | marketing, campaign, seo, content, advertising | Megaphone |
| **project** | project, milestone, sprint, agile, scrum, kanban | FolderKanban |
| **general** | (default fallback) | AppWindow |

## Category Inference Algorithm

1. **Collect Text Sources**:
   - package.json: `name`, `description`, `keywords`
   - README.md or similar docs (first 500 chars)

2. **Score Each Category**:
   - Each keyword match = +1 point
   - Case-insensitive matching
   - Partial word matches (e.g., "invoice" matches "invoicing")

3. **Select Winner**:
   - Category with highest score wins
   - Ties go to first match (deterministic)
   - No matches = "general"

## Examples

### Invoice App

**Input:**
```json
{
  "name": "@acme/invoice-generator",
  "description": "Generate professional invoices",
  "keywords": ["invoice", "billing", "payment"]
}
```

**Output:**
```go
{
  Name:        "Invoice Generator",
  Category:    "finance",      // Matched: invoice, billing, payment
  Icon:        "DollarSign",
  Description: "Generate professional invoices",
  Keywords:    ["invoice", "billing", "payment"]
}
```

### Chat App with README Analysis

**Input:**
```json
{
  "name": "chat-app",
  "keywords": []
}
```

**README.md:**
```markdown
# Real-time Chat
A messaging platform with WebSocket support...
```

**Output:**
```go
{
  Name:        "Chat App",
  Category:    "communication",  // Inferred from README: "chat", "messaging"
  Icon:        "MessageSquare",
  Description: "A communication and messaging tool",
  Keywords:    []
}
```

### Conflicting Keywords

**Input:**
```json
{
  "name": "productivity-app",
  "description": "Manage tasks and organize your day",
  "keywords": ["task", "todo", "chat"]
}
```

**Scoring:**
- productivity: 4 matches (task, todo, organize, manage)
- communication: 1 match (chat)

**Output:**
```go
{
  Category: "productivity"  // Highest score wins
}
```

### Missing package.json

**Input:**
```go
bundleContent := map[string]string{
    "index.html": "<html>...</html>",
}
appPath := "/tmp/my-generated-app"
```

**Output:**
```go
{
  Name:        "My Generated App",  // Cleaned from path
  Category:    "general",
  Icon:        "AppWindow",
  Description: "A custom application",
  Keywords:    []
}
```

## Error Handling

The service is designed to **never fail**:
- Missing package.json → Returns defaults
- Malformed JSON → Returns defaults
- No README → Uses package.json only
- Empty keywords → Still attempts category inference from name/description
- Long description → Automatically truncated to 200 chars at word boundary

## Testing

Comprehensive test suite covers:
- All 11 categories
- README-based inference
- Scoring algorithm
- Name cleaning (npm scopes, kebab-case)
- Description truncation
- Monorepo support (frontend/package.json)
- Graceful fallbacks
- Helper function edge cases

Run tests:
```bash
go test -v ./internal/services/metadata_extractor_test.go ./internal/services/metadata_extractor.go
```

## Performance

- **Fast**: O(n) where n = total text length
- **Memory Efficient**: Processes README in 500-char chunks
- **No External Dependencies**: Pure Go, no API calls

## Maintenance

To add a new category:

1. Add to `inferCategory` patterns map:
   ```go
   "new_category": {"keyword1", "keyword2", "keyword3"},
   ```

2. Add to `categoryToIcon` map:
   ```go
   "new_category": "LucideIconName",
   ```

3. Add to `generateDescription` map:
   ```go
   "new_category": "Description template",
   ```

4. Add test case to `metadata_extractor_test.go`

## Architecture

```
ExtractAppMetadata
    ├── Parse package.json (from bundle or filesystem)
    ├── inferCategory (scoring algorithm)
    │   ├── Analyze package.json fields
    │   ├── Analyze README.md (first 500 chars)
    │   └── Score and select winner
    ├── categoryToIcon (lookup table)
    ├── generateDescription (if missing)
    └── truncateDescription (max 200 chars)
```

## Future Enhancements

Potential improvements:
- [ ] Machine learning-based category prediction
- [ ] Support for more file types (Dockerfile, docker-compose.yml)
- [ ] Tech stack detection (React, Vue, Express, FastAPI)
- [ ] Dependency analysis (security, licensing)
- [ ] Screenshot extraction for preview
- [ ] Multi-language support for descriptions
