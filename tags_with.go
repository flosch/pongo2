package pongo2

// tagWithNode represents the {% with %} tag.
//
// The with tag creates a new scope with additional context variables.
// This is useful for caching complex expressions or simplifying nested
// attribute access.
//
// New style syntax (recommended):
//
//	{% with name=user.profile.full_name email=user.contact.email %}
//	    <p>{{ name }}</p>
//	    <p>{{ email }}</p>
//	{% endwith %}
//
// Old style syntax (Django-compatible):
//
//	{% with user.profile.full_name as name %}
//	    <p>{{ name }}</p>
//	{% endwith %}
//
// Multiple variables (new style):
//
//	{% with
//	    total=cart.items|length
//	    subtotal=cart.subtotal
//	    tax=cart.subtotal * 0.1
//	%}
//	    <p>Items: {{ total }}</p>
//	    <p>Subtotal: ${{ subtotal }}</p>
//	    <p>Tax: ${{ tax }}</p>
//	{% endwith %}
//
// Caching expensive computations:
//
//	{% with expensive_result=some_object.compute_heavy_operation %}
//	    <p>Result: {{ expensive_result }}</p>
//	    <p>Again: {{ expensive_result }}</p>
//	{% endwith %}
//
// Note: Variables defined in {% with %} are only available within the
// {% with %}...{% endwith %} block and do not affect the outer context.
type tagWithNode struct {
	withPairs map[string]IEvaluator
	wrapper   *NodeWrapper
}

// Execute creates a child context with the with-pairs and executes the
// wrapped content. The with-pairs are only visible within the block.
func (node *tagWithNode) Execute(ctx *ExecutionContext, writer TemplateWriter) error {
	// new context for block
	withctx := NewChildExecutionContext(ctx)

	// Put all custom with-pairs into the context
	for key, value := range node.withPairs {
		val, err := value.Evaluate(ctx)
		if err != nil {
			return err
		}
		withctx.Private[key] = val
	}

	return node.wrapper.Execute(withctx, writer)
}

// tagWithParser parses the {% with %} tag. It supports both new style (name=value)
// and old style (value as name) syntax for defining context variables.
func tagWithParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	withNode := &tagWithNode{
		withPairs: make(map[string]IEvaluator),
	}

	if arguments.Count() == 0 {
		return nil, arguments.Error("Tag 'with' requires at least one argument.", nil)
	}

	wrapper, endargs, err := doc.WrapUntilTag("endwith")
	if err != nil {
		return nil, err
	}
	withNode.wrapper = wrapper

	if endargs.Count() > 0 {
		return nil, endargs.Error("Arguments not allowed here.", nil)
	}

	// Scan through all arguments to see which style the user uses (old or new style).
	// If we find any "as" keyword we will enforce old style; otherwise we will use new style.
	oldStyle := false // by default we're using the new_style
	for i := 0; i < arguments.Count(); i++ {
		if arguments.PeekN(i, TokenKeyword, "as") != nil {
			oldStyle = true
			break
		}
	}

	for arguments.Remaining() > 0 {
		if oldStyle {
			valueExpr, err := arguments.ParseExpression()
			if err != nil {
				return nil, err
			}
			if arguments.Match(TokenKeyword, "as") == nil {
				return nil, arguments.Error("Expected 'as' keyword.", nil)
			}
			keyToken := arguments.MatchType(TokenIdentifier)
			if keyToken == nil {
				return nil, arguments.Error("Expected an identifier", nil)
			}
			withNode.withPairs[keyToken.Val] = valueExpr
		} else {
			keyToken := arguments.MatchType(TokenIdentifier)
			if keyToken == nil {
				return nil, arguments.Error("Expected an identifier", nil)
			}
			if arguments.Match(TokenSymbol, "=") == nil {
				return nil, arguments.Error("Expected '='.", nil)
			}
			valueExpr, err := arguments.ParseExpression()
			if err != nil {
				return nil, err
			}
			withNode.withPairs[keyToken.Val] = valueExpr
		}
	}

	return withNode, nil
}

func init() {
	mustRegisterTag("with", tagWithParser)
}
