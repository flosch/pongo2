package pongo2

// tagIncludeNode represents the {% include %} tag.
//
// The include tag renders another template and inserts its output at the
// current location. The included template has access to the current context.
//
// Basic usage:
//
//	{% include "header.html" %}
//	<main>Content here</main>
//	{% include "footer.html" %}
//
// Using "if_exists" to silently skip missing templates:
//
//	{% include "optional_sidebar.html" if_exists %}
//
// Passing additional context with "with":
//
//	{% include "user_card.html" with username=user.name avatar=user.avatar %}
//
// Using "only" to exclude parent context (included template only sees with variables):
//
//	{% include "widget.html" with title="My Widget" only %}
//
// Dynamic template names (lazy evaluation):
//
//	{% include template_name %}
//	{% include "partials/"|add:partial_name|add:".html" %}
//
// The "only" keyword must come after all with pairs:
//
//	{% include "card.html" with title="Hello" subtitle="World" only %}
//
// Note: Static filenames (strings) are parsed at compile time for better
// performance. Dynamic filenames are resolved at runtime.
type tagIncludeNode struct {
	tpl               *Template
	filenameEvaluator IEvaluator
	lazy              bool
	only              bool
	filename          string
	withPairs         map[string]IEvaluator
	ifExists          bool
}

func (node *tagIncludeNode) Execute(ctx *ExecutionContext, writer TemplateWriter) error {
	// Building the context for the template
	includeCtx := make(Context)

	// Fill the context with all data from the parent
	if !node.only {
		includeCtx.Update(ctx.Public)
		includeCtx.Update(ctx.Private)
	}

	// Put all custom with-pairs into the context
	for key, value := range node.withPairs {
		val, err := value.Evaluate(ctx)
		if err != nil {
			return err
		}
		includeCtx[key] = val
	}

	// Execute the template
	if node.lazy {
		// Evaluate the filename
		filename, err := node.filenameEvaluator.Evaluate(ctx)
		if err != nil {
			return err
		}

		if filename.String() == "" {
			return ctx.Error("Filename for 'include'-tag evaluated to an empty string.", nil)
		}

		// Get include-filename
		includedFilename := ctx.template.set.resolveFilename(ctx.template, filename.String())

		includedTpl, err2 := ctx.template.set.FromFile(includedFilename)
		if err2 != nil {
			// if this is ReadFile error, and "if_exists" flag is enabled
			if e, ok := err2.(*Error); ok && node.ifExists && e.Sender == "fromfile" {
				return nil
			}
			return err2
		}
		err2 = includedTpl.ExecuteWriter(includeCtx, writer)
		if err2 != nil {
			return err2
		}
		return nil
	}
	// Template is already parsed with static filename
	err := node.tpl.ExecuteWriter(includeCtx, writer)
	if err != nil {
		return err
	}
	return nil
}

type tagIncludeEmptyNode struct{}

func (node *tagIncludeEmptyNode) Execute(ctx *ExecutionContext, writer TemplateWriter) error {
	return nil
}

func tagIncludeParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	includeNode := &tagIncludeNode{
		withPairs: make(map[string]IEvaluator),
	}

	if filenameToken := arguments.MatchType(TokenString); filenameToken != nil {
		// prepared, static template

		// "if_exists" flag
		ifExists := arguments.Match(TokenIdentifier, "if_exists") != nil

		// Get include-filename
		includedFilename := doc.template.set.resolveFilename(doc.template, filenameToken.Val)

		// Parse the parent
		includeNode.filename = includedFilename
		includedTpl, err := doc.template.set.FromFile(includedFilename)
		if err != nil {
			// if this is ReadFile error, and "if_exists" token presents we should create and empty node
			if e, ok := err.(*Error); ok && e.Sender == "fromfile" && ifExists {
				return &tagIncludeEmptyNode{}, nil
			}
			return nil, updateErrorToken(err, doc.template, filenameToken)
		}
		includeNode.tpl = includedTpl
	} else {
		// No String, then the user wants to use lazy-evaluation (slower, but possible)
		filenameEvaluator, err := arguments.ParseExpression()
		if err != nil {
			return nil, updateErrorToken(err, doc.template, filenameToken)
		}
		includeNode.filenameEvaluator = filenameEvaluator
		includeNode.lazy = true
		includeNode.ifExists = arguments.Match(TokenIdentifier, "if_exists") != nil // "if_exists" flag
	}

	// After having parsed the filename we're gonna parse the with+only options
	if arguments.Match(TokenIdentifier, "with") != nil {
		for arguments.Remaining() > 0 {
			// We have at least one key=expr pair (because of starting "with")
			keyToken := arguments.MatchType(TokenIdentifier)
			if keyToken == nil {
				return nil, arguments.Error("Expected an identifier", nil)
			}
			if arguments.Match(TokenSymbol, "=") == nil {
				return nil, arguments.Error("Expected '='.", nil)
			}
			valueExpr, err := arguments.ParseExpression()
			if err != nil {
				return nil, updateErrorToken(err, doc.template, keyToken)
			}

			includeNode.withPairs[keyToken.Val] = valueExpr

			// Only?
			if arguments.Match(TokenIdentifier, "only") != nil {
				includeNode.only = true
				break // stop parsing arguments because it's the last option
			}
		}
	}

	if arguments.Remaining() > 0 {
		return nil, arguments.Error("Malformed 'include'-tag arguments.", nil)
	}

	return includeNode, nil
}

func init() {
	mustRegisterTag("include", tagIncludeParser)
}
