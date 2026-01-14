package pongo2

import (
	"errors"
	"fmt"
)

var autoescape = true

// SetAutoescape configures the global default autoescaping behavior for templates.
// When enabled (true), template output will be automatically HTML-escaped for safety.
// This setting affects all newly created ExecutionContexts.
func SetAutoescape(newValue bool) {
	autoescape = newValue
}

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
	for k, v := range other {
		c[k] = v
	}
	return c
}

// ExecutionContext contains all data important for the current rendering state.
//
// If you're writing a custom tag, your tag's Execute()-function will
// have access to the ExecutionContext. This struct stores anything
// about the current rendering process's Context including
// the Context provided by the user (field Public).
// You can safely use the Private context to provide data to the user's
// template (like a 'forloop'-information). The Shared-context is used
// to share data between tags. All ExecutionContexts share this context.
//
// Please be careful when accessing the Public data.
// PLEASE DO NOT MODIFY THE PUBLIC CONTEXT (read-only).
//
// To create your own execution context within tags, use the
// NewChildExecutionContext(parent) function.
type ExecutionContext struct {
	template   *Template
	macroDepth int

	Autoescape bool
	Public     Context
	Private    Context
	Shared     Context
}

var pongo2MetaContext = Context{
	"version": Version,
}

func newExecutionContext(tpl *Template, ctx Context) *ExecutionContext {
	privateCtx := make(Context)

	// Make the pongo2-related funcs/vars available to the context
	privateCtx["pongo2"] = pongo2MetaContext

	return &ExecutionContext{
		template: tpl,

		Public:     ctx,
		Private:    privateCtx,
		Autoescape: autoescape,
	}
}

// NewChildExecutionContext creates a new execution context that inherits from
// a parent context. The child context shares the same Public context and Shared
// context as the parent, but gets its own Private context (pre-populated with
// copies of the parent's private data). This is useful for custom tags that need
// to create isolated scopes while maintaining access to the template's data.
func NewChildExecutionContext(parent *ExecutionContext) *ExecutionContext {
	newctx := &ExecutionContext{
		template: parent.template,

		Public:     parent.Public,
		Private:    make(Context),
		Autoescape: parent.Autoescape,
	}
	newctx.Shared = parent.Shared

	// Copy all existing private items
	newctx.Private.Update(parent.Private)

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
