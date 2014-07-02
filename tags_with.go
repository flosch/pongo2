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
	wrapper, err := doc.WrapUntilTag("endwith")
	include_node := &tagWithNode{
		with_pairs: make(map[string]IEvaluator),
		wrapper:    wrapper,
	}

	if err != nil {
		return nil, err
	}

	if arguments.Count() == 0 {
		return nil, arguments.Error("Tag 'with' requires an argument.", nil)
	}

	for arguments.Remaining() > 0 {

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

	return include_node, nil
}

func init() {
	RegisterTag("with", tagWithParser)
}
