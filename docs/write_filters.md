# Writing Custom Filters

Filters transform values in templates. This guide explains how to create your own custom filters for pongo2.

## Filter Function Signature

All filters must implement the `FilterFunction` type:

```go
type FilterFunction func(in *Value, param *Value) (out *Value, err *Error)
```

- **in** - The input value being filtered (left side of the pipe)
- **param** - The filter parameter (after the colon), or nil if no parameter
- **out** - The transformed output value
- **err** - Error if something went wrong, or nil on success

## Registering a Filter

Use `RegisterFilter` to register your filter, typically in an `init()` function:

```go
package main

import "github.com/flosch/pongo2/v7"

func init() {
    pongo2.RegisterFilter("double", filterDouble)
}

func filterDouble(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
    return pongo2.AsValue(in.Integer() * 2), nil
}
```

Usage in template:

```django
{{ 21|double }}  {# Output: 42 #}
```

## The Value Type

The `*Value` type wraps Go values and provides methods for type checking and conversion.

### Type Checking Methods

```go
func (v *Value) IsString() bool    // Is the value a string?
func (v *Value) IsInteger() bool   // Is the value an integer (int, int8, ..., uint, uint8, ...)?
func (v *Value) IsFloat() bool     // Is the value a float (float32, float64)?
func (v *Value) IsNumber() bool    // Is the value a number (integer or float)?
func (v *Value) IsBool() bool      // Is the value a boolean?
func (v *Value) IsNil() bool       // Is the value nil?
func (v *Value) IsTime() bool      // Is the value a time.Time?
func (v *Value) IsTrue() bool      // Is the value "truthy" (Pythonic evaluation)?
func (v *Value) CanSlice() bool    // Can the value be sliced (array, slice, string)?
```

### Type Conversion Methods

```go
func (v *Value) String() string    // Convert to string
func (v *Value) Integer() int      // Convert to int (returns 0 if not convertible)
func (v *Value) Float() float64    // Convert to float64 (returns 0.0 if not convertible)
func (v *Value) Bool() bool        // Get bool value (returns false if not bool)
func (v *Value) Time() time.Time   // Get time.Time value (returns zero time if not time)
func (v *Value) Interface() any    // Get the underlying Go value
```

### Collection Methods

```go
func (v *Value) Len() int                           // Length of array, slice, map, chan, or string
func (v *Value) Index(i int) *Value                 // Get i-th element
func (v *Value) Slice(i, j int) *Value              // Slice from i to j
func (v *Value) Contains(other *Value) bool         // Check if value contains another
func (v *Value) Iterate(fn, empty func())           // Iterate over collection
func (v *Value) IterateOrder(fn, empty, rev, sort)  // Iterate with ordering
```

### Creating Values

```go
// Regular value (will be HTML-escaped if autoescape is on)
pongo2.AsValue("hello")
pongo2.AsValue(42)
pongo2.AsValue(3.14)
pongo2.AsValue(true)
pongo2.AsValue([]string{"a", "b", "c"})

// Safe value (will NOT be HTML-escaped)
pongo2.AsSafeValue("<strong>bold</strong>")
```

## Filter Examples

### Simple Filter (No Parameter)

A filter that reverses a string:

```go
func init() {
    pongo2.RegisterFilter("reverse", filterReverse)
}

func filterReverse(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
    s := in.String()
    runes := []rune(s)
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    return pongo2.AsValue(string(runes)), nil
}
```

Usage:

```django
{{ "hello"|reverse }}  {# Output: olleh #}
```

### Filter with Parameter

A filter that repeats a string:

```go
func init() {
    pongo2.RegisterFilter("repeat", filterRepeat)
}

func filterRepeat(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
    s := in.String()
    times := param.Integer()
    if times <= 0 {
        times = 1
    }

    result := strings.Repeat(s, times)
    return pongo2.AsValue(result), nil
}
```

Usage:

```django
{{ "ha"|repeat:3 }}  {# Output: hahaha #}
```

### Filter with Type Checking

A filter that formats numbers with a thousands separator:

```go
func init() {
    pongo2.RegisterFilter("thousands", filterThousands)
}

func filterThousands(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
    if !in.IsNumber() {
        // Return input unchanged if not a number
        return in, nil
    }

    sep := ","
    if !param.IsNil() && param.String() != "" {
        sep = param.String()
    }

    n := in.Integer()
    negative := n < 0
    if negative {
        n = -n
    }

    str := strconv.Itoa(n)
    var result strings.Builder

    for i, c := range str {
        if i > 0 && (len(str)-i)%3 == 0 {
            result.WriteString(sep)
        }
        result.WriteRune(c)
    }

    if negative {
        return pongo2.AsValue("-" + result.String()), nil
    }
    return pongo2.AsValue(result.String()), nil
}
```

Usage:

```django
{{ 1234567|thousands }}        {# Output: 1,234,567 #}
{{ 1234567|thousands:"." }}    {# Output: 1.234.567 #}
```

### Filter Returning Safe HTML

A filter that wraps text in a tag (output should not be escaped):

```go
func init() {
    pongo2.RegisterFilter("wrap", filterWrap)
}

func filterWrap(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
    text := in.String()
    tag := "span"
    if !param.IsNil() && param.String() != "" {
        tag = param.String()
    }

    // Use AsSafeValue since we're returning HTML
    html := fmt.Sprintf("<%s>%s</%s>", tag, html.EscapeString(text), tag)
    return pongo2.AsSafeValue(html), nil
}
```

Usage:

```django
{{ "Hello"|wrap }}          {# Output: <span>Hello</span> #}
{{ "World"|wrap:"strong" }} {# Output: <strong>World</strong> #}
```

### Filter with Error Handling

A filter that parses JSON:

```go
func init() {
    pongo2.RegisterFilter("fromjson", filterFromJSON)
}

func filterFromJSON(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
    var result interface{}
    err := json.Unmarshal([]byte(in.String()), &result)
    if err != nil {
        return nil, &pongo2.Error{
            Sender:    "filter:fromjson",
            OrigError: fmt.Errorf("invalid JSON: %w", err),
        }
    }
    return pongo2.AsValue(result), nil
}
```

Usage:

```django
{% with data='{"name": "John", "age": 30}'|fromjson %}
  Name: {{ data.name }}, Age: {{ data.age }}
{% endwith %}
```

### Filter Working with Collections

A filter that shuffles a slice:

```go
func init() {
    pongo2.RegisterFilter("shuffle", filterShuffle)
}

func filterShuffle(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
    if !in.CanSlice() {
        return in, nil
    }

    // Build a slice of values
    var items []interface{}
    in.Iterate(func(idx, count int, key, value *pongo2.Value) bool {
        items = append(items, key.Interface())
        return true
    }, func() {})

    // Shuffle using Fisher-Yates
    for i := len(items) - 1; i > 0; i-- {
        j := rand.Intn(i + 1)
        items[i], items[j] = items[j], items[i]
    }

    return pongo2.AsValue(items), nil
}
```

Usage:

```django
{% for item in items|shuffle %}
  {{ item }}
{% endfor %}
```

### Filter with Default Parameter

A filter with an optional parameter that has a default:

```go
func init() {
    pongo2.RegisterFilter("prefix", filterPrefix)
}

func filterPrefix(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
    prefix := ">>> " // default
    if !param.IsNil() {
        prefix = param.String()
    }
    return pongo2.AsValue(prefix + in.String()), nil
}
```

Usage:

```django
{{ "Hello"|prefix }}           {# Output: >>> Hello #}
{{ "Hello"|prefix:"** " }}     {# Output: ** Hello #}
```

## API Reference

### RegisterFilter

```go
func RegisterFilter(name string, fn FilterFunction) error
```

Registers a new filter. Returns an error if a filter with that name already exists.

```go
err := pongo2.RegisterFilter("myfilter", myFilterFunc)
if err != nil {
    panic(err) // Filter already exists
}
```

### ReplaceFilter

```go
func ReplaceFilter(name string, fn FilterFunction) error
```

Replaces an existing filter. Returns an error if the filter doesn't exist.

```go
// Override the built-in upper filter
err := pongo2.ReplaceFilter("upper", myUpperFilter)
```

### FilterExists

```go
func FilterExists(name string) bool
```

Checks if a filter is registered.

```go
if pongo2.FilterExists("myfilter") {
    // Filter is available
}
```

### ApplyFilter

```go
func ApplyFilter(name string, value *Value, param *Value) (*Value, *Error)
```

Applies a filter programmatically.

```go
value := pongo2.AsValue("hello")
param := pongo2.AsValue(nil)

result, err := pongo2.ApplyFilter("upper", value, param)
if err != nil {
    // Handle error
}
fmt.Println(result.String()) // HELLO
```

### MustApplyFilter

```go
func MustApplyFilter(name string, value *Value, param *Value) *Value
```

Like `ApplyFilter` but panics on error.

```go
result := pongo2.MustApplyFilter("upper", pongo2.AsValue("hello"), nil)
```

## The Error Type

When returning errors from filters, create a `*pongo2.Error`:

```go
return nil, &pongo2.Error{
    Sender:    "filter:myfilter",        // Identifies the source
    OrigError: errors.New("something went wrong"),
}
```

The error will include template location information when displayed to users.

## Best Practices

### 1. Handle Nil Values Gracefully

```go
func filterSafe(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
    if in.IsNil() {
        return pongo2.AsValue(""), nil // Return empty string for nil
    }
    // ... rest of filter
}
```

### 2. Check Types Before Converting

```go
func filterDouble(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
    if !in.IsNumber() {
        return in, nil // Return unchanged if not a number
    }
    return pongo2.AsValue(in.Integer() * 2), nil
}
```

### 3. Use AsSafeValue for HTML Output

```go
// BAD: HTML might be double-escaped
return pongo2.AsValue("<b>text</b>"), nil

// GOOD: HTML will be preserved
return pongo2.AsSafeValue("<b>text</b>"), nil
```

### 4. Validate Parameters

```go
func filterTruncate(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
    if param.IsNil() {
        return nil, &pongo2.Error{
            Sender:    "filter:truncate",
            OrigError: errors.New("truncate requires a length parameter"),
        }
    }
    length := param.Integer()
    if length < 0 {
        return nil, &pongo2.Error{
            Sender:    "filter:truncate",
            OrigError: errors.New("length must be non-negative"),
        }
    }
    // ... rest of filter
}
```

### 5. Register in init()

```go
func init() {
    // Filters are registered at package initialization
    pongo2.RegisterFilter("myfilter", myFilterFunc)
}
```

### 6. Document Your Filters

```go
// filterMarkdown converts markdown text to HTML.
//
// Usage:
//   {{ text|markdown }}
//
// The output is marked as safe and will not be escaped.
func filterMarkdown(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
    // ...
}
```

## Complete Example: Filter Package

Create a reusable filter package:

```go
// filters/text.go
package filters

import (
    "strings"
    "unicode"

    "github.com/flosch/pongo2/v7"
)

func init() {
    pongo2.RegisterFilter("initials", filterInitials)
    pongo2.RegisterFilter("slugify", filterSlugify)
    pongo2.RegisterFilter("wordcount", filterWordCount)
}

// filterInitials extracts initials from a name.
// Usage: {{ "John Doe"|initials }} -> "JD"
func filterInitials(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
    words := strings.Fields(in.String())
    var initials strings.Builder
    for _, word := range words {
        if len(word) > 0 {
            initials.WriteRune(unicode.ToUpper(rune(word[0])))
        }
    }
    return pongo2.AsValue(initials.String()), nil
}

// filterSlugify converts text to a URL-friendly slug.
// Usage: {{ "Hello World!"|slugify }} -> "hello-world"
func filterSlugify(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
    s := strings.ToLower(in.String())
    var result strings.Builder
    lastWasHyphen := true // Prevent leading hyphen

    for _, r := range s {
        if unicode.IsLetter(r) || unicode.IsDigit(r) {
            result.WriteRune(r)
            lastWasHyphen = false
        } else if !lastWasHyphen {
            result.WriteRune('-')
            lastWasHyphen = true
        }
    }

    slug := strings.TrimSuffix(result.String(), "-")
    return pongo2.AsValue(slug), nil
}

// filterWordCount counts the number of words in text.
// Usage: {{ "Hello World"|wordcount }} -> 2
func filterWordCount(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
    words := strings.Fields(in.String())
    return pongo2.AsValue(len(words)), nil
}
```

Usage:

```go
// main.go
package main

import (
    _ "myapp/filters" // Import for side effects (registers filters)

    "github.com/flosch/pongo2/v7"
)

func main() {
    tpl := pongo2.Must(pongo2.FromString(`
        {{ name|initials }}
        {{ title|slugify }}
        {{ content|wordcount }} words
    `))
    // ...
}
```
