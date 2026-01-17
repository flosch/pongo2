# Tags Reference

Tags provide logic and control flow in templates. They are enclosed in `{% %}`:

```django
{% tagname %}
{% tagname argument %}
{% tagname %}...{% endtagname %}
```

## Control Flow Tags

### if / elif / else / endif

Conditional display based on expressions.

```django
{% if user.is_admin %}
  <p>Welcome, admin!</p>
{% elif user.is_member %}
  <p>Welcome, member!</p>
{% else %}
  <p>Welcome, guest!</p>
{% endif %}
```

Supports complex expressions:

```django
{% if user.age >= 18 and user.verified %}...{% endif %}
{% if count > 0 or show_empty %}...{% endif %}
{% if not user.is_blocked %}...{% endif %}
{% if (a or b) and c %}...{% endif %}
{% if item in list %}...{% endif %}
```

### for / empty / endfor

Iterates over sequences.

```django
{% for item in items %}
  <li>{{ item }}</li>
{% empty %}
  <li>No items found.</li>
{% endfor %}
```

**Iterating over maps:**

```django
{% for key, value in mymap %}
  {{ key }}: {{ value }}
{% endfor %}
```

**Loop modifiers:**

```django
{% for item in items reversed %}...{% endfor %}
{% for item in items sorted %}...{% endfor %}
{% for item in items reversed sorted %}...{% endfor %}
```

**Loop variables (forloop):**

| Variable | Description |
|----------|-------------|
| `forloop.Counter` | 1-indexed counter |
| `forloop.Counter0` | 0-indexed counter |
| `forloop.Revcounter` | Reverse counter (ends at 1) |
| `forloop.Revcounter0` | Reverse counter (ends at 0) |
| `forloop.First` | True on first iteration |
| `forloop.Last` | True on last iteration |
| `forloop.Parentloop` | Parent loop in nested loops |

```django
{% for item in items %}
  {% if forloop.First %}<ul>{% endif %}
  <li>{{ forloop.Counter }}. {{ item }}</li>
  {% if forloop.Last %}</ul>{% endif %}
{% endfor %}
```

### ifequal / endifequal

Compares two values for equality. (Prefer `{% if a == b %}` instead.)

```django
{% ifequal user.name "admin" %}
  Admin user
{% endifequal %}
```

### ifnotequal / endifnotequal

Compares two values for inequality.

```django
{% ifnotequal status "active" %}
  Inactive
{% endifnotequal %}
```

### ifchanged / endifchanged

Outputs content only when value changes within a loop.

```django
{% for item in items %}
  {% ifchanged item.category %}
    <h2>{{ item.category }}</h2>
  {% endifchanged %}
  <p>{{ item.name }}</p>
{% endfor %}
```

## Template Inheritance Tags

### extends

Inherits from a parent template. Must be the first tag in the template.

```django
{% extends "base.html" %}
```

### block / endblock

Defines overridable sections in templates.

**Parent template (base.html):**

```django
<!DOCTYPE html>
<html>
<head>
  <title>{% block title %}Default Title{% endblock %}</title>
</head>
<body>
  {% block content %}{% endblock %}
</body>
</html>
```

**Child template:**

```django
{% extends "base.html" %}

{% block title %}My Page{% endblock %}

{% block content %}
  <h1>Hello World</h1>
{% endblock %}
```

**Accessing parent block content:**

```django
{% block sidebar %}
  {{ block.Super }}
  <p>Additional sidebar content</p>
{% endblock %}
```

Optional block name in endblock:

```django
{% block content %}
  ...
{% endblock content %}
```

### include

Includes another template.

```django
{% include "header.html" %}
```

**With custom context:**

```django
{% include "user_card.html" with user=current_user show_email=true %}
```

**Isolating context (only keyword):**

```django
{% include "widget.html" with title="Hello" only %}
```

**Conditional include:**

```django
{% include "optional.html" if_exists %}
{% include template_name if_exists %}
```

**Dynamic templates:**

```django
{% include template_variable %}
{% include "partials/"|add:partial_name|add:".html" %}
```

### ssi (Server-Side Include)

Includes a file from the filesystem.

```django
{% ssi "/path/to/file.txt" %}
```

## Macro Tags

### macro / endmacro

Defines reusable template fragments.

```django
{% macro button(text, url, class="primary") %}
  <a href="{{ url }}" class="btn btn-{{ class }}">{{ text }}</a>
{% endmacro %}

{{ button("Click me", "/action") }}
{{ button("Delete", "/delete", class="danger") }}
```

**Exported macros:**

```django
{% macro user_avatar(user) export %}
  <img src="{{ user.avatar_url }}" alt="{{ user.name }}">
{% endmacro %}
```

### import

Imports macros from another file.

```django
{% import "macros.html" button, form_field %}
{% import "macros.html" button as btn %}

{{ btn("Click", "/") }}
```

## Variable Tags

### set

Creates or updates a variable.

```django
{% set name = "John" %}
{% set full_name = first_name + " " + last_name %}
{% set total = price * quantity %}
```

### with / endwith

Creates scoped variables.

**New style:**

```django
{% with total=price*quantity tax=total*0.1 %}
  Total: {{ total }}
  Tax: {{ tax }}
{% endwith %}
```

**Old style (Django-compatible):**

```django
{% with price*quantity as total %}
  Total: {{ total }}
{% endwith %}
```

### firstof

Outputs the first non-false value.

```django
{% firstof var1 var2 var3 "fallback" %}
```

Equivalent to:

```django
{% if var1 %}{{ var1 }}{% elif var2 %}{{ var2 }}{% elif var3 %}{{ var3 }}{% else %}fallback{% endif %}
```

### cycle

Cycles through values on each iteration.

```django
{% for item in items %}
  <tr class="{% cycle "odd" "even" %}">
    <td>{{ item }}</td>
  </tr>
{% endfor %}
```

**Named cycles:**

```django
{% cycle "red" "green" "blue" as rowcolor silent %}
<tr class="{{ rowcolor }}">...</tr>
{% cycle rowcolor %}
```

## Output Control Tags

### autoescape / endautoescape

Controls automatic HTML escaping.

```django
{% autoescape off %}
  {{ raw_html }}
{% endautoescape %}

{% autoescape on %}
  {{ user_input }}  {# Will be escaped #}
{% endautoescape %}
```

### filter / endfilter

Applies filters to an entire block.

```django
{% filter upper %}
  This text will be uppercase.
{% endfilter %}

{% filter escape|linebreaks %}
  Line 1
  Line 2
{% endfilter %}
```

### spaceless / endspaceless

Removes whitespace between HTML tags.

```django
{% spaceless %}
  <ul>
    <li>One</li>
    <li>Two</li>
  </ul>
{% endspaceless %}
```

Output: `<ul><li>One</li><li>Two</li></ul>`

### verbatim / endverbatim

Outputs content without processing template syntax.

```django
{% verbatim %}
  {{ this will not be processed }}
  {% neither will this %}
{% endverbatim %}
```

Useful for client-side templates (Vue.js, Angular, etc.).

## Utility Tags

### comment / endcomment

Multi-line comments that don't appear in output.

```django
{% comment %}
  This is a comment.
  It can span multiple lines.
  {% if something %}This won't be evaluated{% endif %}
{% endcomment %}
```

### now

Outputs the current date/time.

```django
{% now "2006-01-02" %}
{% now "Monday, January 2, 2006 at 15:04" %}
```

Uses Go's time format strings.

### lorem

Generates placeholder text.

```django
{% lorem %}                 {# One paragraph #}
{% lorem 3 %}               {# Three paragraphs #}
{% lorem 5 w %}             {# Five words #}
{% lorem 2 p %}             {# Two HTML paragraphs #}
{% lorem 3 b %}             {# Three plain paragraphs (default) #}
{% lorem 2 p random %}      {# Random paragraphs #}
```

Methods:
- `w` - words
- `p` - HTML paragraphs (`<p>...</p>`)
- `b` - plain text paragraphs (default)

### templatetag

Outputs template syntax characters.

```django
{% templatetag openblock %}     {# {% #}
{% templatetag closeblock %}    {# %} #}
{% templatetag openvariable %}  {# {{ #}
{% templatetag closevariable %} {# }} #}
{% templatetag openbrace %}     {# { #}
{% templatetag closebrace %}    {# } #}
{% templatetag opencomment %}   {# {# #}
{% templatetag closecomment %}  {# #} #}
```

### widthratio

Calculates a ratio for creating bar charts, etc.

```django
<div style="width: {% widthratio current_value max_value 100 %}px;"></div>
```

Formula: `(value / max_value) * max_width`

```django
{% widthratio 50 100 200 %}  {# Outputs: 100 #}
```

## Tags Not Implemented

The following Django tags are not implemented in pongo2:

- **csrf_token** - Web framework specific
- **load** - Python specific
- **url** - Web framework specific
- **debug** - Not yet implemented
- **regroup** - Python specific
