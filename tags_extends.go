package pongo2

// tagExtendsNode represents the {% extends %} tag.
//
// The extends tag indicates that this template extends a parent template.
// It must be the first tag in the template if used. The child template can
// override blocks defined in the parent template using {% block %} tags.
//
// Usage:
//
//	{% extends "base.html" %}
//	{% block content %}
//	    My custom content that overrides the parent's content block.
//	{% endblock %}
//
// Example parent template (base.html):
//
//	<!DOCTYPE html>
//	<html>
//	<head><title>{% block title %}Default{% endblock %}</title></head>
//	<body>
//	    <header>{% block header %}Default Header{% endblock %}</header>
//	    <main>{% block content %}{% endblock %}</main>
//	    <footer>{% block footer %}Default Footer{% endblock %}</footer>
//	</body>
//	</html>
//
// Example child template (page.html):
//
//	{% extends "base.html" %}
//	{% block title %}My Page{% endblock %}
//	{% block content %}
//	    <h1>Welcome to my page!</h1>
//	{% endblock %}
//
// Note: Only one extends tag is allowed per template, and it must be at the root level.
type tagExtendsNode struct {
	filename string
}

func (node *tagExtendsNode) Execute(ctx *ExecutionContext, writer TemplateWriter) error {
	return nil
}

func tagExtendsParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	extendsNode := &tagExtendsNode{}

	if doc.template.level > 1 {
		return nil, arguments.Error("The 'extends' tag can only defined on root level.", start)
	}

	if doc.template.parent != nil {
		// Already one parent
		return nil, arguments.Error("This template has already one parent.", start)
	}

	if filenameToken := arguments.MatchType(TokenString); filenameToken != nil {
		// prepared, static template

		// Get parent's filename
		parentFilename := doc.template.set.resolveFilename(doc.template, filenameToken.Val)

		// Parse the parent
		parentTemplate, err := doc.template.set.FromFile(parentFilename)
		if err != nil {
			return nil, err
		}

		// Keep track of things
		parentTemplate.child = doc.template
		doc.template.parent = parentTemplate
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
	mustRegisterTag("extends", tagExtendsParser)
}
