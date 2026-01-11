package pongo2

// tagCommentNode represents the {% comment %} tag.
//
// The comment tag ignores everything between {% comment %} and {% endcomment %}.
// This is useful for commenting out code or adding notes that should not appear
// in the rendered output. Unlike HTML comments, template comments are completely
// removed from the output.
//
// Usage:
//
//	{% comment %}
//	    This text will not appear in the output.
//	    You can write multiple lines here.
//	    {{ variables }} and {% tags %} are also ignored.
//	{% endcomment %}
//
// Example:
//
//	<p>Visible content</p>
//	{% comment %}
//	    TODO: Add more features here later
//	    {{ debug_var }}
//	{% endcomment %}
//	<p>More visible content</p>
//
// Output:
//
//	<p>Visible content</p>
//	<p>More visible content</p>
type tagCommentNode struct{}

func (node *tagCommentNode) Execute(ctx *ExecutionContext, writer TemplateWriter) error {
	return nil
}

func tagCommentParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	commentNode := &tagCommentNode{}

	// TODO: Process the endtag's arguments (see django 'comment'-tag documentation)
	err := doc.SkipUntilTag("endcomment")
	if err != nil {
		return nil, err
	}

	if arguments.Count() != 0 {
		return nil, arguments.Error("Tag 'comment' does not take any argument.", nil)
	}

	return commentNode, nil
}

func init() {
	RegisterTag("comment", tagCommentParser)
}
