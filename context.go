package pongo2

import (
	"errors"
	"fmt"
	"regexp"
)

var reIdentifiers = regexp.MustCompile("^[a-zA-Z0-9_]+$")

type Context map[string]interface{}

func (c Context) checkForValidIdentifiers() error {
	for k, v := range c {
		if !reIdentifiers.MatchString(k) {
			return errors.New(fmt.Sprintf("Context-key '%s' (value: '%+v') is not a valid identifier.", k, v))
		}
	}
	return nil
}

func (c Context) Update(other Context) Context {
	for k, v := range other {
		c[k] = v
	}
	return c
}

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
// NewExecutionContext(parent) function.
type ExecutionContext struct {
	template   *Template
	Autoescape bool
	Public     Context
	Private    Context
	Shared     Context
}

func newExecutionContext(tpl *Template, ctx Context) *ExecutionContext {
	return &ExecutionContext{
		template:   tpl,
		Public:     ctx,
		Private:    make(Context),
		Autoescape: true,
	}
}

func NewChildExecutionContext(parent *ExecutionContext) *ExecutionContext {
	newctx := &ExecutionContext{
		template:   parent.template,
		Public:     parent.Public,
		Private:    make(Context),
		Autoescape: parent.Autoescape,
	}
	newctx.Shared = parent.Shared

	// Copy all existing private items
	for key, value := range parent.Private {
		newctx.Private[key] = value
	}

	return newctx
}

func (ctx *ExecutionContext) Error(msg string, token *Token) error {
	pos := ""
	filename := ctx.template.name
	if token != nil {
		// No tokens available
		// TODO: Add location (from where?)
		filename = token.Filename
		pos = fmt.Sprintf(" | Line %d Col %d (%s)",
			token.Line, token.Col, token.String())
	}
	return errors.New(
		fmt.Sprintf("[Execution Error in %s%s] %s",
			filename, pos, msg,
		))
}
