package pongo2

import (
	"bytes"
	"io/ioutil"
)

type tagSSINode struct {
	filename string
	content  string
}

func (node *tagSSINode) Execute(ctx *ExecutionContext, buffer *bytes.Buffer) error {
	buffer.WriteString(node.content)
	return nil
}

func tagSSIParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	ssi_node := &tagSSINode{}

	if file_token := arguments.MatchType(TokenString); file_token != nil {
		ssi_node.filename = file_token.Val

		if arguments.Match(TokenIdentifier, "parsed") != nil {
			// parsed
			temporary_tpl, err := FromFile(file_token.Val)
			if err != nil {
				return nil, err
			}
			buf, err := temporary_tpl.Execute(nil)
			if err != nil {
				return nil, err
			}
			ssi_node.content = buf
		} else {
			// plaintext
			buf, err := ioutil.ReadFile(file_token.Val)
			if err != nil {
				return nil, err
			}
			ssi_node.content = string(buf)
		}
	} else {
		return nil, arguments.Error("First argument must be a string.", nil)
	}

	if arguments.Remaining() > 0 {
		return nil, arguments.Error("Malformed SSI-tag argument.", nil)
	}

	return ssi_node, nil
}

func init() {
	RegisterTag("ssi", tagSSIParser)
}
