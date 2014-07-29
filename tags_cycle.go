package pongo2

import (
	"bytes"
)

type tagCycleNode struct {
	position *Token
	args     []INodeEvaluator
	idx      int
	as_name  string
}

func (node *tagCycleNode) Execute(ctx *ExecutionContext, buffer *bytes.Buffer) error {
	item := node.args[node.idx%len(node.args)]
	node.idx++

	val, err := item.Evaluate(ctx)
	if err != nil {
		return err
	}

	if node.as_name != "" {
		ctx.Private[node.as_name] = val
	} else {
		buffer.WriteString(val.String())
	}

	return nil
}

func tagCycleParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	cycle_node := &tagCycleNode{
		position: start,
	}

	for arguments.Remaining() > 0 {
		node, err := arguments.ParseExpression()
		if err != nil {
			return nil, err
		}
		cycle_node.args = append(cycle_node.args, node)

		if arguments.MatchOne(TokenKeyword, "as") != nil {
			// as

			name_token := arguments.MatchType(TokenIdentifier)
			if name_token == nil {
				return nil, arguments.Error("Name (identifier) expected after 'as'.", nil)
			}
			cycle_node.as_name = name_token.Val

			// Now we're finished
			break
		}
	}

	if arguments.Remaining() > 0 {
		return nil, arguments.Error("Malformed cycle-tag.", nil)
	}

	return cycle_node, nil
}

func init() {
	RegisterTag("cycle", tagCycleParser)
}
