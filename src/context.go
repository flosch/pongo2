package pongo2

import (
	"errors"
	"fmt"
	"regexp"
)

var reIdentifiers = regexp.MustCompile("^[a-zA-Z0-9_]+$")

type Context map[string]interface{}

func (c *Context) checkForValidIdentifiers() error {
	for k, _ := range *c {
		if !reIdentifiers.MatchString(k) {
			return errors.New(fmt.Sprintf("Context-key '%s' is not a valid identifier.", k))
		}
	}
	return nil
}

type ExecutionContext struct {
	template    *Template
	Public      *Context
	Private     *Context
	StringStore map[string]string
}

func (ctx *ExecutionContext) Error(msg string, token *Token) error {
	pos := ""
	if token != nil {
		// No tokens available
		// TODO: Add location (from where?)
		pos = fmt.Sprintf(" | Line %d Col %d (%s)",
			token.Line, token.Col, token.String())
	}
	return errors.New(
		fmt.Sprintf("[Execution Error in %s%s] %s",
			ctx.template.name, pos, msg,
		))
}
