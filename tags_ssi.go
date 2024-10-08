package pongo2

import (
	"fmt"
	"os"
)

type tagSSINode struct {
	filename string
	content  string
	template *Template
}

func (node *tagSSINode) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
	if node.template != nil {
		// Execute the template within the current context
		includeCtx := make(Context)
		includeCtx.Update(ctx.Public)
		includeCtx.Update(ctx.Private)

		err := node.template.execute(includeCtx, writer)
		if err != nil {
			return err.(*Error)
		}
	} else {
		// Just print out the content
		if _, err := writer.WriteString(node.content); err != nil {
			return ctx.Error(err, nil)
		}
	}
	return nil
}

func tagSSIParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, *Error) {
	SSINode := &tagSSINode{}

	if fileToken := arguments.MatchType(TokenString); fileToken != nil {
		SSINode.filename = fileToken.Val

		if arguments.Match(TokenIdentifier, "parsed") != nil {
			// parsed
			temporaryTpl, err := doc.template.set.FromFile(doc.template.set.resolveFilename(doc.template, fileToken.Val))
			if err != nil {
				return nil, err.(*Error).updateFromTokenIfNeeded(doc.template, fileToken)
			}
			SSINode.template = temporaryTpl
		} else {
			// plaintext
			buf, err := os.ReadFile(doc.template.set.resolveFilename(doc.template, fileToken.Val))
			if err != nil {
				return nil, (&Error{
					Sender:    "tag:ssi",
					OrigError: err,
				}).updateFromTokenIfNeeded(doc.template, fileToken)
			}
			SSINode.content = string(buf)
		}
	} else {
		return nil, arguments.Error(fmt.Errorf("First argument must be a string."), nil)
	}

	if arguments.Remaining() > 0 {
		return nil, arguments.Error(fmt.Errorf("Malformed SSI-tag argument."), nil)
	}

	return SSINode, nil
}

func init() {
	MustRegisterTag("ssi", tagSSIParser)
}
