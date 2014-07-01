package pongo2

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
)

var reIdentifiers = regexp.MustCompile("^[a-zA-Z0-9_]+$")

type Context map[string]interface{}

func (c *Context) checkForValidIdentifiers() error {
	for k, v := range *c {
		if !reIdentifiers.MatchString(k) {
			return errors.New(fmt.Sprintf("Context-key '%s' (value: '%+v') is not a valid identifier.", k, v))
		}
	}
	return nil
}

type ExecutionContext struct {
	template    *Template
	Public      *Context
	Private     *Context
	StringStore map[string]string

	internalResolveValueStack []*reflect.Value // used within nodeResolver
}

func (ctx *ExecutionContext) isInternalResolveValueSet() bool {
	return ctx.internalResolveValueStack[len(ctx.internalResolveValueStack)-1] != nil
}

func (ctx *ExecutionContext) emptyInternalResolveValue() {
	ctx.internalResolveValueStack[len(ctx.internalResolveValueStack)-1] = nil
}

func (ctx *ExecutionContext) getInternalResolveValue() reflect.Value {
	return *ctx.internalResolveValueStack[len(ctx.internalResolveValueStack)-1]
}

func (ctx *ExecutionContext) setInternalResolveValue(v reflect.Value) {
	ctx.internalResolveValueStack[len(ctx.internalResolveValueStack)-1] = &v
}

func (ctx *ExecutionContext) pushInternalResolveValue() {
	ctx.internalResolveValueStack = append(ctx.internalResolveValueStack, nil)
}

func (ctx *ExecutionContext) popInternalResolveValue() {
	ctx.internalResolveValueStack = ctx.internalResolveValueStack[:len(ctx.internalResolveValueStack)-1]
}

func (ctx *ExecutionContext) Error(msg string, token *Token) error {
	pos := ""
	if token != nil {
		// No tokens available
		// TODO: Add location (from where?)
		pos = fmt.Sprintf(" | Line %d Col %d Value '%s' (%s)",
			token.Line, token.Col, token.Val, token.Type())
	}
	return errors.New(
		fmt.Sprintf("[Execution Error in '%s'%s] %s",
			ctx.template.name, pos, msg,
		))
}
