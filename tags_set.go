package pongo2

// tagSetNode represents the {% set %} tag.
//
// The set tag assigns a value to a variable in the current context.
// This is useful for creating temporary variables or caching computed values.
//
// Basic usage:
//
//	{% set greeting = "Hello, World!" %}
//	{{ greeting }}
//
// Output: "Hello, World!"
//
// Setting a variable from an expression:
//
//	{% set full_name = user.first_name + " " + user.last_name %}
//	Welcome, {{ full_name }}!
//
// Using with filters:
//
//	{% set slug = title|slugify %}
//	<a href="/posts/{{ slug }}">{{ title }}</a>
//
// Setting computed values:
//
//	{% set total = price * quantity %}
//	{% set discounted = total * 0.9 %}
//	Total: ${{ total }}, After discount: ${{ discounted }}
//
// Note: Variables set with {% set %} are only available in the current
// template context and do not persist across template includes.
type tagSetNode struct {
	name       string
	expression IEvaluator
}

func (node *tagSetNode) Execute(ctx *ExecutionContext, writer TemplateWriter) error {
	// Evaluate expression
	value, err := node.expression.Evaluate(ctx)
	if err != nil {
		return err
	}

	ctx.Private[node.name] = value
	return nil
}

func tagSetParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	node := &tagSetNode{}

	// Parse variable name
	typeToken := arguments.MatchType(TokenIdentifier)
	if typeToken == nil {
		return nil, arguments.Error("Expected an identifier.", nil)
	}
	node.name = typeToken.Val

	if arguments.Match(TokenSymbol, "=") == nil {
		return nil, arguments.Error("Expected '='.", nil)
	}

	// Variable expression
	keyExpression, err := arguments.ParseExpression()
	if err != nil {
		return nil, err
	}
	node.expression = keyExpression

	// Remaining arguments
	if arguments.Remaining() > 0 {
		return nil, arguments.Error("Malformed 'set'-tag arguments.", nil)
	}

	return node, nil
}

func init() {
	mustRegisterTag("set", tagSetParser)
}
