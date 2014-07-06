package pongo2

type tagIfNode struct {
	condition   IEvaluator
	thenWrapper *NodeWrapper
	elseWrapper *NodeWrapper
}

func (node *tagIfNode) Execute(ctx *ExecutionContext) (string, error) {
	result, err := node.condition.Evaluate(ctx)
	if err != nil {
		return "", err
	}

	if result.IsTrue() {
		return node.thenWrapper.Execute(ctx)
	} else {
		if node.elseWrapper != nil {
			return node.elseWrapper.Execute(ctx)
		}
	}
	return "", nil
}

func tagIfParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	if_node := &tagIfNode{}

	// Parse condition
	condition, err := arguments.ParseExpression()
	if err != nil {
		return nil, err
	}
	if_node.condition = condition

	if arguments.Remaining() > 0 {
		return nil, arguments.Error("If-condition is malformed.", nil)
	}

	// Wrap then/else-blocks
	wrapper, endargs, err := doc.WrapUntilTag("else", "endif")
	if err != nil {
		return nil, err
	}
	if_node.thenWrapper = wrapper

	if endargs.Count() > 0 {
		return nil, endargs.Error("Arguments not allowed here.", nil)
	}

	if wrapper.Endtag == "else" {
		// if there's an else in the if-statement, we need the else-Block as well
		wrapper, endargs, err = doc.WrapUntilTag("endif")
		if err != nil {
			return nil, err
		}
		if_node.elseWrapper = wrapper

		if endargs.Count() > 0 {
			return nil, endargs.Error("Arguments not allowed here.", nil)
		}
	}

	return if_node, nil
}

func init() {
	RegisterTag("if", tagIfParser)
}
