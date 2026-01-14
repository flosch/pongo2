# Writing Custom Tags

Tags are template constructs that control logic, flow, and output. Unlike filters which transform values, tags can access the parser, wrap content, modify the execution context, and implement complex control structures.

## Overview

Creating a custom tag requires two components:

1. **Tag Parser Function** - Parses the tag syntax at compile time
2. **Tag Node** - Executes the tag logic at render time

```go
// Tag parser function signature
type TagParser func(doc *Parser, start *Token, arguments *Parser) (INodeTag, *Error)

// Tag node must implement INodeTag (which extends INode)
type INodeTag interface {
    Execute(ctx *ExecutionContext, writer TemplateWriter) *Error
}
```

## Registering Tags

### RegisterTag

Register a new tag with a unique name:

```go
func init() {
    err := pongo2.RegisterTag("mytag", myTagParser)
    if err != nil {
        panic(err) // Tag name already exists
    }
}
```

### ReplaceTag

Replace an existing tag (built-in or custom):

```go
func init() {
    err := pongo2.ReplaceTag("for", myCustomForParser)
    if err != nil {
        panic(err) // Tag doesn't exist
    }
}
```

### TagExists

Check if a tag is registered:

```go
if pongo2.TagExists("cache") {
    // Tag is available
}
```

## Tag Parser Function

The parser function is called at compile time when pongo2 encounters your tag:

```go
func myTagParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, *Error)
```

### Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `doc` | `*Parser` | Document parser for parsing nested content |
| `start` | `*Token` | Token containing the tag name (for error reporting) |
| `arguments` | `*Parser` | Parser for the tag's arguments |

### Return Values

- `INodeTag` - The compiled tag node (implements `Execute`)
- `*Error` - Parse error if syntax is invalid, otherwise `nil`

## Parsing Arguments

The `arguments` parser provides methods to consume and match tokens.

### Token Types

```go
const (
    TokenError TokenType = iota
    EOF
    TokenHTML
    TokenKeyword      // "in", "and", "or", "not", "true", "false", "as", etc.
    TokenIdentifier   // Variable names: user, item, counter
    TokenString       // "hello", 'world'
    TokenNumber       // 42, 3.14
    TokenSymbol       // (, ), [, ], {, }, ,, :, =, +, -, etc.
    TokenNil          // nil, None
)
```

### MatchType

Match any token of a specific type:

```go
// Match any string token
if token := arguments.MatchType(pongo2.TokenString); token != nil {
    value := token.Val  // The string content (without quotes)
}

// Match any identifier
if token := arguments.MatchType(pongo2.TokenIdentifier); token != nil {
    name := token.Val
}

// Match any number
if token := arguments.MatchType(pongo2.TokenNumber); token != nil {
    numStr := token.Val  // "42" or "3.14"
}
```

### Match

Match a specific token type and value:

```go
// Match the keyword "as"
if token := arguments.Match(pongo2.TokenKeyword, "as"); token != nil {
    // Found "as" keyword
}

// Match a specific symbol
if arguments.Match(pongo2.TokenSymbol, "=") != nil {
    // Found "=" symbol
}

// Match a specific identifier
if arguments.Match(pongo2.TokenIdentifier, "reversed") != nil {
    // Found "reversed" identifier
}
```

### MatchOne

Match a token type with any of several values:

```go
// Match either "asc" or "desc"
if token := arguments.MatchOne(pongo2.TokenIdentifier, "asc", "desc"); token != nil {
    direction := token.Val  // "asc" or "desc"
}

// Match opening bracket or parenthesis
if token := arguments.MatchOne(pongo2.TokenSymbol, "(", "["); token != nil {
    opener := token.Val
}
```

### Peek Methods

Look at tokens without consuming them:

```go
// Peek at next token type and value
if token := arguments.Peek(pongo2.TokenKeyword, "as"); token != nil {
    // Next token is "as" (not consumed)
}

// Peek at next token type only
if token := arguments.PeekType(pongo2.TokenString); token != nil {
    // Next token is a string (not consumed)
}

// Peek one of multiple values
if token := arguments.PeekOne(pongo2.TokenSymbol, ",", ")"); token != nil {
    // Next token is "," or ")"
}

// Peek N tokens ahead (0-indexed)
if token := arguments.PeekN(1, pongo2.TokenSymbol, "="); token != nil {
    // Token after next is "="
}
```

### Consume and Navigation

```go
// Get current token without advancing
current := arguments.Current()

// Consume and return current token
token := arguments.Consume()

// Consume N tokens
tokens := arguments.ConsumeN(3)

// Get token at specific index
token := arguments.Get(0)  // First token

// Get token from end (negative index)
token := arguments.GetR(-1)  // Last token

// Check remaining token count
if arguments.Remaining() > 0 {
    // More tokens to parse
}

// Get total token count
count := arguments.Count()
```

### ParseExpression

Parse a full expression (variables, operators, function calls):

```go
expr, err := arguments.ParseExpression()
if err != nil {
    return nil, err
}
// expr implements IEvaluator - call expr.Evaluate(ctx) at runtime
```

### Error Reporting

Create descriptive parse errors:

```go
// Error with current position
return nil, arguments.Error("Expected identifier after 'as'", nil)

// Error at specific token
return nil, arguments.Error("Invalid argument", someToken)
```

## Block Tags with WrapUntilTag

For tags that wrap content (like `{% if %}...{% endif %}`), use `WrapUntilTag`:

```go
func myBlockParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, *Error) {
    node := &myBlockNode{}

    // Parse until "endmyblock" tag
    wrapper, endtagArgs, err := doc.WrapUntilTag("endmyblock")
    if err != nil {
        return nil, err
    }
    node.wrapper = wrapper

    // Check that end tag has no arguments
    if endtagArgs.Count() > 0 {
        return nil, endtagArgs.Error("endmyblock takes no arguments", nil)
    }

    return node, nil
}
```

### Multiple End Tags

Handle tags with alternatives (like if/elif/else):

```go
wrapper, endArgs, err := doc.WrapUntilTag("elif", "else", "endif")
if err != nil {
    return nil, err
}

// Check which end tag was found
switch wrapper.Endtag {
case "elif":
    // Parse elif condition, continue parsing
case "else":
    // Parse else block
case "endif":
    // Done parsing
}
```

### SkipUntilTag

Skip content without parsing (for optimization):

```go
err := doc.SkipUntilTag("endcomment")
if err != nil {
    return nil, err
}
```

## Execution Context

The `ExecutionContext` provides access to template state during rendering.

### Structure

```go
type ExecutionContext struct {
    template   *Template           // Current template
    macroDepth int                 // Recursion depth for macros

    Autoescape bool                // HTML auto-escaping enabled
    Public     Context             // User-provided context (read-only by convention)
    Private    Context             // Internal variables (forloop, macro args, etc.)
    Shared     Context             // Shared across all templates in execution
}
```

### Reading Context Values

```go
func (node *myNode) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
    // Read from public context (user-provided data)
    if user, ok := ctx.Public["user"]; ok {
        // Use user data
    }

    // Read from private context (internal variables)
    if counter, ok := ctx.Private["forloop"]; ok {
        // Inside a for loop
    }

    // Read from shared context (persists across includes)
    if cache, ok := ctx.Shared["_cache"]; ok {
        // Use shared cache
    }

    return nil
}
```

### Writing to Context

```go
func (node *myNode) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
    // Set private variable (for internal use)
    ctx.Private["my_counter"] = 0

    // Set shared variable (available in included templates)
    ctx.Shared["breadcrumbs"] = breadcrumbList

    // Update context with map
    ctx.Private.Update(pongo2.Context{
        "item":  currentItem,
        "index": currentIndex,
    })

    return nil
}
```

### Child Contexts

Create isolated scopes for nested content:

```go
func (node *myNode) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
    // Create child context (inherits parent's variables)
    childCtx := pongo2.NewChildExecutionContext(ctx)

    // Add scoped variables (only visible in child)
    childCtx.Private["scoped_var"] = someValue

    // Execute wrapped content with child context
    err := node.wrapper.Execute(childCtx, writer)
    if err != nil {
        return err
    }

    // Parent context unchanged
    return nil
}
```

### Error Handling

Create errors with template location information:

```go
func (node *myNode) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
    if somethingWrong {
        return ctx.Error("Something went wrong", node.position)
    }
    return nil
}
```

### Logging

Debug logging (only active when template set has `Debug: true`):

```go
func (node *myNode) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
    ctx.Logf("Processing item: %v", item)
    return nil
}
```

### Autoescape

Check and respect the autoescape setting:

```go
func (node *myNode) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
    content := "<b>Bold</b>"

    if ctx.Autoescape {
        // HTML will be escaped by default
        writer.WriteString(content)  // Shows as &lt;b&gt;Bold&lt;/b&gt;
    } else {
        writer.WriteString(content)  // Shows as <b>Bold</b>
    }

    return nil
}
```

## Writing Output

The `TemplateWriter` interface:

```go
type TemplateWriter interface {
    io.Writer
    WriteString(s string) (int, error)
}
```

### Writing Strings

```go
func (node *myNode) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
    writer.WriteString("Hello, World!")
    return nil
}
```

### Writing Bytes

```go
func (node *myNode) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
    data := []byte("binary data")
    writer.Write(data)
    return nil
}
```

### Capturing Output

Use a buffer to capture and transform output:

```go
func (node *myNode) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
    var buf bytes.Buffer

    // Execute wrapped content into buffer
    err := node.wrapper.Execute(ctx, &buf)
    if err != nil {
        return err
    }

    // Transform the captured content
    transformed := strings.ToUpper(buf.String())

    // Write transformed content
    writer.WriteString(transformed)
    return nil
}
```

## Complete Examples

### Simple Tag: Current Time

A tag that outputs the current time:

```django
{% now %}
{% now "2006-01-02" %}
```

```go
package main

import (
    "time"

    "github.com/flosch/pongo2/v6"
)

type tagNowNode struct {
    position *pongo2.Token
    format   string
}

func (node *tagNowNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
    writer.WriteString(time.Now().Format(node.format))
    return nil
}

func tagNowParser(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
    node := &tagNowNode{
        position: start,
        format:   time.RFC3339, // Default format
    }

    // Optional: custom format string
    if formatToken := arguments.MatchType(pongo2.TokenString); formatToken != nil {
        node.format = formatToken.Val
    }

    // No more arguments allowed
    if arguments.Remaining() > 0 {
        return nil, arguments.Error("now tag takes at most one argument", nil)
    }

    return node, nil
}

func init() {
    pongo2.RegisterTag("now", tagNowParser)
}
```

### Block Tag: Uppercase

A tag that transforms content to uppercase:

```django
{% uppercase %}
  Hello, World!
{% enduppercase %}
```

```go
package main

import (
    "bytes"
    "strings"

    "github.com/flosch/pongo2/v6"
)

type tagUppercaseNode struct {
    wrapper *pongo2.NodeWrapper
}

func (node *tagUppercaseNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
    // Capture block content
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
    // No arguments for this tag
    if arguments.Remaining() > 0 {
        return nil, arguments.Error("uppercase tag takes no arguments", nil)
    }

    node := &tagUppercaseNode{}

    // Parse until enduppercase
    wrapper, endArgs, err := doc.WrapUntilTag("enduppercase")
    if err != nil {
        return nil, err
    }
    node.wrapper = wrapper

    // enduppercase shouldn't have arguments
    if endArgs.Count() > 0 {
        return nil, endArgs.Error("enduppercase takes no arguments", nil)
    }

    return node, nil
}

func init() {
    pongo2.RegisterTag("uppercase", tagUppercaseParser)
}
```

### Tag with Expressions: Repeat

A tag that repeats content a variable number of times:

```django
{% repeat 3 %}Hello {% endrepeat %}
{% repeat count %}Item {% endrepeat %}
```

```go
package main

import (
    "github.com/flosch/pongo2/v6"
)

type tagRepeatNode struct {
    position  *pongo2.Token
    countExpr pongo2.IEvaluator
    wrapper   *pongo2.NodeWrapper
}

func (node *tagRepeatNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
    // Evaluate the count expression at runtime
    countVal, err := node.countExpr.Evaluate(ctx)
    if err != nil {
        return err
    }

    count := countVal.Integer()
    if count < 0 {
        return ctx.Error("repeat count cannot be negative", node.position)
    }

    // Repeat the content
    for i := 0; i < count; i++ {
        err := node.wrapper.Execute(ctx, writer)
        if err != nil {
            return err
        }
    }

    return nil
}

func tagRepeatParser(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
    node := &tagRepeatNode{
        position: start,
    }

    // Parse the count expression (required)
    countExpr, err := arguments.ParseExpression()
    if err != nil {
        return nil, err
    }
    node.countExpr = countExpr

    // No more arguments
    if arguments.Remaining() > 0 {
        return nil, arguments.Error("repeat tag takes exactly one argument", nil)
    }

    // Parse until endrepeat
    wrapper, endArgs, err := doc.WrapUntilTag("endrepeat")
    if err != nil {
        return nil, err
    }
    node.wrapper = wrapper

    if endArgs.Count() > 0 {
        return nil, endArgs.Error("endrepeat takes no arguments", nil)
    }

    return node, nil
}

func init() {
    pongo2.RegisterTag("repeat", tagRepeatParser)
}
```

### Tag with Loop Variables: Each

A tag that iterates with access to loop metadata:

```django
{% each item in items %}
  {{ eachloop.Counter }}: {{ item }}
{% endeach %}
```

```go
package main

import (
    "github.com/flosch/pongo2/v6"
)

type eachLoop struct {
    Counter     int
    Counter0    int
    First       bool
    Last        bool
    Revcounter  int
    Revcounter0 int
}

type tagEachNode struct {
    position    *pongo2.Token
    itemName    string
    listExpr    pongo2.IEvaluator
    wrapper     *pongo2.NodeWrapper
}

func (node *tagEachNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
    // Evaluate the list expression
    listVal, err := node.listExpr.Evaluate(ctx)
    if err != nil {
        return err
    }

    if !listVal.CanSlice() {
        return ctx.Error("each requires an iterable", node.position)
    }

    length := listVal.Len()

    for i := 0; i < length; i++ {
        // Create child context for each iteration
        childCtx := pongo2.NewChildExecutionContext(ctx)

        // Set item variable
        childCtx.Private[node.itemName] = listVal.Index(i).Interface()

        // Set loop metadata
        childCtx.Private["eachloop"] = &eachLoop{
            Counter:     i + 1,
            Counter0:    i,
            First:       i == 0,
            Last:        i == length-1,
            Revcounter:  length - i,
            Revcounter0: length - i - 1,
        }

        // Execute body
        err := node.wrapper.Execute(childCtx, writer)
        if err != nil {
            return err
        }
    }

    return nil
}

func tagEachParser(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
    node := &tagEachNode{
        position: start,
    }

    // Parse: item in list
    itemToken := arguments.MatchType(pongo2.TokenIdentifier)
    if itemToken == nil {
        return nil, arguments.Error("expected item variable name", nil)
    }
    node.itemName = itemToken.Val

    if arguments.Match(pongo2.TokenKeyword, "in") == nil {
        return nil, arguments.Error("expected 'in' keyword", nil)
    }

    listExpr, err := arguments.ParseExpression()
    if err != nil {
        return nil, err
    }
    node.listExpr = listExpr

    if arguments.Remaining() > 0 {
        return nil, arguments.Error("unexpected arguments after list", nil)
    }

    // Parse body
    wrapper, endArgs, err := doc.WrapUntilTag("endeach")
    if err != nil {
        return nil, err
    }
    node.wrapper = wrapper

    if endArgs.Count() > 0 {
        return nil, endArgs.Error("endeach takes no arguments", nil)
    }

    return node, nil
}

func init() {
    pongo2.RegisterTag("each", tagEachParser)
}
```

### Tag with Named Arguments: Widget

A tag that accepts keyword arguments:

```django
{% widget "button" text="Click me" style="primary" disabled=false %}
```

```go
package main

import (
    "fmt"

    "github.com/flosch/pongo2/v6"
)

type tagWidgetNode struct {
    position   *pongo2.Token
    widgetType string
    args       map[string]pongo2.IEvaluator
}

func (node *tagWidgetNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
    // Evaluate all arguments
    evaluatedArgs := make(map[string]interface{})
    for name, expr := range node.args {
        val, err := expr.Evaluate(ctx)
        if err != nil {
            return err
        }
        evaluatedArgs[name] = val.Interface()
    }

    // Render widget (example implementation)
    text, _ := evaluatedArgs["text"].(string)
    style, _ := evaluatedArgs["style"].(string)
    if style == "" {
        style = "default"
    }

    disabled := false
    if d, ok := evaluatedArgs["disabled"].(bool); ok {
        disabled = d
    }

    disabledAttr := ""
    if disabled {
        disabledAttr = " disabled"
    }

    html := fmt.Sprintf(`<button class="widget-%s btn-%s"%s>%s</button>`,
        node.widgetType, style, disabledAttr, text)

    writer.WriteString(html)
    return nil
}

func tagWidgetParser(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
    node := &tagWidgetNode{
        position: start,
        args:     make(map[string]pongo2.IEvaluator),
    }

    // Parse widget type (required string)
    typeToken := arguments.MatchType(pongo2.TokenString)
    if typeToken == nil {
        return nil, arguments.Error("widget requires a type string", nil)
    }
    node.widgetType = typeToken.Val

    // Parse keyword arguments: name=value
    for arguments.Remaining() > 0 {
        nameToken := arguments.MatchType(pongo2.TokenIdentifier)
        if nameToken == nil {
            return nil, arguments.Error("expected argument name", nil)
        }

        if arguments.Match(pongo2.TokenSymbol, "=") == nil {
            return nil, arguments.Error("expected '=' after argument name", nil)
        }

        valueExpr, err := arguments.ParseExpression()
        if err != nil {
            return nil, err
        }

        node.args[nameToken.Val] = valueExpr
    }

    return node, nil
}

func init() {
    pongo2.RegisterTag("widget", tagWidgetParser)
}
```

### Tag with Conditional Branches: Switch

A tag with multiple conditional branches:

```django
{% switch status %}
{% case "active" %}
  Active user
{% case "pending" %}
  Pending approval
{% default %}
  Unknown status
{% endswitch %}
```

```go
package main

import (
    "github.com/flosch/pongo2/v6"
)

type switchCase struct {
    value   pongo2.IEvaluator
    wrapper *pongo2.NodeWrapper
}

type tagSwitchNode struct {
    position     *pongo2.Token
    expr         pongo2.IEvaluator
    cases        []switchCase
    defaultCase  *pongo2.NodeWrapper
}

func (node *tagSwitchNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
    // Evaluate switch expression
    switchVal, err := node.expr.Evaluate(ctx)
    if err != nil {
        return err
    }

    // Find matching case
    for _, c := range node.cases {
        caseVal, err := c.value.Evaluate(ctx)
        if err != nil {
            return err
        }

        // Compare values
        if switchVal.Interface() == caseVal.Interface() {
            return c.wrapper.Execute(ctx, writer)
        }
    }

    // Execute default if no match
    if node.defaultCase != nil {
        return node.defaultCase.Execute(ctx, writer)
    }

    return nil
}

func tagSwitchParser(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
    node := &tagSwitchNode{
        position: start,
        cases:    make([]switchCase, 0),
    }

    // Parse switch expression
    expr, err := arguments.ParseExpression()
    if err != nil {
        return nil, err
    }
    node.expr = expr

    if arguments.Remaining() > 0 {
        return nil, arguments.Error("unexpected arguments", nil)
    }

    // Parse cases
    for {
        wrapper, endArgs, err := doc.WrapUntilTag("case", "default", "endswitch")
        if err != nil {
            return nil, err
        }

        switch wrapper.Endtag {
        case "case":
            // Parse case value
            caseExpr, err := endArgs.ParseExpression()
            if err != nil {
                return nil, err
            }

            if endArgs.Remaining() > 0 {
                return nil, endArgs.Error("unexpected arguments after case value", nil)
            }

            node.cases = append(node.cases, switchCase{
                value:   caseExpr,
                wrapper: wrapper,
            })

        case "default":
            if endArgs.Count() > 0 {
                return nil, endArgs.Error("default takes no arguments", nil)
            }

            // Parse default block
            defaultWrapper, defaultEnd, err := doc.WrapUntilTag("endswitch")
            if err != nil {
                return nil, err
            }

            if defaultEnd.Count() > 0 {
                return nil, defaultEnd.Error("endswitch takes no arguments", nil)
            }

            node.defaultCase = defaultWrapper
            return node, nil

        case "endswitch":
            if endArgs.Count() > 0 {
                return nil, endArgs.Error("endswitch takes no arguments", nil)
            }
            return node, nil
        }
    }
}

func init() {
    pongo2.RegisterTag("switch", tagSwitchParser)
    // Note: "case" and "default" are handled internally by switch, not registered separately
}
```

### Tag with Caching: Cache

A tag that caches rendered content:

```django
{% cache "sidebar" 300 %}
  {{ expensive_computation() }}
{% endcache %}
```

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
    position *pongo2.Token
    keyExpr  pongo2.IEvaluator
    ttl      time.Duration
    wrapper  *pongo2.NodeWrapper
}

func (node *tagCacheNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
    // Evaluate cache key
    keyVal, err := node.keyExpr.Evaluate(ctx)
    if err != nil {
        return err
    }
    key := keyVal.String()

    // Check cache
    cacheMutex.RLock()
    entry, exists := cache[key]
    cacheMutex.RUnlock()

    if exists && time.Now().Before(entry.expiresAt) {
        writer.WriteString(entry.content)
        return nil
    }

    // Render content
    var buf bytes.Buffer
    execErr := node.wrapper.Execute(ctx, &buf)
    if execErr != nil {
        return execErr
    }

    content := buf.String()

    // Store in cache
    cacheMutex.Lock()
    cache[key] = cacheEntry{
        content:   content,
        expiresAt: time.Now().Add(node.ttl),
    }
    cacheMutex.Unlock()

    writer.WriteString(content)
    return nil
}

func tagCacheParser(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
    node := &tagCacheNode{
        position: start,
        ttl:      5 * time.Minute, // Default TTL
    }

    // Parse cache key expression
    keyExpr, err := arguments.ParseExpression()
    if err != nil {
        return nil, err
    }
    node.keyExpr = keyExpr

    // Optional TTL in seconds
    if ttlToken := arguments.MatchType(pongo2.TokenNumber); ttlToken != nil {
        seconds := pongo2.AsValue(ttlToken.Val).Integer()
        node.ttl = time.Duration(seconds) * time.Second
    }

    if arguments.Remaining() > 0 {
        return nil, arguments.Error("unexpected arguments", nil)
    }

    // Parse content
    wrapper, endArgs, err := doc.WrapUntilTag("endcache")
    if err != nil {
        return nil, err
    }
    node.wrapper = wrapper

    if endArgs.Count() > 0 {
        return nil, endArgs.Error("endcache takes no arguments", nil)
    }

    return node, nil
}

func init() {
    pongo2.RegisterTag("cache", tagCacheParser)
}
```

## Error Handling

### Parse-Time Errors

Return errors during parsing for syntax issues:

```go
func myTagParser(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
    // Missing required argument
    if arguments.Remaining() == 0 {
        return nil, arguments.Error("mytag requires an argument", nil)
    }

    // Wrong token type
    token := arguments.MatchType(pongo2.TokenString)
    if token == nil {
        return nil, arguments.Error("expected string argument", nil)
    }

    // Invalid value
    if token.Val == "" {
        return nil, arguments.Error("string cannot be empty", token)
    }

    return node, nil
}
```

### Execution-Time Errors

Return errors during execution for runtime issues:

```go
func (node *myNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
    val, err := node.expr.Evaluate(ctx)
    if err != nil {
        return err
    }

    if val.IsNil() {
        return ctx.Error("value cannot be nil", node.position)
    }

    if val.Integer() < 0 {
        return ctx.Error("value must be non-negative", node.position)
    }

    return nil
}
```

### Error Structure

```go
type Error struct {
    Template  *Template    // Template where error occurred
    Filename  string       // Template filename
    Line      int          // Line number
    Column    int          // Column number
    Token     *Token       // Related token
    Sender    string       // Component that generated error
    OrigError error        // Underlying Go error
}
```

Update error with token position:

```go
err := someOperation()
if err != nil {
    return err.(*pongo2.Error).updateFromTokenIfNeeded(ctx.template, node.position)
}
```

## Best Practices

### 1. Store Position for Error Reporting

Always store the start token for error messages:

```go
type myTagNode struct {
    position *pongo2.Token  // For error reporting
    // ... other fields
}
```

### 2. Validate Arguments Thoroughly

Check all arguments at parse time:

```go
func myTagParser(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
    // Check for required arguments
    if arguments.Remaining() == 0 {
        return nil, arguments.Error("mytag requires at least one argument", nil)
    }

    // Parse and validate each argument
    // ...

    // Check for unexpected extra arguments
    if arguments.Remaining() > 0 {
        return nil, arguments.Error("unexpected extra arguments", nil)
    }

    return node, nil
}
```

### 3. Use Child Contexts for Scoped Variables

Prevent variable leakage:

```go
func (node *myNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
    childCtx := pongo2.NewChildExecutionContext(ctx)
    childCtx.Private["local_var"] = value

    return node.wrapper.Execute(childCtx, writer)
    // local_var is not visible in parent context
}
```

### 4. Document Your Tags

Include usage examples:

```go
// TagCache caches rendered content for a specified duration.
//
// Usage:
//   {% cache "key" %}content{% endcache %}
//   {% cache key_var 300 %}content{% endcache %}
//
// Arguments:
//   - key: Cache key (string or expression)
//   - ttl: Optional TTL in seconds (default: 300)
func tagCacheParser(...) { ... }
```

### 5. Handle Edge Cases

Consider empty content, nil values, and type mismatches:

```go
func (node *myNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
    val, err := node.expr.Evaluate(ctx)
    if err != nil {
        return err
    }

    // Handle nil
    if val.IsNil() {
        return nil  // Or write default content
    }

    // Handle wrong type gracefully
    if !val.CanSlice() {
        ctx.Logf("expected iterable, got %T", val.Interface())
        return nil
    }

    // Handle empty collections
    if val.Len() == 0 {
        return nil  // Nothing to iterate
    }

    // ... process items
    return nil
}
```

### 6. Clean Up Resources

If your tag acquires resources, ensure cleanup:

```go
func (node *myNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
    resource := acquireResource()
    defer resource.Release()

    return node.wrapper.Execute(ctx, writer)
}
```

## API Reference

### Registration Functions

| Function | Description |
|----------|-------------|
| `RegisterTag(name, parser)` | Register a new tag |
| `ReplaceTag(name, parser)` | Replace an existing tag |
| `TagExists(name)` | Check if tag is registered |

### Parser Methods

| Method | Description |
|--------|-------------|
| `Match(type, val)` | Match specific token |
| `MatchOne(type, vals...)` | Match one of several values |
| `MatchType(type)` | Match any token of type |
| `Peek(type, val)` | Look ahead without consuming |
| `PeekOne(type, vals...)` | Peek one of several values |
| `PeekType(type)` | Peek token type |
| `PeekN(n, type, val)` | Peek N tokens ahead |
| `Consume()` | Consume current token |
| `ConsumeN(n)` | Consume N tokens |
| `Current()` | Get current token |
| `Get(i)` | Get token at index |
| `GetR(i)` | Get token from end |
| `Remaining()` | Count remaining tokens |
| `Count()` | Total token count |
| `ParseExpression()` | Parse full expression |
| `Error(msg, token)` | Create parse error |
| `WrapUntilTag(names...)` | Parse until end tag |
| `SkipUntilTag(name)` | Skip until end tag |

### ExecutionContext Methods

| Method/Field | Description |
|--------------|-------------|
| `Public` | User-provided context |
| `Private` | Internal variables |
| `Shared` | Shared across templates |
| `Autoescape` | HTML escaping enabled |
| `Error(msg, token)` | Create execution error |
| `Logf(format, args...)` | Debug logging |
| `NewChildExecutionContext(ctx)` | Create child context |

### Token Types

| Constant | Description | Example |
|----------|-------------|---------|
| `TokenIdentifier` | Variable/function names | `user`, `count` |
| `TokenString` | String literals | `"hello"`, `'world'` |
| `TokenNumber` | Numeric literals | `42`, `3.14` |
| `TokenKeyword` | Reserved words | `in`, `as`, `and`, `or` |
| `TokenSymbol` | Operators and punctuation | `(`, `)`, `=`, `,` |
| `TokenNil` | Nil value | `nil`, `None` |
