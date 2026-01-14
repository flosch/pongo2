# pongo2 Documentation

> **Note:** Parts of this documentation have been generated with the assistance of LLMs and may contain inaccuracies. If you encounter any errors or issues, please [create an issue](https://github.com/flosch/pongo2/issues) on GitHub.

pongo2 is a Django-syntax compatible template engine for Go.

## Features

- **Django Compatible** - Familiar syntax for Django developers
- **Advanced Expressions** - C-like expressions with arithmetic, comparison, and logical operators
- **Template Inheritance** - `extends`, `block`, and `include` for DRY templates
- **Macros** - Reusable template fragments with arguments and default values
- **Extensible** - Custom filters and tags
- **Sandbox Mode** - Restrict available tags and filters
- **Multiple Loaders** - Load templates from filesystem, embedded files, HTTP, or custom sources

## Quick Start

### Installation

```bash
go get -u github.com/flosch/pongo2/v6
```

### Hello World

```go
package main

import (
    "fmt"
    "github.com/flosch/pongo2/v6"
)

func main() {
    // These functions use the DefaultSet (convenient for simple use cases)
    tpl, _ := pongo2.FromString("Hello {{ name }}!")
    out, _ := tpl.Execute(pongo2.Context{"name": "World"})
    fmt.Println(out) // Hello World!
}
```

### Template File

```go
// Using DefaultSet (simple, but has unrestricted filesystem access)
tpl := pongo2.Must(pongo2.FromFile("templates/page.html"))
err := tpl.ExecuteWriter(pongo2.Context{
    "title": "My Page",
    "items": []string{"one", "two", "three"},
}, responseWriter)
```

> **Note:** The examples above use package-level functions that operate on a `DefaultSet` with unrestricted filesystem access. For production applications requiring sandboxing, see [Template Sets](template-sets.md) and [Security and Sandboxing](security-sandboxing.md).

## Documentation

### Core Concepts

- [Getting Started](getting-started.md) - Installation and basic usage
- [Template Syntax](template-syntax.md) - Variables, expressions, comments

### Reference

- [Filters](filters.md) - Built-in filter reference
- [Tags](tags.md) - Built-in tag reference

### Advanced Topics

- [Template Sets](template-sets.md) - Loaders, caching, globals, sandbox
- [Custom Extensions](custom-extensions.md) - Creating filters and tags
- [Security and Sandboxing](security-sandboxing.md) - Autoescape, sandboxing, best practices
- [Changelog](../CHANGELOG.md) - Version history and release notes

## Differences from Django

### Go-Specific Behavior

| Feature | Django | pongo2 |
|---------|--------|--------|
| Date format | `"d/m/Y"` | `"02/01/2006"` (Go format) |
| stringformat | `"%.2f"` (Python) | `"%.2f"` (Go fmt.Sprintf) |
| forloop variables | `forloop.counter` | `forloop.Counter` (capitalized) |
| Struct fields | N/A | Case-sensitive, exported only |

### Additional Features

- **Macros** - `{% macro %}` with `export` keyword
- **set tag** - `{% set var = value %}`
- **Expressions** - `{{ a + b * c }}`
- **Array literals** - `{% for x in [1, 2, 3] %}`
- **reversed/sorted** - `{% for x in items reversed sorted %}`

### Not Implemented

- `csrf_token` - Web framework specific
- `load` - Python specific
- `url` - Web framework specific

## Add-ons

### Official

- [pongo2-addons](https://github.com/flosch/pongo2-addons) - Additional filters:
  - `filesizeformat`, `slugify`, `markdown`
  - `timesince`, `timeuntil`, `naturaltime`, `naturalday`
  - `intcomma`, `ordinal`
  - `truncatesentences`, `truncatesentences_html`

### Third-Party Integrations

- [beego-pongo2](https://github.com/oal/beego-pongo2) - Beego integration
- [ginpongo2](https://github.com/ngerakines/ginpongo2) - Gin middleware
- [pongo2gin](https://gitlab.com/go-box/pongo2gin) - Alternative Gin renderer
- [macaron-pongo2](https://github.com/macaron-contrib/pongo2) - Macaron support
- [Iris](https://github.com/kataras/iris) - Built-in pongo2 support
- [pongorenderer](https://github.com/siredwin/pongorenderer) - Echo renderer
- [p2cli](https://github.com/wrouesnel/p2cli) - Command-line templating

## Example Template

```django
{% extends "base.html" %}

{% block title %}{{ page.title }}{% endblock %}

{% block content %}
<article>
  <h1>{{ page.title|title }}</h1>
  <p class="meta">
    By {{ page.author.name }} on {{ page.date|date:"January 2, 2006" }}
  </p>

  {% if page.tags %}
  <ul class="tags">
    {% for tag in page.tags %}
    <li>{{ tag }}</li>
    {% endfor %}
  </ul>
  {% endif %}

  <div class="content">
    {{ page.content|safe }}
  </div>
</article>

{% include "comments.html" with post=page %}
{% endblock %}
```

## Version

Current version: **6.0.0**

## Support

- [GitHub Issues](https://github.com/flosch/pongo2/issues)
- [GoDoc](https://pkg.go.dev/github.com/flosch/pongo2/v6)

## License

MIT License
