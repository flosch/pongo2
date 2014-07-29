package pongo2

import (
	"bytes"
	"fmt"
)

type tagMacroNode struct {
	position   *Token
	name       string
	args_order []string
	args       map[string]INodeEvaluator

	wrapper *NodeWrapper
}

func (node *tagMacroNode) Execute(ctx *ExecutionContext, buffer *bytes.Buffer) error {

	args_ctx := make(Context)

	for k, v := range node.args {
		if v == nil {
			// User did not provided a default value
			args_ctx[k] = nil
		} else {
			// Evaluate the default value
			value_expr, err := v.Evaluate(ctx)
			if err != nil {
				return err
			}

			args_ctx[k] = value_expr
		}
	}

	macro_func := func(args ...*Value) string {
		if len(args) > len(node.args_order) {
			// Too many arguments, we're ignoring them and just logging into debug mode.
			logf(ctx.Error(fmt.Sprintf("Macro '%s' called with too many arguments (%d instead of %d).",
				node.name, len(args), len(node.args_order)), node.position).Error())
			return "(pongo2 error: macro called with too many arguments [see debug mode])"
		}

		// Make a context for the macro execution
		macroCtx := NewChildExecutionContext(ctx)

		// Register all arguments in the private context
		macroCtx.Private.Update(args_ctx)

		for idx, arg_value := range args {
			macroCtx.Private[node.args_order[idx]] = arg_value.Interface()
		}

		var b bytes.Buffer
		err := node.wrapper.Execute(macroCtx, &b)
		if err != nil {
			logf(ctx.Error(fmt.Sprintf("Error occured during execution of macro '%s': %s",
				err.Error()), node.position).Error())
			return "(pongo2 error: macro execution error [see debug mode])"
		}
		return b.String()
	}

	// Register the created macro function into the context
	ctx.Private[node.name] = macro_func

	return nil
}

func tagMacroParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	macro_node := &tagMacroNode{
		position: start,
		args:     make(map[string]INodeEvaluator),
	}

	name_token := arguments.MatchType(TokenIdentifier)
	if name_token == nil {
		return nil, arguments.Error("Macro-tag needs at least an identifier as name.", nil)
	}
	macro_node.name = name_token.Val

	if arguments.MatchOne(TokenSymbol, "(") == nil {
		return nil, arguments.Error("Expected '('.", nil)
	}

	for arguments.Match(TokenSymbol, ")") == nil {
		arg_name_token := arguments.MatchType(TokenIdentifier)
		if arg_name_token == nil {
			return nil, arguments.Error("Expected argument name as identifier.", nil)
		}
		macro_node.args_order = append(macro_node.args_order, arg_name_token.Val)

		if arguments.Match(TokenSymbol, "=") != nil {
			// Default expression follows
			arg_default_expr, err := arguments.ParseExpression()
			if err != nil {
				return nil, err
			}
			macro_node.args[arg_name_token.Val] = arg_default_expr
		} else {
			// No default expression
			macro_node.args[arg_name_token.Val] = nil
		}

		if arguments.Match(TokenSymbol, ")") != nil {
			break
		}
		if arguments.Match(TokenSymbol, ",") == nil {
			return nil, arguments.Error("Expected ',' or ')'.", nil)
		}
	}

	// Body wrapping
	wrapper, endargs, err := doc.WrapUntilTag("endmacro")
	if err != nil {
		return nil, err
	}
	macro_node.wrapper = wrapper

	if endargs.Count() > 0 {
		return nil, endargs.Error("Arguments not allowed here.", nil)
	}

	return macro_node, nil
}

func init() {
	RegisterTag("macro", tagMacroParser)
}
