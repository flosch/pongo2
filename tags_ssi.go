package pongo2

import "os"

// tagSSINode represents the {% ssi %} tag.
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
// Use cases:
//   - Including static text files (plain text mode)
//   - Including template fragments that need context (parsed mode)
//   - Server-side includes similar to Apache SSI
//
// Note: Unlike {% include %}, the ssi tag reads the file at parse time.
// The "parsed" keyword is required if you want the included content to be
// processed as a template.
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
		writer.WriteString(node.content)
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
			// plaintext
			buf, err := os.ReadFile(doc.template.set.resolveFilename(doc.template, fileToken.Val))
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
	RegisterTag("ssi", tagSSIParser)
}
