package pongo2

import (
	"fmt"
)

type FilterFunction func(in *Value, param *Value) (out *Value, err error)

var filters map[string]FilterFunction

func init() {
	filters = make(map[string]FilterFunction)
}

func RegisterFilter(name string, fn FilterFunction) {
	_, existing := filters[name]
	if existing {
		panic(fmt.Sprintf("Filter with name '%s' is already registered.", name))
	}
	filters[name] = fn
}

const (
	filterParamTypeIdx = iota
	filterParamTypeString
	filterParamTypeVariable
)

type filterCall struct {
	token *Token

	name      string
	parameter IEvaluator

	filterFunc FilterFunction
}

type nodeFilter struct {
	location_token *Token
	expr           IEvaluator
	filterChain    []*filterCall
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

	filtered_value, err := fc.filterFunc(v, param)
	if err != nil {
		return nil, ctx.Error(fmt.Sprintf("Error executing filter '%s': %s", fc.name, err.Error()), fc.token)
	}
	return filtered_value, nil
}

func (f *nodeFilter) Evaluate(ctx *ExecutionContext) (*Value, error) {
	value, err := f.expr.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	safe := false
	for _, filter := range f.filterChain {
		if filter.name == "safe" {
			safe = true
		}
		value, err = filter.Execute(value, ctx)
		if err != nil {
			return nil, err
		}
	}

	if !safe && value.IsString() {
		// apply escape filter
		value, err = filters["escape"](value, nil)
		if err != nil {
			return nil, err
		}
	}

	return value, nil
}

// Filter = IDENT | IDENT ":" FilterArg | IDENT "|" Filter
func (p *Parser) parseSingleFilter() (*filterCall, error) {
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

	// Check for filter-argument (2 tokens needed: ':' ARG)
	if p.Match(TokenSymbol, ":") != nil {
		param_token := p.Current()

		if param_token == nil {
			return nil, p.Error("Filter parameter required after ':'.", nil)
		}

		// Get filter argument
		v, err := p.parseFactor()
		if err != nil {
			return nil, err
		}
		filter.parameter = v
	}

	return filter, nil
}

func (p *Parser) parseFilter(expr IEvaluator) (IEvaluator, error) {
	// Check for filter (if there's none, we don't need a nodeFilter here)
	if p.Peek(TokenSymbol, "|") == nil {
		return expr, nil
	}

	f := &nodeFilter{
		expr: expr,
	}

	// Are there filters applied to the expression?
	// Parse all the filters
filterLoop:
	for p.Match(TokenSymbol, "|") != nil {
		// Parse one single filter
		filter, err := p.parseSingleFilter()
		if err != nil {
			return nil, err
		}
		f.filterChain = append(f.filterChain, filter)

		continue filterLoop
	}

	return f, nil
}
