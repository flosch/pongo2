# Getting Started with pongo2

pongo2 is a Django-syntax compatible template engine for Go. This guide will help you get up and running quickly.

## Installation

```bash
go get -u github.com/flosch/pongo2/v7
```

## Basic Usage

### Rendering a Template from a String

```go
package main

import (
    "fmt"
    "github.com/flosch/pongo2/v7"
)

func main() {
    // Compile the template
    tpl, err := pongo2.FromString("Hello {{ name|capfirst }}!")
    if err != nil {
        panic(err)
    }

    // Execute with context
    out, err := tpl.Execute(pongo2.Context{"name": "florian"})
    if err != nil {
        panic(err)
    }

    fmt.Println(out) // Output: Hello Florian!
}
```

### Rendering a Template from a File

```go
package main

import (
    "github.com/flosch/pongo2/v7"
    "net/http"
)

// Pre-compile templates at startup for better performance
var tplExample = pongo2.Must(pongo2.FromFile("templates/example.html"))

func examplePage(w http.ResponseWriter, r *http.Request) {
    err := tplExample.ExecuteWriter(pongo2.Context{
        "query": r.FormValue("query"),
    }, w)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func main() {
    http.HandleFunc("/", examplePage)
    http.ListenAndServe(":8080", nil)
}
```

## The Context

A `Context` is a map that provides variables to your template:

```go
ctx := pongo2.Context{
    "name":     "John",
    "age":      30,
    "is_admin": true,
    "items":    []string{"apple", "banana", "cherry"},
    "user": struct {
        Name  string
        Email string
    }{"John Doe", "john@example.com"},
}

out, err := tpl.Execute(ctx)
```

In your template, access these values:

```django
Hello {{ name }}!
You are {{ age }} years old.
{% if is_admin %}You are an admin.{% endif %}

Items:
{% for item in items %}
  - {{ item }}
{% endfor %}

User: {{ user.Name }} ({{ user.Email }})
```

## Template Syntax Overview

### Variables

Variables are enclosed in double curly braces:

```django
{{ variable }}
{{ user.name }}
{{ items.0 }}
{{ dict.key }}
```

### Filters

Filters transform variable output:

```django
{{ name|upper }}
{{ text|truncatewords:30 }}
{{ price|floatformat:2 }}
{{ items|join:", " }}
```

### Tags

Tags provide logic and control flow:

```django
{% if condition %}...{% endif %}
{% for item in items %}...{% endfor %}
{% include "header.html" %}
{% extends "base.html" %}
```

### Comments

```django
{# This is a comment #}

{% comment %}
  This is a
  multi-line comment
{% endcomment %}
```

## Template Caching

For production, use `FromCache` to avoid recompiling templates on every request:

```go
tpl, err := pongo2.FromCache("templates/page.html")
```

In debug mode, set `Debug` to true to disable caching:

```go
pongo2.DefaultSet.Debug = true
```

## Execution Methods

pongo2 provides several ways to execute templates:

```go
// Returns rendered template as string
out, err := tpl.Execute(ctx)

// Returns rendered template as []byte
bytes, err := tpl.ExecuteBytes(ctx)

// Writes to an io.Writer (buffered, safe on error)
err := tpl.ExecuteWriter(ctx, w)

// Writes to an io.Writer (unbuffered, faster but partial output on error)
err := tpl.ExecuteWriterUnbuffered(ctx, w)

// Execute specific blocks only
blocks, err := tpl.ExecuteBlocks(ctx, []string{"content", "sidebar"})
```

## Global Variables

Set variables available to all templates in a set:

```go
pongo2.Globals["site_name"] = "My Website"
pongo2.Globals["current_year"] = time.Now().Year()

// Or use a custom TemplateSet
mySet := pongo2.NewSet("custom", pongo2.DefaultLoader)
mySet.Globals["api_url"] = "https://api.example.com"
```

## Autoescape

By default, pongo2 automatically escapes HTML in variable output. Control this behavior:

```go
// Disable globally
pongo2.SetAutoescape(false)

// Or in templates
{% autoescape off %}
  {{ raw_html }}
{% endautoescape %}

// Or mark individual values as safe
{{ html_content|safe }}
```

## Error Handling

pongo2 provides detailed error information:

```go
tpl, err := pongo2.FromString("{{ invalid syntax }}")
if err != nil {
    // err contains file, line, column information
    fmt.Printf("Template error: %v\n", err)
}
```

## Next Steps

- [Template Syntax](template-syntax.md) - Complete syntax reference
- [Filters](filters.md) - All available filters
- [Tags](tags.md) - All available tags
- [Template Sets](template-sets.md) - Advanced configuration
- [Custom Extensions](custom-extensions.md) - Create your own filters and tags
