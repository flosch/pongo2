# Filters Reference

Filters transform the value of a variable. Apply them using the pipe symbol (`|`):

```django
{{ value|filtername }}
{{ value|filtername:argument }}
{{ value|filter1|filter2|filter3 }}
```

## String Filters

### upper

Converts a string to uppercase.

```django
{{ "hello"|upper }}  {# HELLO #}
```

### lower

Converts a string to lowercase.

```django
{{ "HELLO"|lower }}  {# hello #}
```

### capfirst

Capitalizes the first character of the string.

```django
{{ "hello world"|capfirst }}  {# Hello world #}
```

### title

Converts a string to title case.

```django
{{ "hello world"|title }}  {# Hello World #}
```

### center

Centers the value in a field of a given width.

```django
{{ "hi"|center:10 }}  {# "    hi    " #}
```

### ljust

Left-aligns the value in a field of a given width.

```django
{{ "hi"|ljust:10 }}  {# "hi        " #}
```

### rjust

Right-aligns the value in a field of a given width.

```django
{{ "hi"|rjust:10 }}  {# "        hi" #}
```

### cut

Removes all occurrences of a substring.

```django
{{ "hello world"|cut:" " }}  {# helloworld #}
```

### truncatechars

Truncates a string to the specified number of characters, appending "..." if truncated.

```django
{{ "Hello World"|truncatechars:8 }}  {# Hello... #}
```

### truncatechars_html

Like `truncatechars`, but preserves HTML tags.

```django
{{ "<p>Hello World</p>"|truncatechars_html:8 }}  {# <p>Hello...</p> #}
```

### truncatewords

Truncates a string after a specified number of words.

```django
{{ "One two three four five"|truncatewords:3 }}  {# One two three... #}
```

### truncatewords_html

Like `truncatewords`, but preserves HTML tags.

```django
{{ "<p>One two three four five</p>"|truncatewords_html:3 }}
{# <p>One two three...</p> #}
```

### wordcount

Returns the number of words.

```django
{{ "Hello World"|wordcount }}  {# 2 #}
```

### wordwrap

Wraps text at the specified character column width. Lines are broken at word
boundaries; long words that exceed the width are not split. Existing newlines
are preserved. `\r\n` and `\r` are normalized to `\n` before processing.

Verified against Django 4.2 `django.utils.text.wrap()`.

```django
{{ "a b c d e f g h"|wordwrap:5 }}
{# a b c
d e f
g h #}
```

### addslashes

Adds backslashes before quotes.

```django
{{ "I'm here"|addslashes }}  {# I\'m here #}
```

### split

Splits a string into a list.

```django
{% for part in "a,b,c"|split:"," %}{{ part }}{% endfor %}
{# abc #}
```

### make_list

Converts a string into a list of characters.

```django
{% for char in "hello"|make_list %}{{ char }}-{% endfor %}
{# h-e-l-l-o- #}
```

### phone2numeric

Converts phone letters to numbers.

```django
{{ "1-800-COLLECT"|phone2numeric }}  {# 1-800-2655328 #}
```

## HTML Filters

### escape (alias: e)

Escapes HTML special characters. Applied automatically unless `safe` is used.

```django
{{ "<script>"|escape }}  {# &lt;script&gt; #}
```

### escapejs

Escapes characters for use in JavaScript strings.

```django
{{ "line1\nline2"|escapejs }}
```

### safe

Marks a value as safe (not requiring HTML escaping).

```django
{{ trusted_html|safe }}
```

### linebreaks

Converts newlines into `<p>` and `<br />` tags.

```django
{{ "Line 1\n\nLine 2"|linebreaks }}
{# <p>Line 1</p><p>Line 2</p> #}
```

### linebreaksbr

Converts newlines into `<br />` tags.

```django
{{ "Line 1\nLine 2"|linebreaksbr }}
{# Line 1<br />Line 2 #}
```

### striptags

Removes all HTML tags.

```django
{{ "<p>Hello</p>"|striptags }}  {# Hello #}
```

### removetags

Removes specified HTML tags.

```django
{{ "<p><b>Hello</b></p>"|removetags:"b" }}  {# <p>Hello</p> #}
```

### linenumbers

Adds line numbers to each line.

```django
{{ "Line 1\nLine 2"|linenumbers }}
{# 1. Line 1
2. Line 2 #}
```

### urlize

Converts URLs and email addresses in text into clickable links.

```django
{{ "Visit https://example.com"|urlize }}
{# Visit <a href="https://example.com" rel="nofollow">https://example.com</a> #}
```

### urlizetrunc

Like `urlize`, but truncates URLs longer than the given limit.

```django
{{ "Visit https://example.com/very/long/path"|urlizetrunc:20 }}
```

### iriencode

Encodes an IRI (Internationalized Resource Identifier).

```django
{{ "path/to file"|iriencode }}  {# path/to%20file #}
```

### urlencode

URL-encodes a string.

```django
{{ "hello world"|urlencode }}  {# hello+world #}
```

## Number Filters

### add

Adds numbers (or concatenates strings).

```django
{{ 4|add:2 }}        {# 6 #}
{{ "hello"|add:" world" }}  {# hello world #}
```

### divisibleby

Returns true if the value is divisible by the argument.

```django
{% if 21|divisibleby:3 %}Yes{% endif %}  {# Yes #}
```

### floatformat

Formats a floating-point number.

```django
{{ 34.23234|floatformat:2 }}   {# 34.23 #}
{{ 34.00|floatformat:"-2" }}   {# 34 (trims zeros) #}
{{ 34.26|floatformat }}        {# 34.3 (default: -1) #}
```

### get_digit

Returns the digit at the specified position (from right, 1-indexed).

```django
{{ 123456|get_digit:2 }}  {# 5 #}
```

### float

Converts value to float.

```django
{{ "3.14"|float }}  {# 3.14 #}
```

### integer

Converts value to integer.

```django
{{ "42"|integer }}  {# 42 #}
{{ 3.7|integer }}   {# 3 #}
```

### stringformat

Formats the value using Go's fmt.Sprintf.

```django
{{ 3.14159|stringformat:"%.2f" }}  {# 3.14 #}
{{ 42|stringformat:"%05d" }}       {# 00042 #}
{{ 255|stringformat:"%x" }}        {# ff #}
```

## List/Collection Filters

### first

Returns the first element.

```django
{{ items|first }}
{{ "hello"|first }}  {# h #}
```

### last

Returns the last element.

```django
{{ items|last }}
{{ "hello"|last }}  {# o #}
```

### length

Returns the length.

```django
{{ items|length }}
{{ "hello"|length }}  {# 5 #}
```

### length_is

Returns true if the length equals the argument.

```django
{% if items|length_is:3 %}Exactly 3 items{% endif %}
```

### join

Joins list elements with a string.

```django
{{ items|join:", " }}  {# apple, banana, cherry #}
```

### slice

Returns a slice of the list.

```django
{{ items|slice:":2" }}   {# First two items #}
{{ items|slice:"1:" }}   {# All but first #}
{{ items|slice:"1:3" }}  {# Items 1 and 2 #}
{{ items|slice:"-2:" }}  {# Last two items #}
```

### random

Returns a random element.

```django
{{ items|random }}
```

## Date/Time Filters

### date

Formats a date using Go's time format.

```django
{{ now|date:"2006-01-02" }}           {# 2024-01-15 #}
{{ now|date:"Monday, January 2" }}    {# Monday, January 15 #}
{{ now|date:"Jan 2, 2006 3:04 PM" }}  {# Jan 15, 2024 2:30 PM #}
```

Go time format reference:
- `2006` - Year
- `01` - Month (01-12)
- `02` - Day (01-31)
- `15` - Hour (00-23)
- `03` - Hour (01-12)
- `04` - Minute (00-59)
- `05` - Second (00-59)
- `PM` - AM/PM
- `Monday` - Day name
- `January` - Month name

### time

Alias for `date`. Formats time values.

```django
{{ now|time:"15:04:05" }}  {# 14:30:45 #}
```

## Default Value Filters

### default

Returns the argument if value is empty/false.

```django
{{ value|default:"nothing" }}
{{ ""|default:"empty" }}     {# empty #}
{{ 0|default:"zero" }}       {# zero #}
{{ false|default:"nope" }}   {# nope #}
```

### default_if_none

Returns the argument only if value is nil.

```django
{{ value|default_if_none:"nothing" }}
{{ 0|default_if_none:"zero" }}  {# 0 (not replaced) #}
```

## Boolean Filters

### yesno

Maps true/false/nil to custom strings.

```django
{{ true|yesno:"yeah,nope,maybe" }}   {# yeah #}
{{ false|yesno:"yeah,nope,maybe" }}  {# nope #}
{{ nil|yesno:"yeah,nope,maybe" }}    {# maybe #}
{{ value|yesno }}                     {# yes/no/maybe (defaults) #}
```

### pluralize

Returns plural suffix based on value.

```django
{{ count }} item{{ count|pluralize }}           {# items if count != 1 #}
{{ count }} cherr{{ count|pluralize:"y,ies" }}  {# cherry/cherries #}
{{ count }} lun{{ count|pluralize:"ch,ches" }}  {# lunch/lunches #}
```

## Additional Filters (pongo2-addons)

The following filters are available via [pongo2-addons](https://github.com/flosch/pongo2-addons):

- **filesizeformat** - Formats file sizes (e.g., "13 KB")
- **slugify** - Converts to URL-friendly slug
- **timesince** - Time since a date
- **timeuntil** - Time until a date
- **naturaltime** - Human-readable time difference
- **naturalday** - "yesterday", "today", "tomorrow"
- **intcomma** - Adds commas to numbers
- **ordinal** - Adds ordinal suffix (1st, 2nd, 3rd)
- **markdown** - Renders Markdown to HTML
- **truncatesentences** - Truncates to N sentences
- **truncatesentences_html** - Same, but preserves HTML
