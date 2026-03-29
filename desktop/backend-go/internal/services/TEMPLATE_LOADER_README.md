# Template Loader Service

## Overview

The Template Loader Service provides a robust system for loading, validating, and rendering YAML-based prompt templates. It's designed to work with OSA (One-Shot App) templates that contain metadata, variables, and Go template strings.

## Architecture

### Components

1. **template_types.go** - Type definitions for templates and variables
2. **template_loader_service.go** - Core service implementation
3. **template_helpers.go** - Helper functions for validation and path resolution
4. **template_loader_service_test.go** - Comprehensive test suite
5. **template_helpers_test.go** - Helper function tests

### Directory Structure

```
internal/
├── services/
│   ├── template_types.go              # Type definitions
│   ├── template_loader_service.go     # Main service
│   ├── template_helpers.go            # Utilities
│   ├── template_loader_service_test.go
│   └── template_helpers_test.go
└── prompts/
    └── templates/
        └── osa/
            ├── bug-fix.yaml
            ├── crm-app-generation.yaml
            ├── dashboard-creation.yaml
            ├── data-pipeline-creation.yaml
            └── feature-addition.yaml
```

## Features

### 1. Template Loading

- **File-based**: Loads YAML templates from the filesystem
- **Caching**: In-memory caching with thread-safe access
- **Validation**: Automatic YAML parsing and structure validation

### 2. Variable Validation

Supports multiple variable types:
- `string` - Text values
- `array` - Slices and arrays
- `object` - Maps and structs
- `number` - Int, float, uint variants
- `boolean` - True/false values

Required vs. optional variables with default values.

### 3. Template Rendering

- **Go templates**: Uses Go's `text/template` engine
- **Variable substitution**: Replace `{{.VarName}}` placeholders
- **Default values**: Automatic application of defaults for optional variables

### 4. Thread Safety

- Uses `sync.RWMutex` for cache access
- Safe for concurrent use across goroutines

## Usage

### Basic Example

```go
package main

import (
    "log"
    "github.com/rhl/businessos-backend/internal/services"
)

func main() {
    // Create service
    templatesDir := services.GetTemplatesDirectory()
    service := services.NewTemplateLoaderService(templatesDir)

    // Load template
    tmpl, err := service.LoadTemplate("bug-fix")
    if err != nil {
        log.Fatalf("Failed to load template: %v", err)
    }

    // Prepare variables
    variables := map[string]interface{}{
        "AppName":           "MyApp",
        "BugDescription":    "Login button not working",
        "ReproductionSteps": "1. Click login\n2. Nothing happens",
    }

    // Render template
    rendered, err := service.RenderTemplate("bug-fix", variables)
    if err != nil {
        log.Fatalf("Failed to render: %v", err)
    }

    log.Printf("Rendered template:\n%s", rendered)
}
```

### List All Templates

```go
templates, err := service.ListTemplates()
if err != nil {
    log.Fatal(err)
}

for _, tmpl := range templates {
    fmt.Printf("%s: %s\n", tmpl.Name, tmpl.Description)
}
```

### Validate Variables

```go
tmpl, _ := service.LoadTemplate("crm-app-generation")

variables := map[string]interface{}{
    "AppType":          "CRM",
    "UserBusiness":     "Real Estate",
    "UserRequirements": "Property management",
}

if err := service.ValidateVariables(tmpl, variables); err != nil {
    log.Printf("Validation failed: %v", err)
}
```

### Cache Management

```go
// Clear cache to reload templates from disk
service.ClearCache()
```

## YAML Template Format

### Structure

```yaml
name: "template-name"
display_name: "Human Readable Name"
description: "What this template does"
category: "category-name"
version: "1.0.0"
tags:
  - "tag1"
  - "tag2"

variables:
  - name: "RequiredVar"
    type: "string"
    required: true
    description: "Description of this variable"

  - name: "OptionalVar"
    type: "string"
    required: false
    default: "default-value"
    description: "Optional variable with default"

template: |
  # Template content here
  Required: {{.RequiredVar}}
  Optional: {{.OptionalVar}}
```

### Variable Types

| Type | Go Type | Example |
|------|---------|---------|
| `string` | `string` | `"Hello"` |
| `array` | `[]interface{}` | `["a", "b", "c"]` |
| `object` | `map[string]interface{}` | `{"key": "value"}` |
| `number` | `int`, `float64`, etc. | `123`, `45.67` |
| `boolean` | `bool` | `true`, `false` |

## API Reference

### TemplateLoaderService

#### Methods

**NewTemplateLoaderService(templatesDir string) *TemplateLoaderService**
- Creates a new service instance
- `templatesDir`: Absolute path to templates directory

**LoadTemplate(name string) (*TemplateDefinition, error)**
- Loads and parses a template by name
- Uses cache if available
- Returns error if file not found or invalid YAML

**ListTemplates() ([]*TemplateDefinition, error)**
- Returns all templates in the templates directory
- Skips invalid templates with warning logs

**RenderTemplate(name string, variables map[string]interface{}) (string, error)**
- Loads template, validates variables, and renders output
- Returns fully rendered template string

**ValidateVariables(tmpl *TemplateDefinition, variables map[string]interface{}) error**
- Validates that all required variables are present
- Checks variable types match definitions
- Returns descriptive error on validation failure

**ClearCache()**
- Clears the in-memory template cache
- Forces reload from disk on next access

### Helper Functions

**GetTemplatesDirectory() string**
- Returns absolute path to OSA templates directory
- Uses runtime.Caller to resolve path

**ValidateVariableType(value interface{}, expectedType string) error**
- Validates a single variable's type
- Returns error if type mismatch

**CoerceToString(value interface{}) string**
- Converts any value to string representation
- Useful for debugging and logging

## Testing

### Run All Tests

```bash
cd desktop/backend-go
go test -v ./internal/services -run "TestLoadTemplate|TestRenderTemplate|TestValidate"
```

### Integration Tests

Tests with real OSA templates:

```bash
go test -v ./internal/services -run "TestLoadRealOSATemplates|TestListRealOSATemplates"
```

### Test Coverage

- Unit tests: Template loading, caching, validation, rendering
- Integration tests: Real OSA templates
- Helper tests: Type validation, path resolution
- Edge cases: Missing files, invalid YAML, missing variables

## Error Handling

### Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| "failed to read template file" | File not found | Check template name and directory path |
| "failed to parse template YAML" | Invalid YAML syntax | Fix YAML formatting in template file |
| "required variable 'X' is missing" | Missing variable | Provide all required variables |
| "expected type 'array', got 'string'" | Type mismatch | Ensure variable type matches definition |

### Error Examples

```go
// Missing required variable
err: "variable validation failed: required variable 'AppName' is missing"

// Invalid type
err: "variable validation failed: variable 'Items': expected type 'array', got 'string'"

// Template not found
err: "failed to read template file .../non-existent.yaml: no such file or directory"
```

## Performance

### Caching

- First load: Reads from disk (~5-10ms)
- Cached load: Memory access (<1ms)
- Thread-safe: Uses RWMutex for concurrent access

### Best Practices

1. **Reuse service instance**: Create once, use many times
2. **Let caching work**: Don't clear cache unnecessarily
3. **Validate early**: Validate variables before rendering
4. **Handle errors**: Always check returned errors

## Logging

Uses `slog` for structured logging:

```go
// Successful operations
slog.Info("template loaded successfully", "name", "bug-fix")
slog.Info("template rendered successfully", "name", "bug-fix", "output_size", 1234)

// Errors
slog.Error("failed to read template file", "name", "missing", "error", err)
slog.Warn("failed to load template, skipping", "name", "invalid", "error", err)
```

## Future Enhancements

Potential improvements:

1. **Hot reload**: Watch template directory for changes
2. **Template inheritance**: Base templates with overrides
3. **Custom functions**: Extend Go template functions
4. **Schema validation**: JSON Schema for variable validation
5. **Metrics**: Track template usage and performance

## Contributing

When adding new templates:

1. Follow the YAML structure defined above
2. Add comprehensive variable definitions
3. Include both required and optional variables with defaults
4. Test with the integration test suite
5. Document template purpose and usage

## License

Part of BusinessOS backend. Internal use only.

---

**Last Updated**: January 2026
**Version**: 1.0.0
**Maintainer**: Backend Team
