package pongo2

import (
	"fmt"
)

type FilterFunction func(in *Value, param *Value) (out *Value, err *Error)

// var filters map[string]FilterFunction

type PongoFilters struct {
	Filters map[string]FilterFunction
}

// func init() {
//     filters =
// }

func NewPongoFilters() PongoFilters {
	return PongoFilters{Filters: make(map[string]FilterFunction)}
}

// Registers a new filter. If there's already a filter with the same
// name, RegisterFilter will panic. You usually want to call this
// function in the filter's init() function:
// http://golang.org/doc/effective_go.html#init
//
// See http://www.florian-schlachter.de/post/pongo2/ for more about
// writing filters and tags.
func (f *PongoFilters) RegisterFilter(name string, fn FilterFunction) {
	_, existing := f.Filters[name]
	if existing {
		panic(fmt.Sprintf("Filter with name '%s' is already registered.", name))
	}
	f.Filters[name] = fn
}

func (f *PongoFilters) SetFilter(name string, fn FilterFunction) {
	f.Filters[name] = fn
}

// Replaces an already registered filter with a new implementation. Use this
// function with caution since it allows you to change existing filter behaviour.
func (f *PongoFilters) ReplaceFilter(name string, fn FilterFunction) {
	_, existing := f.Filters[name]
	if !existing {
		panic(fmt.Sprintf("Filter with name '%s' does not exist (therefore cannot be overridden).", name))
	}
	f.Filters[name] = fn
}

func (f *PongoFilters) GetFilter(name string) (FilterFunction, *Error) {
	filter, existing := f.Filters[name]
	if !existing {
		return nil, &Error{
			Sender:   "getfilter",
			ErrorMsg: fmt.Sprintf("Filter with name '%s' not found.", name),
		}
	}

	return filter, nil
}

// Like ApplyFilter, but panics on an error
func (f *PongoFilters) MustApplyFilter(name string, value *Value, param *Value) *Value {
	val, err := f.ApplyFilter(name, value, param)
	if err != nil {
		panic(err)
	}
	return val
}

// Applies a filter to a given value using the given parameters. Returns a *pongo2.Value or an error.
func (f *PongoFilters) ApplyFilter(name string, value *Value, param *Value) (*Value, *Error) {
	fn, existing := f.Filters[name]
	if !existing {
		return nil, &Error{
			Sender:   "applyfilter",
			ErrorMsg: fmt.Sprintf("Filter with name '%s' not found.", name),
		}
	}

	// Make sure param is a *Value
	if param == nil {
		param = AsValue(nil)
	}

	return fn(value, param)
}

type filterCall struct {
	token *Token

	name      string
	parameter IEvaluator

	filterFunc string
}

func (fc *filterCall) Execute(v *Value, ctx *ExecutionContext) (*Value, *Error) {
	var param *Value
	var err *Error

	if fc.parameter != nil {
		param, err = fc.parameter.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		param = AsValue(nil)
	}
	// Get filter func
	filterFunc, err := ctx.GetFilterFunc(fc.filterFunc)
	if err != nil {
		return nil, err.updateFromTokenIfNeeded(ctx.template, fc.token)
	}
	// Apply filter func
	filtered_value, err := filterFunc(v, param)
	if err != nil {
		return nil, err.updateFromTokenIfNeeded(ctx.template, fc.token)
	}
	return filtered_value, nil
}

// Filter = IDENT | IDENT ":" FilterArg | IDENT "|" Filter
func (p *Parser) parseFilter() (*filterCall, *Error) {
	ident_token := p.MatchType(TokenIdentifier)

	// Check filter ident
	if ident_token == nil {
		return nil, p.Error("Filter name must be an identifier.", nil)
	}

	filter := &filterCall{
		token: ident_token,
		name:  ident_token.Val,
	}

	filter.filterFunc = ident_token.Val

	// Check for filter-argument (2 tokens needed: ':' ARG)
	if p.Match(TokenSymbol, ":") != nil {
		if p.Peek(TokenSymbol, "}}") != nil {
			return nil, p.Error("Filter parameter required after ':'.", nil)
		}

		// Get filter argument expression
		v, err := p.parseVariableOrLiteral()
		if err != nil {
			return nil, err
		}
		filter.parameter = v
	}

	return filter, nil
}
