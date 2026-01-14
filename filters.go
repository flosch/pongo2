package pongo2

import (
	"fmt"
	"maps"
)

// FilterFunction is the type filter functions must fulfil
type FilterFunction func(in *Value, param *Value) (out *Value, err error)

var builtinFilters = make(map[string]FilterFunction)

// copyFilters creates a shallow copy of a filter map.
func copyFilters(src map[string]FilterFunction) map[string]FilterFunction {
	dst := make(map[string]FilterFunction, len(src))
	maps.Copy(dst, src)
	return dst
}

// BuiltinFilterExists returns true if the given filter is a built-in filter.
// Use TemplateSet.FilterExists to check filters in a specific template set.
func BuiltinFilterExists(name string) bool {
	_, existing := builtinFilters[name]
	return existing
}

// BuiltinTagExists returns true if the given tag is registered in builtinTags.
// Use TemplateSet.TagExists to check tags in a specific template set.
func BuiltinTagExists(name string) bool {
	_, existing := builtinTags[name]
	return existing
}

// registerFilterBuiltin registers a new filter to the global filter map.
// This is used during package initialization to register builtin filters.
func registerFilterBuiltin(name string, fn FilterFunction) error {
	if BuiltinFilterExists(name) {
		return fmt.Errorf("filter with name '%s' is already registered", name)
	}
	builtinFilters[name] = fn
	return nil
}

// MustApplyFilter behaves like ApplyFilter, but panics on an error.
// This function uses builtinFilters. Use TemplateSet.MustApplyFilter for set-specific filters.
func MustApplyFilter(name string, value *Value, param *Value) *Value {
	val, err := ApplyFilter(name, value, param)
	if err != nil {
		panic(err)
	}
	return val
}

// ApplyFilter applies a built-infilter to a given value using the given
// parameters. Returns a *pongo2.Value or an error. Use TemplateSet.ApplyFilter
// for set-specific filters.
func ApplyFilter(name string, value *Value, param *Value) (*Value, error) {
	fn, existing := builtinFilters[name]
	if !existing {
		return nil, &Error{
			Sender:    "applyfilter",
			OrigError: fmt.Errorf("filter with name '%s' not found", name),
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

	filterFunc FilterFunction
}

func (fc *filterCall) Execute(v *Value, ctx *ExecutionContext) (*Value, error) {
	var param *Value
	var err error

	if fc.parameter != nil {
		param, err = fc.parameter.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		param = AsValue(nil)
	}

	filteredValue, err := fc.filterFunc(v, param)
	if err != nil {
		return nil, updateErrorToken(err, ctx.template, fc.token)
	}
	return filteredValue, nil
}

// Filter = IDENT | IDENT ":" FilterArg | IDENT "|" Filter
func (p *Parser) parseFilter() (*filterCall, error) {
	identToken := p.MatchType(TokenIdentifier)

	// Check filter ident
	if identToken == nil {
		return nil, p.Error("Filter name must be an identifier.", nil)
	}

	filter := &filterCall{
		token: identToken,
		name:  identToken.Val,
	}

	// Check sandbox filter restriction
	if _, isBanned := p.template.set.bannedFilters[identToken.Val]; isBanned {
		return nil, p.Error(fmt.Sprintf("Usage of filter '%s' is not allowed (sandbox restriction active).", identToken.Val), identToken)
	}

	// Get the appropriate filter function and bind it
	filterFn, exists := p.template.set.filters[identToken.Val]
	if !exists {
		return nil, p.Error(fmt.Sprintf("Filter '%s' does not exist.", identToken.Val), identToken)
	}

	filter.filterFunc = filterFn

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
