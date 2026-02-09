package pongo2

type tagInlineExtendNode struct {
	filename string
	blocks   map[string]*NodeWrapper
	wrapper  *NodeWrapper
}

func (node *tagInlineExtendNode) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
	parentTpl, err := ctx.template.set.FromFile(node.filename)
	if err != nil {
		return err.(*Error)
	}

	childTpl := &Template{
		blocks: node.blocks,
		parent: parentTpl,
		set:    ctx.template.set,
		root:   &nodeDocument{Nodes: node.wrapper.nodes},
	}
	parentTpl.child = childTpl

	newCtx := NewChildExecutionContext(ctx)
	newCtx.template = childTpl

	return parentTpl.root.Execute(newCtx, writer)
}

func tagInlineExtendParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, *Error) {
	node := &tagInlineExtendNode{
		blocks: make(map[string]*NodeWrapper),
	}

	filenameToken := arguments.MatchType(TokenString)
	if filenameToken == nil {
		return nil, arguments.Error("Expected a string as first argument for 'inlineextend' tag.", start)
	}
	node.filename = doc.template.set.resolveFilename(doc.template, filenameToken.Val)

	if arguments.Remaining() > 0 {
		return nil, arguments.Error("Malformed 'inlineextend'-tag arguments.", start)
	}

	// hold on to the blocks
	tempTemplate := &Template{
		blocks: make(map[string]*NodeWrapper),
		set:    doc.template.set,
	}

	// temporarily replace the parser's template
	originalTemplate := doc.template
	doc.template = tempTemplate
	defer func() { doc.template = originalTemplate }()

	// parse the body of the inlineextend tag
	wrapper, endtagargs, err := doc.WrapUntilTag("endinlineextend")
	if err != nil {
		return nil, err
	}
	if endtagargs.Remaining() > 0 {
		return nil, endtagargs.Error("Arguments not allowed for 'endinlineextend'.", nil)
	}

	node.blocks = tempTemplate.blocks
	node.wrapper = wrapper

	return node, nil
}

func init() {
	RegisterTag("inlineextend", tagInlineExtendParser)
}
