package pongo2

// tagTemplateTagNode represents the {% templatetag %} tag.
//
// The templatetag tag outputs special template syntax characters that would
// otherwise be interpreted by the template engine. This is useful when you
// need to display template tags as literal text.
//
// Available arguments:
//   - openblock: outputs "{%"
//   - closeblock: outputs "%}"
//   - openvariable: outputs "{{"
//   - closevariable: outputs "}}"
//   - openbrace: outputs "{"
//   - closebrace: outputs "}"
//   - opencomment: outputs "{#"
//   - closecomment: outputs "#}"
//
// Examples:
//
//	{% templatetag openblock %} for item in items {% templatetag closeblock %}
//
// Output: "{% for item in items %}"
//
//	{% templatetag openvariable %} item {% templatetag closevariable %}
//
// Output: "{{ item }}"
//
// Displaying a comment tag:
//
//	{% templatetag opencomment %} This is a comment {% templatetag closecomment %}
//
// Output: "{# This is a comment #}"
//
// Use cases:
//   - Documenting template syntax in templates
//   - Generating template code dynamically
//   - Escaping template syntax in output
type tagTemplateTagNode struct {
	content string
}

// templateTagMapping maps argument names to their literal template syntax output.
var templateTagMapping = map[string]string{
	"openblock":     "{%",
	"closeblock":    "%}",
	"openvariable":  "{{",
	"closevariable": "}}",
	"openbrace":     "{",
	"closebrace":    "}",
	"opencomment":   "{#",
	"closecomment":  "#}",
}

// Execute outputs the literal template syntax string (e.g., "{{" or "%}").
func (node *tagTemplateTagNode) Execute(ctx *ExecutionContext, writer TemplateWriter) error {
	_, err := writer.WriteString(node.content)
	return err
}

// tagTemplateTagParser parses the {% templatetag %} tag. It requires one
// identifier argument from the templateTagMapping (e.g., "openblock").
func tagTemplateTagParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	ttNode := &tagTemplateTagNode{}

	if argToken := arguments.MatchType(TokenIdentifier); argToken != nil {
		output, found := templateTagMapping[argToken.Val]
		if !found {
			return nil, arguments.Error("Argument not found", argToken)
		}
		ttNode.content = output
	} else {
		return nil, arguments.Error("Identifier expected.", nil)
	}

	if arguments.Remaining() > 0 {
		return nil, arguments.Error("Malformed templatetag-tag argument.", nil)
	}

	return ttNode, nil
}

func init() {
	mustRegisterTag("templatetag", tagTemplateTagParser)
}
