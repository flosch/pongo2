package pongo2

import "fmt"

type tagTemplateTagNode struct {
	content string
}

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

func (node *tagTemplateTagNode) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
	if _, err := writer.WriteString(node.content); err != nil {
		return ctx.Error(err, nil)
	}
	return nil
}

func tagTemplateTagParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, *Error) {
	ttNode := &tagTemplateTagNode{}

	if argToken := arguments.MatchType(TokenIdentifier); argToken != nil {
		output, found := templateTagMapping[argToken.Val]
		if !found {
			return nil, arguments.Error(fmt.Errorf("Argument not found"), argToken)
		}
		ttNode.content = output
	} else {
		return nil, arguments.Error(fmt.Errorf("Identifier expected."), nil)
	}

	if arguments.Remaining() > 0 {
		return nil, arguments.Error(fmt.Errorf("Malformed templatetag-tag argument."), nil)
	}

	return ttNode, nil
}

func init() {
	MustRegisterTag("templatetag", tagTemplateTagParser)
}
