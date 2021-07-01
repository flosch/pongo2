package pongo2

type tagExtendsNode struct {
	filename string
}

func (node *tagExtendsNode) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
	return nil
}

func tagExtendsParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, *Error) {
	extendsNode := &tagExtendsNode{}

	if doc.Template.level > 1 {
		return nil, arguments.Error("The 'extends' tag can only defined on root level.", start)
	}

	if doc.Template.parent != nil {
		// Already one parent
		return nil, arguments.Error("This template has already one parent.", start)
	}

	if filenameToken := arguments.MatchType(TokenString); filenameToken != nil {
		// prepared, static template

		// Get parent's filename
		parentFilename := doc.Template.Set.resolveFilename(doc.Template, filenameToken.Val)

		// Parse the parent
		parentTemplate, err := doc.Template.Set.FromFile(parentFilename)
		if err != nil {
			return nil, err.(*Error)
		}

		// Keep track of things
		parentTemplate.child = doc.Template
		doc.Template.parent = parentTemplate
		extendsNode.filename = parentFilename
	} else {
		return nil, arguments.Error("Tag 'extends' requires a template filename as string.", nil)
	}

	if arguments.Remaining() > 0 {
		return nil, arguments.Error("Tag 'extends' does only take 1 argument.", nil)
	}

	return extendsNode, nil
}

func init() {
	RegisterTag("extends", tagExtendsParser)
}
