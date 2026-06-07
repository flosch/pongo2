package pongo2

import (
	"errors"
	"fmt"
	"maps"
)

// A Context type provides constants, variables, instances or functions to a template.
//
// pongo2 automatically provides meta-information or functions through the "pongo2"-key.
// Currently, context["pongo2"] contains the following keys:
//  1. version: returns the version string
//
// Template examples for accessing items from your context:
//
//	{{ myconstant }}
//	{{ myfunc("test", 42) }}
//	{{ user.name }}
//	{{ pongo2.version }}
type Context map[string]any

func (c Context) checkForValidIdentifiers() error {
	for k, v := range c {
		if !isValidIdentifier(k) {
			return &Error{
				Sender:    "checkForValidIdentifiers",
				OrigError: fmt.Errorf("context-key '%s' (value: '%+v') is not a valid identifier", k, v),
			}
		}
	}
	return nil
}

func isValidIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i := range s {
		if !isValidIdentifierChar(s[i]) {
			return false
		}
	}
	return true
}

func isValidIdentifierChar(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '_'
}

// Update updates this context with the key/value-pairs from another context.
func (c Context) Update(other Context) Context {
	maps.Copy(c, other)
	return c
}

// ExecutionContext holds the runtime state during template rendering.
//
// Custom tags receive this in their Execute() method. Use NewChildExecutionContext
// to create scoped child contexts within tags.
//
// Context hierarchy:
//   - Public: User data + set globals (READ-ONLY view)
//   - Private: Scoped engine data (layered per child context, not copied)
//   - Shared: Global state (same instance across all contexts)
type ExecutionContext struct {
	// The template being executed (provides config, inheritance, and TemplateSet access).
	template *Template

	// Tracks recursive macro call depth; errors if exceeding maxMacroDepth.
	macroDepth int

	// When true, {{ variable }} output is HTML-escaped. Toggle with {% autoescape %}.
	// The |safe filter bypasses escaping.
	Autoescape bool

	// User-provided data from Execute() overlaying the set's globals. Read-only:
	// access via Public.Get/Has/Range, never by indexing. The caller's map is
	// used directly (not copied), so it must not be mutated during rendering.
	Public PublicContext

	// parent links a child execution context to the one it was derived from,
	// forming the scope chain walked by Private. nil for a root context.
	parent *ExecutionContext

	// privateVars holds this context's own private layer. Allocated lazily on
	// first write. Parent layers are reached through parent, never copied.
	privateVars Context

	// Engine-managed scoped data (e.g., "forloop" from {% for %}, variables
	// from {% set %}, or macros). A child context layers its own data over the
	// parent's without copying; writes touch only the child's layer. Access via
	// Private.Get/Set/Delete/Range, never by indexing.
	Private Scope

	// Data shared across all contexts during a single render. Use for cross-scope
	// tag communication.
	Shared Context

	// tagState stores per-execution mutable state for tags that need it
	// (e.g., cycle index, ifchanged last values). Keyed by the tag node
	// pointer to ensure each tag instance has its own state. This map is
	// shared across all child contexts within a single execution, so tags
	// maintain consistent state regardless of nesting depth.
	tagState map[any]any
}

var pongo2MetaContext = Context{
	"version": Version,
}

func newExecutionContext(tpl *Template, ctx Context) *ExecutionContext {
	newctx := &ExecutionContext{
		template: tpl,

		// Make the pongo2-related funcs/vars available to the context
		privateVars: Context{"pongo2": pongo2MetaContext},
		Public:      PublicContext{vars: ctx, globals: tpl.set.Globals},
		Shared:      make(Context),
		Autoescape:  tpl.set.autoescape,
		tagState:    make(map[any]any),
	}
	newctx.Private = Scope{ec: newctx}
	return newctx
}

// NewChildExecutionContext creates a new execution context that inherits from
// a parent context. The child shares the same Public and Shared context as the
// parent, and layers a fresh, empty Private scope over the parent's (parent data
// is reached by walking the scope chain, never copied). Writes to the child's
// Private touch only the child's layer, leaving the parent untouched. Useful for
// custom tags that need an isolated scope while keeping access to outer data.
func NewChildExecutionContext(parent *ExecutionContext) *ExecutionContext {
	newctx := &ExecutionContext{
		template: parent.template,
		parent:   parent,

		Public:     parent.Public,
		Shared:     parent.Shared,
		Autoescape: parent.Autoescape,
		tagState:   parent.tagState,
	}
	newctx.Private = Scope{ec: newctx}
	return newctx
}

func (ctx *ExecutionContext) Error(msg string, token *Token) error {
	return ctx.OrigError(errors.New(msg), token)
}

func (ctx *ExecutionContext) OrigError(err error, token *Token) error {
	filename := ctx.template.name
	var line, col int
	if token != nil {
		// No tokens available
		// TODO: Add location (from where?)
		filename = token.Filename
		line = token.Line
		col = token.Col
	}
	return &Error{
		Template:  ctx.template,
		Filename:  filename,
		Line:      line,
		Column:    col,
		Token:     token,
		Sender:    "execution",
		OrigError: err,
	}
}

func (ctx *ExecutionContext) Logf(format string, args ...any) {
	ctx.template.set.logf(format, args...)
}
