package pongo2

import (
	"fmt"
	"strconv"
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

type filterParameter struct {
	typ int
	i   int
	s   string
	v   IEvaluator
}

type filterCall struct {
	token *Token

	name      string
	parameter *filterParameter

	filterFunc FilterFunction
}

func (fc *filterCall) Execute(v *Value, ctx *ExecutionContext) (*Value, error) {
	var param *Value
	var err error

	if fc.parameter != nil {
		switch fc.parameter.typ {
		case filterParamTypeVariable:
			// First get variable content
			param, err = fc.parameter.v.Evaluate(ctx)
			if err != nil {
				return nil, err
			}
		case filterParamTypeIdx:
			param = AsValue(fc.parameter.i)
		case filterParamTypeString:
			param = AsValue(fc.parameter.s)
		default:
			panic("unimplemented")
		}
	} else {
		// No parameter available
		param = AsValue(0)
	}

	return fc.filterFunc(v, param)
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

	// Check for filter-argument (2 tokens needed: ':' ARG)
	if p.Match(TokenSymbol, ":") != nil {
		param_token := p.Current()

		if param_token == nil {
			return nil, p.Error("Filter parameter required after ':'.", nil)
		}

		// Check argument type
		// Filter arguments allowed: IDENTIFIER (beginning of a variable), STRING, NUMBER
		switch param_token.Typ {
		case TokenIdentifier:
			v, err := p.ParseExpression()
			if err != nil {
				return nil, err
			}
			filter.parameter = &filterParameter{
				typ: filterParamTypeVariable,
				v:   v,
			}
		case TokenString:
			filter.parameter = &filterParameter{
				typ: filterParamTypeString,
				s:   param_token.Val,
			}
			p.Consume() // consume STRING
		case TokenNumber:
			i, err := strconv.Atoi(param_token.Val)
			if err != nil {
				return nil, p.Error(err.Error(), param_token)
			}
			filter.parameter = &filterParameter{
				typ: filterParamTypeIdx,
				i:   i,
			}
			p.Consume() // consume NUMBER
		default:
			return nil, p.Error("Filter parameter invalid.", nil)
		}
	}

	return filter, nil
}
