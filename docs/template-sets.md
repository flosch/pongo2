# Template Sets

Template sets allow you to group templates with shared configuration, globals, and loaders. This is useful for separating different types of templates (e.g., web templates vs. email templates).

## Understanding Template Sets

A `TemplateSet` is a container that holds:

- **Template loaders** - How templates are located and read
- **Global variables** - Variables available to all templates in the set
- **Tag registry** - Available tags (built-in + custom)
- **Filter registry** - Available filters (built-in + custom)
- **Banned tags/filters** - Sandbox restrictions
- **Configuration** - Debug mode, autoescape, whitespace options
- **Template cache** - Compiled templates for performance

Each template set is **fully isolated** from others. Custom tags, filters, globals, and sandbox restrictions in one set do not affect other sets.

## DefaultSet and Package-Level Functions

pongo2 provides a `DefaultSet` for convenience. The package-level functions are shortcuts that operate on this default set:

```go
// These package-level functions all use DefaultSet internally:
pongo2.FromFile("template.html")       // → DefaultSet.FromFile()
pongo2.FromString("Hello {{ name }}!") // → DefaultSet.FromString()
pongo2.FromBytes([]byte("..."))        // → DefaultSet.FromBytes()
pongo2.FromCache("template.html")      // → DefaultSet.FromCache()
pongo2.RenderTemplateString(s, ctx)    // → DefaultSet.RenderTemplateString()
pongo2.RenderTemplateFile(fn, ctx)     // → DefaultSet.RenderTemplateFile()

// These also operate on DefaultSet:
pongo2.RegisterFilter("name", fn)      // → DefaultSet.RegisterFilter()
pongo2.ReplaceFilter("name", fn)       // → DefaultSet.ReplaceFilter()
pongo2.RegisterTag("name", parser)     // → DefaultSet.RegisterTag()
pongo2.ReplaceTag("name", parser)      // → DefaultSet.ReplaceTag()
pongo2.SetAutoescape(false)            // → DefaultSet.SetAutoescape()

// Global variables for DefaultSet
pongo2.Globals["site_name"] = "My Website"
```

### DefaultSet Configuration

The `DefaultSet` is initialized with:

- **`DefaultLoader`** - A `LocalFilesystemLoader` with an empty base directory (uses current working directory)
- **Autoescape enabled** - HTML escaping is on by default
- **All built-in tags and filters** - No restrictions

```go
// This is essentially how DefaultSet is created internally:
DefaultLoader = MustNewLocalFileSystemLoader("")
DefaultSet = NewSet("default", DefaultLoader)
```

### Security Consideration: DefaultSet Uses Local Filesystem

> **Important:** The `DefaultSet` uses a `LocalFilesystemLoader` that has unrestricted access to the local filesystem. For production applications, especially those handling user-generated content or requiring sandboxing, consider creating a custom template set with appropriate restrictions.

```go
// For better security, create your own template set:
loader := pongo2.NewFSLoader(embeddedFS)  // Use embedded templates
set := pongo2.NewSet("secure", loader)
set.BanTag("include")                      // Restrict file inclusion
set.BanTag("ssi")
set.BanFilter("safe")                      // Prevent autoescape bypass
```

See [Security and Sandboxing](security-sandboxing.md) for comprehensive security guidance.

## Creating Custom Template Sets

```go
// Create a new set with a loader
loader := pongo2.MustNewLocalFileSystemLoader("/path/to/templates")
mySet := pongo2.NewSet("email-templates", loader)

// Configure the set
mySet.Debug = true
mySet.Globals["company_name"] = "ACME Corp"

// Load templates from this set
tpl, err := mySet.FromFile("welcome.html")
```

## Template Loaders

Loaders define how templates are located and read.

### LocalFilesystemLoader

Loads templates from the local filesystem:

```go
// Without base directory (uses working directory)
loader, err := pongo2.NewLocalFileSystemLoader("")

// With base directory
loader, err := pongo2.NewLocalFileSystemLoader("/var/templates")

// Panics on error (useful for init)
loader := pongo2.MustNewLocalFileSystemLoader("/var/templates")
```

With a base directory, all template paths are resolved relative to it:

```go
loader := pongo2.MustNewLocalFileSystemLoader("/var/templates")
set := pongo2.NewSet("web", loader)

// Loads /var/templates/pages/home.html
tpl, err := set.FromFile("pages/home.html")
```

### FSLoader

Supports Go's `fs.FS` interface (Go 1.16+):

```go
import "embed"

//go:embed templates/*
var templateFS embed.FS

func main() {
    loader := pongo2.NewFSLoader(templateFS)
    set := pongo2.NewSet("embedded", loader)

    tpl, err := set.FromFile("templates/page.html")
}
```

Works with any `fs.FS` implementation:

```go
// os.DirFS
loader := pongo2.NewFSLoader(os.DirFS("/var/templates"))

// embed.FS
//go:embed templates
var templates embed.FS
loader := pongo2.NewFSLoader(templates)
```

### Multiple Loaders

A template set can have multiple loaders. Templates are resolved in order:

```go
loader1 := pongo2.MustNewLocalFileSystemLoader("/var/templates/custom")
loader2 := pongo2.MustNewLocalFileSystemLoader("/var/templates/default")

set := pongo2.NewSet("multi", loader1, loader2)
set.AddLoader(anotherLoader) // Add more later

// Tries loader1 first, then loader2
tpl, err := set.FromFile("page.html")
```

### Custom Loaders

Implement the `TemplateLoader` interface:

```go
type TemplateLoader interface {
    // Abs calculates the absolute path to a template
    // base is the parent template's path (for includes/extends)
    // name is the requested template name
    Abs(base, name string) string

    // Get returns a reader for the template content
    Get(path string) (io.Reader, error)
}
```

Example database loader:

```go
type DBLoader struct {
    db *sql.DB
}

func (l *DBLoader) Abs(base, name string) string {
    return name // Templates are identified by name only
}

func (l *DBLoader) Get(path string) (io.Reader, error) {
    var content string
    err := l.db.QueryRow("SELECT content FROM templates WHERE name = ?", path).Scan(&content)
    if err != nil {
        return nil, err
    }
    return strings.NewReader(content), nil
}
```

## Template Set Options

### Debug Mode

When enabled:
- Templates are recompiled on every request (no caching)
- `ExecutionContext.Logf()` outputs to STDOUT

```go
set.Debug = true
```

### TrimBlocks and LStripBlocks

Control whitespace around block tags:

```go
set.Options.TrimBlocks = true   // Remove first newline after block tags
set.Options.LStripBlocks = true // Strip leading whitespace before block tags
```

Example:

```django
{% if true %}
Hello
{% endif %}
```

Without options: `\nHello\n`
With TrimBlocks: `Hello\n`
With LStripBlocks: `\nHello\n`
With both: `Hello\n`

## Global Variables

Variables available to all templates in a set:

```go
set.Globals["site_name"] = "My Website"
set.Globals["current_year"] = time.Now().Year()
set.Globals["version"] = "1.0.0"

// Functions are also allowed
set.Globals["format_price"] = func(price float64) string {
    return fmt.Sprintf("$%.2f", price)
}
```

Access in templates:

```django
<footer>&copy; {{ current_year }} {{ site_name }}</footer>
{{ format_price(19.99) }}
```

## Template Caching

### FromCache

Caches compiled templates for production use:

```go
// First call compiles and caches
tpl, err := set.FromCache("page.html")

// Subsequent calls return cached template
tpl, err := set.FromCache("page.html")
```

When `Debug` is true, caching is disabled.

### CleanCache

Clear the template cache:

```go
// Clear specific templates
set.CleanCache("page.html", "header.html")

// Clear entire cache
set.CleanCache()
```

## Autoescape

Control automatic HTML escaping:

```go
// Configure autoescape for a template set
set.SetAutoescape(false)  // Disable HTML escaping
set.SetAutoescape(true)   // Enable HTML escaping (default)

// Or use the global function for DefaultSet
pongo2.SetAutoescape(false)
```

## Per-Set Tags and Filters

Each template set has its own tag and filter registries. This allows different template sets to have different custom extensions.

### Registering Set-Specific Filters

```go
// Register a filter only for this set
err := set.RegisterFilter("custom_filter", myFilterFunc)

// Replace an existing filter in this set only
err := set.ReplaceFilter("upper", myUpperFilter)

// Check if a filter exists in this set
if set.FilterExists("custom_filter") {
    // Filter is available
}

// Apply a filter programmatically using this set's registry
result, err := set.ApplyFilter("upper", pongo2.AsValue("hello"), nil)

// Panic version
result := set.MustApplyFilter("upper", pongo2.AsValue("hello"), nil)
```

### Registering Set-Specific Tags

```go
// Register a tag only for this set
err := set.RegisterTag("custom_tag", myTagParser)

// Replace an existing tag in this set only
err := set.ReplaceTag("for", myForParser)

// Check if a tag exists in this set
if set.TagExists("custom_tag") {
    // Tag is available
}
```

### Isolation Example

```go
// Create two isolated template sets
webSet := pongo2.NewSet("web", webLoader)
emailSet := pongo2.NewSet("email", emailLoader)

// Register a filter only for web templates
webSet.RegisterFilter("asset_url", assetUrlFilter)

// Register a tag only for email templates
emailSet.RegisterTag("unsubscribe", unsubscribeTagParser)

// asset_url filter is NOT available in emailSet
// unsubscribe tag is NOT available in webSet
```

## Sandbox Features

Restrict what templates can do by banning specific tags and filters. For comprehensive security guidance, see [Security and Sandboxing](security-sandboxing.md).

### Banning Tags

```go
set := pongo2.NewSet("restricted", loader)

// Ban dangerous tags before loading any templates
set.BanTag("include")  // Disable file inclusion
set.BanTag("import")   // Disable macro imports
set.BanTag("ssi")      // Disable server-side includes

// Now load templates
tpl, err := set.FromFile("user-content.html")
```

### Banning Filters

```go
set.BanFilter("safe")  // Prevent bypassing autoescape
```

### Restrictions

- Tags and filters must be banned BEFORE the first template is loaded
- Once a template is loaded, the set is "locked"
- Attempting to ban after loading returns an error

```go
set := pongo2.NewSet("test", loader)

// This works
set.BanTag("ssi")

// Load a template
tpl, _ := set.FromFile("page.html")

// This returns an error - too late!
err := set.BanTag("include")
// err: "you cannot ban any tags after you've added your first template"
```

## Built-in pongo2 Context

Templates have access to pongo2 metadata via `pongo2`:

```django
pongo2 version: {{ pongo2.version }}
```

## Complete Example

```go
package main

import (
    "log"
    "net/http"
    "time"

    "github.com/flosch/pongo2/v7"
)

var templates *pongo2.TemplateSet

func init() {
    // Create loader
    loader := pongo2.MustNewLocalFileSystemLoader("./templates")

    // Create template set
    templates = pongo2.NewSet("web", loader)

    // Configure options
    templates.Options.TrimBlocks = true
    templates.Options.LStripBlocks = true

    // Set global variables
    templates.Globals["site_name"] = "My App"
    templates.Globals["current_year"] = time.Now().Year()

    // Helper functions
    templates.Globals["asset_url"] = func(path string) string {
        return "/static/" + path
    }

    // Sandbox restrictions
    templates.BanTag("ssi")
    templates.BanTag("include") // Force explicit template paths
}

func handler(w http.ResponseWriter, r *http.Request) {
    tpl, err := templates.FromCache("pages/home.html")
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    err = tpl.ExecuteWriter(pongo2.Context{
        "user":  getCurrentUser(r),
        "items": getItems(),
    }, w)
    if err != nil {
        log.Printf("Template error: %v", err)
    }
}

func main() {
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}
```
