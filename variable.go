package pongo2

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	varTypeInt = iota
	varTypeIdent
)

type variablePart struct {
	typ int
	s   string
	i   int

	is_function_call bool
	calling_args     []functionCallArgument // needed for a function call, represents all argument nodes (INode supports nested function calls)
}

type functionCallArgument interface {
	Evaluate(*ExecutionContext) (*Value, error)
}

type stringResolver string
type numberResolver int
type boolResolver bool

type variableResolver struct {
	location_token *Token // TODO: Use it in Evaluate()/Execute() for proper in-execution error messages

	parts []*variablePart
}

type NodeVariable struct {
	location_token *Token // TODO: Use it in Evaluate()/Execute() for proper in-execution error messages

	resolver    IEvaluator
	filterChain []*filterCall
}

func (s *stringResolver) Evaluate(ctx *ExecutionContext) (*Value, error) {
	return AsValue(string(*s)), nil
}

func (n *numberResolver) Evaluate(ctx *ExecutionContext) (*Value, error) {
	return AsValue(int(*n)), nil
}

func (b *boolResolver) Evaluate(ctx *ExecutionContext) (*Value, error) {
	return AsValue(bool(*b)), nil
}

func (vr *variableResolver) String() string {
	parts := make([]string, 0, len(vr.parts))
	for _, p := range vr.parts {
		switch p.typ {
		case varTypeInt:
			parts = append(parts, strconv.Itoa(p.i))
		case varTypeIdent:
			parts = append(parts, p.s)
		default:
			panic("unimplemented")
		}
	}
	return strings.Join(parts, ".")
}

func (vr *variableResolver) resolve(ctx *ExecutionContext) (*Value, error) {
	var current reflect.Value

	for idx, part := range vr.parts {
		if idx == 0 {
			// First part, get it from public context
			current = reflect.ValueOf((*ctx.Public)[vr.parts[0].s]) // Get the initial value
		} else {
			// Next parts, resolve it from current

			// Before resolving the pointer, let's see if we have a method to call
			// Problem with resolving the pointer is we're changing the receiver
			is_func := false
			if part.typ == varTypeIdent {
				func_value := current.MethodByName(part.s)
				if func_value.IsValid() {
					current = func_value
					is_func = true
				}
			}

			if !is_func {
				// If current a pointer, resolve it
				if current.Kind() == reflect.Ptr {
					current = current.Elem()
					if !current.IsValid() {
						// Value is not valid (anymore)
						return AsValue(nil), nil
					}
				}

				// Look up which part must be called now
				switch part.typ {
				case varTypeInt:
					// Calling an index is only possible for:
					// * slices/arrays/strings
					switch current.Kind() {
					case reflect.String, reflect.Array, reflect.Slice:
						current = current.Index(part.i)
					default:
						return nil, errors.New(fmt.Sprintf("Can't access an index on type %s (variable %s)", current.Kind().String(), vr.String()))
					}
				case varTypeIdent:
					// debugging:
					// fmt.Printf("now = %s (kind: %s)\n", part.s, current.Kind().String())

					// Calling a field or key
					switch current.Kind() {
					case reflect.Struct:
						current = current.FieldByName(part.s)
					case reflect.Map:
						current = current.MapIndex(reflect.ValueOf(part.s))
					default:
						return nil, errors.New(fmt.Sprintf("Can't access a field by name on type %s (variable %s)", current.Kind().String(), vr.String()))
					}
				default:
					panic("unimplemented")
				}
			}
		}

		if !current.IsValid() {
			// Value is not valid (anymore)
			return AsValue(nil), nil
		}

		// If current is a reflect.ValueOf(pongo2.Value), then unpack it
		// Happens in function calls (as a return value) or by injecting
		// into the execution context (e.g. in a for-loop)
		if current.Type() == reflect.TypeOf(&Value{}) {
			current = current.Interface().(*Value).v
		}

		// Check whether this is an interface and resolve it where required
		if current.Kind() == reflect.Interface {
			current = reflect.ValueOf(current.Interface())
		}

		// Check if the part is a function call
		if part.is_function_call || current.Kind() == reflect.Func {
			// Check for callable
			if current.Kind() != reflect.Func {
				return nil, errors.New(fmt.Sprintf("'%s' is not a function (it is %s).", vr.String(), current.Kind().String()))
			}

			// Check for correct function syntax and types
			// func(*Value, ...) *Value
			t := current.Type()

			// Input arguments
			for i := 0; i < t.NumIn(); i++ {
				if t.In(i) != reflect.TypeOf(new(Value)) {
					return nil, errors.New(fmt.Sprintf("Function input argument %d of '%s' must be of type *Value.", i, vr.String()))
				}
			}
			if len(part.calling_args) != t.NumIn() {
				return nil,
					errors.New(fmt.Sprintf("Function input argument count (%d) of '%s' must be equal to the calling argument count (%d).",
						t.NumIn(), vr.String(), len(part.calling_args)))
			}

			// Output arguments
			if t.NumOut() != 1 {
				return nil, errors.New(fmt.Sprintf("'%s' must have exactly 1 output argument.", vr.String()))
			}
			if t.Out(0) != reflect.TypeOf(new(Value)) {
				return nil, errors.New(fmt.Sprintf("Function return type of '%s' must be of type *Value.", vr.String()))
			}

			// Evaluate all parameters
			parameters := make([]reflect.Value, 0)
			for _, arg := range part.calling_args {
				pv, err := arg.Evaluate(ctx)
				if err != nil {
					return nil, err
				}
				parameters = append(parameters, reflect.ValueOf(pv))
			}

			// Call it
			rv := current.Call(parameters)

			// Return the function call value
			//return rv[0].Interface().(*Value), nil
			current = rv[0].Interface().(*Value).v
			// fmt.Printf("=> %+v %+T\n\n", current, current)
		}
	}

	if !current.IsValid() {
		// Value is not valid (e. g. NIL value)
		return AsValue(nil), nil
	}

	return &Value{current}, nil
}

// Is being used within a function call to get the argument
func (vr *variableResolver) Evaluate(ctx *ExecutionContext) (*Value, error) {
	value, err := vr.resolve(ctx)
	if err != nil {
		return AsValue(nil), ctx.Error(err.Error(), vr.location_token)
	}
	return value, nil
}

// Is being used within a function call to get the argument
func (v *NodeVariable) Evaluate(ctx *ExecutionContext) (*Value, error) {
	value, err := v.resolver.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	safe := false
	for _, filter := range v.filterChain {
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

// IDENT | IDENT.(IDENT|NUMBER)...
func (p *Parser) parseVariableOrLiteral() (IEvaluator, error) {
	t := p.Current()

	if t == nil {
		return nil, p.Error("Expected a number, string, keyword or identifier.", nil)
	}

	// Is first part a number or a string, there's nothing to resolve (because there's only to return the value then)
	switch t.Typ {
	case TokenNumber:
		p.Consume()
		i, err := strconv.Atoi(t.Val)
		if err != nil {
			return nil, err
		}
		nr := numberResolver(i)
		return &nr, nil
	case TokenString:
		p.Consume()
		sr := stringResolver(t.Val)
		return &sr, nil
	case TokenKeyword:
		p.Consume()
		switch t.Val {
		case "true":
			br := boolResolver(true)
			return &br, nil
		case "false":
			br := boolResolver(false)
			return &br, nil
		default:
			return nil, p.Error("This keyword is not allowed here.", nil)
		}
	}

	resolver := &variableResolver{
		location_token: t,
	}

	// First part of a variable MUST be an identifier
	if t.Typ != TokenIdentifier {
		return nil, p.Error(fmt.Sprintf("Variable (1st part, '%s') must be an identifier", t.Val), t)
	}

	resolver.parts = append(resolver.parts, &variablePart{
		typ: varTypeIdent,
		s:   t.Val,
	})

	p.Consume() // we consumed the first identifier of the variable name

variableLoop:
	for p.Remaining() > 0 {
		t = p.Current()

		if p.Match(TokenSymbol, ".") != nil {
			// Next variable part (can be either NUMBER or IDENT)
			t2 := p.Current()
			if t2 != nil {
				switch t2.Typ {
				case TokenIdentifier:
					resolver.parts = append(resolver.parts, &variablePart{
						typ: varTypeIdent,
						s:   t2.Val,
					})
					p.Consume() // consume: IDENT
					continue variableLoop
				case TokenNumber:
					i, err := strconv.Atoi(t2.Val)
					if err != nil {
						return nil, p.Error(err.Error(), t2)
					}
					resolver.parts = append(resolver.parts, &variablePart{
						typ: varTypeInt,
						i:   i,
					})
					p.Consume() // consume: NUMBER
					continue variableLoop
				default:
					return nil, p.Error("This token is not allowed within a variable name.", t2)
				}
			} else {
				// EOF
				return nil, p.Error("Unexpected EOF, expected either IDENTIFIER or NUMBER after DOT.", t)
			}
		} else if p.Match(TokenSymbol, "(") != nil {
			// Function call
			// FunctionName '(' Comma-separated list of expressions ')'
			part := resolver.parts[len(resolver.parts)-1]
			part.is_function_call = true
		argumentLoop:
			for {
				if p.Remaining() == 0 {
					return nil, p.Error("Unexpected EOF, expected function call argument list.", nil)
				}

				if p.Peek(TokenSymbol, ")") == nil {
					// No closing bracket, so we're parsing an expression
					expr_arg, err := p.ParseExpression()
					if err != nil {
						return nil, err
					}
					part.calling_args = append(part.calling_args, expr_arg)

					if p.Match(TokenSymbol, ")") != nil {
						// If there's a closing bracket after an expression, we will stop parsing the arguments
						break argumentLoop
					} else {
						// If there's NO closing bracket, there MUST be an comma
						if p.Match(TokenSymbol, ",") == nil {
							return nil, p.Error("Missing comma or closing bracket after argument.", nil)
						}
					}
				} else {
					// We got a closing bracket, so stop parsing arguments
					p.Consume()
					break argumentLoop
				}

			}
			// We're done parsing the function call, next variable part
			continue variableLoop
		}

		// No dot or function call? Then we're done with the variable parsing
		break
	}

	return resolver, nil
}

func (p *Parser) parseVariableOrLiteralWithFilter() (*NodeVariable, error) {
	v := &NodeVariable{
		location_token: p.Current(),
	}

	// Parse the variable name
	resolver, err := p.parseVariableOrLiteral()
	if err != nil {
		return nil, err
	}
	v.resolver = resolver

	// Parse all the filters
filterLoop:
	for p.Match(TokenSymbol, "|") != nil {
		// Parse one single filter
		filter, err := p.parseFilter()
		if err != nil {
			return nil, err
		}
		v.filterChain = append(v.filterChain, filter)

		continue filterLoop

		return nil, p.Error("This token is not allowed within a variable.", nil)
	}

	return v, nil
}

func (p *Parser) parseVariableElement() (INode, error) {
	p.Consume() // consume '{{'

	expr, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}

	if p.Match(TokenSymbol, "}}") == nil {
		return nil, p.Error("'}}' expected", nil)
	}

	return expr, nil
}
