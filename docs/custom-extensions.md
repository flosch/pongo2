# Custom Extensions

pongo2 allows you to extend the template engine with custom filters and tags.

## Custom Filters

Filters transform variable values. They are functions with the signature:

```go
type FilterFunction func(in *Value, param *Value) (out *Value, err *Error)
```

### Registering a Filter

```go
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

### Filter with Parameter

```go
func init() {
    pongo2.RegisterFilter("multiply", filterMultiply)
}

func filterMultiply(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
    return pongo2.AsValue(in.Integer() * param.Integer()), nil
}
```

Usage:

```django
{{ 7|multiply:6 }}  {# Output: 42 #}
```

### Returning Errors

```go
func filterDivide(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
    divisor := param.Integer()
    if divisor == 0 {
        return nil, &pongo2.Error{
            Sender:    "filter:divide",
            OrigError: errors.New("division by zero"),
        }
    }
    return pongo2.AsValue(in.Integer() / divisor), nil
}
```

### Working with the Value Type

The `Value` type wraps Go values and provides helper methods:

```go
func myFilter(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
    // Type checking
    if in.IsString() { /* ... */ }
    if in.IsInteger() { /* ... */ }
    if in.IsFloat() { /* ... */ }
    if in.IsBool() { /* ... */ }
    if in.IsNil() { /* ... */ }
    if in.IsNumber() { /* ... */ } // Integer or Float

    // Type conversion
    s := in.String()      // Convert to string
    i := in.Integer()     // Convert to int
    f := in.Float()       // Convert to float64
    b := in.Bool()        // Convert to bool

    // Collection operations
    if in.CanSlice() {
        length := in.Len()
        first := in.Index(0)
        slice := in.Slice(0, 5)
    }

    // Check truthiness (for conditionals)
    if in.IsTrue() { /* ... */ }

    // Get underlying interface{}
    raw := in.Interface()

    // Return values
    return pongo2.AsValue("result"), nil       // Regular value
    return pongo2.AsSafeValue("<b>html</b>"), nil  // Safe (no escaping)
}
```

### Replacing Built-in Filters

```go
func init() {
    // Override the built-in upper filter
    pongo2.ReplaceFilter("upper", myUpperFilter)
}
```

### Check if Filter Exists

```go
if pongo2.FilterExists("myfilter") {
    // Filter is registered
}
```

### Apply Filter Programmatically

```go
value := pongo2.AsValue("hello")
param := pongo2.AsValue(nil)

result, err := pongo2.ApplyFilter("upper", value, param)
// result.String() == "HELLO"

// Panic version
result = pongo2.MustApplyFilter("upper", value, param)
```

## Custom Tags

Tags are more complex than filters. They can:
- Access and modify the parser
- Wrap content (block tags)
- Execute logic during rendering

### Tag Parser Function

```go
type TagParser func(doc *Parser, start *Token, arguments *Parser) (INodeTag, *Error)
```

Parameters:
- `doc` - The document parser (for parsing nested content)
- `start` - The token containing the tag name
- `arguments` - Parser for the tag's arguments

### Simple Tag Example

A tag that outputs the current time:

```go
func init() {
    pongo2.RegisterTag("current_time", tagCurrentTimeParser)
}

type tagCurrentTimeNode struct {
    format string
}

func (node *tagCurrentTimeNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
    writer.WriteString(time.Now().Format(node.format))
    return nil
}

func tagCurrentTimeParser(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
    node := &tagCurrentTimeNode{
        format: time.RFC3339, // default format
    }

    // Parse optional format argument
    if formatToken := arguments.MatchType(pongo2.TokenString); formatToken != nil {
        node.format = formatToken.Val
    }

    // Check for extra arguments
    if arguments.Remaining() > 0 {
        return nil, arguments.Error("Malformed current_time tag", nil)
    }

    return node, nil
}
```

Usage:

```django
{% current_time %}
{% current_time "2006-01-02" %}
```

### Block Tag Example

A tag that wraps content:

```go
func init() {
    pongo2.RegisterTag("uppercase", tagUppercaseParser)
}

type tagUppercaseNode struct {
    wrapper *pongo2.NodeWrapper
}

func (node *tagUppercaseNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
    // Capture the block content
    var buf bytes.Buffer
    err := node.wrapper.Execute(ctx, &buf)
    if err != nil {
        return err
    }

    // Transform and output
    writer.WriteString(strings.ToUpper(buf.String()))
    return nil
}

func tagUppercaseParser(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
    node := &tagUppercaseNode{}

    // Parse until enduppercase
    wrapper, endargs, err := doc.WrapUntilTag("enduppercase")
    if err != nil {
        return nil, err
    }
    node.wrapper = wrapper

    // enduppercase shouldn't have arguments
    if endargs.Count() > 0 {
        return nil, endargs.Error("enduppercase takes no arguments", nil)
    }

    return node, nil
}
```

Usage:

```django
{% uppercase %}
  hello world
{% enduppercase %}
{# Output: HELLO WORLD #}
```

### Tag with Expression Arguments

Parse expressions that can contain variables:

```go
type tagRepeatNode struct {
    countExpr pongo2.IEvaluator
    wrapper   *pongo2.NodeWrapper
}

func (node *tagRepeatNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
    // Evaluate the count expression
    countVal, err := node.countExpr.Evaluate(ctx)
    if err != nil {
        return err
    }

    count := countVal.Integer()
    for i := 0; i < count; i++ {
        err := node.wrapper.Execute(ctx, writer)
        if err != nil {
            return err
        }
    }
    return nil
}

func tagRepeatParser(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
    node := &tagRepeatNode{}

    // Parse the count expression
    countExpr, err := arguments.ParseExpression()
    if err != nil {
        return nil, err
    }
    node.countExpr = countExpr

    // Parse until endrepeat
    wrapper, _, err := doc.WrapUntilTag("endrepeat")
    if err != nil {
        return nil, err
    }
    node.wrapper = wrapper

    return node, nil
}
```

Usage:

```django
{% repeat 3 %}Hello {% endrepeat %}
{# Output: Hello Hello Hello #}

{% repeat count %}Item {% endrepeat %}
```

### Tag with Multiple End Tags

Handle tags like if/elif/else/endif:

```go
wrapper, endargs, err := doc.WrapUntilTag("elif", "else", "endif")
if err != nil {
    return nil, err
}

switch wrapper.Endtag {
case "elif":
    // Parse elif condition and continue
case "else":
    // Parse else block
case "endif":
    // Done
}
```

### Parser Methods

Common methods for parsing arguments:

```go
// Match specific token type and value
if token := arguments.Match(pongo2.TokenKeyword, "as"); token != nil {
    // Matched "as" keyword
}

// Match any of several values
if token := arguments.MatchOne(pongo2.TokenIdentifier, "asc", "desc"); token != nil {
    // Matched either "asc" or "desc"
}

// Match by type only
if token := arguments.MatchType(pongo2.TokenString); token != nil {
    value := token.Val
}
if token := arguments.MatchType(pongo2.TokenNumber); token != nil {
    // ...
}
if token := arguments.MatchType(pongo2.TokenIdentifier); token != nil {
    name := token.Val
}

// Parse a full expression
expr, err := arguments.ParseExpression()

// Check remaining arguments
if arguments.Remaining() > 0 {
    return nil, arguments.Error("Too many arguments", nil)
}

// Peek without consuming
if arguments.Peek(pongo2.TokenSymbol, "=") != nil {
    // Next token is "="
}
```

### Execution Context

Access template context during execution:

```go
func (node *myNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
    // Read from public context (user-provided)
    user := ctx.Public["user"]

    // Read/write private context (internal use)
    ctx.Private["my_counter"] = 0

    // Check autoescape setting
    if ctx.Autoescape {
        // HTML escaping is enabled
    }

    // Log debug messages (only when Debug=true)
    ctx.Logf("Processing item %d", itemNum)

    // Create error with template location
    return ctx.Error("Something went wrong", node.token)
}
```

### Creating Child Contexts

For tags that create new variable scopes:

```go
func (node *myNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
    // Create child context (inherits parent's variables)
    childCtx := pongo2.NewChildExecutionContext(ctx)

    // Add scoped variables
    childCtx.Private["loop_var"] = someValue

    // Execute wrapped content with child context
    return node.wrapper.Execute(childCtx, writer)
}
```

### Replacing Built-in Tags

```go
func init() {
    pongo2.ReplaceTag("for", myCustomForParser)
}
```

## Complete Example: Cache Tag

A tag that caches rendered content:

```go
package main

import (
    "bytes"
    "sync"
    "time"

    "github.com/flosch/pongo2/v6"
)

var (
    cache      = make(map[string]cacheEntry)
    cacheMutex sync.RWMutex
)

type cacheEntry struct {
    content   string
    expiresAt time.Time
}

type tagCacheNode struct {
    key      string
    duration time.Duration
    wrapper  *pongo2.NodeWrapper
}

func (node *tagCacheNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
    // Check cache
    cacheMutex.RLock()
    entry, exists := cache[node.key]
    cacheMutex.RUnlock()

    if exists && time.Now().Before(entry.expiresAt) {
        writer.WriteString(entry.content)
        return nil
    }

    // Render content
    var buf bytes.Buffer
    err := node.wrapper.Execute(ctx, &buf)
    if err != nil {
        return err
    }

    content := buf.String()

    // Store in cache
    cacheMutex.Lock()
    cache[node.key] = cacheEntry{
        content:   content,
        expiresAt: time.Now().Add(node.duration),
    }
    cacheMutex.Unlock()

    writer.WriteString(content)
    return nil
}

func tagCacheParser(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
    node := &tagCacheNode{
        duration: 5 * time.Minute, // default
    }

    // Parse cache key
    keyToken := arguments.MatchType(pongo2.TokenString)
    if keyToken == nil {
        return nil, arguments.Error("cache tag requires a key string", nil)
    }
    node.key = keyToken.Val

    // Parse optional duration
    if durationToken := arguments.MatchType(pongo2.TokenNumber); durationToken != nil {
        seconds := pongo2.AsValue(durationToken.Val).Integer()
        node.duration = time.Duration(seconds) * time.Second
    }

    // Parse content
    wrapper, _, err := doc.WrapUntilTag("endcache")
    if err != nil {
        return nil, err
    }
    node.wrapper = wrapper

    return node, nil
}

func init() {
    pongo2.RegisterTag("cache", tagCacheParser)
}
```

Usage:

```django
{% cache "sidebar" 300 %}
  <div class="sidebar">
    {{ expensive_query() }}
  </div>
{% endcache %}
```
