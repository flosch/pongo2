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
	varTypeSubscript
	varTypeArray
	varTypeNil
)

var (
	typeOfValuePtr   = reflect.TypeFor[*Value]()
	typeOfExecCtxPtr = reflect.TypeFor[*ExecutionContext]()
)

type variablePart struct {
	typ       int
	s         string
	i         int
	subscript IEvaluator
	isNil     bool

	isFunctionCall bool
	callingArgs    []functionCallArgument // needed for a function call, represents all argument nodes (INode supports nested function calls)
}

func (p *variablePart) String() string {
	switch p.typ {
	case varTypeInt:
		return strconv.Itoa(p.i)
	case varTypeIdent:
		return p.s
	case varTypeSubscript:
		return "[subscript]"
	case varTypeArray:
		return "[array]"
	}

	panic("unimplemented")
}

type functionCallArgument interface {
	Evaluate(*ExecutionContext) (*Value, error)
}

// TODO: Add location tokens
type stringResolver struct {
	locationToken *Token
	val           string
}

type intResolver struct {
	locationToken *Token
	val           int
}

type floatResolver struct {
	locationToken *Token
	val           float64
}

type boolResolver struct {
	locationToken *Token
	val           bool
}

type variableResolver struct {
	locationToken *Token

	parts []*variablePart
}

type nodeFilteredVariable struct {
	locationToken *Token

	resolver    IEvaluator
	filterChain []*filterCall
}

type nodeVariable struct {
	locationToken *Token
	expr          IEvaluator
}

type executionCtxEval struct{}

// executeEvaluator is a helper that evaluates and writes a value to the writer.
func executeEvaluator(e IEvaluator, ctx *ExecutionContext, writer TemplateWriter) error {
	value, err := e.Evaluate(ctx)
	if err != nil {
		return err
	}
	_, err = writer.WriteString(value.String())
	return err
}

func (v *nodeFilteredVariable) Execute(ctx *ExecutionContext, writer TemplateWriter) error {
	return executeEvaluator(v, ctx, writer)
}

func (vr *variableResolver) Execute(ctx *ExecutionContext, writer TemplateWriter) error {
	return executeEvaluator(vr, ctx, writer)
}

func (s *stringResolver) Execute(ctx *ExecutionContext, writer TemplateWriter) error {
	return executeEvaluator(s, ctx, writer)
}

func (i *intResolver) Execute(ctx *ExecutionContext, writer TemplateWriter) error {
	return executeEvaluator(i, ctx, writer)
}

func (f *floatResolver) Execute(ctx *ExecutionContext, writer TemplateWriter) error {
	return executeEvaluator(f, ctx, writer)
}

func (b *boolResolver) Execute(ctx *ExecutionContext, writer TemplateWriter) error {
	return executeEvaluator(b, ctx, writer)
}

func (v *nodeFilteredVariable) GetPositionToken() *Token {
	return v.locationToken
}

func (vr *variableResolver) GetPositionToken() *Token {
	return vr.locationToken
}

func (s *stringResolver) GetPositionToken() *Token {
	return s.locationToken
}

func (i *intResolver) GetPositionToken() *Token {
	return i.locationToken
}

func (f *floatResolver) GetPositionToken() *Token {
	return f.locationToken
}

func (b *boolResolver) GetPositionToken() *Token {
	return b.locationToken
}

func (s *stringResolver) Evaluate(ctx *ExecutionContext) (*Value, error) {
	return AsValue(s.val), nil
}

func (i *intResolver) Evaluate(ctx *ExecutionContext) (*Value, error) {
	return AsValue(i.val), nil
}

func (f *floatResolver) Evaluate(ctx *ExecutionContext) (*Value, error) {
	return AsValue(f.val), nil
}

func (b *boolResolver) Evaluate(ctx *ExecutionContext) (*Value, error) {
	return AsValue(b.val), nil
}

func (s *stringResolver) FilterApplied(name string) bool {
	return false
}

func (i *intResolver) FilterApplied(name string) bool {
	return false
}

func (f *floatResolver) FilterApplied(name string) bool {
	return false
}

func (b *boolResolver) FilterApplied(name string) bool {
	return false
}

func (nv *nodeVariable) FilterApplied(name string) bool {
	return nv.expr.FilterApplied(name)
}

func (nv *nodeVariable) Execute(ctx *ExecutionContext, writer TemplateWriter) error {
	value, err := nv.expr.Evaluate(ctx)
	if err != nil {
		return err
	}

	if !nv.expr.FilterApplied("safe") && !value.safe && value.IsString() && ctx.Autoescape {
		// apply escape filter
		escapeFn := ctx.template.set.filters["escape"]
		if escapeFn != nil {
			value, err = escapeFn(value, nil)
			if err != nil {
				return err
			}
		}
	}

	_, err = writer.WriteString(value.String())
	return err
}

func (executionCtxEval) Evaluate(ctx *ExecutionContext) (*Value, error) {
	return AsValue(ctx), nil
}

func (vr *variableResolver) FilterApplied(name string) bool {
	return false
}

func (vr *variableResolver) String() string {
	parts := make([]string, 0, len(vr.parts))
	for _, p := range vr.parts {
		parts = append(parts, p.String())
	}

	return strings.Join(parts, ".")
}

func (vr *variableResolver) resolve(ctx *ExecutionContext) (*Value, error) {
	// Handle in-template array definition
	if len(vr.parts) > 0 && vr.parts[0].typ == varTypeArray {
		return vr.resolveArrayDefinition(ctx)
	}

	var current reflect.Value
	var isSafe bool

	for idx, part := range vr.parts {
		if idx == 0 {
			current = vr.lookupInitialValue(ctx)
		} else {
			resolved, isNil, err := vr.resolveNextPart(ctx, current, part)
			if err != nil {
				return nil, err
			}
			if isNil {
				return AsValue(nil), nil
			}
			current = resolved
		}

		if !current.IsValid() {
			return AsValue(nil), nil
		}

		// Unpack *Value if needed
		current, isSafe = vr.unpackValue(current, isSafe)

		// Resolve interface to concrete value
		if current.Kind() == reflect.Interface {
			current = reflect.ValueOf(current.Interface())
		}

		// Handle function call
		if part.isFunctionCall || current.Kind() == reflect.Func {
			result, err := vr.handleFunctionCall(ctx, current, part)
			if err != nil {
				return nil, err
			}
			current = result.value
			isSafe = result.isSafe
		}

		if !current.IsValid() {
			return AsValue(nil), nil
		}
	}

	return &Value{val: current, safe: isSafe}, nil
}

// resolveArrayDefinition handles in-template array definitions like [a, b, c].
func (vr *variableResolver) resolveArrayDefinition(ctx *ExecutionContext) (*Value, error) {
	items := make([]*Value, 0, len(vr.parts))
	for _, part := range vr.parts {
		v, ok := part.subscript.(*nodeFilteredVariable)
		if !ok {
			return nil, errors.New("unknown variable type is given")
		}
		item, err := v.resolver.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return &Value{val: reflect.ValueOf(items), safe: true}, nil
}

// lookupInitialValue looks up the first part of the variable in the context.
func (vr *variableResolver) lookupInitialValue(ctx *ExecutionContext) reflect.Value {
	val, inPrivate := ctx.Private[vr.parts[0].s]
	if !inPrivate {
		val = ctx.Public[vr.parts[0].s]
	}
	return reflect.ValueOf(val)
}

// unpackValue unpacks a *Value if the current value is of that type.
func (vr *variableResolver) unpackValue(current reflect.Value, isSafe bool) (reflect.Value, bool) {
	if current.Type() == typeOfValuePtr {
		tmpValue := current.Interface().(*Value)
		return tmpValue.val, tmpValue.safe
	}
	return current, isSafe
}

// resolveNextPart resolves the next part of a variable path from the current value.
// Returns (resolved value, isNil, error).
func (vr *variableResolver) resolveNextPart(
	ctx *ExecutionContext,
	current reflect.Value,
	part *variablePart,
) (reflect.Value, bool, error) {
	// Check for method call first
	if part.typ == varTypeIdent {
		funcValue := current.MethodByName(part.s)
		if funcValue.IsValid() {
			return funcValue, false, nil
		}
	}

	// Resolve pointer
	if current.Kind() == reflect.Ptr {
		current = current.Elem()
		if !current.IsValid() {
			return reflect.Value{}, true, nil
		}
	}

	return vr.resolvePartByType(ctx, current, part)
}

// resolvePartByType resolves a variable part based on its type.
func (vr *variableResolver) resolvePartByType(
	ctx *ExecutionContext,
	current reflect.Value,
	part *variablePart,
) (reflect.Value, bool, error) {
	switch part.typ {
	case varTypeInt:
		return vr.resolveIntIndex(current, part)
	case varTypeIdent:
		return vr.resolveIdentifier(current, part)
	case varTypeSubscript:
		return vr.resolveSubscript(ctx, current, part)
	default:
		panic("unimplemented")
	}
}

// resolveIntIndex resolves an integer index access on a slice/array/string.
func (vr *variableResolver) resolveIntIndex(current reflect.Value, part *variablePart) (reflect.Value, bool, error) {
	switch current.Kind() {
	case reflect.String:
		// For strings, return the character (rune) at the index (Django-compatible behavior)
		runes := []rune(current.String())
		if part.i >= 0 && len(runes) > part.i {
			return reflect.ValueOf(string(runes[part.i])), false, nil
		}
		return reflect.Value{}, true, nil
	case reflect.Array, reflect.Slice:
		if part.i >= 0 && current.Len() > part.i {
			return current.Index(part.i), false, nil
		}
		return reflect.Value{}, true, nil
	default:
		return reflect.Value{}, false, fmt.Errorf("can't access an index on type %s (variable %s)",
			current.Kind().String(), vr.String())
	}
}

// resolveIdentifier resolves a field or map key access by name.
func (vr *variableResolver) resolveIdentifier(current reflect.Value, part *variablePart) (reflect.Value, bool, error) {
	switch current.Kind() {
	case reflect.Struct:
		return current.FieldByName(part.s), false, nil
	case reflect.Map:
		return current.MapIndex(reflect.ValueOf(part.s)), false, nil
	default:
		return reflect.Value{}, false, fmt.Errorf("can't access a field by name on type %s (variable %s)",
			current.Kind().String(), vr.String())
	}
}

// resolveSubscript resolves a subscript access (e.g., foo[bar]).
func (vr *variableResolver) resolveSubscript(
	ctx *ExecutionContext,
	current reflect.Value,
	part *variablePart,
) (reflect.Value, bool, error) {
	sv, err := part.subscript.Evaluate(ctx)
	if err != nil {
		return reflect.Value{}, false, err
	}

	switch current.Kind() {
	case reflect.String:
		// For strings, return the character (rune) at the index (Django-compatible behavior)
		runes := []rune(current.String())
		si := sv.Integer()
		if si >= 0 && len(runes) > si {
			return reflect.ValueOf(string(runes[si])), false, nil
		}
		return reflect.Value{}, true, nil
	case reflect.Array, reflect.Slice:
		si := sv.Integer()
		if si >= 0 && current.Len() > si {
			return current.Index(si), false, nil
		}
		return reflect.Value{}, true, nil
	case reflect.Struct:
		return current.FieldByName(sv.String()), false, nil
	case reflect.Map:
		if sv.IsNil() {
			return reflect.Value{}, true, nil
		}
		if sv.val.Type().AssignableTo(current.Type().Key()) {
			return current.MapIndex(sv.val), false, nil
		}
		return reflect.Value{}, true, nil
	default:
		return reflect.Value{}, false, fmt.Errorf("can't access an index on type %s (variable %s)",
			current.Kind().String(), vr.String())
	}
}

// callResult holds the result of a function call resolution.
type callResult struct {
	value  reflect.Value
	isSafe bool
}

// handleFunctionCall processes a function call on the current value and returns the result.
func (vr *variableResolver) handleFunctionCall(
	ctx *ExecutionContext,
	current reflect.Value,
	part *variablePart,
) (*callResult, error) {
	if current.Kind() != reflect.Func {
		return nil, fmt.Errorf("'%s' is not a function (it is %s)", vr.String(), current.Kind().String())
	}

	t := current.Type()
	currArgs := part.callingArgs

	// If an implicit ExecCtx is needed
	if t.NumIn() > 0 && t.In(0) == typeOfExecCtxPtr {
		currArgs = append([]functionCallArgument{executionCtxEval{}}, currArgs...)
	}

	// Validate input argument count
	if len(currArgs) != t.NumIn() && (len(currArgs) < t.NumIn()-1 || !t.IsVariadic()) {
		return nil, fmt.Errorf("function input argument count (%d) of '%s' must be equal to the calling argument count (%d)",
			t.NumIn(), vr.String(), len(currArgs))
	}

	// Validate output argument count
	if t.NumOut() != 1 && t.NumOut() != 2 {
		return nil, fmt.Errorf("'%s' must have exactly 1 or 2 output arguments, the second argument must be of type error", vr.String())
	}

	// Evaluate and prepare parameters
	parameters, err := vr.prepareCallParameters(ctx, t, currArgs)
	if err != nil {
		return nil, err
	}

	// Execute the function call
	return vr.executeCall(current, t, parameters)
}

// prepareCallParameters evaluates arguments and prepares them for function call.
func (vr *variableResolver) prepareCallParameters(
	ctx *ExecutionContext,
	t reflect.Type,
	currArgs []functionCallArgument,
) ([]reflect.Value, error) {
	var parameters []reflect.Value
	numArgs := t.NumIn()
	isVariadic := t.IsVariadic()

	for idx, arg := range currArgs {
		pv, err := arg.Evaluate(ctx)
		if err != nil {
			return nil, err
		}

		fnArg := vr.getFnArgType(t, idx, numArgs, isVariadic)

		param, err := vr.convertArgToParam(pv, fnArg, idx, isVariadic)
		if err != nil {
			return nil, err
		}
		parameters = append(parameters, param)
	}

	// Validate all parameters
	for _, p := range parameters {
		if p.Kind() == reflect.Invalid {
			return nil, fmt.Errorf("calling a function using an invalid parameter")
		}
	}

	return parameters, nil
}

// getFnArgType returns the expected type for a function argument at the given index.
func (vr *variableResolver) getFnArgType(t reflect.Type, idx, numArgs int, isVariadic bool) reflect.Type {
	if isVariadic && idx >= t.NumIn()-1 {
		return t.In(numArgs - 1).Elem()
	}
	return t.In(idx)
}

// convertArgToParam converts an evaluated Value to a reflect.Value suitable for function call.
func (vr *variableResolver) convertArgToParam(pv *Value, fnArg reflect.Type, idx int, isVariadic bool) (reflect.Value, error) {
	if fnArg == typeOfValuePtr {
		return reflect.ValueOf(pv), nil
	}

	// Check type compatibility
	if fnArg != reflect.TypeOf(pv.Interface()) && fnArg.Kind() != reflect.Interface {
		if isVariadic {
			return reflect.Value{}, fmt.Errorf("function variadic input argument of '%s' must be of type %s or *pongo2.Value (not %T)",
				vr.String(), fnArg.String(), pv.Interface())
		}
		return reflect.Value{}, fmt.Errorf("function input argument %d of '%s' must be of type %s or *pongo2.Value (not %T)",
			idx, vr.String(), fnArg.String(), pv.Interface())
	}

	if pv.IsNil() {
		var empty any
		return reflect.ValueOf(&empty).Elem(), nil
	}
	return reflect.ValueOf(pv.Interface()), nil
}

// executeCall performs the actual function call and processes the result.
func (vr *variableResolver) executeCall(
	current reflect.Value,
	t reflect.Type,
	parameters []reflect.Value,
) (*callResult, error) {
	values := current.Call(parameters)
	rv := values[0]

	// Check for error return value
	if t.NumOut() == 2 {
		if e := values[1].Interface(); e != nil {
			err, ok := e.(error)
			if !ok {
				return nil, fmt.Errorf("the second return value is not an error")
			}
			if err != nil {
				return nil, err
			}
		}
	}

	result := &callResult{}
	if rv.Type() != typeOfValuePtr {
		result.value = reflect.ValueOf(rv.Interface())
	} else {
		val := rv.Interface().(*Value)
		result.value = val.val
		result.isSafe = val.safe
	}

	return result, nil
}

func (vr *variableResolver) Evaluate(ctx *ExecutionContext) (*Value, error) {
	value, err := vr.resolve(ctx)
	if err != nil {
		return AsValue(nil), ctx.Error(err.Error(), vr.locationToken)
	}
	return value, nil
}

func (v *nodeFilteredVariable) FilterApplied(name string) bool {
	for _, filter := range v.filterChain {
		if filter.name == name {
			return true
		}
	}
	return false
}

func (v *nodeFilteredVariable) Evaluate(ctx *ExecutionContext) (*Value, error) {
	value, err := v.resolver.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	for _, filter := range v.filterChain {
		value, err = filter.Execute(value, ctx)
		if err != nil {
			return nil, err
		}
	}

	return value, nil
}

// "[" [expr {, expr}] "]"
func (p *Parser) parseArray() (IEvaluator, error) {
	resolver := &variableResolver{
		locationToken: p.Current(),
	}
	p.Consume() // We consume '['

	// We allow an empty list, so check for a closing bracket.
	if p.Match(TokenSymbol, "]") != nil {
		return resolver, nil
	}

	// parsing an array declaration with at least one expression
	for {
		if p.Remaining() == 0 {
			return nil, p.Error("Unexpected EOF, unclosed array list.", p.lastToken)
		}

		// No closing bracket, so we're parsing an expression
		exprArg, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}

		resolver.parts = append(resolver.parts, &variablePart{
			typ:       varTypeArray,
			subscript: exprArg,
		})

		if p.Match(TokenSymbol, "]") != nil {
			// If there's a closing bracket after an expression, we will stop parsing the arguments
			break
		}

		// If there's NO closing bracket, there MUST be an comma
		if p.Match(TokenSymbol, ",") == nil {
			return nil, p.Error("Missing comma or closing bracket after argument.", p.Current())
		}
	}

	return resolver, nil
}

func (p *Parser) parseNumberLiteral(sign int, numToken *Token, locToken *Token) (IEvaluator, error) {
	// One exception to the rule that we don't have float64 literals is at the beginning
	// of an expression (or a variable name). Since we know we started with an integer
	// which can't obviously be a variable name, we can check whether the first number
	// is followed by dot (and then a number again). If so we're converting it to a float64.
	if p.Match(TokenSymbol, ".") != nil {
		t2 := p.MatchType(TokenNumber)
		if t2 == nil {
			return nil, p.Error("Expected a number after the '.'.", nil)
		}
		f, err := strconv.ParseFloat(fmt.Sprintf("%s.%s", numToken.Val, t2.Val), 64)
		if err != nil {
			return nil, p.Error(err.Error(), locToken)
		}
		return &floatResolver{locationToken: locToken, val: float64(sign) * f}, nil
	}
	i, err := strconv.Atoi(numToken.Val)
	if err != nil {
		return nil, p.Error(err.Error(), numToken)
	}
	return &intResolver{locationToken: locToken, val: sign * i}, nil
}

// IDENT | IDENT.(IDENT|NUMBER)... | IDENT[expr]... | "[" [ expr {, expr}] "]"
//
//nolint:gocyclo,cyclop,funlen // parser for variable expressions handles many token types
func (p *Parser) parseVariableOrLiteral() (IEvaluator, error) {
	t := p.Current()

	if t == nil {
		return nil, p.Error("Unexpected EOF, expected a number, string, keyword or identifier.", p.lastToken)
	}

	// Is first part a number or a string, there's nothing to resolve (because there's only to return the value then)
	switch t.Typ {
	case TokenNumber:
		p.Consume()
		return p.parseNumberLiteral(1, t, t)
	case TokenString:
		p.Consume()
		sr := &stringResolver{
			locationToken: t,
			val:           t.Val,
		}
		return sr, nil
	case TokenKeyword:
		p.Consume()
		switch t.Val {
		case "true":
			br := &boolResolver{
				locationToken: t,
				val:           true,
			}
			return br, nil
		case "false":
			br := &boolResolver{
				locationToken: t,
				val:           false,
			}
			return br, nil
		default:
			return nil, p.Error("This keyword is not allowed here.", nil)
		}
	case TokenSymbol:
		if t.Val == "[" {
			// Parsing an array literal [expr {, expr}]
			return p.parseArray()
		}
		if t.Val == "-" {
			// Negative number literal
			p.Consume() // consume '-'
			t2 := p.Current()
			if t2 == nil || t2.Typ != TokenNumber {
				return nil, p.Error("Expected a number after '-'.", t)
			}
			p.Consume() // consume the number
			return p.parseNumberLiteral(-1, t2, t)
		}
	}

	resolver := &variableResolver{
		locationToken: t,
	}

	if t.Typ != TokenIdentifier {
		// First part of a variable MUST be an identifier
		return nil, p.Error("Expected either a number, string, keyword or identifier.", t)
	}

	resolver.parts = append(resolver.parts, &variablePart{
		typ: varTypeIdent,
		s:   t.Val,
	})
	p.Consume() // we consumed the first identifier of the variable name

variableLoop:
	for p.Remaining() > 0 {
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
				case TokenNil:
					resolver.parts = append(resolver.parts, &variablePart{
						typ:   varTypeNil,
						isNil: true,
					})
					p.Consume() // consume: NIL
					continue variableLoop
				default:
					return nil, p.Error("This token is not allowed within a variable name.", t2)
				}
			} else {
				// EOF
				return nil, p.Error("Unexpected EOF, expected either IDENTIFIER or NUMBER after DOT.",
					p.lastToken)
			}
		} else if p.Match(TokenSymbol, "[") != nil {
			// Variable subscript
			if p.Remaining() == 0 {
				return nil, p.Error("Unexpected EOF, expected subscript subscript.", p.lastToken)
			}

			exprSubscript, err := p.ParseExpression()
			if err != nil {
				return nil, err
			}
			resolver.parts = append(resolver.parts, &variablePart{
				typ:       varTypeSubscript,
				subscript: exprSubscript,
			})
			if p.Match(TokenSymbol, "]") == nil {
				return nil, p.Error("Missing closing bracket after subscript argument.", nil)
			}

		} else if p.Match(TokenSymbol, "(") != nil {
			// Function call
			// FunctionName '(' Comma-separated list of expressions ')'
			part := resolver.parts[len(resolver.parts)-1]
			part.isFunctionCall = true
		argumentLoop:
			for {
				if p.Remaining() == 0 {
					return nil, p.Error("Unexpected EOF, expected function call argument list.", p.lastToken)
				}

				if p.Peek(TokenSymbol, ")") == nil {
					// No closing bracket, so we're parsing an expression
					exprArg, err := p.ParseExpression()
					if err != nil {
						return nil, err
					}
					part.callingArgs = append(part.callingArgs, exprArg)

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

		// No dot, subscript or function call? Then we're done with the variable parsing
		break
	}

	return resolver, nil
}

func (p *Parser) parseVariableOrLiteralWithFilter() (*nodeFilteredVariable, error) {
	v := &nodeFilteredVariable{
		locationToken: p.Current(),
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

		// Check sandbox filter restriction
		if _, isBanned := p.template.set.bannedFilters[filter.name]; isBanned {
			return nil, p.Error(fmt.Sprintf("Usage of filter '%s' is not allowed (sandbox restriction active).", filter.name), nil)
		}

		v.filterChain = append(v.filterChain, filter)

		continue filterLoop
	}

	return v, nil
}

func (p *Parser) parseVariableElement() (INode, error) {
	node := &nodeVariable{
		locationToken: p.Current(),
	}

	p.Consume() // consume '{{'

	expr, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}
	node.expr = expr

	if p.Match(TokenSymbol, "}}") == nil {
		return nil, p.Error("'}}' expected", nil)
	}

	return node, nil
}
