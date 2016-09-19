package pongo2

import (
	"fmt"
	"regexp"
	"sync"
)

var reIdentifiers = regexp.MustCompile("^[a-zA-Z0-9_]+$")

type ContextMap map[string]interface{}

// A Context type provides constants, variables, instances or functions to a template.
//
// pongo2 automatically provides meta-information or functions through the "pongo2"-key.
// Currently, context["pongo2"] contains the following keys:
//  1. version: returns the version string
//
// Template examples for accessing items from your context:
//     {{ myconstant }}
//     {{ myfunc("test", 42) }}
//     {{ user.name }}
//     {{ pongo2.version }}
type Context struct {
	context ContextMap
	mut     *sync.RWMutex
}

func (c *Context) checkForValidIdentifiers() *Error {
	for k, v := range c.context {
		if !reIdentifiers.MatchString(k) {
			return &Error{
				Sender:   "checkForValidIdentifiers",
				ErrorMsg: fmt.Sprintf("Context-key '%s' (value: '%+v') is not a valid identifier.", k, v),
			}
		}
	}
	return nil
}

// Update updates this context with the key/value-pairs from another context.
func (c *Context) Update(other *Context) *Context {
	c.mut.Lock()
	for k, v := range other.context {
		c.context[k] = v
	}
	c.mut.Unlock()
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
	template *Template

	Autoescape bool
	Public     *Context
	Private    *Context
	Shared     *Context
}

func NewContext() *Context {
	context := make(ContextMap)
	return &Context{context, &sync.RWMutex{}}
}

func (c *Context) Set(key string, value interface{}) *Context {
	c.mut.Lock()
	c.context[key] = value
	c.mut.Unlock()
	return c
}

func (c *Context) SetMap(context ContextMap) *Context {
	c.mut.Lock()
	c.context = context
	c.mut.Unlock()
	return c
}

func (c *Context) GetMap() ContextMap {
	return c.context
}

func (c *Context) Get(key string) interface{} {
	return c.context[key]
}

func (c *Context) GetString(key string) string {
	return c.context[key].(string)
}

var pongo2MetaMap = ContextMap{"version": Version}

func newExecutionContext(tpl *Template, ctx *Context) *ExecutionContext {
	privateCtx := NewContext()

	// Make the pongo2-related funcs/vars available to the context
	privateCtx.Set("pongo2", pongo2MetaMap)

	return &ExecutionContext{
		template: tpl,

		Public:     ctx,
		Private:    privateCtx,
		Autoescape: true,
	}
}

func NewChildExecutionContext(parent *ExecutionContext) *ExecutionContext {
	newctx := &ExecutionContext{
		template: parent.template,

		Public:     parent.Public,
		Private:    NewContext(),
		Autoescape: parent.Autoescape,
	}
	newctx.Shared = parent.Shared

	// Copy all existing private items
	newctx.Private.Update(parent.Private)

	return newctx
}

func (ctx *ExecutionContext) Error(msg string, token *Token) *Error {
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
		Template: ctx.template,
		Filename: filename,
		Line:     line,
		Column:   col,
		Token:    token,
		Sender:   "execution",
		ErrorMsg: msg,
	}
}

func (ctx *ExecutionContext) Logf(format string, args ...interface{}) {
	ctx.template.set.logf(format, args...)
}
