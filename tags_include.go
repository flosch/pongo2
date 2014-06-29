package pongo2

/*
type tagIncludeNode struct {
	filename string
}

func (node *tagIncludeNode) Execute(ctx *ExecutionContext) (string, error) {
	return "", nil
}

func tagIncludeParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	include_node := &tagIncludeNode{}

	filename_evaluator, err := arguments.ParseExpression()
	if err != nil {
		return nil, err
	}
	if filename_evaluator.e

	if arguments.Remaining() > 0 {
		return nil, arguments.Error("Tag 'include' does only take 1 argument.", nil)
	}

	return include_node, nil
}

func init() {
	RegisterTag("include", tagIncludeParser)
}
*/
