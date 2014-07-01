package pongo2

import (
	"fmt"
	"reflect"
	"strconv"
)

const (
	varTypeInt = iota
	varTypeIdent
)

/*
type variablePart struct {
	typ int
	s   string
	i   int

	is_function_call bool
	calling_args     []functionCallArgument // needed for a function call, represents all argument nodes (INode supports nested function calls)
}
*/

type stringResolver string
type identResolver string
type numberResolver int
type boolResolver bool

type nodeResolver struct {
	location_token *Token // TODO: Use it in Evaluate()/Execute() for proper in-execution error messages
	expr           INodeEvaluator
	resolver       IEvaluator
	args           []IEvaluator
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

func (i *identResolver) Evaluate(ctx *ExecutionContext) (*Value, error) {
	return AsValue(string(*i)), nil
}

func (i *identResolver) Execute(ctx *ExecutionContext) (string, error) {
	return string(*i), nil
}

func (vr *nodeResolver) Evaluate(ctx *ExecutionContext) (*Value, error) {
	lookup_value, err := vr.expr.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	if !ctx.isInternalResolveValueSet() {
		//fmt.Printf("fresh start with lookup name=%s\n", lookup_value.String())
		// No current value, then do a lookup in the Public context and fill
		// ctx.internalResolveValue with an initial value

		// Key will be of type identifier (name is accessable by String())
		ctx.setInternalResolveValue(reflect.ValueOf((*ctx.Public)[lookup_value.String()]))
	} else {
		if !ctx.getInternalResolveValue().IsValid() {
			// Value is not valid (anymore)
			return AsValue(nil), nil
		}

		//fmt.Printf("next lookup name=%s\n", lookup_value.String())
		// Before resolving the pointer, let's see if we have a method to call
		// Problem with resolving the pointer is we're changing the receiver
		is_func := false

		// There are only functions with a NAME (as string; not an integer)
		if lookup_value.IsString() {
			func_value := ctx.getInternalResolveValue().MethodByName(lookup_value.String())
			if func_value.IsValid() {
				ctx.setInternalResolveValue(func_value)
				is_func = true
			}
		}

		// No function? Okay, then let's resolve the part
		if !is_func {
			// If current a pointer, resolve it
			if ctx.getInternalResolveValue().Kind() == reflect.Ptr {
				ctx.setInternalResolveValue(ctx.getInternalResolveValue().Elem())
				if !ctx.getInternalResolveValue().IsValid() {
					// Value is not valid (anymore)
					return AsValue(nil), nil
				}
			}

			// Look up which part must be called now
			if lookup_value.IsInteger() {
				// Calling an index is only possible for:
				// * slices/arrays/strings
				switch ctx.getInternalResolveValue().Kind() {
				case reflect.String, reflect.Array, reflect.Slice:
					ctx.setInternalResolveValue(ctx.getInternalResolveValue().Index(lookup_value.Integer()))
				default:
					return nil, ctx.Error(fmt.Sprintf("Can't access an index on type %s (lookup-index: %d)",
						ctx.getInternalResolveValue().Kind().String(), lookup_value.Integer()), vr.location_token)
				}
			} else if lookup_value.IsString() {
				// debugging:
				// fmt.Printf("now = %s (kind: %s)\n", part.s, current.Kind().String())

				// Calling a field or key
				switch ctx.getInternalResolveValue().Kind() {
				case reflect.Struct:
					ctx.setInternalResolveValue(ctx.getInternalResolveValue().FieldByName(lookup_value.String()))
				case reflect.Map:
					ctx.setInternalResolveValue(ctx.getInternalResolveValue().MapIndex(reflect.ValueOf(lookup_value.String())))
				default:
					return nil, ctx.Error(fmt.Sprintf("Can't access a field by name on type %s (lookup-name: '%s')",
						ctx.getInternalResolveValue().Kind().String(), lookup_value.String()), vr.location_token)
				}
			} else {
				panic("resolver type unimplemented")
			}
		}
	}

	if !ctx.getInternalResolveValue().IsValid() {
		// Value is not valid (anymore)
		return AsValue(nil), nil
	}

	// If current is a reflect.ValueOf(pongo2.Value), then unpack it
	// Happens in function calls (as a return value) or by injecting
	// into the execution context (e.g. in a for-loop)
	if ctx.getInternalResolveValue().Type() == reflect.TypeOf(&Value{}) {
		ctx.setInternalResolveValue(ctx.getInternalResolveValue().Interface().(*Value).v)
	}

	// Check whether this is an interface and resolve it where required
	if ctx.getInternalResolveValue().Kind() == reflect.Interface {
		ctx.setInternalResolveValue(reflect.ValueOf(ctx.getInternalResolveValue().Interface()))
	}

	// Check if the part is a function call
	if ctx.getInternalResolveValue().Kind() == reflect.Func {
		// Check for correct function syntax and types
		// func(*Value, ...) *Value
		t := ctx.getInternalResolveValue().Type()

		// Input arguments
		for i := 0; i < t.NumIn(); i++ {
			if t.In(i) != reflect.TypeOf(new(Value)) {
				return nil, ctx.Error(fmt.Sprintf("Function input argument %d must be of type *Value.", i), vr.location_token)
			}
		}
		if len(vr.args) != t.NumIn() {
			return nil,
				ctx.Error(fmt.Sprintf("Function input argument count (%d) must be equal to the calling argument count (%d).",
					t.NumIn(), len(vr.args)), vr.location_token)
		}

		// Output arguments
		if t.NumOut() != 1 {
			return nil, ctx.Error(fmt.Sprintf("'%s' must have exactly 1 output argument.",
				lookup_value.String()), vr.location_token)
		}
		if t.Out(0) != reflect.TypeOf(new(Value)) {
			return nil, ctx.Error(fmt.Sprintf("Function return type of '%s' must be of type *Value.",
				lookup_value.String()), vr.location_token)
		}

		// Evaluate all parameters
		parameters := make([]reflect.Value, 0)
		for _, arg := range vr.args {
			ctx.pushInternalResolveValue()
			pv, err := arg.Evaluate(ctx)
			if err != nil {
				return nil, err
			}
			ctx.popInternalResolveValue()
			parameters = append(parameters, reflect.ValueOf(pv))
		}

		// Call it
		rv := ctx.getInternalResolveValue().Call(parameters)

		// Return the function call value
		//return rv[0].Interface().(*Value), nil
		ctx.setInternalResolveValue(rv[0].Interface().(*Value).v)
		// fmt.Printf("=> %+v %+T\n\n", current, current)
	}

	if vr.resolver == nil {
		v := ctx.getInternalResolveValue()
		ctx.emptyInternalResolveValue()
		return &Value{v}, nil
	}
	return vr.resolver.Evaluate(ctx)
}

func (p *Parser) parseLiteral() (IEvaluator, error) {
	t := p.Current()

	if t == nil {
		return nil, p.Error("Expected a number, string, keyword or identifier.", nil)
	}

	// Literals (Number [integer], String, Keyword [true/false])
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

	return nil, p.Error("Unrecognized literal.", nil)
}

// $expr '.' $expr
func (p *Parser) parseResolver() (IEvaluator, error) {

	if p.PeekType(TokenNumber) != nil || p.PeekType(TokenString) != nil ||
		p.PeekType(TokenKeyword) != nil {
		// parse literal
		return p.parseLiteral()
	}

	nr := &nodeResolver{
		location_token: p.Current(),
	}

	if ident_token := p.MatchType(TokenIdentifier); ident_token != nil {
		ir := identResolver(ident_token.Val)
		nr.expr = &ir
	} else {
		// Otherwise we're parsing an expression
		if p.Match(TokenSymbol, "(") != nil {
			br_expr, err := p.ParseExpression()
			if err != nil {
				return nil, err
			}
			if p.Match(TokenSymbol, ")") == nil {
				return nil, p.Error("Closing bracket expected after expression", nil)
			}
			/*filter_expr, err := p.parseFilter(br_expr)
			if err != nil {
				return nil, err
			}*/
			nr.expr = br_expr
		} else {
			expr, err := p.ParseExpression()
			if err != nil {
				return nil, err
			}
			nr.expr = expr
		}
	}

	// Is there anything to resolve?
	if p.Match(TokenSymbol, ".") != nil {
		resolver, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}
		nr.resolver = resolver
	} else if p.Match(TokenSymbol, "(") != nil {
		// Function call
		// FunctionName '(' Comma-separated list of expressions ')'

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
				nr.args = append(nr.args, expr_arg)

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
	}

	return p.parseFilter(nr)
}

func (p *Parser) parseLiteralWithFilter() (IEvaluator, error) {
	val, err := p.parseResolver()
	if err != nil {
		return nil, err
	}

	return p.parseFilter(val)
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
