package pongo2

// tagFirstofNode represents the {% firstof %} tag.
//
// The firstof tag outputs the first variable that is "true" (not empty, not zero,
// not nil, not false). If all variables are false, nothing is output.
// This is useful for displaying fallback values.
//
// Usage:
//
//	{% firstof var1 var2 var3 %}
//
// Example with fallback values:
//
//	{% firstof user.nickname user.username "Anonymous" %}
//
// If user.nickname is "Johnny", output: "Johnny"
// If user.nickname is empty but username is "john_doe", output: "john_doe"
// If both are empty, output: "Anonymous"
//
// Example in practice:
//
//	<p>Welcome, {% firstof user.display_name user.email "Guest" %}!</p>
//
// Note: Output is automatically HTML-escaped when autoescape is enabled.
// Use the |safe filter if you need unescaped output.
type tagFirstofNode struct {
	position *Token
	args     []IEvaluator
}

func (node *tagFirstofNode) Execute(ctx *ExecutionContext, writer TemplateWriter) error {
	for _, arg := range node.args {
		val, err := arg.Evaluate(ctx)
		if err != nil {
			return err
		}

		if val.IsTrue() {
			if ctx.Autoescape && !arg.FilterApplied("safe") {
				val, err = ApplyFilter("escape", val, nil)
				if err != nil {
					return err
				}
			}

			_, err = writer.WriteString(val.String())
			return err
		}
	}

	return nil
}

func tagFirstofParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	firstofNode := &tagFirstofNode{
		position: start,
	}

	// Django requires at least one argument
	if arguments.Count() == 0 {
		return nil, arguments.Error("Tag 'firstof' requires at least one argument.", nil)
	}

	for arguments.Remaining() > 0 {
		node, err := arguments.ParseExpression()
		if err != nil {
			return nil, err
		}
		firstofNode.args = append(firstofNode.args, node)
	}

	return firstofNode, nil
}

func init() {
	mustRegisterTag("firstof", tagFirstofParser)
}
