# Macros

Macros are reusable template fragments that work like functions. They accept arguments, can have default values, and return rendered content. Macros are one of pongo2's most powerful features for creating DRY (Don't Repeat Yourself) templates.

## Defining Macros

Use the `{% macro %}` tag to define a macro:

```django
{% macro button(text, url) %}
  <a href="{{ url }}" class="btn">{{ text }}</a>
{% endmacro %}
```

### Syntax

```django
{% macro name(arg1, arg2, ...) %}
  ...content...
{% endmacro %}
```

- **name** - The macro's identifier (must be a valid identifier)
- **arguments** - Comma-separated list of parameter names
- **content** - The template content to render when called

## Calling Macros

Call macros like functions using `{{ }}`:

```django
{% macro greeting(name) %}
  Hello, {{ name }}!
{% endmacro %}

{{ greeting("World") }}
{{ greeting("Alice") }}
```

Output:

```
Hello, World!
Hello, Alice!
```

### Passing Variables

You can pass any expression as an argument:

```django
{% macro user_card(user) %}
  <div class="card">
    <h3>{{ user.name }}</h3>
    <p>{{ user.email }}</p>
  </div>
{% endmacro %}

{{ user_card(current_user) }}
{{ user_card(admin) }}
```

### Multiple Arguments

```django
{% macro link(text, url, class) %}
  <a href="{{ url }}" class="{{ class }}">{{ text }}</a>
{% endmacro %}

{{ link("Home", "/", "nav-link") }}
{{ link("About", "/about", "nav-link active") }}
```

## Default Values

Arguments can have default values using `=`:

```django
{% macro button(text, url, style="primary", size="md") %}
  <a href="{{ url }}" class="btn btn-{{ style }} btn-{{ size }}">
    {{ text }}
  </a>
{% endmacro %}

{{ button("Click", "/action") }}
{{ button("Delete", "/delete", style="danger") }}
{{ button("Small", "/small", size="sm") }}
{{ button("Large Danger", "/alert", style="danger", size="lg") }}
```

Output:

```html
<a href="/action" class="btn btn-primary btn-md">Click</a>
<a href="/delete" class="btn btn-danger btn-md">Delete</a>
<a href="/small" class="btn btn-primary btn-sm">Small</a>
<a href="/alert" class="btn btn-danger btn-lg">Large Danger</a>
```

### Default Value Expressions

Default values can be any valid expression:

```django
{% macro greet(name, greeting="Hello", punctuation="!") %}
  {{ greeting }}, {{ name }}{{ punctuation }}
{% endmacro %}

{% macro repeat(text, times=3) %}
  {% for i in range(times) %}{{ text }}{% endfor %}
{% endmacro %}

{% macro format_price(amount, currency="$", decimals=2) %}
  {{ currency }}{{ amount|floatformat:decimals }}
{% endmacro %}
```

### Arguments with No Default

Arguments without defaults are required. If not provided, they will be `nil`:

```django
{% macro required_example(required_arg, optional_arg="default") %}
  Required: {{ required_arg|default:"(not provided)" }}
  Optional: {{ optional_arg }}
{% endmacro %}

{{ required_example("value") }}
{{ required_example("value", optional_arg="custom") }}
```

## Exporting Macros

To use macros across multiple templates, mark them with `export`:

```django
{% macro button(text, url) export %}
  <a href="{{ url }}" class="btn">{{ text }}</a>
{% endmacro %}

{% macro icon(name, size="16") export %}
  <svg class="icon icon-{{ name }}" width="{{ size }}" height="{{ size }}">
    <use xlink:href="#icon-{{ name }}"></use>
  </svg>
{% endmacro %}
```

The `export` keyword must come after the argument list:

```django
{% macro name(args...) export %}
  ...
{% endmacro %}
```

### What Export Does

- Makes the macro available for `{% import %}` in other templates
- Without `export`, macros are only available in the template where they're defined
- Exported macros are registered at parse time, not execution time

## Importing Macros

Use `{% import %}` to bring exported macros into another template:

```django
{% import "macros/ui.html" button, icon %}

{{ button("Submit", "/submit") }}
{{ icon("check") }}
```

### Import Syntax

```django
{% import "filename" macro1, macro2, macro3 %}
```

- **filename** - Path to the template containing exported macros
- **macros** - Comma-separated list of macro names to import

### Aliasing Imported Macros

Use `as` to rename macros when importing:

```django
{% import "macros/buttons.html" button as btn, submit_button as submit %}

{{ btn("Click", "/") }}
{{ submit("Save") }}
```

This is useful when:
- Avoiding name conflicts with local macros
- Creating shorter names for frequently used macros
- Making code more readable

### Multiple Import Statements

You can have multiple import statements:

```django
{% import "macros/buttons.html" button, link %}
{% import "macros/forms.html" input, select, checkbox %}
{% import "macros/icons.html" icon as i %}

<form>
  {{ input("username", "Username") }}
  {{ input("password", "Password", type="password") }}
  {{ button("Login", "/login") }}
</form>
```

## Macro Scope

### Local Variables

Variables defined inside a macro are scoped to that macro:

```django
{% macro counter() %}
  {% set count = 0 %}
  {% for i in range(5) %}
    {% set count = count + 1 %}
    {{ count }}
  {% endfor %}
{% endmacro %}

{{ counter() }}  {# 1 2 3 4 5 #}
{{ counter() }}  {# 1 2 3 4 5 (starts fresh) #}
```

### Accessing Context

Macros have access to the template's context:

```django
{% macro site_header() %}
  <header>
    <h1>{{ site_name }}</h1>  {# From context #}
    <p>{{ site_tagline }}</p>
  </header>
{% endmacro %}
```

However, relying on context makes macros less reusable. Prefer passing data as arguments:

```django
{# Better: explicit arguments #}
{% macro site_header(name, tagline) %}
  <header>
    <h1>{{ name }}</h1>
    <p>{{ tagline }}</p>
  </header>
{% endmacro %}

{{ site_header(site_name, site_tagline) }}
```

## Recursive Macros

Macros can call themselves recursively:

```django
{% macro tree(items, level=0) %}
  <ul class="level-{{ level }}">
    {% for item in items %}
      <li>
        {{ item.name }}
        {% if item.children %}
          {{ tree(item.children, level=level+1) }}
        {% endif %}
      </li>
    {% endfor %}
  </ul>
{% endmacro %}

{{ tree(menu_items) }}
```

### Recursion Limit

pongo2 limits macro recursion depth to **1000 calls** to prevent infinite loops. If exceeded, an error is returned:

```
maximum recursive macro call depth reached (max is 1000)
```

## Return Value

Macros return their rendered content as a **safe string** (HTML is not escaped):

```django
{% macro bold(text) %}
  <strong>{{ text }}</strong>
{% endmacro %}

{{ bold("Hello") }}  {# Output: <strong>Hello</strong> #}
```

The output is marked as safe, so HTML tags are preserved. If you need escaping, apply it inside the macro:

```django
{% macro user_comment(text) %}
  <div class="comment">{{ text|escape }}</div>
{% endmacro %}
```

## Best Practices

### 1. Keep Macros Focused

Each macro should do one thing well:

```django
{# Good: Single responsibility #}
{% macro avatar(user, size="md") %}
  <img src="{{ user.avatar }}" alt="{{ user.name }}" class="avatar avatar-{{ size }}">
{% endmacro %}

{% macro username(user, link=true) %}
  {% if link %}
    <a href="{{ user.profile_url }}">{{ user.name }}</a>
  {% else %}
    {{ user.name }}
  {% endif %}
{% endmacro %}

{# Usage: Compose them #}
<div class="user">
  {{ avatar(user) }}
  {{ username(user) }}
</div>
```

### 2. Use Meaningful Default Values

```django
{% macro pagination(current, total, url_pattern="/page/{page}") %}
  <nav class="pagination">
    {% if current > 1 %}
      <a href="{{ url_pattern|replace:"{page}":(current-1)|string }}">Previous</a>
    {% endif %}
    <span>Page {{ current }} of {{ total }}</span>
    {% if current < total %}
      <a href="{{ url_pattern|replace:"{page}":(current+1)|string }}">Next</a>
    {% endif %}
  </nav>
{% endmacro %}
```

### 3. Document Complex Macros

Use comments to document macro purpose and arguments:

```django
{#
  Renders a form input field with label and error handling.

  Arguments:
    name     - Field name (required)
    label    - Display label (required)
    type     - Input type (default: "text")
    value    - Current value (default: "")
    error    - Error message to display (default: none)
    required - Whether field is required (default: false)
#}
{% macro form_field(name, label, type="text", value="", error="", required=false) %}
  <div class="form-group{% if error %} has-error{% endif %}">
    <label for="{{ name }}">
      {{ label }}
      {% if required %}<span class="required">*</span>{% endif %}
    </label>
    <input
      type="{{ type }}"
      id="{{ name }}"
      name="{{ name }}"
      value="{{ value }}"
      {% if required %}required{% endif %}
    >
    {% if error %}
      <span class="error">{{ error }}</span>
    {% endif %}
  </div>
{% endmacro %}
```

### 4. Organize Macros in Dedicated Files

Create macro library files:

```
templates/
  macros/
    buttons.html     # Button macros
    forms.html       # Form field macros
    icons.html       # Icon macros
    layout.html      # Layout component macros
  pages/
    home.html
    about.html
```

### 5. Prefer Arguments Over Context

```django
{# Avoid: Relies on implicit context #}
{% macro bad_header() %}
  <h1>{{ page_title }}</h1>
{% endmacro %}

{# Better: Explicit argument #}
{% macro good_header(title) %}
  <h1>{{ title }}</h1>
{% endmacro %}

{{ good_header(page_title) }}
```

## Complete Example

**macros/ui.html:**

```django
{% macro card(title, content, footer="") export %}
  <div class="card">
    <div class="card-header">
      <h3>{{ title }}</h3>
    </div>
    <div class="card-body">
      {{ content|safe }}
    </div>
    {% if footer %}
    <div class="card-footer">
      {{ footer|safe }}
    </div>
    {% endif %}
  </div>
{% endmacro %}

{% macro alert(message, type="info", dismissible=false) export %}
  <div class="alert alert-{{ type }}{% if dismissible %} alert-dismissible{% endif %}">
    {{ message }}
    {% if dismissible %}
      <button type="button" class="close" data-dismiss="alert">&times;</button>
    {% endif %}
  </div>
{% endmacro %}

{% macro badge(text, color="primary") export %}
  <span class="badge badge-{{ color }}">{{ text }}</span>
{% endmacro %}
```

**pages/dashboard.html:**

```django
{% extends "base.html" %}
{% import "macros/ui.html" card, alert, badge %}

{% block content %}
  {% if error_message %}
    {{ alert(error_message, type="danger", dismissible=true) }}
  {% endif %}

  <div class="row">
    {{ card("Users", user_count|string + " registered users") }}
    {{ card("Orders", order_count|string + " orders today",
            footer=badge("+" + new_orders|string + " new", color="success")) }}
    {{ card("Revenue", "$" + revenue|floatformat:2,
            footer=badge(growth|floatformat:1 + "%", color="info")) }}
  </div>
{% endblock %}
```

## Error Handling

### Common Errors

**Macro not found:**
```
Macro 'button' not found (or not exported) in 'macros.html'.
```
Solution: Add `export` keyword to the macro definition.

**Too many arguments:**
```
Macro 'button' called with too many arguments (5 instead of 3).
```
Solution: Check the macro's argument count.

**Missing required argument:**
Arguments without defaults will be `nil`. Use `default` filter:
```django
{% macro greet(name) %}
  Hello, {{ name|default:"Guest" }}!
{% endmacro %}
```

**Maximum recursion depth:**
```
maximum recursive macro call depth reached (max is 1000)
```
Solution: Add a base case to stop recursion.
