package pongo2

import (
	"bytes"
	"fmt"
)

// maxMacroDepth limits the maximum depth of recursive macro calls.
// This prevents infinite recursion (e.g., a macro calling itself without
// a base case) from causing a stack overflow. When a macro is called,
// macroDepth in ExecutionContext is incremented; if it exceeds this limit,
// an error is returned. The limit of 1000 allows for reasonable nesting
// while protecting against runaway recursion.
const maxMacroDepth = 1000

// tagMacroNode represents the {% macro %} tag.
//
// The macro tag defines reusable template fragments that can be called like
// functions. Macros can accept arguments with optional default values.
//
// Basic macro definition:
//
//	{% macro greeting(name) %}
//	    Hello, {{ name }}!
//	{% endmacro %}
//
// Calling a macro:
//
//	{{ greeting("World") }}
//
// Output: "Hello, World!"
//
// Macro with default argument values:
//
//	{% macro button(text, type="primary", disabled=false) %}
//	    <button class="btn-{{ type }}"{% if disabled %} disabled{% endif %}>
//	        {{ text }}
//	    </button>
//	{% endmacro %}
//
//	{{ button("Click me") }}
//	{{ button("Submit", type="success") }}
//	{{ button("Disabled", disabled=true) }}
//
// Exporting macros for use in other templates:
//
//	{% macro input_field(name, label) export %}
//	    <label for="{{ name }}">{{ label }}</label>
//	    <input type="text" id="{{ name }}" name="{{ name }}">
//	{% endmacro %}
//
// Exported macros can be imported using the {% import %} tag:
//
//	{% import "forms/macros.html" input_field %}
//	{{ input_field("email", "Email Address") }}
//
// Note: Recursive macro calls are limited to a depth of 1000 to prevent
// infinite recursion.
type tagMacroNode struct {
	position  *Token
	name      string
	argsOrder []string
	args      map[string]IEvaluator
	exported  bool

	wrapper *NodeWrapper
}

// Execute registers the macro as a callable function in the private context.
// The macro can then be called like {{ macro_name(args) }}.
func (node *tagMacroNode) Execute(ctx *ExecutionContext, writer TemplateWriter) error {
	ctx.Private[node.name] = func(args ...*Value) (*Value, error) {
		ctx.macroDepth++
		defer func() {
			ctx.macroDepth--
		}()

		if ctx.macroDepth > maxMacroDepth {
			return nil, ctx.Error(fmt.Sprintf("maximum recursive macro call depth reached (max is %v)", maxMacroDepth), node.position)
		}

		return node.call(ctx, args...)
	}

	return nil
}

// call executes the macro body with the provided arguments and returns the
// rendered output as a safe value. It creates an isolated context for execution.
func (node *tagMacroNode) call(ctx *ExecutionContext, args ...*Value) (*Value, error) {
	argsCtx := make(Context)

	for k, v := range node.args {
		if v == nil {
			// User did not provided a default value
			argsCtx[k] = nil
		} else {
			// Evaluate the default value
			valueExpr, err := v.Evaluate(ctx)
			if err != nil {
				ctx.Logf(err.Error())
				return AsSafeValue(""), err
			}

			argsCtx[k] = valueExpr.Interface()
		}
	}

	if len(args) > len(node.argsOrder) {
		// Too many arguments, we're ignoring them and just logging into debug mode.
		err := ctx.Error(fmt.Sprintf("Macro '%s' called with too many arguments (%d instead of %d).",
			node.name, len(args), len(node.argsOrder)), node.position)

		return AsSafeValue(""), err
	}

	// Make a context for the macro execution
	macroCtx := NewChildExecutionContext(ctx)

	// Register all arguments in the private context
	macroCtx.Private.Update(argsCtx)

	for idx, argValue := range args {
		macroCtx.Private[node.argsOrder[idx]] = argValue.Interface()
	}

	var b bytes.Buffer
	err := node.wrapper.Execute(macroCtx, &b)
	if err != nil {
		return AsSafeValue(""), updateErrorToken(err, ctx.template, node.position)
	}

	return AsSafeValue(b.String()), nil
}

// tagMacroParser parses the {% macro %} tag. It requires a name, argument list
// with optional defaults, and optionally "export" to make it available via import.
func tagMacroParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	macroNode := &tagMacroNode{
		position: start,
		args:     make(map[string]IEvaluator),
	}

	nameToken := arguments.MatchType(TokenIdentifier)
	if nameToken == nil {
		return nil, arguments.Error("Macro-tag needs at least an identifier as name.", nil)
	}
	macroNode.name = nameToken.Val

	if arguments.MatchOne(TokenSymbol, "(") == nil {
		return nil, arguments.Error("Expected '('.", nil)
	}

	for arguments.Match(TokenSymbol, ")") == nil {
		argNameToken := arguments.MatchType(TokenIdentifier)
		if argNameToken == nil {
			return nil, arguments.Error("Expected argument name as identifier.", nil)
		}
		macroNode.argsOrder = append(macroNode.argsOrder, argNameToken.Val)

		if arguments.Match(TokenSymbol, "=") != nil {
			// Default expression follows
			argDefaultExpr, err := arguments.ParseExpression()
			if err != nil {
				return nil, err
			}
			macroNode.args[argNameToken.Val] = argDefaultExpr
		} else {
			// No default expression
			macroNode.args[argNameToken.Val] = nil
		}

		if arguments.Match(TokenSymbol, ")") != nil {
			break
		}
		if arguments.Match(TokenSymbol, ",") == nil {
			return nil, arguments.Error("Expected ',' or ')'.", nil)
		}
	}

	if arguments.Match(TokenKeyword, "export") != nil {
		macroNode.exported = true
	}

	if arguments.Remaining() > 0 {
		return nil, arguments.Error("Malformed macro-tag.", nil)
	}

	// Body wrapping
	wrapper, endargs, err := doc.WrapUntilTag("endmacro")
	if err != nil {
		return nil, err
	}
	macroNode.wrapper = wrapper

	if endargs.Count() > 0 {
		return nil, endargs.Error("Arguments not allowed here.", nil)
	}

	if macroNode.exported {
		// Now register the macro if it wants to be exported
		_, has := doc.template.exportedMacros[macroNode.name]
		if has {
			return nil, doc.Error(fmt.Sprintf("another macro with name '%s' already exported", macroNode.name), start)
		}
		doc.template.exportedMacros[macroNode.name] = macroNode
	}

	return macroNode, nil
}

func init() {
	mustRegisterTag("macro", tagMacroParser)
}
