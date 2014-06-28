package pongo2

import (
	"fmt"
)

type tagBlockNode struct {
	name    string
	wrapper *NodeWrapper
}

func (node *tagBlockNode) Execute(ctx *ExecutionContext) (string, error) {
	// Check for internal 'block:$name:content'-key
	key := fmt.Sprintf("block:%s:content", node.name)
	override, key_exists := ctx.StringStore[key]
	if !key_exists {
		// This could be the child template which extends the base template, so we're executing
		// the wrapped elements and saving the result
		rv, err := node.wrapper.Execute(ctx)
		if err != nil {
			return "", err
		}
		ctx.StringStore[key] = rv
		return rv, nil
	}
	return override, nil
}

func tagBlockParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	block_node := &tagBlockNode{}

	wrapper, err := doc.WrapUntilTag("endblock")
	if err != nil {
		return nil, err
	}
	block_node.wrapper = wrapper

	if arguments.Count() == 0 {
		return nil, arguments.Error("Tag 'block' requires an identifier.", nil)
	}

	name_token := arguments.MatchType(TokenIdentifier)
	if name_token == nil {
		return nil, arguments.Error("First argument for tag 'block' must be an identifier.", nil)
	}
	block_node.name = name_token.Val

	if arguments.Remaining() != 0 {
		return nil, arguments.Error("Tag 'block' takes exactly 1 argument (an identifier).", nil)
	}

	return block_node, nil
}

func init() {
	RegisterTag("block", tagBlockParser)
}
