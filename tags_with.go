package pongo2

type tagWithNode struct {
	with_pairs map[string]IEvaluator
	wrapper    *NodeWrapper
}

func (node *tagWithNode) Execute(ctx *ExecutionContext) (string, error) {
	// Building the context for the template
	include_ctx := make(Context)
	//new context for block
	myctx := new(ExecutionContext)
	// Fill the context with all data from the parent
	for key, value := range *ctx.Public {
		include_ctx[key] = value
	}

	// Put all custom with-pairs into the context
	for key, value := range node.with_pairs {
		val, err := value.Evaluate(ctx)
		if err != nil {
			return "", err
		}
		include_ctx[key] = val
	}

	myctx.Public = &include_ctx
	myctx.Private = ctx.Private
	myctx.StringStore = ctx.StringStore

	return node.wrapper.Execute(myctx)
}

func tagWithParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	include_node := &tagWithNode{
		with_pairs: make(map[string]IEvaluator),
	}

	wrapper, err := doc.WrapUntilTag("endwith")
	if err != nil {
		return nil, err
	}
	include_node.wrapper = wrapper

	if arguments.Count() == 0 {
		return nil, arguments.Error("Tag 'with' requires at least one argument.", nil)
	}

	// Scan through all arguments to see which style the user uses (old or new style).
	// If we find any "as" keyword we will enforce old style; otherwise we will use new style.
	old_style := false // by default we're using the new_style
	for i := 0; i < arguments.Count(); i++ {
		if arguments.PeekN(i, TokenKeyword, "as") != nil {
			old_style = true
			break
		}
	}

	for arguments.Remaining() > 0 {
		if old_style {
			value_expr, err := arguments.ParseExpression()
			if err != nil {
				return nil, err
			}
			if arguments.Match(TokenKeyword, "as") == nil {
				return nil, arguments.Error("Expected 'as' keyword.", nil)
			}
			key_token := arguments.MatchType(TokenIdentifier)
			if key_token == nil {
				return nil, arguments.Error("Expected an identifier", nil)
			}
			include_node.with_pairs[key_token.Val] = value_expr
		} else {
			key_token := arguments.MatchType(TokenIdentifier)
			if key_token == nil {
				return nil, arguments.Error("Expected an identifier", nil)
			}
			if arguments.Match(TokenSymbol, "=") == nil {
				return nil, arguments.Error("Expected '='.", nil)
			}
			value_expr, err := arguments.ParseExpression()
			if err != nil {
				return nil, err
			}
			include_node.with_pairs[key_token.Val] = value_expr
		}
	}

	return include_node, nil
}

func init() {
	RegisterTag("with", tagWithParser)
}
