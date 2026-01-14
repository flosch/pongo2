# Template Syntax

pongo2 uses Django-compatible template syntax. This document covers all syntax elements.

## Variables

Variables output values from the context:

```django
{{ variable }}
```

### Accessing Attributes

Access struct fields, map keys, and slice/array indices:

```django
{{ user.name }}          {# Struct field or map key #}
{{ user.Name }}          {# Go struct fields are case-sensitive #}
{{ items.0 }}            {# First element of slice/array #}
{{ dict.key }}           {# Map key access #}
{{ user["name"] }}       {# Subscript notation #}
{{ items[index] }}       {# Dynamic index #}
```

### Calling Methods

Call methods on objects:

```django
{{ user.GetFullName() }}
{{ items.Len() }}
{{ now.Format("2006-01-02") }}
```

Methods can receive arguments:

```django
{{ math.Add(1, 2) }}
{{ string.Replace("hello", "h", "j") }}
```

**Method requirements:**
- Must be exported (capitalized)
- Must return 1 or 2 values
- If 2 values, second must be `error`
- Can accept `*pongo2.Value` or concrete types
- Can optionally accept `*pongo2.ExecutionContext` as first parameter

### Filters

Filters modify variable output:

```django
{{ name|upper }}                    {# JOHN #}
{{ name|lower }}                    {# john #}
{{ text|truncatewords:5 }}          {# First five words... #}
{{ price|floatformat:2 }}           {# 19.99 #}
{{ list|join:", " }}                {# apple, banana, cherry #}
```

Chain multiple filters:

```django
{{ name|lower|capfirst }}           {# John #}
{{ html|striptags|truncatechars:100 }}
```

## Expressions

pongo2 supports C-like expressions in templates.

### Arithmetic Operators

```django
{{ 1 + 2 }}                         {# 3 #}
{{ 10 - 3 }}                        {# 7 #}
{{ 4 * 5 }}                         {# 20 #}
{{ 20 / 4 }}                        {# 5 #}
{{ 17 % 5 }}                        {# 2 (modulo) #}
{{ 2 ^ 10 }}                        {# 1024 (power) #}
```

### Comparison Operators

```django
{{ x == y }}
{{ x != y }}
{{ x < y }}
{{ x > y }}
{{ x <= y }}
{{ x >= y }}
```

Also supports `<>` as alternative to `!=`.

### Logical Operators

```django
{{ x and y }}
{{ x or y }}
{{ not x }}
{{ x && y }}                        {# Alternative syntax #}
{{ x || y }}                        {# Alternative syntax #}
{{ !x }}                            {# Alternative syntax #}
```

### The `in` Operator

Check if a value is contained in another:

```django
{% if "a" in "abc" %}yes{% endif %}             {# String contains #}
{% if key in dict %}yes{% endif %}              {# Map contains key #}
{% if item in list %}yes{% endif %}             {# Slice contains #}
{% if !(key in dict) %}not found{% endif %}     {# Negation #}
```

### Grouping with Parentheses

```django
{{ (1 + 2) * 3 }}                   {# 9 #}
{% if (a or b) and c %}...{% endif %}
```

### Array Literals

Define inline arrays:

```django
{% for item in [1, 2, 3] %}{{ item }}{% endfor %}
{% for name in ["Alice", "Bob", "Charlie"] %}...{% endfor %}
```

## Tags

Tags control template logic and are enclosed in `{% %}`:

```django
{% if condition %}
  Content shown if true
{% endif %}
```

### Block Tags

Many tags have opening and closing parts:

```django
{% if condition %}
  ...
{% elif other_condition %}
  ...
{% else %}
  ...
{% endif %}

{% for item in items %}
  ...
{% empty %}
  No items found.
{% endfor %}
```

## Comments

### Single-line Comments

```django
{# This is a comment and won't appear in output #}
```

### Multi-line Comments

```django
{% comment %}
  This entire block is a comment.
  It can span multiple lines.
  Useful for temporarily disabling template code.
{% endcomment %}
```

## Whitespace Control

### TrimBlocks and LStripBlocks

Control whitespace around block tags:

```go
set := pongo2.NewSet("name", pongo2.DefaultLoader)
set.Options.TrimBlocks = true   // Remove first newline after block tags
set.Options.LStripBlocks = true // Strip leading whitespace from block tags
```

### Manual Whitespace Control

Use `-` to trim whitespace:

```django
{%- if condition -%}   {# Trims whitespace on both sides #}
  content
{%- endif -%}

{{- variable -}}        {# Trims whitespace around variable #}
```

### Spaceless Tag

Remove whitespace between HTML tags:

```django
{% spaceless %}
  <ul>
    <li>Item 1</li>
    <li>Item 2</li>
  </ul>
{% endspaceless %}
```

Output: `<ul><li>Item 1</li><li>Item 2</li></ul>`

## Escaping

### Automatic HTML Escaping

By default, pongo2 escapes HTML special characters:

```django
{{ "<script>alert('xss')</script>" }}
{# Output: &lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt; #}
```

### Disabling Escaping

```django
{# For a single variable #}
{{ trusted_html|safe }}

{# For a block #}
{% autoescape off %}
  {{ raw_html }}
{% endautoescape %}
```

### JavaScript Escaping

For JavaScript contexts:

```django
<script>
var data = "{{ user_input|escapejs }}";
</script>
```

## Literals

### Strings

```django
{{ "hello world" }}
{{ 'single quotes work too' }}
```

### Numbers

```django
{{ 42 }}
{{ 3.14 }}
{{ -17 }}
{{ -2.5 }}
```

### Booleans

```django
{{ true }}
{{ false }}
```

## Differences from Django

### forloop Variables

In pongo2, forloop variables are capitalized:

```django
{% for item in items %}
  {{ forloop.Counter }}      {# 1-indexed counter #}
  {{ forloop.Counter0 }}     {# 0-indexed counter #}
  {{ forloop.Revcounter }}   {# Reverse counter #}
  {{ forloop.Revcounter0 }}  {# Reverse 0-indexed #}
  {{ forloop.First }}        {# True on first iteration #}
  {{ forloop.Last }}         {# True on last iteration #}
  {{ forloop.Parentloop }}   {# Access parent loop #}
{% endfor %}
```

### Date/Time Format

pongo2 uses Go's time format, not Python's:

```django
{{ now|date:"2006-01-02" }}
{{ now|time:"15:04:05" }}
{{ now|date:"Monday, January 2, 2006" }}
```

Reference: https://golang.org/pkg/time/#Time.Format

### stringformat Filter

Uses Go's `fmt.Sprintf` format:

```django
{{ 3.14159|stringformat:"%.2f" }}   {# 3.14 #}
{{ 42|stringformat:"%05d" }}         {# 00042 #}
```

## Reserved Keywords

The following are reserved keywords:

- `in`
- `and`
- `or`
- `not`
- `true`
- `false`
- `as`
- `export`
