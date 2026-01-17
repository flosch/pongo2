package pongo2

// tagAutoescapeNode represents the {% autoescape %} tag.
//
// The autoescape tag controls automatic HTML escaping for a block of template content.
// When autoescape is "on", all variable output is HTML-escaped. When "off", variables
// are output as-is (use with caution for trusted content only).
//
// Usage:
//
//	{% autoescape on %}
//	    {{ user_input }}  {# This will be HTML-escaped #}
//	{% endautoescape %}
//
//	{% autoescape off %}
//	    {{ trusted_html }}  {# This will NOT be escaped - use with caution! #}
//	{% endautoescape %}
//
// Example with dangerous input:
//
//	{% autoescape on %}
//	    {{ "<script>alert('XSS')</script>" }}
//	{% endautoescape %}
//
// Output: "&lt;script&gt;alert('XSS')&lt;/script&gt;"
type tagAutoescapeNode struct {
	wrapper    *NodeWrapper
	autoescape bool
}

// Execute renders the block content with the configured autoescape setting.
// It temporarily changes the autoescape state in the context, executes the
// wrapped content, and restores the original autoescape state afterward.
func (node *tagAutoescapeNode) Execute(ctx *ExecutionContext, writer TemplateWriter) error {
	old := ctx.Autoescape
	ctx.Autoescape = node.autoescape

	err := node.wrapper.Execute(ctx, writer)
	if err != nil {
		return err
	}

	ctx.Autoescape = old

	return nil
}

// tagAutoescapeParser parses the {% autoescape %} tag.
// It expects a single argument "on" or "off" to control HTML escaping.
func tagAutoescapeParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	autoescapeNode := &tagAutoescapeNode{}

	wrapper, _, err := doc.WrapUntilTag("endautoescape")
	if err != nil {
		return nil, err
	}
	autoescapeNode.wrapper = wrapper

	modeToken := arguments.MatchType(TokenIdentifier)
	if modeToken == nil {
		return nil, arguments.Error("A mode is required for autoescape-tag.", nil)
	}
	switch modeToken.Val {
	case "on":
		autoescapeNode.autoescape = true
	case "off":
		autoescapeNode.autoescape = false
	default:
		return nil, arguments.Error("Only 'on' or 'off' is valid as an autoescape-mode.", nil)
	}

	if arguments.Remaining() > 0 {
		return nil, arguments.Error("Malformed autoescape-tag arguments.", nil)
	}

	return autoescapeNode, nil
}

func init() {
	mustRegisterTag("autoescape", tagAutoescapeParser)
}
