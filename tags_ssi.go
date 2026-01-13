package pongo2

import "io"

// tagSSINode represents the {% ssi %} tag.
//
// DEPRECATED: This tag was removed from Django in version 1.10.
// Use {% include %} instead, which provides better functionality and security.
//
// See: https://code.djangoproject.com/ticket/24022
//
// The ssi (Server Side Include) tag includes the contents of another file
// into the template. It can include files as plain text or as parsed templates.
//
// Including a file as plain text (content is not parsed):
//
//	{% ssi "static/robots.txt" %}
//
// Output: Contents of robots.txt displayed as-is
//
// Including a file as a parsed template:
//
//	{% ssi "includes/header.html" parsed %}
//
// With "parsed", the file is treated as a template and has access to
// the current context variables.
//
// Deprecated: Use {% include %} instead.
type tagSSINode struct {
	filename string
	content  string
	template *Template
}

func (node *tagSSINode) Execute(ctx *ExecutionContext, writer TemplateWriter) error {
	if node.template != nil {
		// Execute the template within the current context
		includeCtx := make(Context)
		includeCtx.Update(ctx.Public)
		includeCtx.Update(ctx.Private)

		err := node.template.execute(includeCtx, writer)
		if err != nil {
			return err
		}
	} else {
		// Just print out the content
		_, err := writer.WriteString(node.content)
		return err
	}
	return nil
}

func tagSSIParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	SSINode := &tagSSINode{}

	if fileToken := arguments.MatchType(TokenString); fileToken != nil {
		SSINode.filename = fileToken.Val

		if arguments.Match(TokenIdentifier, "parsed") != nil {
			// parsed
			temporaryTpl, err := doc.template.set.FromFile(doc.template.set.resolveFilename(doc.template, fileToken.Val))
			if err != nil {
				return nil, updateErrorToken(err, doc.template, fileToken)
			}
			SSINode.template = temporaryTpl
		} else {
			// plaintext - use the template loader to support virtual filesystems
			_, _, fd, err := doc.template.set.resolveTemplate(doc.template, fileToken.Val)
			if err != nil {
				return nil, updateErrorToken(&Error{
					Sender:    "tag:ssi",
					OrigError: err,
				}, doc.template, fileToken)
			}
			buf, err := io.ReadAll(fd)
			if closer, ok := fd.(io.Closer); ok {
				if closeErr := closer.Close(); closeErr != nil && err == nil {
					err = closeErr
				}
			}
			if err != nil {
				return nil, updateErrorToken(&Error{
					Sender:    "tag:ssi",
					OrigError: err,
				}, doc.template, fileToken)
			}
			SSINode.content = string(buf)
		}
	} else {
		return nil, arguments.Error("First argument must be a string.", nil)
	}

	if arguments.Remaining() > 0 {
		return nil, arguments.Error("Malformed SSI-tag argument.", nil)
	}

	return SSINode, nil
}

func init() {
	mustRegisterTag("ssi", tagSSIParser)
}
