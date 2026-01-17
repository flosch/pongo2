# Security and Sandboxing

pongo2 provides several security features to protect against template injection attacks, restrict template capabilities, and isolate template sets from each other.

## DefaultSet Security Warning

> **Important:** The package-level convenience functions (`pongo2.FromFile()`, `pongo2.FromString()`, etc.) use a `DefaultSet` that is configured with a `LocalFilesystemLoader`. This loader has **unrestricted access to the local filesystem**.
>
> For production applications, especially those that:
> - Handle user-generated template content
> - Need to restrict template capabilities
> - Require sandboxing
>
> **Create a custom template set** with appropriate loaders and restrictions instead of using the DefaultSet.

```go
// Instead of using pongo2.FromFile() (which uses DefaultSet), do this:

// Option 1: Use embedded templates (most secure)
//go:embed templates/*
var templateFS embed.FS
loader := pongo2.NewFSLoader(templateFS)
set := pongo2.NewSet("secure", loader)

// Option 2: Use LocalFilesystemLoader with sandbox restrictions
loader := pongo2.MustNewLocalFileSystemLoader("/var/templates")
set := pongo2.NewSet("restricted", loader)
set.BanTag("include")   // Prevent file inclusion
set.BanTag("import")    // Prevent macro imports
set.BanTag("ssi")       // Prevent server-side includes
set.BanTag("extends")   // Prevent template inheritance
set.BanFilter("safe")   // Prevent autoescape bypass

// Now use the custom set
tpl, err := set.FromFile("page.html")
```

See [Template Sets](template-sets.md) for more information on creating and configuring template sets.

## Automatic HTML Escaping

By default, pongo2 automatically escapes HTML special characters in variable output to prevent XSS (Cross-Site Scripting) attacks.

### How Autoescape Works

```django
{{ user_input }}
```

If `user_input` contains `<script>alert('xss')</script>`, the output will be:

```html
&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;
```

### Configuring Autoescape

**Global setting (DefaultSet):**

```go
// Disable autoescape globally
pongo2.SetAutoescape(false)

// Enable autoescape globally (default)
pongo2.SetAutoescape(true)
```

**Per-set setting:**

```go
set := pongo2.NewSet("my-set", loader)
set.SetAutoescape(false)  // Disable for this set only
```

**Per-template block:**

```django
{% autoescape off %}
  {{ trusted_html }}  {# Will NOT be escaped #}
{% endautoescape %}

{% autoescape on %}
  {{ user_input }}  {# Will be escaped #}
{% endautoescape %}
```

**Per-variable:**

```django
{{ trusted_html|safe }}  {# Mark as safe, skip escaping #}
```

### Best Practices for Escaping

1. **Keep autoescape enabled** (default) for user-facing templates
2. **Use `|safe` sparingly** and only for content you fully control
3. **Use `|escapejs`** for JavaScript contexts:
   ```django
   <script>
   var data = "{{ user_input|escapejs }}";
   </script>
   ```
4. **Validate and sanitize** user input before it reaches templates

## Sandbox Features

Template sets support sandboxing to restrict what templates can do.

### Banning Tags

Prevent templates from using specific tags:

```go
set := pongo2.NewSet("sandboxed", loader)

// Ban potentially dangerous tags
set.BanTag("include")   // Prevent file inclusion
set.BanTag("import")    // Prevent macro imports from other files
set.BanTag("ssi")       // Prevent server-side includes
set.BanTag("extends")   // Prevent template inheritance

// Now load templates - they can't use banned tags
tpl, err := set.FromFile("user-template.html")
```

### Banning Filters

Prevent templates from using specific filters:

```go
set := pongo2.NewSet("restricted", loader)

// Ban filters that bypass security
set.BanFilter("safe")      // Prevent bypassing autoescape

// Ban filters based on your security requirements
set.BanFilter("escapejs")  // If you don't want JS output
```

### Important Timing Restriction

**Tags and filters must be banned BEFORE the first template is loaded:**

```go
set := pongo2.NewSet("test", loader)

// This works - no templates loaded yet
set.BanTag("ssi")

// Load a template
tpl, _ := set.FromFile("page.html")

// This FAILS - too late!
err := set.BanTag("include")
// err: "you cannot ban any tags after you've added your first template..."
```

This restriction exists because:
1. Bans are checked at parse time for efficiency
2. Once a template is parsed, it's cached
3. Allowing late bans would be confusing (some templates might have used the tag already)

### Sandbox Example: User-Generated Templates

For user-submitted templates (e.g., email templates, CMS content):

```go
func createSandboxedSet(loader pongo2.TemplateLoader) *pongo2.TemplateSet {
    set := pongo2.NewSet("user-content", loader)

    // Prevent file system access
    set.BanTag("include")
    set.BanTag("import")
    set.BanTag("ssi")
    set.BanTag("extends")

    // Prevent autoescape bypass
    set.BanFilter("safe")

    // Limit to basic control flow only
    // (all other tags remain available: if, for, with, set, etc.)

    return set
}
```

## Template Set Isolation

Each `TemplateSet` is fully isolated with its own:

- **Tag registry** - Custom tags are per-set
- **Filter registry** - Custom filters are per-set
- **Globals** - Global variables are per-set
- **Cache** - Template cache is per-set
- **Banned tags/filters** - Sandbox restrictions are per-set
- **Options** - TrimBlocks, LStripBlocks are per-set
- **Autoescape setting** - Can be configured per-set

### Per-Set Custom Extensions

Register tags and filters specific to a template set:

```go
// Create isolated sets
webSet := pongo2.NewSet("web", webLoader)
emailSet := pongo2.NewSet("email", emailLoader)

// Register filters only for web templates
webSet.RegisterFilter("asset_url", assetUrlFilter)

// Register tags only for email templates
emailSet.RegisterTag("unsubscribe_link", unsubscribeLinkParser)

// These filters/tags are NOT available in the other set
```

### Per-Set Globals

```go
webSet.Globals["site_url"] = "https://example.com"
webSet.Globals["current_year"] = time.Now().Year()

emailSet.Globals["company_name"] = "ACME Corp"
emailSet.Globals["support_email"] = "support@example.com"
```

## Template Loader Security

### Understanding LocalFileSystemLoader

**Important:** The `LocalFileSystemLoader`'s base directory is **NOT a security feature**. It only serves as the root path for resolving relative template paths to absolute paths. Templates can still access files outside the base directory using absolute paths.

```go
// The base directory is for path resolution, NOT security
loader := pongo2.MustNewLocalFileSystemLoader("/var/templates")
set := pongo2.NewSet("app", loader)

// Relative paths are resolved from the base directory
tpl, _ := set.FromFile("pages/home.html")  // Loads /var/templates/pages/home.html

// WARNING: Absolute paths bypass the base directory entirely
tpl, _ := set.FromFile("/etc/passwd")  // This would work if not restricted!
```

**For actual security**, use sandbox features to restrict file inclusion:

```go
set := pongo2.NewSet("sandboxed", loader)
set.BanTag("include")   // Prevent {% include %}
set.BanTag("import")    // Prevent {% import %}
set.BanTag("ssi")       // Prevent {% ssi %}
set.BanTag("extends")   // Prevent {% extends %}
```

### Custom Loaders for Real Security

Implement a custom loader with additional restrictions:

```go
type SecureLoader struct {
    baseDir    string
    allowedExt []string
}

func (l *SecureLoader) Abs(base, name string) string {
    // Resolve path relative to base directory
    resolved := filepath.Join(filepath.Dir(base), name)

    // Ensure result is within allowed directory
    if !strings.HasPrefix(resolved, l.baseDir) {
        return "" // Return empty to indicate not found
    }

    return resolved
}

func (l *SecureLoader) Get(path string) (io.Reader, error) {
    // Check file extension
    ext := filepath.Ext(path)
    allowed := false
    for _, e := range l.allowedExt {
        if ext == e {
            allowed = true
            break
        }
    }
    if !allowed {
        return nil, fmt.Errorf("file extension not allowed: %s", ext)
    }

    return os.Open(path)
}
```

## Macro Recursion Protection

pongo2 limits macro recursion to prevent stack overflow:

```django
{% macro infinite() %}
  {{ infinite() }}  {# This will eventually fail #}
{% endmacro %}
```

The maximum recursion depth is **1000 calls**. When exceeded:

```
maximum recursive macro call depth reached (max is 1000)
```

This protects against:
- Accidental infinite recursion in user templates
- Denial of service via deeply nested macro calls

## Context Security

### Public vs Private Context

```go
type ExecutionContext struct {
    Public  Context  // User-provided data (read-only by convention)
    Private Context  // Internal engine data (forloop, macro args, etc.)
    Shared  Context  // Shared across included templates
}
```

- **Public**: Data you provide via `tpl.Execute(ctx)`. Templates can read but shouldn't modify.
- **Private**: Internal variables like `forloop`. Templates can access but not override user data.
- **Shared**: Persists across `{% include %}` calls.

### Context Identifier Validation

Context keys must be valid identifiers:

```go
// Valid
ctx := pongo2.Context{
    "user":     user,
    "item_1":   item,
}

// Invalid - will panic
ctx := pongo2.Context{
    "'invalid": value,  // Can't start with quote
    "foo-bar":  value,  // Can't contain hyphen
    "foo.bar":  value,  // Can't contain dot
}
```

## Error Handling and Information Leakage

### Production Error Handling

Avoid exposing template internals to users:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    tpl, err := templates.FromCache("page.html")
    if err != nil {
        // Log detailed error internally
        log.Printf("Template error: %v", err)

        // Return generic error to user
        http.Error(w, "Internal Server Error", 500)
        return
    }

    err = tpl.ExecuteWriter(ctx, w)
    if err != nil {
        // Log detailed error internally
        log.Printf("Execution error: %v", err)

        // Don't expose template paths or variable names
        http.Error(w, "Error rendering page", 500)
    }
}
```

### Debug Mode

Set `Debug = false` in production:

```go
set := pongo2.NewSet("production", loader)
set.Debug = false  // Disable debug logging, enable caching

// In development
devSet := pongo2.NewSet("development", loader)
devSet.Debug = true  // Enable debug logging, disable caching
```

## Security Checklist

### For User-Facing Templates

- [ ] Keep autoescape enabled (default)
- [ ] Use `|escapejs` for JavaScript contexts
- [ ] Validate user input before passing to templates
- [ ] Use parameterized queries, not template string concatenation for SQL
- [ ] Review all uses of `|safe` filter

### For User-Generated Templates

- [ ] Create a dedicated sandboxed `TemplateSet`
- [ ] Ban `include`, `import`, `ssi`, `extends` tags
- [ ] Ban `safe` filter
- [ ] Use a restricted template loader
- [ ] Set resource limits (template size, execution time) at the application level
- [ ] Consider banning complex expressions if not needed

### For Production Deployment

- [ ] Set `Debug = false`
- [ ] Don't expose detailed template errors to users
- [ ] Use `FromCache` for performance and to prevent repeated parsing
- [ ] Monitor template execution times
- [ ] Use separate template sets for different trust levels

## Common Vulnerabilities and Mitigations

### Server-Side Template Injection (SSTI)

**Risk**: If user input is directly concatenated into template strings:

```go
// DANGEROUS - NEVER DO THIS
template := "Hello " + userInput + "!"
tpl, _ := pongo2.FromString(template)
```

**Mitigation**: Always pass user input via context:

```go
// SAFE
tpl, _ := pongo2.FromString("Hello {{ name }}!")
tpl.Execute(pongo2.Context{"name": userInput})
```

### Cross-Site Scripting (XSS)

**Risk**: Rendering unsanitized user input.

**Mitigation**:
1. Keep autoescape enabled
2. Use `|escapejs` for JavaScript
3. Validate/sanitize input before it reaches templates

### Path Traversal

**Risk**: User-controlled template names could access unintended files.

**Mitigation**:
1. **Never pass user input directly to `FromFile`** - the `LocalFileSystemLoader` base directory does NOT prevent path traversal (absolute paths bypass it entirely)
2. **Ban file inclusion tags** if templates come from untrusted sources:
   ```go
   set.BanTag("include")
   set.BanTag("import")
   set.BanTag("extends")
   set.BanTag("ssi")
   ```
3. **Use an allowlist** of permitted template names in your application code:
   ```go
   allowedTemplates := map[string]bool{
       "home.html": true,
       "about.html": true,
   }
   if !allowedTemplates[templateName] {
       return errors.New("template not allowed")
   }
   ```
4. **Implement a secure custom loader** that validates and restricts paths (see "Custom Loaders for Real Security" above)

### Denial of Service

**Risk**: Complex templates, deep recursion, or large contexts could exhaust resources.

**Mitigation**:
1. Macro recursion is automatically limited to 1000 calls
2. Implement application-level timeouts for template execution
3. Limit context data size
4. Use template caching (`FromCache`)
