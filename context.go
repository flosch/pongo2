package pongo2

import (
	"errors"
	"fmt"
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

// If you're writing a custom tag, your tag's Execute()-function will
// have access to the ExecutionContext. This struct stores anything
// about the current rendering process's Context including
// the Context provided by the user (field Public).
// You can safely use the Private context and StringStore to exchange
// data between two tags etc. It is good practice to add a prefix to all your keys.
// Please be careful when modifying/accessing the Public data.
// Changes could lead to erroneous behaviour within pongo2 (e. g.
// if you're corrupting forloop-information). Make sure you're
// leaving the Public context as it was before when you're done
// with executing/rendering your tag.
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
