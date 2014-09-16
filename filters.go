package pongo2

import (
	"fmt"
)

type FilterFunction func(in *Value, params ...*Value) (out *Value, err error)

var filters map[string]FilterFunction

func init() {
	filters = make(map[string]FilterFunction)
}

// Registers a new filter. If there's already a filter with the same
// name, RegisterFilter will panic. You usually want to call this
// function in the filter's init() function:
// http://golang.org/doc/effective_go.html#init
//
// See http://www.florian-schlachter.de/post/pongo2/ for more about
// writing filters and tags.
func RegisterFilter(name string, fn FilterFunction) {
	_, existing := filters[name]
	if existing {
		panic(fmt.Sprintf("Filter with name '%s' is already registered.", name))
	}
	filters[name] = fn
}

// Replaces an already registered filter with a new implementation. Use this
// function with caution since it allows you to change existing filter behaviour.
func ReplaceFilter(name string, fn FilterFunction) {
	_, existing := filters[name]
	if !existing {
		panic(fmt.Sprintf("Filter with name '%s' does not exist (therefore cannot be overridden).", name))
	}
	filters[name] = fn
}

// Like ApplyFilter, but panics on an error
func MustApplyFilter(name string, value *Value, params ...*Value) *Value {
	val, err := ApplyFilter(name, value, params...)
	if err != nil {
		panic(err)
	}
	return val
}

// Applies a filter to a given value using the given parameters. Returns a *pongo2.Value or an error.
func ApplyFilter(name string, value *Value, params ...*Value) (*Value, error) {
	fn, existing := filters[name]
	if !existing {
		return nil, fmt.Errorf("Filter with name '%s' not found.", name)
	}
	return fn(value, params...)
}

type filterCall struct {
	token		*Token

	name		string
	parameters	[]IEvaluator

	filterFunc	FilterFunction
}

func (fc *filterCall) Execute(v *Value, ctx *ExecutionContext) (*Value, error) {
	var params[] *Value

	if len(fc.parameters) > 0 {
		for _, paramEvo := range fc.parameters {
			param, err := paramEvo.Evaluate(ctx)
			if err != nil {
				return nil, err
			}
			params = append(params, param)
		}
	}

	filtered_value, err := fc.filterFunc(v, params...)
	if err != nil {
		return nil, ctx.Error(fmt.Sprintf("Error executing filter '%s': %s", fc.name, err.Error()), fc.token)
	}
	return filtered_value, nil
}

// Filter = IDENT | IDENT ":" FilterArg | IDENT "|" Filter
func (p *Parser) parseFilter() (*filterCall, error) {
	ident_token := p.MatchType(TokenIdentifier)

	// Check filter ident
	if ident_token == nil {
		return nil, p.Error("Filter name must be an identifier.", nil)
	}

	filter := &filterCall{
		token: ident_token,
		name:  ident_token.Val,
	}

	// Get the appropriate filter function and bind it
	filterFn, exists := filters[ident_token.Val]
	if !exists {
		return nil, p.Error(fmt.Sprintf("Filter '%s' does not exist.", ident_token.Val), ident_token)
	}

	filter.filterFunc = filterFn

	for {
		// Start reading parameter
		if p.Match(TokenSymbol, ":") == nil {
			break
		}

		if token := p.PeekOne(TokenSymbol, "}}", "|", ":"); token != nil {
			// Any way we got empty parameter
			filter.parameters = append(filter.parameters, &nilVariable{})

			// It is the end of tag, so lets break the loop
			if token.Val != ":" {
				break
			}
		} else {
			// Get filter argument expression
			v, err := p.parseVariableOrLiteral()
			if err != nil {
				return nil, err
			}
			filter.parameters = append(filter.parameters, v)
		}
	}

	return filter, nil
}
